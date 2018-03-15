package metric

import (
	"errors"

	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/data"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type k8sPopulator struct {
}

// MultipleErrs represents a bunch of errs.
// Recoverable == true means that you can keep working with those errors.
// Recoverable == false means you must handle the errors or panic.
type MultipleErrs struct {
	Recoverable bool
	Errs        []error
}

// Error implements error interface
func (e MultipleErrs) Error() string {
	s := "multiple errors:"

	for _, err := range e.Errs {
		s = fmt.Sprintf("%s\n%s", s, err)
	}
	return s
}

// Populate populates k8s raw data to sdk metrics.
func (p *k8sPopulator) Populate(groups definition.RawGroups, specGroups definition.SpecGroups, i *sdk.IntegrationProtocol2, clusterName string) (bool, error) {
	populatorFunc := definition.IntegrationProtocol2PopulateFunc(i, clusterName, K8sMetricSetTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)
	ok, errs := populatorFunc(groups, specGroups)

	if len(errs) > 0 {
		return true, MultipleErrs{true, errs}
	}

	if !ok {
		// TODO better error
		return false, errors.New("no data was populated")
	}

	return true, nil
}

// NewK8sPopulator creates a Kubernetes aware populator.
func NewK8sPopulator() data.Populator {
	return &k8sPopulator{}
}
