package endpoints

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/rest"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"k8s.io/api/core/v1"
)

const timeout = 1000

func allOkConnectionChecker(_ *http.Client, _, _ string) error {
	return nil
}

func failonInsecureConnection(_ *http.Client, URL, _ string) error {
	urlObj, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if urlObj.Scheme != "https" {
		return fmt.Errorf("the connection can't be established")
	}
	return nil
}

func TestKubeletDiscoveryHTTP(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "1.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: defaultInsecureKubeletPort,
					},
				},
			},
		}, nil)

	// and an Discoverer implementation
	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kurl.NodeIP())
	assert.Equal(t, "1.2.3.4:10255", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTP_NotFoundByName(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	// That doesn't find the pod by name
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(true)

	// But finds it by hostname
	client.On("FindPodsByHostname", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
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
		}, nil)

	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.22.33.44", kurl.NodeIP())
	assert.Equal(t, "11.22.33.44:5432", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTPS_DefaultSecurePort(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "1.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: defaultSecureKubeletPort,
					},
				},
			},
		}, nil)

	// and an Discoverer implementation
	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kurl.NodeIP())
	assert.Equal(t, "1.2.3.4:10250", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTP_CheckingConnection(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "1.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: 55332, // configured without any default port
					},
				},
			},
		}, nil)

	// and an Discoverer implementation
	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kurl.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTPS_CheckingConnection(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{Spec: v1.PodSpec{NodeName: "the-node-name"}}}}, nil)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    "InternalIP",
						Address: "1.2.3.4",
					},
				},
				DaemonEndpoints: v1.NodeDaemonEndpoints{
					KubeletEndpoint: v1.DaemonEndpoint{
						Port: 55332, // configured without any default port
					},
				},
			},
		}, nil)

	// and an Discoverer implementation whose connection check connection fails because it is a secure connection
	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: failonInsecureConnection,
		logger:      logrus.StandardLogger(),
	}

	// When retrieving the Kubelet URL
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kurl.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscovery_NodeNotFoundError(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	// That doesn't find the pod neither by name nor hostname
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindNode", "the-node-name").Return(nil, fmt.Errorf("Node not found"))

	discoverer := kubeletDiscoverer{
		apiClient: client,
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestKubeletDiscovery_NilNodeError(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	// That doesn't find the pod neither by name nor hostname
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindNode", "the-node-name").Return(nil, nil)

	discoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestKubeletDiscovery_HTTP(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Spec:   v1.PodSpec{NodeName: "the-node-name"},
			Status: v1.PodStatus{HostIP: "5.5.5.5"},
		}}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(false)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
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
		}, nil)
	d := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}
	discoverer := &d

	// When retrieving the Kubelet URL for a non-secure discovered port
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.2.3.4", kurl.NodeIP())
	assert.Equal(t, "11.2.3.4:4445", kurl.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kurl.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoverer_NodeIP(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Spec: v1.PodSpec{NodeName: "the-node-name"},
		}}}, nil)
	client.On("IsHTTPS", mock.Anything).Return(true)
	client.On("FindNode", "the-node-name").
		Return(&v1.Node{
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
		}, nil)

	// and an Discoverer implementation
	d := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
	}
	discoverer := &d

	// When retrieving the Kubelet Node IP
	kurl, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered node IP is returned
	assert.Equal(t, "1.2.3.4", kurl.NodeIP())
}
