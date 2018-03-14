package endpoints

import (
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/client"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"github.com/sirupsen/logrus"
)

const cachedKSMKey = "ksm-client"

// cachedKSM holds the data to be cached for a KSM client.
// Its fields must be public to make them visible for the JSON Marshaller.
type cachedKSM struct {
	Endpoint url.URL
	NodeIP   string
}

// composeKSM implements the ClientComposer function signature
func composeKSM(source interface{}, cacher *client.DiscoveryCacher, timeout time.Duration) (client.HTTPClient, error) {
	cached := source.(*cachedKSM)
	return &ksm{
		nodeIP:   cached.NodeIP,
		endpoint: cached.Endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: cacher.Logger,
	}, nil
}

// composeKSM implements the ClientDecomposer function signature
func decomposeKSM(source client.HTTPClient) (interface{}, error) {
	ksm := source.(*ksm)
	return &cachedKSM{
		Endpoint: ksm.endpoint,
		NodeIP:   ksm.nodeIP,
	}, nil
}

// NewKSMDiscoveryCacher creates a new DiscoveryCacher that wraps a ksmDiscoverer and caches the data into the
// specified storage
func NewKSMDiscoveryCacher(ksmDiscoverer client.Discoverer, storage storage.Storage, ttl time.Duration, logger *logrus.Logger) client.Discoverer {
	return &client.DiscoveryCacher{
		CachedDataPtr: &cachedKSM{},
		StorageKey:    cachedKSMKey,
		Discoverer:    ksmDiscoverer,
		Storage:       storage,
		TTL:           ttl,
		Logger:        logger,
		Compose:       composeKSM,
		Decompose:     decomposeKSM,
	}
}
