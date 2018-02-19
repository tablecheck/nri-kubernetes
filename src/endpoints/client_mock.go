package endpoints

import (
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// MockedClient is a Mock for the Kubernetes Client interface to be used only in tests
type MockedClient struct {
	mock.Mock
}

// Config mocks Kubernetes Config
func (m *MockedClient) Config() *rest.Config {
	args := m.Called()
	return args.Get(0).(*rest.Config)
}

// SecureHTTPClient mocks KubernetesClient
func (m *MockedClient) SecureHTTPClient(timeout time.Duration) (*http.Client, error) {
	args := m.Called(timeout)
	return args.Get(0).(*http.Client), args.Error(1)
}

// FindNode mocks KubernetesClient
func (m *MockedClient) FindNode(name string) (*v1.Node, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.Node), args.Error(1)
}

// FindPodsByLabel mocks KubernetesClient
func (m *MockedClient) FindPodsByLabel(name, value string) (*v1.PodList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindPodByName mocks KubernetesClient
func (m *MockedClient) FindPodByName(name string) (*v1.PodList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindPodsByHostname mocks KubernetesClient
func (m *MockedClient) FindPodsByHostname(hostname string) (*v1.PodList, error) {
	args := m.Called(hostname)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

// FindServiceByLabel mocks KubernetesClient
func (m *MockedClient) FindServiceByLabel(name, value string) (*v1.ServiceList, error) {
	args := m.Called(name, value)
	return args.Get(0).(*v1.ServiceList), args.Error(1)
}
