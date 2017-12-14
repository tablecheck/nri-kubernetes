package endpoints

import "net/url"

// Discoverer allows discovering the endpoints from different services in the Kubernetes ecosystem.
type Discoverer interface {
	// Discover returns the Endpoint of the Kubelet service that is located in the same node as the invoking pod.
	Discover() (url.URL, error)

	// GetNodeIP returns the IP of the Node that contains the service
	GetNodeIP() (string, error)
}
