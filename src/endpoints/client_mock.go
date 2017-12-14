package endpoints

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

// MockedClient is a Mock for the Kubernetes Client interface to be used only in tests
type MockedClient struct {
	mock.Mock
}

// FindNode mocks KubernetesClient
func (m MockedClient) FindNode(name string) (*v1.NodeList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.NodeList), args.Error(1)
}

// FindPodsByLabel mocks KubernetesClient
func (m MockedClient) FindPodsByLabel(name, value string) (*v1.PodList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindPodByName mocks KubernetesClient
func (m MockedClient) FindPodByName(name string) (*v1.PodList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindPodsByHostname mocks KubernetesClient
func (m MockedClient) FindPodsByHostname(hostname string) (*v1.PodList, error) {
	args := m.Called(hostname)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindServiceByLabel mocks KubernetesClient
func (m MockedClient) FindServiceByLabel(name, value string) (*v1.ServiceList, error) {
	args := m.Called(name, value)
	return args.Get(0).(*v1.ServiceList), args.Error(1)
}

// IsHTTPS mocks KubernetesClient
func (m MockedClient) IsHTTPS(url string) bool {
	args := m.Called(url)
	return args.Bool(0)
}
