package endpoints

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

const timeout = time.Second

var logger = logrus.StandardLogger()

func allOkConnectionChecker(_ *http.Client, _ url.URL, _, _ string) error {
	return nil
}

func failOnInsecureConnection(_ *http.Client, URL url.URL, _, _ string) error {
	if URL.Scheme != "https" {
		return fmt.Errorf("the connection can't be established")
	}
	return nil
}

func onlyAPIConnectionChecker(_ *http.Client, URL url.URL, _, _ string) error {
	if URL.Host == apiHost {
		return nil
	}
	return fmt.Errorf("the connection can't be established")
}

func mockStatusCodeHandler(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}
}

func TestDiscoverHTTP_DefaultInsecurePort(t *testing.T) {
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
		logger:      logger,
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

func TestDiscoverHTTP_NotFoundByName(t *testing.T) {
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
		logger:      logger,
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

func TestDiscoverHTTPS_DefaultSecurePort(t *testing.T) {
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
		logger:      logger,
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

func TestDiscoverHTTP_CheckingConnection(t *testing.T) {
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
		logger:      logger,
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

func TestDiscoverHTTPS_CheckingConnection(t *testing.T) {
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
		logger:      logger,
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

func TestDiscoverHTTPS_ApiConnection(t *testing.T) {
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
		logger:      logger,
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

func TestDiscover_NodeNotFoundError(t *testing.T) {
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
		logger:    logger,
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestDiscover_NilNodeError(t *testing.T) {
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
		logger:      logger,
	}

	// When retrieving the Kubelet URL
	_, err := discoverer.Discover(timeout)
	// The system returns an error
	assert.NotNil(t, err, "should return error")
}

func TestDo_HTTP(t *testing.T) {
	s := httptest.NewServer(mockStatusCodeHandler(http.StatusOK))
	defer s.Close()
	s.URL = "http://example.com/"

	endpoint, err := url.Parse(s.URL)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	var c = &kubelet{
		nodeIP:     "1.2.3.4",
		config:     rest.Config{BearerToken: "Foo"},
		nodeName:   "nodeFoo",
		endpoint:   *endpoint,
		httpClient: s.Client(),
		logger:     logger,
	}

	resp, err := c.Do("GET", "foo")

	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/foo", resp.Request.URL.String())
	assert.Equal(t, "", resp.Request.Header.Get("Authorization"))
	assert.Equal(t, "GET", resp.Request.Method)
	assert.Equal(t, "http://example.com/", endpoint.String())
}

func TestDo_HTTPS(t *testing.T) {
	s := httptest.NewServer(mockStatusCodeHandler(http.StatusOK))
	defer s.Close()
	s.URL = "https://example.com"

	endpoint, err := url.Parse(s.URL)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	var c = &kubelet{
		nodeIP:     "1.2.3.4",
		config:     rest.Config{BearerToken: "Foo"},
		nodeName:   "nodeFoo",
		endpoint:   *endpoint,
		httpClient: s.Client(),
		logger:     logger,
	}

	resp, err := c.Do("GET", "foo")

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/foo", resp.Request.URL.String())
	assert.Equal(t, fmt.Sprintf("Bearer %s", c.config.BearerToken), resp.Request.Header.Get("Authorization"))
	assert.Equal(t, "GET", resp.Request.Method)
	assert.Equal(t, "https://example.com", endpoint.String())
}

func TestCheckCall(t *testing.T) {
	s := httptest.NewServer(mockStatusCodeHandler(http.StatusOK))
	defer s.Close()

	endpoint, err := url.Parse(s.URL)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = checkCall(s.Client(), *endpoint, "foo", "foo token")
	assert.NoError(t, err)
}

func TestCheckCall_ErrorNotSuccessStatusCode(t *testing.T) {
	s := httptest.NewServer(mockStatusCodeHandler(http.StatusNotFound))
	defer s.Close()
	s.URL = "https://example.com/"

	endpoint, err := url.Parse(s.URL)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	expectedCalledURL := "https://example.com/foo"

	err = checkCall(s.Client(), *endpoint, "foo", "foo token")
	assert.EqualError(t, err, fmt.Sprintf("error calling endpoint %s. Got status code: %d", expectedCalledURL, http.StatusNotFound))
}

func TestCheckCall_ErrorConnecting(t *testing.T) {
	err := checkCall(http.DefaultClient, url.URL{}, "foo", "foo token")
	assert.NotNil(t, err)
}
