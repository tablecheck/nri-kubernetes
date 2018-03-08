package endpoints

import (
	"fmt"
	"io/ioutil"
	"testing"

	endpoints2 "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
	"k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscover_CachedKSM(t *testing.T) {
	// Setup cache directory
	tmpDir, err := ioutil.TempDir("", "test_discover")
	assert.NoError(t, err)

	// Setup Kubernetes API client
	client := new(endpoints2.MockedClient)
	client.On("FindPodsByLabel", mock.Anything, mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Status: v1.PodStatus{HostIP: "6.7.8.9"},
		}}}, nil)

	// Setup storage
	store := storage.NewJSONDiskStorage(tmpDir)

	// Given a KSM discoverer
	wrappedDiscoverer := ksmDiscoverer{
		lookupSRV: fakeLookupSRV,
		apiClient: client,
		logger:    logger,
	}
	// That is wrapped into a Cached Discoverer
	cacher := NewKSMDiscoveryCacher(&wrappedDiscoverer, &store, logger)

	// And previously has discovered the KSM endpoint
	caClient, err := cacher.Discover(timeout)

	// When the discovery process is invoked again
	wrappedDiscoverer.lookupSRV = failingLookupSRV
	caClient, err = cacher.Discover(timeout)
	assert.NoError(t, err)

	// The cached value has been retrieved, instead of triggered the discovery
	// (otherwise it would have failed when invoking the `failedLookupSRV` and the unconfigured mock
	assert.NoError(t, err)
	ksmClient := endpoints2.WrappedClient(caClient)
	assert.Equal(t, fmt.Sprintf("%s:%v", ksmQualifiedName, 11223), ksmClient.(*ksm).endpoint.Host)
	assert.Equal(t, "http", ksmClient.(*ksm).endpoint.Scheme)
	assert.Equal(t, "6.7.8.9", ksmClient.(*ksm).nodeIP)

	assert.Equal(t, "6.7.8.9", caClient.NodeIP())
}

func TestDiscover_CachedKSM_BothFail(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_discover")
	assert.NoError(t, err)

	// Given a client that is unable to discover the endpoint
	client := new(endpoints2.MockedClient)
	client.On("FindPodsByLabel", mock.Anything, mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{}}, fmt.Errorf("error invoking Kubernetes API"))
	client.On("FindServiceByLabel", mock.Anything, mock.Anything).
		Return(&v1.ServiceList{Items: []v1.Service{}}, fmt.Errorf("error invoking Kubernetes API"))

	// And a cache that does not store any cached copy
	store := storage.NewJSONDiskStorage(tmpDir)

	// And a Cached KSM discoverer
	cacher := NewKSMDiscoveryCacher(
		&ksmDiscoverer{
			lookupSRV: fakeLookupSRV,
			apiClient: client,
			logger:    logger,
		}, &store, logger)

	// The Discover invocation should return error
	_, err = cacher.Discover(timeout)
	assert.Error(t, err)
}

func TestDiscover_LoadCacheFail(t *testing.T) {
	// Setup cache directory
	tmpDir, err := ioutil.TempDir("", "test_discover")
	assert.NoError(t, err)

	// Setup Kubernetes API client
	client := new(endpoints2.MockedClient)
	client.On("FindPodsByLabel", mock.Anything, mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Status: v1.PodStatus{HostIP: "6.7.8.9"},
		}}}, nil)

	// Setup storage
	store := storage.NewJSONDiskStorage(tmpDir)

	// Given a KSM discoverer
	wrappedDiscoverer := ksmDiscoverer{
		lookupSRV: fakeLookupSRV,
		apiClient: client,
		logger:    logger,
	}
	// That is wrapped into a Cached Discoverer
	cacher := NewKSMDiscoveryCacher(&wrappedDiscoverer, &store, logger)

	// And previously has discovered the KSM endpoint
	caClient, err := cacher.Discover(timeout)

	// But the cache stored data is corrupted
	assert.Nil(t, store.Write(cachedKSMKey, "corrupt-data"))

	// When the discovery process is invoked again
	caClient, err = cacher.Discover(timeout)

	// The discovery process has been triggered again
	assert.NoError(t, err)
	ksmClient := endpoints2.WrappedClient(caClient)
	assert.Equal(t, fmt.Sprintf("%s:%v", ksmQualifiedName, 11223), ksmClient.(*ksm).endpoint.Host)
	assert.Equal(t, "http", ksmClient.(*ksm).endpoint.Scheme)
	assert.Equal(t, "6.7.8.9", caClient.NodeIP())

}
