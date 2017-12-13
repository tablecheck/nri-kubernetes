package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

type mockedClient struct {
	mock.Mock
}

func (m mockedClient) FindPodByName(name string) (*v1.PodList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

func (m mockedClient) FindPodsByHostname(hostname string) (*v1.PodList, error) {
	args := m.Called(hostname)
	return args.Get(0).(*v1.PodList), args.Error(1)
}

func (m mockedClient) FindNode(name string) (*v1.NodeList, error) {
	args := m.Called(name)
	return args.Get(0).(*v1.NodeList), args.Error(1)
}

func (m mockedClient) IsHTTPS(url string) bool {
	args := m.Called(url)
	return args.Bool(0)
}

func TestKubelet(t *testing.T) {
	// Given a client
	client := new(mockedClient)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(true)
	client.On("FindNode", "the-node-name").
		Return(&v1.NodeList{Items: []v1.Node{{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "1.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: 12345,
					},
				},
			},
		}}}, nil)

	// and an Discoverer implementation
	endpoints := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	kurl, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4:12345", kurl.Host)
	assert.Equal(t, "https", kurl.Scheme)
}

func TestKubelet_NotFoundByName(t *testing.T) {
	// Given a client
	client := new(mockedClient)
	// That doesn't find the pod by name
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(true)

	// But finds it by hostname
	client.On("FindPodsByHostname", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.NodeList{Items: []v1.Node{{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "11.22.33.44",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: 5432,
					},
				},
			},
		}}}, nil)

	endpoints := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	kurl, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.22.33.44:5432", kurl.Host)
	assert.Equal(t, "https", kurl.Scheme)
}

func TestKubelet_NotFoundError(t *testing.T) {
	// Given a client
	client := new(mockedClient)
	// That doesn't find the pod neither by name nor hostname
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindNode", "the-node-name").Return(&v1.NodeList{Items: []v1.Node{}}, nil)

	endpoints := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	_, err := endpoints.Discover()
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestKubelet_HTTP(t *testing.T) {
	// Given a client
	client := new(mockedClient)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(false)
	client.On("FindNode", "the-node-name").
		Return(&v1.NodeList{Items: []v1.Node{{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "11.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: 4445,
					},
				},
			},
		}}}, nil)
	endpoints := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL for a non-secure discovered port
	kurl, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.2.3.4:4445", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}
