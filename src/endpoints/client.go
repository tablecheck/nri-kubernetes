package endpoints

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesClient provides an interface to common Kubernetes API operations
type KubernetesClient interface {
	// FindNode returns a NodeList reference containing the pod named as the argument, if any
	FindNode(name string) (*v1.NodeList, error)
	// FindPodsByLabel returns a PodList reference containing the pods matching the provided name/value label pair
	FindPodsByLabel(name, value string) (*v1.PodList, error)
	// FindPodByName returns a PodList reference that should contain the pod whose name matches with the name argument
	FindPodByName(name string) (*v1.PodList, error)
	// FindPodsByHostname returns a Podlist reference containing the pod or pods whose hostname matches the argument
	FindPodsByHostname(hostname string) (*v1.PodList, error)
	// FindServiceByLabel returns a ServiceList containing the services matching the provided name/value label pair
	// name/value pairs
	FindServiceByLabel(name, value string) (*v1.ServiceList, error)
	// IsHTTPS checks whether a connection to a URL is secure or not
	IsHTTPS(url string) bool
}

type goClientImpl struct {
	client *kubernetes.Clientset
}

func (ka goClientImpl) FindNode(name string) (*v1.NodeList, error) {
	return ka.client.CoreV1().Nodes().List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
}

func (ka goClientImpl) FindPodsByLabel(name, value string) (*v1.PodList, error) {
	return ka.client.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", name, value),
	})
}

func (ka goClientImpl) FindPodByName(name string) (*v1.PodList, error) {
	return ka.client.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
}

func (ka goClientImpl) FindPodsByHostname(hostname string) (*v1.PodList, error) {
	return ka.client.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.hostname=%s", hostname),
	})
}

func (ka goClientImpl) FindServiceByLabel(name, value string) (*v1.ServiceList, error) {
	return ka.client.CoreV1().Services("").List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", name, value),
	})
}

func (ka goClientImpl) IsHTTPS(url string) bool {
	// We ignore certificates only for checking
	netClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := netClient.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.TLS != nil
}

// NewKubernetesClient instantiates a Kubernetes API client
func NewKubernetesClient() (KubernetesClient, error) {
	var ka goClientImpl

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	ka.client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return ka, nil
}
