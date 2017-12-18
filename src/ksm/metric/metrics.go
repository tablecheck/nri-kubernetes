package metric

import (
	"fmt"

	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
)

// K8sMetricSetTypeGuesser is the metric set type guesser for k8s integrations.
func K8sMetricSetTypeGuesser(groupLabel, _ string, _ definition.RawGroups) string {
	return fmt.Sprintf("K8s%vSample", strings.Title(groupLabel))
}

type namespaceFetcher func(groupLabel, entityId string, groups definition.RawGroups) string

// KubeletNamespaceFetcher fetches the namespace from a Kubelet RawGroups information
func KubeletNamespaceFetcher(groupLabel, entityId string, groups definition.RawGroups) string {
	namespace := groups[groupLabel][entityId]["namespace"]
	if namespace == nil {
		return ""
	}
	return namespace.(string)
}

// KSMNamespaceFetcher fetches the namespace from a KSM RawGroups information
func KSMNamespaceFetcher(groupLabel, entityId string, groups definition.RawGroups) string {
	ns, _ := FromPrometheusLabelValue("kube_deployment_labels", "namespace")(groupLabel, entityId, groups)
	return ns.(string)
}

func K8sMetricSetEntityTypeGuesserMalo(groupLabel, entityId string, groups definition.RawGroups) string {
	if groupLabel == "container" {
		groupLabel = "pod"
	}

	if groupLabel == "namespace" {
		return fmt.Sprintf("k8s:namespace")
	}
	return fmt.Sprintf("k8s:%s:%s", groups[groupLabel][entityId]["namespace"], groupLabel)
}

// K8sMetricSetEntityTypeGuesser guesses the Entity Type given a group name, entity Id and a namespace fetcher function
func K8sMetricSetEntityTypeGuesser(fetcher namespaceFetcher) func(groupLabel, entityId string, groups definition.RawGroups) string {
	return func(groupLabel, entityId string, groups definition.RawGroups) string {
		if groupLabel == "container" {
			groupLabel = "pod"
		}

		if groupLabel == "namespace" {
			return fmt.Sprintf("k8s:namespace")
		}
		return fmt.Sprintf("k8s:%s:%s", fetcher(groupLabel, entityId, groups), groupLabel)
	}
}

// FromPrometheusLabelValueEntityIDGenerator generates an entityID from the pod name. It's only used for k8s containers.
func FromPrometheusLabelValueEntityIDGenerator(key, label string) definition.MetricSetEntityIDGeneratorFunc {
	return func(groupLabel string, rawEntityID string, g definition.RawGroups) (string, error) {
		v, err := FromPrometheusLabelValue(key, label)(groupLabel, rawEntityID, g)
		if v == nil {
			return "", fmt.Errorf("error generating metric set entity id from prometheus label value. Key: %v, Label: %v", key, label)
		}

		return v.(string), err
	}
}

// GroupPrometheusMetricsBySpec groups metrics coming from Prometheus by a given metric spec.
// Example: grouping by K8s pod, container, etc.
func GroupPrometheusMetricsBySpec(specs definition.SpecGroups, families []prometheus.MetricFamily) (g definition.RawGroups, errs []error) {
	g = make(definition.RawGroups)
	for groupLabel := range specs {
		for _, f := range families {
			for _, m := range f.Metrics {
				if !m.Labels.Has(groupLabel) {
					continue
				}

				var rawEntityID string
				switch groupLabel {
				case "namespace":
					rawEntityID = m.Labels[groupLabel]
				case "container":
					rawEntityID = fmt.Sprintf("%v_%v_%v", m.Labels["namespace"], m.Labels["pod"], m.Labels[groupLabel])
				default:
					rawEntityID = fmt.Sprintf("%v_%v", m.Labels["namespace"], m.Labels[groupLabel])
				}

				if _, ok := g[groupLabel]; !ok {
					g[groupLabel] = make(map[string]definition.RawMetrics)
				}

				if _, ok := g[groupLabel][rawEntityID]; !ok {
					g[groupLabel][rawEntityID] = make(definition.RawMetrics)
				}

				g[groupLabel][rawEntityID][f.Name] = m
			}
		}

		if len(g[groupLabel]) == 0 {
			errs = append(errs, fmt.Errorf("no data found for %s object", groupLabel))
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
		metricKey, r := getRandomMetric(groups[groupLabel][entityID])
		m, ok := r.(prometheus.Metric)

		if !ok {
			return "", fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", m)
		}

		relatedMetricID, ok := m.Labels[group]
		if !ok {
			return nil, fmt.Errorf("label not found. Label: %s, Metric: %s", group, metricKey)
		}

		parent, err := definition.FromRaw(relatedMetricKey)(group, relatedMetricID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, group, relatedMetricID)
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
		metricKey, r := getRandomMetric(groups[groupLabel][entityID])
		m, ok := r.(prometheus.Metric)
		if !ok {
			return "", fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", m)
		}

		relatedMetricID, ok := m.Labels[parentGroupLabel]
		if !ok {
			return nil, fmt.Errorf("label not found. Label: %s, Metric: %s", parentGroupLabel, metricKey)
		}

		parent, err := fetchPrometheusMetric(relatedMetricKey)(parentGroupLabel, relatedMetricID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, parentGroupLabel, relatedMetricID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(prometheus.Metric).Labels {
			multiple[fmt.Sprintf("label.%v", k)] = v
		}

		return multiple, nil
	}
}

func getRandomMetric(metrics definition.RawMetrics) (metricKey string, value definition.RawValue) {
	for metricKey, value = range metrics {
		// We just want 1.
		break
	}

	return
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
