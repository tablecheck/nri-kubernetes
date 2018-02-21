package endpoints

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	logSDK "github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

const timeout = time.Second

func allOkConnectionChecker(_ *http.Client, _, _ string) error {
	return nil
}

func failOnInsecureConnection(_ *http.Client, URL, _ string) error {
	urlObj, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if urlObj.Scheme != "https" {
		return fmt.Errorf("the connection can't be established")
	}
	return nil
}

func onlyAPIConnectionChecker(_ *http.Client, URL, _ string) error {
	purl, _ := url.Parse(URL)
	if purl.Host == apiHost {
		return nil
	}
	return fmt.Errorf("the connection can't be established")
}

func TestKubeletDiscoveryHTTP_DefaultInsecurePort(t *testing.T) {
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
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:10255", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTP_NotFoundByName(t *testing.T) {
	// Given a client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(&http.Client{}, nil)

	// That doesn't find the pod by name
	client.On("FindPodByName", mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{}}, nil)

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
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "11.22.33.44", kclient.NodeIP())
	assert.Equal(t, "11.22.33.44:5432", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
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
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:10250", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
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
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
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
		connChecker: failOnInsecureConnection,
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
}

func TestKubeletDiscoveryHTTPS_ApiConnection(t *testing.T) {
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
		connChecker: onlyAPIConnectionChecker,
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	kclient, err := discoverer.Discover(timeout)
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, apiHost, kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
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
		logger:    logSDK.New(false),
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
		logger:      logSDK.New(false),
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}
