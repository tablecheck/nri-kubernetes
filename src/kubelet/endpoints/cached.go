package endpoints

import (
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/client"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"github.com/sirupsen/logrus"
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
func composeKubelet(source interface{}, cacher *client.DiscoveryCacher, timeout time.Duration) (client.HTTPClient, error) {
	cached := source.(*cachedKubelet)
	kd := cacher.Discoverer.(*kubeletDiscoverer)
	var c *http.Client
	switch cached.HTTPType {
	case httpInsecure:
		c = client.InsecureHTTPClient(timeout)
	case httpSecure:
		api, err := kd.connectionAPIHTTPS(cached.NodeName, timeout)
		if err != nil {
			return nil, err
		}
		c = api.client
	default:
		c = client.BasicHTTPClient(timeout)
	}
	return newKubelet(cached.NodeIP, cached.NodeName, cached.Endpoint, cached.BearerToken, c, cached.HTTPType, kd.logger), nil
}

// decomposeKubelet implements the ClientDecomposer function signature
func decomposeKubelet(source client.HTTPClient) (interface{}, error) {
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
func NewKubeletDiscoveryCacher(discoverer client.Discoverer, storage storage.Storage, ttl time.Duration, logger *logrus.Logger) *client.DiscoveryCacher {
	return &client.DiscoveryCacher{
		CachedDataPtr: &cachedKubelet{},
		StorageKey:    cachedKubeletKey,
		Discoverer:    discoverer,
		Storage:       storage,
		TTL:           ttl,
		Logger:        logger,
		Compose:       composeKubelet,
		Decompose:     decomposeKubelet,
	}
}
