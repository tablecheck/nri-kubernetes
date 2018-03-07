package endpoints

import (
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"github.com/sirupsen/logrus"
)

// DiscoveryCacher implements the Discoverer API to read endpoints from a cache storage. It also wraps another
// Discoverer and uses it to discover endpoints when the data is not found in the cache.
// This type is not thread-safe.
type DiscoveryCacher struct {
	// CachedDataPtr must be a pointer to an object where the data will be unmarshalled to
	CachedDataPtr interface{}
	// StorageKey is the key for the Storage Cache
	StorageKey string
	// Discoverer points to the wrapped Discovered used to resolve endpoints when they are not found in the cache
	Discoverer Discoverer
	// Storage for cached data
	Storage   storage.Storage
	Logger    *logrus.Logger
	Compose   ClientComposer
	Decompose ClientDecomposer
}

// ClientDecomposer implementors must convert a Client into a data structure that can be Stored in the cache.
type ClientDecomposer func(source Client) (interface{}, error)

// ClientComposer implementors must convert the data from the cached entities to a Client.
type ClientComposer func(source interface{}, cacher *DiscoveryCacher, timeout time.Duration) (Client, error)

// Discover tries to retrieve a Client from the cache, and otherwise engage the discovery process from the wrapped
// Discoverer
func (d *DiscoveryCacher) Discover(timeout time.Duration) (Client, error) {
	ts, err := d.Storage.Read(d.StorageKey, d.CachedDataPtr)
	if err == nil {
		d.Logger.Debug("Found cached copy of %q stored at %s", d.StorageKey, time.Unix(ts, 0))
		return d.Compose(d.CachedDataPtr, d, timeout)
	}
	d.Logger.Debug("Cached %q not found. Triggering discovery process", d.StorageKey)
	// If the load-from-caching process failed, we trigger the discovery process
	client, err := d.Discoverer.Discover(timeout)
	if err != nil {
		return nil, err
	}
	// and store the discovered data into the cache
	toCache, err := d.Decompose(client)
	if err == nil {
		err = d.Storage.Write(d.StorageKey, toCache)
	}
	if err != nil {
		d.Logger.WithError(err).Warn("while storing %q in the cache", d.StorageKey)
	}
	return client, nil
}
