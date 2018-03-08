package endpoints

import (
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
)

const cachedKubeletKey = "kubelet-client"

// cachedKubelet holds the data to be cached for a Kubelet client.
// Its fields must be public to make them visible for the JSON Marshaller.
type cachedKubelet struct {
	Endpoint    url.URL
	NodeIP      string
	NodeName    string
	HTTPType    int
	BearerToken string
}

// composeKubelet implements the ClientComposer function signature
func composeKubelet(source interface{}, cacher *endpoints.DiscoveryCacher, timeout time.Duration) (endpoints.Client, error) {
	cached := source.(*cachedKubelet)
	kd := cacher.Discoverer.(*kubeletDiscoverer)
	var client *http.Client
	switch cached.HTTPType {
	case httpInsecure:
		client = endpoints.InsecureHTTPClient(timeout)
	case httpSecure:
		api, err := kd.connectionAPIHTTPS(cached.NodeName, timeout)
		if err != nil {
			return nil, err
		}
		client = api.client
	default:
		client = endpoints.BasicHTTPClient(timeout)
	}
	return newKubelet(cached.NodeIP, cached.NodeName, cached.Endpoint, cached.BearerToken, client, cached.HTTPType, kd.logger), nil
}

// decomposeKubelet implements the ClientDecomposer function signature
func decomposeKubelet(source endpoints.Client) (interface{}, error) {
	kc := source.(*kubelet)
	return &cachedKubelet{
		Endpoint:    kc.endpoint,
		NodeIP:      kc.nodeIP,
		NodeName:    kc.nodeName,
		HTTPType:    kc.httpType,
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
