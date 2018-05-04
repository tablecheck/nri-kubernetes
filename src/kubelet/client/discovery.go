package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/client"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/prometheus"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// discoverer implements Discoverer interface by using official Kubernetes' Go client
type discoverer struct {
	apiClient   client.Kubernetes
	logger      *logrus.Logger
	connChecker connectionChecker
}

const (
	healthzPath                = "/healthz"
	apiHost                    = "kubernetes.default"
	defaultInsecureKubeletPort = 10255
	defaultSecureKubeletPort   = 10250
)

// client type (if you need to add new values, do it at the end of the list)
const (
	httpBasic = iota
	httpInsecure
	httpSecure
)

// kubelet implements Client interface
type kubelet struct {
	httpClient *http.Client
	endpoint   url.URL
	config     rest.Config
	nodeIP     string
	nodeName   string
	httpType   int // httpBasic, httpInsecure, httpSecure
	logger     *logrus.Logger
}

type connectionParams struct {
	url      url.URL
	client   *http.Client
	httpType int // httpBasic, httpInsecure, httpSecure
}

type connectionChecker func(client *http.Client, URL url.URL, path, token string) error

func (c *kubelet) NodeIP() string {
	return c.nodeIP
}

// Do method calls discovered kubelet endpoint with specified method and path, i.e. "/stats/summary
func (c *kubelet) Do(method, path string) (*http.Response, error) {
	e := c.endpoint
	e.Path = filepath.Join(c.endpoint.Path, path)

	var r *http.Request
	var err error

	// TODO Create a new discoverer and client for cadvisor
	if path == metric.KubeletCAdvisorMetricsPath {
		if port := os.Getenv("CADVISOR_PORT"); port != "" {
			// We force to call the standalone cadvisor because k8s < 1.7.6 do not have /metrics/cadvisor kubelet endpoint.
			e.Scheme = "http"
			e.Host = fmt.Sprintf("%s:%s", c.nodeIP, port)
			e.Path = metric.StandaloneCAdvisorMetricsPath

			c.logger.Debugf("Using standalone cadvisor on port %s", port)
		}

		r, err = prometheus.NewRequest(method, e.String())
	} else {
		r, err = http.NewRequest(method, e.String(), nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating %s request to: %s. Got error: %s ", method, e.String(), err)
	}

	if c.endpoint.Scheme == "https" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.BearerToken))
	}

	c.logger.Debugf("Calling Kubelet endpoint: %s", r.URL.String())

	return c.httpClient.Do(r)
}

func (sd *discoverer) Discover(timeout time.Duration) (client.HTTPClient, error) {
	pod, err := sd.getPod()
	if err != nil {
		return nil, err
	}
	nodeName := getNodeName(pod)

	node, err := sd.getNode(nodeName)
	if err != nil {
		return nil, err
	}

	port, err := getPort(node)
	if err != nil {
		return nil, err
	}
	host, err := getHost(node)
	if err != nil {
		return nil, err
	}

	config := sd.apiClient.Config()
	hostURL := fmt.Sprintf("%s:%d", host, port)

	connectionAPIHTTPS, secErr := sd.connectionAPIHTTPS(nodeName, timeout)

	usedConnectionCases := make([]connectionParams, 0)
	switch port {
	case defaultInsecureKubeletPort:
		usedConnectionCases = append(usedConnectionCases, connectionHTTP(hostURL, timeout), connectionAPIHTTPS)
	case defaultSecureKubeletPort:
		usedConnectionCases = append(usedConnectionCases, connectionHTTPS(hostURL, timeout), connectionAPIHTTPS)
	default:
		usedConnectionCases = append(usedConnectionCases, connectionHTTP(hostURL, timeout), connectionHTTPS(hostURL, timeout), connectionAPIHTTPS)
	}

	for _, c := range usedConnectionCases {
		if secErr != nil && c.url.Host == apiHost {
			return nil, secErr
		}

		err = sd.connChecker(c.client, c.url, healthzPath, config.BearerToken)
		if err != nil {
			sd.logger.Debug(err.Error())
			continue
		}

		return newKubelet(host, nodeName, c.url, config.BearerToken, c.client, c.httpType, sd.logger), nil
	}
	return nil, err
}

func newKubelet(nodeIP string, nodeName string, endpoint url.URL, bearerToken string, client *http.Client, httpType int, logger *logrus.Logger) *kubelet {
	return &kubelet{
		nodeIP: nodeIP,
		endpoint: url.URL{
			Host:   endpoint.Host,
			Path:   endpoint.Path,
			Scheme: endpoint.Scheme,
		},
		httpClient: client,
		httpType:   httpType,
		config: rest.Config{
			BearerToken: bearerToken,
		},
		nodeName: nodeName,
		logger:   logger,
	}
}

func connectionHTTP(host string, timeout time.Duration) connectionParams {
	return connectionParams{
		url: url.URL{
			Host:   host,
			Scheme: "http",
		},
		client:   client.BasicHTTPClient(timeout),
		httpType: httpBasic,
	}
}

func connectionHTTPS(host string, timeout time.Duration) connectionParams {
	return connectionParams{
		url: url.URL{
			Host:   host,
			Scheme: "https",
		},
		client:   client.InsecureHTTPClient(timeout),
		httpType: httpInsecure,
	}
}

func (sd *discoverer) connectionAPIHTTPS(nodeName string, timeout time.Duration) (connectionParams, error) {
	secureClient, err := sd.apiClient.SecureHTTPClient(timeout)
	if err != nil {
		return connectionParams{}, err
	}
	return connectionParams{
		url: url.URL{
			Host:   apiHost,
			Path:   fmt.Sprintf("/api/v1/nodes/%s/proxy/", nodeName),
			Scheme: "https",
		},
		client:   secureClient,
		httpType: httpSecure,
	}, nil
}

func checkCall(client *http.Client, URL url.URL, path, token string) error {
	URL.Path = filepath.Join(URL.Path, path)

	r, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating request to: %s. Got error: %s ", URL.String(), err)
	}
	if URL.Scheme == "https" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("error trying to connect to: %s. Got error: %s ", URL.String(), err)
	}
	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("error calling endpoint %s. Got status code: %d", URL.String(), resp.StatusCode)
}

// NewDiscoverer instantiates a new Discoverer
func NewDiscoverer(logger *logrus.Logger) (client.Discoverer, error) {
	var discoverer discoverer
	var err error

	discoverer.apiClient, err = client.NewKubernetes()
	if err != nil {
		return nil, err
	}
	discoverer.logger = logger
	discoverer.connChecker = checkCall

	return &discoverer, nil
}

func (sd *discoverer) getPod() (v1.Pod, error) {
	var pod v1.Pod
	hostname, _ := os.Hostname()

	// get current pod whose name is equal to hostname and get the Node name
	pods, err := sd.apiClient.FindPodByName(hostname)
	if err != nil {
		return pod, err
	}

	// If not found by name, looking for the pod whose hostname annotation coincides (if unique in the cluster)
	if len(pods.Items) == 0 {
		pods, err = sd.apiClient.FindPodsByHostname(hostname)
		if err != nil {
			return pod, err
		}
		if len(pods.Items) == 0 {
			return pod, fmt.Errorf("no pods found whose name or hostname is %q", hostname)
		}
		if len(pods.Items) > 1 {
			return pod, fmt.Errorf("multiple pods sharing the hostname %q, can't apply autodiscovery", hostname)
		}
	}

	pod = pods.Items[0]
	return pod, nil
}

func getNodeName(pod v1.Pod) string {
	return pod.Spec.NodeName
}

func (sd *discoverer) getNode(nodeName string) (*v1.Node, error) {
	var node = new(v1.Node)
	var err error
	// Get the containing node and discover the InternalIP and Kubelet port
	node, err = sd.apiClient.FindNode(nodeName)
	if err != nil {
		return nil, fmt.Errorf("could not find node named %q", nodeName)
	}

	return node, nil
}

func getPort(node *v1.Node) (int, error) {
	port := int(node.Status.DaemonEndpoints.KubeletEndpoint.Port)
	if port == 0 {
		return 0, fmt.Errorf("could not get Kubelet port")
	}

	return port, nil
}

func getHost(node *v1.Node) (string, error) {
	var host string

	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			host = address.Address
			break
		}
	}

	if host == "" {
		return "", fmt.Errorf("could not get Kubelet host IP")
	}

	return host, nil
}
