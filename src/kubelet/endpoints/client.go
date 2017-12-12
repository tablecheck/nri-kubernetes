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

// kubernetesClient provides an interface to common Kubernetes API operations
type kubernetesClient interface {
	// findPodsByName returns a PodList reference that should contain the pod whose name matches with the name argument
	findPodByName(name string) (*v1.PodList, error)
	// fundPodsByHostname returns a Podlist reference containing the pod or pods whose hostname matches the argument
	findPodsByHostname(hostname string) (*v1.PodList, error)
	// findNode returns a NodeList reference containing the pod named as the argument, if any
	findNode(name string) (*v1.NodeList, error)
	// isHTTPS checks whether a connection to a URL is secure or not
	isHTTPS(url string) bool
}

type goClientImpl struct {
	client *kubernetes.Clientset
}

func (ka goClientImpl) findPodByName(name string) (*v1.PodList, error) {
	return ka.client.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
}

func (ka goClientImpl) findPodsByHostname(hostname string) (*v1.PodList, error) {
	return ka.client.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.hostname=%s", hostname),
	})
}

func (ka goClientImpl) findNode(name string) (*v1.NodeList, error) {
	return ka.client.CoreV1().Nodes().List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
}

func (ka goClientImpl) isHTTPS(url string) bool {
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

func newKubernetesClient() (kubernetesClient, error) {
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
