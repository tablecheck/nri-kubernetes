package endpoints

import "net/url"

// Discoverer allows discovering the endpoints from different services in the Kubernetes ecosystem.
// There is no guarantee that the implementations of this interface cache the endpoints from previous invocations of
// their methods: every time the methods of this interface are invoked, the discovery process is completely repeated,
// since services endpoints could change during the lifetime of an application.
type Discoverer interface {
	// Discover returns the Endpoint of the Kubelet service that is located in the same node as the invoking pod.
	Discover() (url.URL, error)
}
