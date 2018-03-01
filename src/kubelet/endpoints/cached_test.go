package endpoints

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var secureHTTPClient = &http.Client{}

func failingClientMock() *endpoints.MockedClient {
	client := new(endpoints.MockedClient)
	client.On("Config").Return(nil)
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{}, errors.New("FindPodByName should not be invoked"))
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{}, errors.New("FindPodsByHostname should not be invoked"))
	client.On("FindNode", mock.Anything).Return(nil, errors.New("FindNode should not be invoked"))
	return client
}

func TestDiscover_CachedKubelet_DefaultInsecurePort(t *testing.T) {
	// Given a Kubernetes API client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the insecure Kubelet URL
	kclient, err := cacher.Discover(timeout)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// And the returned cached instance should be correctly configured
	assert.Nil(t, err)
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:10255", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "the-node-name", kclient.(*kubelet).nodeName)
	assert.Nil(t, kclient.(*kubelet).httpClient.Transport)
}

func TestDiscover_CachedKubelet_NotFoundByName(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)

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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the insecure Kubelet URL
	kclient, err := cacher.Discover(timeout)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// And the returned cached instance should be correctly configured
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "11.22.33.44", kclient.NodeIP())
	assert.Equal(t, "11.22.33.44:5432", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "the-node-name", kclient.(*kubelet).nodeName)
	assert.Nil(t, kclient.(*kubelet).httpClient.Transport)
}

func TestDiscover_CachedKubelet_DefaultSecurePort(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the secure Kubelet URL
	kclient, err := cacher.Discover(timeout)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the cached host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:10250", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "the-node-name", kclient.(*kubelet).nodeName)
	assert.True(t, kclient.(*kubelet).httpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
}

func TestDiscover_CachedKubelet_HTTP_CheckingConnection(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: allOkConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the insecure Kubelet URL
	kclient, err := cacher.Discover(timeout)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the cached host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "http", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "the-node-name", kclient.(*kubelet).nodeName)
	assert.Nil(t, kclient.(*kubelet).httpClient.Transport)
}

func TestDiscover_CachedKubelet_HTTPS_CheckingConnection(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: failOnInsecureConnection,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the secure Kubelet URL
	kclient, err := cacher.Discover(timeout)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the cached host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, "1.2.3.4:55332", kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
	assert.Equal(t, "the-node-name", kclient.(*kubelet).nodeName)
	assert.True(t, kclient.(*kubelet).httpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
}

func TestDiscover_CachedKubelet_ApiConnection(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)
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

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: onlyAPIConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// That successfully retrieved the secure Kubelet API URL
	kclient, err := cacher.Discover(timeout)

	// When invoking again the discovery process, it should not use the API client
	wrappedDiscoverer.apiClient = failingClientMock()
	kclient, err = cacher.Discover(timeout)

	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the cached host:port of the Kubelet is returned
	assert.Equal(t, "1.2.3.4", kclient.NodeIP())
	assert.Equal(t, apiHost, kclient.(*kubelet).endpoint.Host)
	assert.Equal(t, "https", kclient.(*kubelet).endpoint.Scheme)
	assert.Equal(t, "/api/v1/nodes/the-node-name/proxy/", kclient.(*kubelet).endpoint.Path)
	assert.Equal(t, "d34db33f", kclient.(*kubelet).config.BearerToken)
}

func TestDiscover_CachedKubelet_DiscoveryError(t *testing.T) {
	// Given a Kubernetes API Client
	client := new(endpoints.MockedClient)
	client.On("Config").Return(&rest.Config{BearerToken: "d34db33f"})
	client.On("SecureHTTPClient", mock.Anything).Return(secureHTTPClient, nil)

	// That doesn't find the pod neither by name nor hostname
	client.On("FindPodByName", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindPodsByHostname", mock.Anything).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
	client.On("FindNode", "the-node-name").Return(nil, fmt.Errorf("Node not found"))

	// And a disk cache storage
	tmpDir, err := ioutil.TempDir("", "test_discover_cached_kubelet")
	assert.Nil(t, err)
	storage := storage.NewJSONDiskStorage(tmpDir)
	// and an Discoverer implementation
	wrappedDiscoverer := kubeletDiscoverer{
		apiClient:   client,
		connChecker: onlyAPIConnectionChecker,
		logger:      logger,
	}

	// And a Kubelet Discovery Cacher without any cached data
	cacher := NewKubeletDiscoveryCacher(&wrappedDiscoverer, storage)

	// When retrieving the Kubelet URL
	_, err = cacher.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}
