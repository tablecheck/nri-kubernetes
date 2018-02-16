package endpoints

import (
	"net"
	"testing"

	"fmt"

	endpoints2 "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

func fakeLookupSRV(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	return "cname", []*net.SRV{{Port: 11223}}, nil
}

func emptyLookupSRV(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	return "cname", []*net.SRV{}, nil
}

func failingLookupSRV(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	return "cname", nil, fmt.Errorf("patapum")
}

func TestKSMDiscover_DNS(t *testing.T) {
	// Given an Discoverer implementation
	endpoints := ksmDiscoverer{lookupSRV: fakeLookupSRV}

	// When retrieving the KSM URL
	kurl, _, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the KSM Service is returned
	assert.Equal(t, fmt.Sprintf("%s:%v", ksmQualifiedName, 11223), kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKSMDiscover_API(t *testing.T) {
	// Given a client
	client := new(endpoints2.MockedClient)
	client.On("FindServiceByLabel", mock.Anything, mock.Anything).
		Return(&v1.ServiceList{Items: []v1.Service{{
			Spec: v1.ServiceSpec{
				ClusterIP: "1.2.3.4",
				Ports: []v1.ServicePort{{
					Name: ksmPortName,
					Port: 8888,
				}},
			},
		},
		}}, nil)

	// and an Discoverer implementation whose DNS returns empty response
	endpoints := ksmDiscoverer{
		lookupSRV: emptyLookupSRV,
		apiClient: client,
	}

	// When retrieving the KSM URL
	kurl, _, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the KSM Service is returned
	assert.Equal(t, "1.2.3.4:8888", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKSMDiscover_API_afterError(t *testing.T) {
	// Given a client
	client := new(endpoints2.MockedClient)
	client.On("FindServiceByLabel", mock.Anything, mock.Anything).
		Return(&v1.ServiceList{Items: []v1.Service{{
			Spec: v1.ServiceSpec{
				ClusterIP: "1.2.3.4",
				Ports: []v1.ServicePort{{
					Name: ksmPortName,
					Port: 8888,
				}},
			},
		},
		}}, nil)

	// and an Discoverer implementation whose DNS returns an error
	endpoints := ksmDiscoverer{
		lookupSRV: failingLookupSRV,
		apiClient: client,
	}

	// When retrieving the KSM URL
	kurl, _, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the KSM Service is returned
	assert.Equal(t, "1.2.3.4:8888", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKSMDiscover_API_GuessTCPPort(t *testing.T) {
	// Given a client
	client := new(endpoints2.MockedClient)
	client.On("FindServiceByLabel", mock.Anything, mock.Anything).
		Return(&v1.ServiceList{Items: []v1.Service{{
			Spec: v1.ServiceSpec{
				ClusterIP: "11.22.33.44",
				Ports: []v1.ServicePort{{
					Name:     "SomeCoolPort",
					Protocol: "UDP",
					Port:     1234,
				}, {
					Name:     "ThisPortShouldWork",
					Protocol: "TCP",
					Port:     8081,
				}},
			}}}}, nil)

	// and an Discoverer implementation
	endpoints := ksmDiscoverer{
		lookupSRV: emptyLookupSRV,
		apiClient: client,
	}

	// When retrieving the KSM URL with no port named 'http-metrics'
	kurl, _, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the first TCP host:port of the KSM Service is returned
	assert.Equal(t, "11.22.33.44:8081", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKsmDiscoverer_NodeIP(t *testing.T) {
	// Given a client
	client := new(endpoints2.MockedClient)
	client.On("FindPodsByLabel", mock.Anything, mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{{
			Status: v1.PodStatus{HostIP: "6.7.8.9"},
		}}}, nil)

	// and an Discoverer implementation
	endpoints := ksmDiscoverer{
		apiClient: client,
	}

	// When getting the Node IP
	nodeIP, err := endpoints.NodeIP()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the nodeIP is correctly returned
	assert.Equal(t, "6.7.8.9", nodeIP)
}

func TestKsmDiscoverer_NodeIP_MultiplePods(t *testing.T) {
	// Given a client
	client := new(endpoints2.MockedClient)
	client.On("FindPodsByLabel", mock.Anything, mock.Anything).
		Return(&v1.PodList{Items: []v1.Pod{
			{Status: v1.PodStatus{HostIP: "6.7.8.9"}},
			{Status: v1.PodStatus{HostIP: "162.178.1.1"}},
			{Status: v1.PodStatus{HostIP: "4.3.2.1"}},
		}}, nil)

	// and an Discoverer implementation
	endpoints := ksmDiscoverer{
		apiClient: client,
	}

	// When getting the Node IP
	nodeIP, err := endpoints.NodeIP()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the nodeIP is correctly returned
	assert.Equal(t, "162.178.1.1", nodeIP)
}
