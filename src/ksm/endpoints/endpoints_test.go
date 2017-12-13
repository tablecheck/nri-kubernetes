package endpoints

import (
	"testing"

	endpoints2 "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

func TestKSMDiscover(t *testing.T) {
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

	// and an Discoverer implementation
	endpoints := ksmDiscoverer{
		client: client,
	}

	// When retrieving the KSM URL
	kurl, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the discovered host:port of the KSM Service is returned
	assert.Equal(t, "1.2.3.4:8888", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}

func TestKSMDiscover_GuessTCPPort(t *testing.T) {
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
		client: client,
	}

	// When retrieving the KSM URL with no port named 'http-metrics'
	kurl, err := endpoints.Discover()
	// The call works correctly
	assert.Nil(t, err, "should not return error")
	// And the first TCP host:port of the KSM Service is returned
	assert.Equal(t, "11.22.33.44:8081", kurl.Host)
	assert.Equal(t, "http", kurl.Scheme)
}
