package endpoints

import (
	"fmt"
	"net/url"
	"os"
)

// Discoverer allows discovering the endpoints from different services in the Kubernetes ecosystem.
// There is no guarantee that the implementations of this interface cache the endpoints from previous invocations of
// their methods: every time the methods of this interface are invoked, the discovery process is completely repeated,
// since services endpoints could change during the lifetime of an application.
type Discoverer interface {
	// Kubelet returns the Endpoint of the Kubelet service that is located in the same node as the invoking pod.
	Discover() (url.URL, error)
}

// kubeletDiscoverer implements Discoverer interface by using official Kubernetes' Go client
type kubeletDiscoverer struct {
	client kubernetesClient
}

func (sd kubeletDiscoverer) Discover() (url.URL, error) {
	var endpoint url.URL
	var port int
	var host string

	hostname, _ := os.Hostname()

	// get current pod whose name is equal to hostname and get the Node name
	pods, err := sd.client.findPodByName(hostname)
	if err != nil {
		return endpoint, err
	}

	// If not found by name, looking for the pod whose hostname annotation coincides (if unique in the cluster)
	if len(pods.Items) == 0 {
		pods, err = sd.client.findPodsByHostname(hostname)
		if err != nil {
			return endpoint, err
		}
		if len(pods.Items) == 0 {
			return endpoint, fmt.Errorf("no pods found whose name or hostname is %q", hostname)
		}
		if len(pods.Items) > 1 {
			return endpoint, fmt.Errorf("multiple pods sharing the hostname %q, can't apply autodiscovery", hostname)
		}
	}

	nodeName := pods.Items[0].Spec.NodeName

	// Get the containing node and discover the InternalIP and Kubelet port
	nodes, _ := sd.client.findNode(nodeName)

	if len(nodes.Items) == 0 {
		return endpoint, fmt.Errorf("could not find node named %q", nodeName)
	}

	port = int(nodes.Items[0].Status.DaemonEndpoints.KubeletEndpoint.Port)
	for _, address := range nodes.Items[0].Status.Addresses {
		if address.Type == "InternalIP" {
			host = address.Address
			break
		}
	}

	if port == 0 {
		return endpoint, fmt.Errorf("could not get Kubelet port")
	}
	if host == "" {
		return endpoint, fmt.Errorf("could not get Kubelet host")
	}
	endpoint.Host = fmt.Sprintf("%s:%d", host, port)

	// Guess whether the connection is HTTP or HTTPS
	endpoint.Scheme = "https"

	if !sd.client.isHTTPS(endpoint.String()) {
		endpoint.Scheme = "http"
	}

	return endpoint, nil
}

// NewKubeletDiscoverer instantiates a new Discoverer
func NewKubeletDiscoverer() (Discoverer, error) {
	var discoverer kubeletDiscoverer
	var err error

	discoverer.client, err = newKubernetesClient()
	if err != nil {
		return nil, err
	}

	return discoverer, nil
}
