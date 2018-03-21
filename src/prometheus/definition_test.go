package prometheus

import (
	"errors"
	"fmt"
	"testing"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/stretchr/testify/assert"
)

var mFamily = []MetricFamily{
	{
		Name: "kube_pod_start_time",
		Metrics: []Metric{
			{
				Value: GaugeValue(1507117436),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "fluentd-elasticsearch-jnqb7",
				},
			},
			{
				Value: GaugeValue(1510579152),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "newrelic-infra-monitoring-cglrn",
				},
			},
		},
	},
	{
		Name: "kube_pod_info",
		Metrics: []Metric{
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "DaemonSet",
					"created_by_name": "fluentd-elasticsearch",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "fluentd-elasticsearch-jnqb7",
				},
			},
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "DaemonSet",
					"created_by_name": "newrelic-infra-monitoring",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "newrelic-infra-monitoring-cglrn",
				},
			},
		},
	},
	{
		Name: "kube_pod_labels",
		Metrics: []Metric{
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"label_app":                      "newrelic-infra-monitoring",
					"label_controller_revision_hash": "1758702902",
					"label_pod_template_generation":  "1",
					"namespace":                      "kube-system",
					"pod":                            "newrelic-infra-monitoring-cglrn",
				},
			},
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"label_name":                     "fluentd-elasticsearch",
					"label_controller_revision_hash": "3534845553",
					"label_pod_template_generation":  "1",
					"namespace":                      "kube-system",
					"pod":                            "fluentd-elasticsearch-jnqb7",
				},
			},
		},
	},
}

var spec = []definition.Spec{
	{"podStartTime", FromPrometheusValue("kube_pod_start_time"), metric.GAUGE},
	{"podInfo.namespace", FromPrometheusLabelValue("kube_pod_info", "namespace"), metric.ATTRIBUTE},
	{"podInfo.pod", FromPrometheusLabelValue("kube_pod_info", "pod"), metric.ATTRIBUTE},
}

var containersSpec = definition.SpecGroups{
	"container": definition.SpecGroup{
		Specs: []definition.Spec{
			{"container", FromPrometheusLabelValue("kube_pod_container_info", "container"), metric.ATTRIBUTE},
			{"image", FromPrometheusLabelValue("kube_pod_container_info", "image"), metric.ATTRIBUTE},
			{"namespace", FromPrometheusLabelValue("kube_pod_container_info", "namespace"), metric.ATTRIBUTE},
			{"pod", FromPrometheusLabelValue("kube_pod_container_info", "pod"), metric.ATTRIBUTE},
		},
	},
}

var specs = definition.SpecGroups{
	"pod": definition.SpecGroup{
		Specs: spec,
	},
}

var metricFamilyContainersWithTheSameName = []MetricFamily{
	{
		Name: "kube_pod_container_info",
		Metrics: []Metric{
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"container": "kube-state-metrics",
					"image":     "gcr.io/google_containers/kube-state-metrics:v1.1.0",
					"namespace": "kube-system",
					"pod":       "newrelic-infra-monitoring-3bxnh",
				},
			},
			{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"container": "kube-state-metrics",
					"image":     "gcr.io/google_containers/kube-state-metrics:v1.1.0",
					"namespace": "kube-system",
					"pod":       "fluentd-elasticsearch-jnqb7",
				},
			},
		},
	},
}

var rawGroups = definition.RawGroups{
	"pod": {
		"fluentd-elasticsearch-jnqb7": definition.RawMetrics{
			"kube_pod_start_time": Metric{
				Value: GaugeValue(1507117436),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "fluentd-elasticsearch-jnqb7",
				},
			},
			"kube_pod_info": Metric{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "ReplicaSet",
					"created_by_name": "fluentd-elasticsearch-fafnoa",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "fluentd-elasticsearch-jnqb7",
				},
			},
		},
		"newrelic-infra-monitoring-cglrn": definition.RawMetrics{
			"kube_pod_start_time": Metric{
				Value: GaugeValue(1510579152),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "newrelic-infra-monitoring-cglrn",
				},
			},
			"kube_pod_info": Metric{
				Value: GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "DaemonSet",
					"created_by_name": "newrelic-infra-monitoring",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "newrelic-infra-monitoring-cglrn",
				},
			},
		},
	},
}

var rawGroupsIncompatibleType = definition.RawGroups{
	"pod": {
		"fluentd-elasticsearch-jnqb7": definition.RawMetrics{
			"kube_pod_start_time": "foo",
		},
	},
}

// --------------- GroupPrometheusMetricsBySpec ---------------
func TestGroupPrometheusMetricsBySpec_CorrectValue(t *testing.T) {
	expectedMetricGroup := definition.RawGroups{
		"pod": {
			"kube-system_fluentd-elasticsearch-jnqb7": definition.RawMetrics{
				"kube_pod_start_time": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "fluentd-elasticsearch-jnqb7",
					},
				},
				"kube_pod_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"created_by_kind": "DaemonSet",
						"created_by_name": "fluentd-elasticsearch",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "fluentd-elasticsearch-jnqb7",
					},
				},
				"kube_pod_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"label_name":                     "fluentd-elasticsearch",
						"label_controller_revision_hash": "3534845553",
						"label_pod_template_generation":  "1",
						"namespace":                      "kube-system",
						"pod":                            "fluentd-elasticsearch-jnqb7",
					},
				},
			},
			"kube-system_newrelic-infra-monitoring-cglrn": definition.RawMetrics{
				"kube_pod_start_time": Metric{
					Value: GaugeValue(1510579152),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"created_by_kind": "DaemonSet",
						"created_by_name": "newrelic-infra-monitoring",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"label_app":                      "newrelic-infra-monitoring",
						"label_controller_revision_hash": "1758702902",
						"label_pod_template_generation":  "1",
						"namespace":                      "kube-system",
						"pod":                            "newrelic-infra-monitoring-cglrn",
					},
				},
			},
		},
	}

	metricGroup, errs := GroupPrometheusMetricsBySpec(specs, mFamily)
	assert.Empty(t, errs)
	assert.Equal(t, expectedMetricGroup, metricGroup)
}

func TestGroupPrometheusMetricsBySpec_CorrectValue_ContainersWithTheSameName(t *testing.T) {
	expectedMetricGroup := definition.RawGroups{
		"container": {
			"kube-system_fluentd-elasticsearch-jnqb7_kube-state-metrics": definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"container": "kube-state-metrics",
						"image":     "gcr.io/google_containers/kube-state-metrics:v1.1.0",
						"namespace": "kube-system",
						"pod":       "fluentd-elasticsearch-jnqb7",
					},
				},
			},
			"kube-system_newrelic-infra-monitoring-3bxnh_kube-state-metrics": definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"container": "kube-state-metrics",
						"image":     "gcr.io/google_containers/kube-state-metrics:v1.1.0",
						"namespace": "kube-system",
						"pod":       "newrelic-infra-monitoring-3bxnh",
					},
				},
			},
		},
	}

	metricGroup, errs := GroupPrometheusMetricsBySpec(containersSpec, metricFamilyContainersWithTheSameName)
	assert.Empty(t, errs)
	assert.Equal(t, expectedMetricGroup, metricGroup)
}

func TestGroupPrometheusMetricsBySpec_EmptyMetricFamily(t *testing.T) {
	var emptyMetricFamily []MetricFamily

	metricGroup, errs := GroupPrometheusMetricsBySpec(specs, emptyMetricFamily)
	assert.Len(t, errs, 1)
	assert.Equal(t, errors.New("no data found for pod object"), errs[0])
	assert.Empty(t, metricGroup)
}

// --------------- FromPrometheusValue ---------------
func TestFromRawPrometheusValue_CorrectValue(t *testing.T) {
	expectedFetchedValue := GaugeValue(1507117436)

	fetchedValue, err := FromPrometheusValue("kube_pod_start_time")("pod", "fluentd-elasticsearch-jnqb7", rawGroups)
	assert.Equal(t, expectedFetchedValue, fetchedValue)
	assert.NoError(t, err)
}

func TestFromRawPrometheusValue_RawMetricNotFound(t *testing.T) {

	fetchedValue, err := FromPrometheusValue("foo")("pod", "fluentd-elasticsearch-jnqb7", rawGroups)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, "FromRaw: metric not found. SpecGroup: pod, EntityID: fluentd-elasticsearch-jnqb7, Metric: foo")
}

func TestFromRawPrometheusValue_IncompatibleType(t *testing.T) {

	fetchedValue, err := FromPrometheusValue("kube_pod_start_time")("pod", "fluentd-elasticsearch-jnqb7", rawGroupsIncompatibleType)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, "incompatible metric type. Expected: Metric. Got: string")
}

// --------------- FromPrometheusLabelValue ---------------
func TestFromRawPrometheusLabelValue_CorrectValue(t *testing.T) {
	expectedFetchedValue := "kube-system"

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "namespace")("pod", "fluentd-elasticsearch-jnqb7", rawGroups)
	assert.Equal(t, expectedFetchedValue, fetchedValue)
	assert.NoError(t, err)
}

func TestFromRawPrometheusLabelValue_RawMetricNotFound(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("foo", "namespace")("pod", "fluentd-elasticsearch-jnqb7", rawGroups)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, "FromRaw: metric not found. SpecGroup: pod, EntityID: fluentd-elasticsearch-jnqb7, Metric: foo")
}

func TestFromRawPrometheusLabelValue_IncompatibleType(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "namespace")("pod", "fluentd-elasticsearch-jnqb7", rawGroupsIncompatibleType)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, "incompatible metric type. Expected: Metric. Got: string")
}

func TestFromRawPrometheusLabelValue_LabelNotFoundInRawMetric(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "foo")("pod", "fluentd-elasticsearch-jnqb7", rawGroups)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, "label 'foo' not found in prometheus metric")
}

// --------------- InheritSpecificPrometheusLabelValuesFrom ---------------

func TestInheritSpecificPrometheusLabelValuesFrom(t *testing.T) {
	containerRawEntityID := "kube-system_kube-addon-manager-minikube_kube-addon-manager"
	raw := definition.RawGroups{
		"pod": {
			"kube-system_kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
		"container": {
			containerRawEntityID: definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod":          "kube-addon-manager-minikube",
						"container_id": "docker://441e4dacbcfb2f012f2221d0f3768552ea1ccb53454da42b7b3eeaf17bbd240a",
						"namespace":    "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"inherited-pod_ip": "pod_ip"})("container", containerRawEntityID, raw)
	assert.NoError(t, err)

	expectedValue := definition.FetchedValues{"inherited-pod_ip": "172.31.248.38"}
	assert.Equal(t, expectedValue, fetchedValue)
}

func TestInheritSpecificPrometheusLabelsFrom_Namespace(t *testing.T) {
	podRawEntityID := "kube-addon-manager-minikube"
	raw := definition.RawGroups{
		"namespace": {
			"kube-system": definition.RawMetrics{
				"kube_namespace_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"namespace": "kube-system",
					},
				},
			},
		},
		"pod": {
			"kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritSpecificPrometheusLabelValuesFrom("namespace", "kube_namespace_labels", map[string]string{"inherited-namespace": "namespace"})("pod", podRawEntityID, raw)
	assert.NoError(t, err)

	expectedValue := definition.FetchedValues{"inherited-namespace": "kube-system"}
	assert.Equal(t, expectedValue, fetchedValue)
}
func TestInheritSpecificPrometheusLabelValuesFrom_RelatedMetricNotFound(t *testing.T) {
	containerRawEntityID := "kube-system_kube-addon-manager-minikube_kube-addon-manager"
	raw := definition.RawGroups{
		"pod": {},
		"container": {
			containerRawEntityID: definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"namespace": "kube-system",
					},
				},
			},
		},
	}

	expectedPodRawEntityID := "kube-system_kube-addon-manager-minikube"
	fetchedValue, err := InheritSpecificPrometheusLabelValuesFrom("pod", "non_existent_metric_key", map[string]string{"inherited-pod_ip": "pod_ip"})("container", containerRawEntityID, raw)
	assert.EqualError(t, err, fmt.Sprintf("related metric not found. Metric: non_existent_metric_key pod:%v", expectedPodRawEntityID))
	assert.Empty(t, fetchedValue)
}

func TestInheritSpecificPrometheusLabelValuesFrom_NamespaceNotFound(t *testing.T) {
	containerRawEntityID := "kube-system_kube-addon-manager-minikube_kube-addon-manager"
	raw := definition.RawGroups{
		"pod": {
			"kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
		"container": {
			containerRawEntityID: definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod": "kube-addon-manager-minikube",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"inherited-pod_ip": "pod_ip"})("container", containerRawEntityID, raw)
	assert.EqualError(t, err, "cannot retrieve the entity ID of metrics to inherit value from, got error: label not found. Label: 'namespace', Metric: kube_pod_container_info")
	assert.Empty(t, fetchedValue)
}

func TestInheritSpecificPrometheusLabelValuesFrom_GroupNotFound(t *testing.T) {
	incorrectContainerRawEntityID := "non-existing-ID"
	raw := definition.RawGroups{
		"pod": {
			"kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
		"container": {
			"kube-addon-manager-minikube_kube-system": definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod":          "kube-addon-manager-minikube",
						"container_id": "docker://441e4dacbcfb2f012f2221d0f3768552ea1ccb53454da42b7b3eeaf17bbd240a",
						"namespace":    "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"inherited-pod_ip": "pod_ip"})("container", incorrectContainerRawEntityID, raw)
	assert.EqualError(t, err, "cannot retrieve the entity ID of metrics to inherit value from, got error: metrics not found for container with entity ID: non-existing-ID")
	assert.Empty(t, fetchedValue)
}

// --------------- InheritAllPrometheusLabelsFrom ---------------
func TestInheritAllPrometheusLabelsFrom(t *testing.T) {
	containerRawEntityID := "kube-system_kube-addon-manager-minikube_kube-addon-manager"
	raw := definition.RawGroups{
		"pod": {
			"kube-system_kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
		"container": {
			containerRawEntityID: definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod":          "kube-addon-manager-minikube",
						"container_id": "docker://441e4dacbcfb2f012f2221d0f3768552ea1ccb53454da42b7b3eeaf17bbd240a",
						"namespace":    "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritAllPrometheusLabelsFrom("pod", "kube_pod_info")("container", containerRawEntityID, raw)
	assert.NoError(t, err)

	expectedValue := definition.FetchedValues{"label.pod_ip": "172.31.248.38", "label.pod": "kube-addon-manager-minikube", "label.namespace": "kube-system"}
	assert.Equal(t, expectedValue, fetchedValue)
}

func TestInheritAllPrometheusLabelsFrom_Namespace(t *testing.T) {
	podRawEntityID := "kube-addon-manager-minikube"
	raw := definition.RawGroups{
		"namespace": {
			"kube-system": definition.RawMetrics{
				"kube_namespace_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"namespace": "kube-system",
					},
				},
			},
		},
		"pod": {
			"kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels")("pod", podRawEntityID, raw)
	assert.NoError(t, err)

	expectedValue := definition.FetchedValues{"label.namespace": "kube-system"}
	assert.Equal(t, expectedValue, fetchedValue)
}

func TestInheritAllPrometheusLabelsFrom_FromTheSameLabelGroup(t *testing.T) {
	deploymentRawEntityID := "kube-public_newrelic-infra-monitoring"
	raw := definition.RawGroups{
		"deployment": {
			deploymentRawEntityID: definition.RawMetrics{
				"kube_deployment_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"deployment": "newrelic-infra-monitoring",
						"label_app":  "newrelic-infra-monitoring",
						"namespace":  "kube-public",
					},
				},
				"kube_deployment_spec_replicas": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"deployment": "newrelic-infra-monitoring",
						"namespace":  "kube-public",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels")("deployment", deploymentRawEntityID, raw)
	assert.NoError(t, err)

	expectedValue := definition.FetchedValues{"label.deployment": "newrelic-infra-monitoring", "label.namespace": "kube-public", "label.app": "newrelic-infra-monitoring"}
	assert.Equal(t, expectedValue, fetchedValue)
}
func TestInheritAllPrometheusLabelsFrom_LabelNotFound(t *testing.T) {
	podRawEntityID := "kube-system_kube-addon-manager-minikube"
	raw := definition.RawGroups{
		"deployment": {
			"newrelic-infra-monitoring": definition.RawMetrics{
				"kube_deployment_labels": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"deployment": "newrelic-infra-monitoring",
						"label_app":  "newrelic-infra-monitoring",
						"namespace":  "kube-public",
					},
				},
			},
		},
		"pod": {
			"kube-system_kube-addon-manager-minikube": definition.RawMetrics{
				"kube_pod_info": Metric{
					Value: GaugeValue(1507117436),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"pod_ip":    "172.31.248.38",
						"namespace": "kube-system",
					},
				},
			},
		},
	}

	fetchedValue, err := InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels")("pod", podRawEntityID, raw)
	assert.Nil(t, fetchedValue)
	assert.EqualError(t, err, fmt.Sprintf("cannot retrieve the entity ID of metrics to inherit labels from, got error: label not found. Label: deployment, Metric: kube_pod_info"))
}

func TestInheritAllPrometheusLabelsFrom_RelatedMetricNotFound(t *testing.T) {
	containerRawEntityID := "kube-system_kube-addon-manager-minikube_kube-addon-manager"
	raw := definition.RawGroups{
		"pod": {},
		"container": {
			containerRawEntityID: definition.RawMetrics{
				"kube_pod_container_info": Metric{
					Value: GaugeValue(1),
					Labels: map[string]string{
						"pod":       "kube-addon-manager-minikube",
						"namespace": "kube-system",
					},
				},
			},
		},
	}

	expectedPodRawEntityID := "kube-system_kube-addon-manager-minikube"
	fetchedValue, err := InheritAllPrometheusLabelsFrom("pod", "non_existent_metric_key")("container", containerRawEntityID, raw)
	assert.EqualError(t, err, fmt.Sprintf("related metric not found. Metric: non_existent_metric_key pod:%v", expectedPodRawEntityID))
	assert.Empty(t, fetchedValue)
}
