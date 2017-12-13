package metric

import (
	"fmt"

	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
)

// K8sMetricSetTypeGuesser is the metric set type guesser for k8s integrations.
func K8sMetricSetTypeGuesser(groupLabel, _ string, _ definition.RawGroups) string {
	return fmt.Sprintf("K8s%vSample", strings.Title(groupLabel))
}

// K8sMetricSetTypeGuesser is the metric set entity type guesser for k8s integrations.
func K8sMetricSetEntityTypeGuesser(groupLabel, _ string, _ definition.RawGroups) string {
	return fmt.Sprintf("k8s/%s", groupLabel)
}

// GroupPrometheusMetricsBySpec groups metrics coming from Prometheus by a given metric spec.
// Example: grouping by K8s pod, container, etc.
func GroupPrometheusMetricsBySpec(specs definition.Specs, families []prometheus.MetricFamily) (g definition.RawGroups, errs []error) {
	g = make(definition.RawGroups)
	for label := range specs {
		for _, f := range families {
			for _, m := range f.Metrics {
				if !m.Labels.Has(label) {
					continue
				}

				objectID := m.Labels[label]

				if _, ok := g[label]; !ok {
					g[label] = make(map[string]definition.RawMetrics)
				}

				if _, ok := g[label][objectID]; !ok {
					g[label][objectID] = make(definition.RawMetrics)
				}

				g[label][objectID][f.Name] = m
			}
		}

		if len(g[label]) == 0 {
			errs = append(errs, fmt.Errorf("no data found for %s object", label))
			continue
		}
	}

	return g, errs
}

// FromPrometheusValue creates a FetchFunc that fetches values from prometheus metrics values.
func FromPrometheusValue(key string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		return v.Value, nil
	}
}

// FromPrometheusLabelValue creates a FetchFunc that fetches values from prometheus metrics labels.
func FromPrometheusLabelValue(key, label string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		l, ok := v.Labels[label]
		if !ok {
			return nil, fmt.Errorf("label '%v' not found in prometheus metric", label)
		}

		return l, nil
	}
}

// InheritSpecificPrometheusLabelValuesFrom gets the specified label values from a related metric.
// Related metric means tany metric you can get with the info that you have in your own metric.
func InheritSpecificPrometheusLabelValuesFrom(group, relatedMetricKey string, labelsToRetrieve map[string]string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		var metricKey string
		var m prometheus.Metric
		for k, v := range groups[groupLabel][entityID] {
			metricKey = k
			pm, ok := v.(prometheus.Metric)
			if !ok {
				return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", pm)
			}

			m = pm

			// We just get 1 randomly.
			break
		}

		relatedMetricID, ok := m.Labels[group]
		if !ok {
			return nil, fmt.Errorf("label not found. Label: %s, Metric: %s", group, metricKey)
		}

		parent, err := definition.FromRaw(relatedMetricKey)(group, relatedMetricID, groups)
		if err != nil {
			return nil, fmt.Errorf("parent metric not found. %s:%s", group, relatedMetricID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(prometheus.Metric).Labels {
			for n, l := range labelsToRetrieve {
				if l == k {
					multiple[n] = v
				}
			}
		}

		return multiple, nil
	}
}

// InheritAllPrometheusLabelsFrom gets all the label values from from a related metric.
// Related metric means tany metric you can get with the info that you have in your own metric.
func InheritAllPrometheusLabelsFrom(parentGroupLabel, relatedMetricKey string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		var metricKey string
		var m prometheus.Metric
		for k, v := range groups[groupLabel][entityID] {
			metricKey = k
			pm, ok := v.(prometheus.Metric)
			if !ok {
				return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", pm)
			}

			m = pm

			// We just get 1 randomly.
			break
		}

		relatedMetricID, ok := m.Labels[parentGroupLabel]
		if !ok {
			return nil, fmt.Errorf("label not found. Label: %s, Metric: %s", parentGroupLabel, metricKey)
		}

		parent, err := fetchPrometheusMetric(relatedMetricKey)(parentGroupLabel, relatedMetricID, groups)
		if err != nil {
			return nil, fmt.Errorf("parent metric not found. %s:%s", parentGroupLabel, relatedMetricID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(prometheus.Metric).Labels {
			multiple[fmt.Sprintf("label.%v", k)] = v
		}

		return multiple, nil
	}
}

func fetchPrometheusMetric(metricKey string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {

		value, err := definition.FromRaw(metricKey)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		return v, nil
	}
}

// GetDeploymentNameForReplicaSet returns the name of the deployment has created
// a ReplicaSet.
func GetDeploymentNameForReplicaSet() definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		replicasetName, err := FromPrometheusLabelValue("kube_replicaset_created", "replicaset")(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		return replicasetNameToDeploymentName(replicasetName.(string)), nil
	}
}

// GetDeploymentNameForPod returns the name of the deployment has created a
// Pod.  It returns an empty string if Pod hasn't been created by a deployment.
func GetDeploymentNameForPod() definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		var deploymentName string
		creatorKind, err := FromPrometheusLabelValue("kube_pod_info", "created_by_kind")(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		creatorName, err := FromPrometheusLabelValue("kube_pod_info", "created_by_name")(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		if creatorKind.(string) == "ReplicaSet" {
			deploymentName = replicasetNameToDeploymentName(creatorName.(string))
		}
		return deploymentName, nil
	}
}

func replicasetNameToDeploymentName(rsName string) string {
	s := strings.Split(rsName, "-")
	return strings.Join(s[:len(s)-1], "-")
}
