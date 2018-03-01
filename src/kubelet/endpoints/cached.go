package endpoints

import (
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"github.com/newrelic/infra-integrations-sdk/log"
)

const cachedKubeletKey = "kubelet-client"

const (
	basicClient     = "basic"
	insecureClient  = "insecure"
	secureAPIClient = "secure-api"
)

// cachedKubelet holds the data to be cached for a Kubelet client.
// Its fields must be public to make them visible for the JSON Marshaller.
type cachedKubelet struct {
	Endpoint    url.URL
	NodeIP      string
	NodeName    string
	Timeout     time.Duration
	ClientType  string // basicClient, insecureClient or secureAPIClient
	BearerToken string
}

// composeKubelet implements the ClientComposer function signature
func composeKubelet(source interface{}, cacher *endpoints.DiscoveryCacher) (endpoints.Client, error) {
	cached := source.(*cachedKubelet)
	kd := cacher.Discoverer.(*kubeletDiscoverer)
	var client *http.Client
	switch cached.ClientType {
	case insecureClient:
		client = endpoints.InsecureHTTPClient(cached.Timeout)
	case secureAPIClient:
		api, err := kd.connectionAPIHTTPS(cached.NodeName, cached.Timeout)
		if err != nil {
			return nil, err
		}
		client = api.client
	default:
		client = endpoints.BasicHTTPClient(cached.Timeout)
	}
	return newKubelet(cached.NodeIP, cached.NodeName, cached.Endpoint, cached.BearerToken, client, kd.logger), nil
}

// decomposeKubelet implements the ClientDecomposer function signature
func decomposeKubelet(source endpoints.Client, cacher *endpoints.DiscoveryCacher) (interface{}, error) {
	kc := source.(*kubelet)
	kcacher := cacher.Discoverer.(*kubeletDiscoverer)
	var clientType string

	// If the client is the SAME as the secure HTTPS Secure API client
	// TODO: instantiate the Secure HTTP Client by ourselves, instead of from Kubernetes API Client,
	// whose implementation may change
	shc, err := kcacher.apiClient.SecureHTTPClient(kc.httpClient.Timeout)
	if err == nil && shc == kc.httpClient {
		clientType = secureAPIClient
	} else {
		// If the client has configured TLS info, then it's the insecure HTTPS
		transport, ok := kc.httpClient.Transport.(*http.Transport)
		if ok && transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
			clientType = insecureClient
		} else {
			// Otherwise, it's the Basic HTTP client
			clientType = basicClient
		}
	}

	log.Debug("Found %q client type in cache", clientType)

	return &cachedKubelet{
		Endpoint:    kc.endpoint,
		NodeIP:      kc.nodeIP,
		NodeName:    kc.nodeName,
		Timeout:     kc.httpClient.Timeout,
		ClientType:  clientType,
		BearerToken: kc.config.BearerToken,
	}, nil
}

// NewKubeletDiscoveryCacher creates a new DiscoveryCacher that wraps a kubeletDiscoverer and caches the data into the
// specified storage
func NewKubeletDiscoveryCacher(discoverer *kubeletDiscoverer, storage storage.Storage) *endpoints.DiscoveryCacher {
	return &endpoints.DiscoveryCacher{
		CachedDataPtr: &cachedKubelet{},
		StorageKey:    cachedKubeletKey,
		Discoverer:    discoverer,
		Storage:       storage,
		Logger:        discoverer.logger,
		Compose:       composeKubelet,
		Decompose:     decomposeKubelet,
	}
}
