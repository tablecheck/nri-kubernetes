package metric

import (
	"testing"

	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/stretchr/testify/assert"
)

func validNamespaceFetcher(expected string) NamespaceFetcher {
	return func(groupLabel, entityID string, groups definition.RawGroups) (string, error) {
		return expected, nil
	}
}

func errorNamespaceFetcher(err error) NamespaceFetcher {
	return func(groupLabel, entityID string, groups definition.RawGroups) (string, error) {
		return UnknownNamespace, err
	}
}

func TestMultipleNamespaceFetcher(t *testing.T) {
	namespace, err := MultipleNamespaceFetcher("", "", nil,
		errorNamespaceFetcher(fmt.Errorf("error")),
		validNamespaceFetcher("validNamespace"),
		validNamespaceFetcher("thisShouldNeverBeReturned"),
	)

	assert.Equal(t, "validNamespace", namespace)
	assert.Nil(t, err)
}

func TestMultipleNamespaceFetcher_Error(t *testing.T) {
	namespace, err := MultipleNamespaceFetcher("", "", nil,
		errorNamespaceFetcher(fmt.Errorf("error1")),
		errorNamespaceFetcher(fmt.Errorf("error2")),
		errorNamespaceFetcher(fmt.Errorf("error3")),
	)

	assert.Equal(t, UnknownNamespace, namespace)
	assert.Equal(t, "error fetching namespace: error1, error2, error3", err.Error())
}
