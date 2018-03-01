package endpoints

import (
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
)

const cachedKSMKey = "ksm-client"

// cachedKSM holds the data to be cached for a KSM client.
// Its fields must be public to make them visible for the JSON Marshaller.
type cachedKSM struct {
	Endpoint url.URL
	NodeIP   string
	Timeout  time.Duration
}

// composeKSM implements the ClientComposer function signature
func composeKSM(source interface{}, cacher *endpoints.DiscoveryCacher) (endpoints.Client, error) {
	cached := source.(*cachedKSM)
	return &ksm{
		nodeIP:   cached.NodeIP,
		endpoint: cached.Endpoint,
		httpClient: &http.Client{
			Timeout: cached.Timeout,
		},
		logger: cacher.Logger,
	}, nil
}

// composeKSM implements the ClientDecomposer function signature
func decomposeKSM(source endpoints.Client, _ *endpoints.DiscoveryCacher) (interface{}, error) {
	ksm := source.(*ksm)
	return &cachedKSM{
		Endpoint: ksm.endpoint,
		NodeIP:   ksm.nodeIP,
		Timeout:  ksm.httpClient.Timeout,
	}, nil
}

// NewKSMDiscoveryCacher creates a new DiscoveryCacher that wraps a ksmDiscoverer and caches the data into the
// specified storage
func NewKSMDiscoveryCacher(ksmDiscoverer *ksmDiscoverer, storage storage.Storage) endpoints.Discoverer {
	return &endpoints.DiscoveryCacher{
		CachedDataPtr: &cachedKSM{},
		StorageKey:    cachedKSMKey,
		Discoverer:    ksmDiscoverer,
		Storage:       storage,
		Logger:        ksmDiscoverer.logger,
		Compose:       composeKSM,
		Decompose:     decomposeKSM,
	}
}
