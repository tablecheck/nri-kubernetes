package metric

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

// NamespaceFetcher fetches the namespace from the provided RawGroups, for the given groupLabel and entityId
type NamespaceFetcher func(groupLabel, entityID string, groups definition.RawGroups) (string, error)

// MultipleNamespaceFetcher asks multiple NamespaceFetcher instances for the namespace. It queries the NamespaceFetchers
// in the provided order and returns the namespace from the first invocation that does not return error. If all the
// invocations fail, this function returns "_unknown" namespace and an error instance.
func MultipleNamespaceFetcher(groupLabel, entityID string, groups definition.RawGroups, fetchers ...NamespaceFetcher) (string, error) {

	errors := make([]string, 0)

	for _, fetcher := range fetchers {
		ns, err := fetcher(groupLabel, entityID, groups)
		if err == nil {
			return ns, nil
		}
		errors = append(errors, err.Error())
	}

	return config.UnknownNamespace, fmt.Errorf("error fetching namespace: %s", strings.Join(errors, ", "))
}

// K8sMetricSetEntityTypeGuesser guesses the Entity Type given a group name, entity Id and a namespace fetcher function
func K8sMetricSetEntityTypeGuesser(clusterName, groupLabel, entityID string, groups definition.RawGroups) (string, error) {
	var actualGroupLabel string
	switch groupLabel {
	case "namespace", "node":
		return fmt.Sprintf("k8s:%s:%s", clusterName, groupLabel), nil
	case "container":
		actualGroupLabel = "pod"
	default:
		actualGroupLabel = groupLabel
	}
	ns, err := MultipleNamespaceFetcher(groupLabel, entityID, groups,
		kubeletMetric.KubeletNamespaceFetcher,
		ksmMetric.KsmNamespaceFetcher)
	return fmt.Sprintf("k8s:%s:%s:%s", clusterName, ns, actualGroupLabel), err
}

// K8sMetricSetTypeGuesser is the metric set type guesser for k8s integrations.
func K8sMetricSetTypeGuesser(_, groupLabel, _ string, _ definition.RawGroups) (string, error) {
	return fmt.Sprintf("K8s%vSample", strings.Title(groupLabel)), nil
}

// K8sClusterMetricsManipulator adds 'clusterName' metric to the MetricSet 'ms',
// taking the value from 'clusterName' argument.
func K8sClusterMetricsManipulator(ms metric.MetricSet, _ sdk.Entity, clusterName string) error {
	return ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
}

// K8sEntityMetricsManipulator adds 'displayName' and 'entityName' metrics to
// the MetricSet, taking values from entity.name and entity.type
func K8sEntityMetricsManipulator(ms metric.MetricSet, entity sdk.Entity, _ string) error {
	err := ms.SetMetric("displayName", entity.Name, metric.ATTRIBUTE)
	if err != nil {
		return err
	}
	return ms.SetMetric("entityName", fmt.Sprintf("%s:%s", entity.Type, entity.Name), metric.ATTRIBUTE)
}
