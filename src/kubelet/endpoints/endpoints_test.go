package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"k8s.io/api/core/v1"
)

func TestKubeletDiscovery(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
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
	discoverer := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4:12345", kurl.Host)
	assert.Equal(t, "https", kurl.Scheme)
}

func TestKubeletDiscovery_NotFoundByName(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
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

	discoverer := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.22.33.44:5432", kurl.Host)
	assert.Equal(t, "https", kurl.Scheme)
}

func TestKubeletDiscovery_NotFoundError(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	// That doesn't find the pod neither by name nor hostname
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindNode", "the-node-name").Return(&v1.NodeList{Items: []v1.Node{}}, nil)

	discoverer := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover()
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestKubeletDiscovery_HTTP(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Spec:   v1.PodSpec{NodeName: "the-node-name"},
			Status: v1.PodStatus{HostIP: "5.5.5.5"},
		}}}, nil)
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
	discoverer := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet URL for a non-secure discovered port
	kurl, err := discoverer.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.2.3.4:4445", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKubeletDiscoverer_GetNodeIP(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Spec: v1.PodSpec{NodeName: "the-node-name"},
		}}}, nil)
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
	discoverer := kubeletDiscoverer{
		client: client,
	}

	// When retrieving the Kubelet Node IP
	nodeIP, err := discoverer.GetNodeIP()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered node IP is returned
	assert.Equal(t, "1.2.3.4", nodeIP)
}
