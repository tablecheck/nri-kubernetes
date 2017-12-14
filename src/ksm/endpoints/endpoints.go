package endpoints

import (
	"fmt"
	"net/url"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
)

const (
	ksmAppLabelName  = "k8s-app"
	ksmAppLabelValue = "kube-state-metrics"
	ksmPortName      = "http-metrics"
	k8sTCP           = "TCP"
)

// ksmDiscoverer implements Discoverer interface by using official Kubernetes' Go client
type ksmDiscoverer struct {
	client endpoints.KubernetesClient
}

// Discover returns the schema://host:port URL part of Kube State Metrics
func (sd ksmDiscoverer) Discover() (url.URL, error) {
	var endpoint url.URL

	services, err := sd.client.FindServiceByLabel(ksmAppLabelName, ksmAppLabelValue)
	if err != nil {
		return endpoint, err
	}

	if len(services.Items) == 0 {
		return endpoint, fmt.Errorf("no service found by label %s=%s", ksmAppLabelName, ksmAppLabelValue)
	}

	// KSM and Prometheus only work with HTTP
	endpoint.Scheme = "http"

	for _, service := range services.Items {
		if service.Spec.ClusterIP != "" && len(service.Spec.Ports) > 0 {
			// Look for a port called "http-metrics"
			for _, port := range service.Spec.Ports {
				if port.Name == ksmPortName {
					endpoint.Host = fmt.Sprintf("%v:%v", service.Spec.ClusterIP, port.Port)
					return endpoint, nil
				}
			}
			// If not found, return the first TCP port
			for _, port := range service.Spec.Ports {
				if port.Protocol == k8sTCP {
					endpoint.Host = fmt.Sprintf("%v:%v", service.Spec.ClusterIP, port.Port)
					return endpoint, nil
				}
			}
		}
	}

	return endpoint, fmt.Errorf("could not guess the Kube State Metrics host/port")
}

func (sd ksmDiscoverer) NodeIP() (string, error) {
	pods, err := sd.client.FindPodsByLabel(ksmAppLabelName, ksmAppLabelValue)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pod found by label %s=%s", ksmAppLabelName, ksmAppLabelValue)
	}
	for _, pod := range pods.Items {
		if pod.Status.HostIP != "" {
			return pod.Status.HostIP, nil
		}
	}
	return "", fmt.Errorf("no InternalIP address found for KSM node")
}

// NewKSMDiscoverer instantiates a new Discoverer
func NewKSMDiscoverer() (endpoints.Discoverer, error) {
	var discoverer ksmDiscoverer
	var err error

	discoverer.client, err = endpoints.NewKubernetesClient()
	if err != nil {
		return nil, err
	}

	return discoverer, nil
}
