package metric

import (
	"fmt"

	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

const (
	unknownNamespace = "_unknown"
)

// K8sMetricSetTypeGuesser is the metric set type guesser for k8s integrations.
func K8sMetricSetTypeGuesser(_, groupLabel, _ string, _ definition.RawGroups) (string, error) {
	return fmt.Sprintf("K8s%vSample", strings.Title(groupLabel)), nil
}

type namespaceFetcher func(groupLabel, entityId string, groups definition.RawGroups) (string, error)

// KubeletKSMNamespaceFetcher fetches the namespace from the RawGroups information trying both Kubelet and KSM
func kubeletKSMNamespaceFetcher(groupLabel, entityId string, groups definition.RawGroups) (string, error) {
	ns, err1 := kubeletNamespaceFetcher(groupLabel, entityId, groups)
	if err1 == nil {
		return ns, nil
	}
	ns, err2 := ksmNamespaceFetcher(groupLabel, entityId, groups)
	if err2 == nil {
		return ns, nil
	}
	return unknownNamespace, fmt.Errorf("error fetching namespace: %q, %q", err1.Error(), err2.Error())
}

// kubeletNamespaceFetcher fetches the namespace from a Kubelet RawGroups information
func kubeletNamespaceFetcher(groupLabel, entityId string, groups definition.RawGroups) (string, error) {
	gl, ok := groups[groupLabel]
	if !ok {
		return unknownNamespace, fmt.Errorf("no grouplabel %q found", groupLabel)
	}
	en, ok := gl[entityId]
	if !ok {
		return unknownNamespace, fmt.Errorf("no entityId %q found for grouplabel %q", entityId, groupLabel)
	}

	ns, ok := en["namespace"]
	if !ok {
		return unknownNamespace, fmt.Errorf("no namespace found for groupLabel %q and entityId %q", groupLabel, entityId)
	}
	return ns.(string), nil
}

var nsKeyForGroup = map[string]string{
	"pod":        "kube_pod_info",
	"replicaset": "kube_replicaset_created",
	"container":  "kube_pod_container_info",
	"namespace":  "kube_namespace_created",
	"deployment": "kube_deployment_labels",
}

// ksmNamespaceFetcher fetches the namespace from a KSM RawGroups information
func ksmNamespaceFetcher(groupLabel, entityId string, groups definition.RawGroups) (string, error) {
	ns, err := FromPrometheusLabelValue(nsKeyForGroup[groupLabel], "namespace")(groupLabel, entityId, groups)
	if err != nil {
		return unknownNamespace, fmt.Errorf("error fetching namespace for groupLabel %q and entityId %q: %v", groupLabel, entityId, err.Error())
	}
	if ns == nil {
		return unknownNamespace, fmt.Errorf("namespace not found for groupLabel %q and entityId %q", groupLabel, entityId)
	}
	return ns.(string), nil
}

// K8sMetricSetEntityTypeGuesser guesses the Entity Type given a group name, entity Id and a namespace fetcher function
func K8sMetricSetEntityTypeGuesser(clusterName, groupLabel, entityId string, groups definition.RawGroups) (string, error) {
	var actualGroupLabel string
	switch groupLabel {
	case "namespace":
		return fmt.Sprintf("k8s:%s:namespace", clusterName), nil
	case "container":
		actualGroupLabel = "pod"
	default:
		actualGroupLabel = groupLabel
	}
	ns, err := kubeletKSMNamespaceFetcher(groupLabel, entityId, groups)
	return fmt.Sprintf("k8s:%s:%s:%s", clusterName, ns, actualGroupLabel), err
}

func K8sClusterMetricsManipulator(ms metric.MetricSet, _ sdk.Entity, clusterName string) {
	ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
}

// K8sEntityMetricsManipulator adds 'displayName' and 'entityName' metrics to
// the MetricSet displayName and entityName, taking values from entity.name and
// entity.type
func K8sEntityMetricsManipulator(ms metric.MetricSet, entity sdk.Entity, _ string) {
	ms.SetMetric("displayName", entity.Name, metric.ATTRIBUTE)
	ms.SetMetric("entityName", fmt.Sprintf("%s:%s", entity.Type, entity.Name), metric.ATTRIBUTE)
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
// Related metric means any metric you can get with the info that you have in your own metric.
func InheritSpecificPrometheusLabelValuesFrom(parentGroupLabel, relatedMetricKey string, labelsToRetrieve map[string]string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		rawEntityID, err := getRawEntityID(parentGroupLabel, groupLabel, entityID, groups)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve the entity ID of metrics to inherit value from, got error: %v", err)
		}
		parent, err := definition.FromRaw(relatedMetricKey)(parentGroupLabel, rawEntityID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, parentGroupLabel, rawEntityID)
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
// Related metric means any metric you can get with the info that you have in your own metric.
func InheritAllPrometheusLabelsFrom(parentGroupLabel, relatedMetricKey string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		rawEntityID, err := getRawEntityID(parentGroupLabel, groupLabel, entityID, groups)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve the entity ID of metrics to inherit labels from, got error: %v", err)
		}

		parent, err := fetchPrometheusMetric(relatedMetricKey)(parentGroupLabel, rawEntityID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, parentGroupLabel, rawEntityID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(prometheus.Metric).Labels {
			multiple[fmt.Sprintf("label.%v", k)] = v
		}

		return multiple, nil
	}
}

func getRawEntityID(parentGroupLabel, groupLabel, entityID string, groups definition.RawGroups) (string, error) {
	group, ok := groups[groupLabel][entityID]
	if !ok {
		return "", fmt.Errorf("metrics not found for %v with entity ID: %v", groupLabel, entityID)
	}
	metricKey, r := getRandomPrometheusMetric(group)
	m, ok := r.(prometheus.Metric)

	if !ok {
		return "", fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", r)
	}

	namespaceID, ok := m.Labels["namespace"]
	if !ok {
		return "", fmt.Errorf("label not found. Label: 'namespace', Metric: %s", metricKey)
	}

	var rawEntityID string
	switch parentGroupLabel {
	case "namespace":
		rawEntityID = namespaceID
	default:
		relatedMetricID, ok := m.Labels[parentGroupLabel]
		if !ok {
			return "", fmt.Errorf("label not found. Label: %s, Metric: %s", parentGroupLabel, metricKey)
		}
		rawEntityID = fmt.Sprintf("%v_%v", namespaceID, relatedMetricID)
	}
	return rawEntityID, nil
}

func getRandomPrometheusMetric(metrics definition.RawMetrics) (metricKey string, value definition.RawValue) {
	for metricKey, value = range metrics {
		if _, ok := value.(prometheus.Metric); !ok {
			continue
		}
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

// GetStatusForContainer returns the status of a container
func GetStatusForContainer() definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		queryValue := prometheus.GaugeValue(1)
		s := []string{"running", "waiting", "terminated"}
		for _, k := range s {
			v, _ := FromPrometheusValue(fmt.Sprintf("kube_pod_container_status_%s", k))(groupLabel, entityID, groups)
			if v == queryValue {
				return strings.Title(k), nil
			}
		}

		return "Unknown", nil
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
		creatorKind, err := FromPrometheusLabelValue("kube_pod_info", "created_by_kind")(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		creatorName, err := FromPrometheusLabelValue("kube_pod_info", "created_by_name")(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		return deploymentNameBasedOnCreator(creatorKind.(string), creatorName.(string)), nil
	}
}

// GetDeploymentNameForContainer returns the name of the deployment has created
// a container. It's providing this information inheriting some metrics from its
// pod. Returns an empty string if its pod hasn't been created by a deployment.
func GetDeploymentNameForContainer() definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		mm := map[string]string{
			"created_by_kind": "created_by_kind",
			"created_by_name": "created_by_name",
		}
		podValues, err := InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", mm)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}
		podMetrics := podValues.(definition.FetchedValues)
		return deploymentNameBasedOnCreator(podMetrics["created_by_kind"].(string), podMetrics["created_by_name"].(string)), nil

	}
}

func deploymentNameBasedOnCreator(creatorKind, creatorName string) string {
	var deploymentName string
	if creatorKind == "ReplicaSet" {
		deploymentName = replicasetNameToDeploymentName(creatorName)
	}
	return deploymentName
}

func replicasetNameToDeploymentName(rsName string) string {
	s := strings.Split(rsName, "-")
	return strings.Join(s[:len(s)-1], "-")
}

// UnscheduledItemsPatcher adds to the destination RawGroups the pods that haven't been scheduled
func UnscheduledItemsPatcher(destination definition.RawGroups, source definition.RawGroups) {
	for podName, pod := range source["pod"] {
		if _, ok := destination["pod"][podName]; !ok {
			podMap := pod["kube_pod_info"].(prometheus.Metric).Labels
			if podMap["node"] == "" {
				destination["pod"][podName] = definition.RawMetrics{}
				destination["pod"][podName]["podName"] = podMap["pod"]
				destination["pod"][podName]["namespace"] = podMap["namespace"]
			}
		}
	}
}
