package endpoints

import (
	"fmt"
	"net/url"
	"os"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
)

// kubeletDiscoverer implements Discoverer interface by using official Kubernetes' Go client
// For efficiency purposes, it caches the URL and the Node name, avoiding to repeat the same request to the K8s API
type kubeletDiscoverer struct {
	endpoint url.URL
	nodeIP   string
	client   endpoints.KubernetesClient
}

func (sd kubeletDiscoverer) Discover() (url.URL, error) {
	var err error
	if sd.nodeIP == "" {
		err = sd.cacheData()
	}
	return sd.endpoint, err
}

func (sd kubeletDiscoverer) GetNodeIP() (string, error) {
	var err error
	if sd.nodeIP == "" {
		err = sd.cacheData()
	}
	return sd.nodeIP, err
}

func (sd *kubeletDiscoverer) cacheData() error {
	var port int
	var host string

	hostname, _ := os.Hostname()

	// get current pod whose name is equal to hostname and get the Node name
	pods, err := sd.client.FindPodByName(hostname)
	if err != nil {
		return err
	}

	// If not found by name, looking for the pod whose hostname annotation coincides (if unique in the cluster)
	if len(pods.Items) == 0 {
		pods, err = sd.client.FindPodsByHostname(hostname)
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			return fmt.Errorf("no pods found whose name or hostname is %q", hostname)
		}
		if len(pods.Items) > 1 {
			return fmt.Errorf("multiple pods sharing the hostname %q, can't apply autodiscovery", hostname)
		}
	}

	nodeName := pods.Items[0].Spec.NodeName

	// Get the containing node and discover the InternalIP and Kubelet port
	nodes, _ := sd.client.FindNode(nodeName)

	if len(nodes.Items) == 0 {
		return fmt.Errorf("could not find node named %q", nodeName)
	}

	port = int(nodes.Items[0].Status.DaemonEndpoints.KubeletEndpoint.Port)
	for _, address := range nodes.Items[0].Status.Addresses {
		if address.Type == "InternalIP" {
			host = address.Address
			sd.nodeIP = address.Address
			break
		}
	}

	if port == 0 {
		return fmt.Errorf("could not get Kubelet port")
	}
	if host == "" {
		return fmt.Errorf("could not get Kubelet host IP")
	}
	sd.endpoint.Host = fmt.Sprintf("%s:%d", host, port)

	// Guess whether the connection is HTTP or HTTPS
	sd.endpoint.Scheme = "https"

	if !sd.client.IsHTTPS(sd.endpoint.String()) {
		sd.endpoint.Scheme = "http"
	}

	return nil
}

// NewKubeletDiscoverer instantiates a new Discoverer
func NewKubeletDiscoverer() (endpoints.Discoverer, error) {
	var discoverer kubeletDiscoverer
	var err error

	discoverer.client, err = endpoints.NewKubernetesClient()
	if err != nil {
		return nil, err
	}

	return discoverer, nil
}
