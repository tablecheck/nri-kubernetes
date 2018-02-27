package endpoints

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// kubeletDiscoverer implements Discoverer interface by using official Kubernetes' Go client
type kubeletDiscoverer struct {
	apiClient   endpoints.KubernetesClient
	logger      *logrus.Logger
	connChecker connectionChecker
}

const (
	healthzPath                = "/healthz"
	apiHost                    = "kubernetes.default"
	defaultInsecureKubeletPort = 10255
	defaultSecureKubeletPort   = 10250
)

// kubelet implements Client interface
type kubelet struct {
	httpClient *http.Client
	endpoint   url.URL
	config     rest.Config
	nodeIP     string
	nodeName   string
	logger     *logrus.Logger
}

type connectionParams struct {
	url    url.URL
	client *http.Client
}

type connectionChecker func(client *http.Client, URL url.URL, path, token string) error

func (c *kubelet) NodeIP() string {
	return c.nodeIP
}

// Do method calls discovered kubelet endpoint with specified method and path, i.e. "/stats/summary
func (c *kubelet) Do(method, path string) (*http.Response, error) {
	e := c.endpoint
	e.Path = filepath.Join(c.endpoint.Path, path)

	r, err := http.NewRequest(method, e.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request to: %s. Got error: %s ", method, e.String(), err)
	}

	if c.endpoint.Scheme == "https" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.BearerToken))
	}

	c.logger.Debugf("Calling Kubelet endpoint: %s", r.URL.String())
	return c.httpClient.Do(r)
}

func (sd *kubeletDiscoverer) Discover(timeout time.Duration) (endpoints.Client, error) {
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

		return &kubelet{
			nodeIP: host,
			endpoint: url.URL{
				Host:   c.url.Host,
				Path:   c.url.Path,
				Scheme: c.url.Scheme,
			},
			httpClient: c.client,
			config: rest.Config{
				BearerToken: config.BearerToken,
			},
			nodeName: nodeName,
			logger:   sd.logger,
		}, nil

	}
	return nil, err
}

func connectionHTTP(host string, timeout time.Duration) connectionParams {
	return connectionParams{
		url: url.URL{
			Host:   host,
			Scheme: "http",
		},
		client: endpoints.BasicHTTPClient(timeout),
	}
}

func connectionHTTPS(host string, timeout time.Duration) connectionParams {
	return connectionParams{
		url: url.URL{
			Host:   host,
			Scheme: "https",
		},
		client: endpoints.InsecureHTTPClient(timeout),
	}
}

func (sd *kubeletDiscoverer) connectionAPIHTTPS(nodeName string, timeout time.Duration) (connectionParams, error) {
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
		client: secureClient,
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

// NewKubeletDiscoverer instantiates a new Discoverer
func NewKubeletDiscoverer(logger *logrus.Logger) (endpoints.Discoverer, error) {
	var discoverer kubeletDiscoverer
	var err error

	discoverer.apiClient, err = endpoints.NewKubernetesClient()
	if err != nil {
		return nil, err
	}
	discoverer.logger = logger
	discoverer.connChecker = checkCall

	return &discoverer, nil
}

func (sd *kubeletDiscoverer) getPod() (v1.Pod, error) {
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

func (sd *kubeletDiscoverer) getNode(nodeName string) (*v1.Node, error) {
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
