package metric

import (
	"errors"
	"testing"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
	"github.com/stretchr/testify/assert"
)

var mFamily = []prometheus.MetricFamily{
	{
		Name: "kube_pod_start_time",
		Metrics: []prometheus.Metric{
			{
				Value: prometheus.GaugeValue(1507117436),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "fluentd-elasticsearch-jnqb7",
				},
			},
			{
				Value: prometheus.GaugeValue(1510579152),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "newrelic-infra-monitoring-cglrn",
				},
			},
		},
	},
	{
		Name: "kube_pod_info",
		Metrics: []prometheus.Metric{
			{
				Value: prometheus.GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "DaemonSet",
					"created_by_name": "fluentd-elasticsearch",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "fluentd-elasticsearch-jnqb7",
				},
			},
			{
				Value: prometheus.GaugeValue(1),
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
		Metrics: []prometheus.Metric{
			{
				Value: prometheus.GaugeValue(1),
				Labels: map[string]string{
					"label_app":                      "newrelic-infra-monitoring",
					"label_controller_revision_hash": "1758702902",
					"label_pod_template_generation":  "1",
					"namespace":                      "kube-system",
					"pod":                            "newrelic-infra-monitoring-cglrn",
				},
			},
			{
				Value: prometheus.GaugeValue(1),
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

var rawMetric = definition.RawMetrics{
	"kube_pod_start_time": prometheus.Metric{
		Value: prometheus.GaugeValue(1507117436),
		Labels: map[string]string{
			"namespace": "kube-system",
			"pod":       "fluentd-elasticsearch-jnqb7",
		},
	},
}

var rawMetricIncompatibleType = definition.RawMetrics{
	"kube_pod_start_time": "foo",
}

var metricGroup = definition.MetricGroups{
	"pod": {
		"fluentd-elasticsearch-jnqb7": definition.RawMetrics{
			"kube_pod_start_time": prometheus.Metric{
				Value: prometheus.GaugeValue(1507117436),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "fluentd-elasticsearch-jnqb7",
				},
			},
			"kube_pod_info": prometheus.Metric{
				Value: prometheus.GaugeValue(1),
				Labels: map[string]string{
					"created_by_kind": "DaemonSet",
					"created_by_name": "fluentd-elasticsearch",
					"namespace":       "kube-system",
					"node":            "minikube",
					"pod":             "fluentd-elasticsearch-jnqb7",
				},
			},
		},
		"newrelic-infra-monitoring-cglrn": definition.RawMetrics{
			"kube_pod_start_time": prometheus.Metric{
				Value: prometheus.GaugeValue(1510579152),
				Labels: map[string]string{
					"namespace": "kube-system",
					"pod":       "newrelic-infra-monitoring-cglrn",
				},
			},
			"kube_pod_info": prometheus.Metric{
				Value: prometheus.GaugeValue(1),
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

var metricDef = []definition.Metric{
	{"podStartTime", FromPrometheusValue("kube_pod_start_time"), sdkMetric.GAUGE},
	{"podInfo.namespace", FromPrometheusLabelValue("kube_pod_info", "namespace"), sdkMetric.ATTRIBUTE},
	{"podInfo.pod", FromPrometheusLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
}

// --------------- Populate ---------------

func TestPopulate_CorrectValue(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s/pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet1 := sdkMetric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
	}
	expectedEntityData1.Metrics = []sdkMetric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s/pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := sdkMetric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1510579152),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "newrelic-infra-monitoring-cglrn",
	}
	expectedEntityData2.Metrics = []sdkMetric.MetricSet{expectedMetricSet2}

	populated, errs := Populate(integration, metricDef, metricGroup)
	assert.True(t, populated)
	assert.Empty(t, errs)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestPopulate_PartialResult(t *testing.T) {
	var metricDefWithIncompatibleType = []definition.Metric{
		{"podStartTime", FromPrometheusValue("kube_pod_start_time"), sdkMetric.GAUGE},
		{"podInfo.namespace", FromPrometheusLabelValue("kube_pod_info", "namespace"), sdkMetric.GAUGE}, // Source type not correct
		{"podInfo.pod", FromPrometheusLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s/pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet1 := sdkMetric.MetricSet{
		"event_type":   "K8sPodSample",
		"podStartTime": prometheus.GaugeValue(1507117436),
		"podInfo.pod":  "fluentd-elasticsearch-jnqb7",
	}
	expectedEntityData1.Metrics = []sdkMetric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s/pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := sdkMetric.MetricSet{
		"event_type":   "K8sPodSample",
		"podStartTime": prometheus.GaugeValue(1510579152),
		"podInfo.pod":  "newrelic-infra-monitoring-cglrn",
	}
	expectedEntityData2.Metrics = []sdkMetric.MetricSet{expectedMetricSet2}

	populated, errs := Populate(integration, metricDefWithIncompatibleType, metricGroup)
	assert.True(t, populated)
	assert.Len(t, errs, 2)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestPopulate_EntitiesDataNotPopulated_EmptyMetricGroups(t *testing.T) {
	var metricGroupEmpty = definition.MetricGroups{}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedData := []*sdk.EntityData{}

	populated, errs := Populate(integration, metricDef, metricGroupEmpty)
	assert.False(t, populated)
	assert.Nil(t, errs)
	assert.Equal(t, expectedData, integration.Data)
}

func TestPopulate_EntitiesDataNotPopulated_ErrorSettingEntities(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	var metricGroupEmptyEntityID = definition.MetricGroups{
		"pod": {
			"": definition.RawMetrics{
				"kube_pod_start_time": prometheus.Metric{
					Value: prometheus.GaugeValue(1507117436),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "fluentd-elasticsearch-jnqb7",
					},
				},
			},
		},
	}
	expectedData := []*sdk.EntityData{}

	populated, errs := Populate(integration, metricDef, metricGroupEmptyEntityID)
	assert.False(t, populated)
	assert.EqualError(t, errs[0], "entity name and type are required when defining one")
	assert.Equal(t, expectedData, integration.Data)
}
func TestPopulate_MetricsSetsNotPopulated_OnlyEntity(t *testing.T) {
	var metricDefIncorrect = []definition.Metric{
		{"podStartTime", FromPrometheusValue("foo"), sdkMetric.GAUGE},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s/pod")
	if err != nil {
		t.Fatal()
	}
	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s/pod")
	if err != nil {
		t.Fatal()
	}

	populated, errs := Populate(integration, metricDefIncorrect, metricGroup)
	assert.False(t, populated)
	assert.Len(t, errs, 2)
	assert.Contains(t, errs, errors.New("entity id: fluentd-elasticsearch-jnqb7: error fetching value for metric podStartTime. Error: raw metric not found with key foo"))
	assert.Contains(t, errs, errors.New("entity id: newrelic-infra-monitoring-cglrn: error fetching value for metric podStartTime. Error: raw metric not found with key foo"))
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)

}

// --------------- GroupPrometheusMetricsByLabel ---------------
func TestGroupPrometheusMetricsByLabel_CorrectValue(t *testing.T) {
	expectedMetricGroup := definition.MetricGroups{
		"pod": {
			"fluentd-elasticsearch-jnqb7": definition.RawMetrics{
				"kube_pod_start_time": prometheus.Metric{
					Value: prometheus.GaugeValue(1507117436),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "fluentd-elasticsearch-jnqb7",
					},
				},
				"kube_pod_info": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
					Labels: map[string]string{
						"created_by_kind": "DaemonSet",
						"created_by_name": "fluentd-elasticsearch",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "fluentd-elasticsearch-jnqb7",
					},
				},
				"kube_pod_labels": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
					Labels: map[string]string{
						"label_name":                     "fluentd-elasticsearch",
						"label_controller_revision_hash": "3534845553",
						"label_pod_template_generation":  "1",
						"namespace":                      "kube-system",
						"pod":                            "fluentd-elasticsearch-jnqb7",
					},
				},
			},
			"newrelic-infra-monitoring-cglrn": definition.RawMetrics{
				"kube_pod_start_time": prometheus.Metric{
					Value: prometheus.GaugeValue(1510579152),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_info": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
					Labels: map[string]string{
						"created_by_kind": "DaemonSet",
						"created_by_name": "newrelic-infra-monitoring",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_labels": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
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

	metricGroup := GroupPrometheusMetricsByLabel("pod", mFamily)
	assert.Equal(t, expectedMetricGroup, metricGroup)
}

// This is wrong behavior, because we create more MetricsGroup than expected, which will not be used.
// The metrics from mFamily that we use belong to pods category, not to namespace.
// Once we fix this behavior, the expectedMetricGroup should be empty.
func TestGroupPrometheusMetricsByLabel_MetricFamilyForPod_LabelForNamespace(t *testing.T) {
	expectedMetricGroup := definition.MetricGroups{
		"namespace": {
			"kube-system": definition.RawMetrics{
				"kube_pod_start_time": prometheus.Metric{
					Value: prometheus.GaugeValue(1510579152),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_info": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
					Labels: map[string]string{
						"created_by_kind": "DaemonSet",
						"created_by_name": "newrelic-infra-monitoring",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "newrelic-infra-monitoring-cglrn",
					},
				},
				"kube_pod_labels": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
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
	metricGroup := GroupPrometheusMetricsByLabel("namespace", mFamily)
	assert.Equal(t, expectedMetricGroup, metricGroup)
}

func TestGroupPrometheusMetricsByLabel_EmptyMetricFamily(t *testing.T) {
	emptyMetricFamily := []prometheus.MetricFamily{}

	metricGroup := GroupPrometheusMetricsByLabel("pod", emptyMetricFamily)
	assert.Empty(t, metricGroup)
}

func TestGroupPrometheusMetricsByLabel_LabelNotFound(t *testing.T) {

	metricGroup := GroupPrometheusMetricsByLabel("replicaset", mFamily)
	assert.Empty(t, metricGroup)
}

// --------------- FromPrometheusValue ---------------
func TestFromRawPrometheusValue_CorrectValue(t *testing.T) {
	expectedFetchedValue := prometheus.GaugeValue(1507117436)

	fetchedValue, err := FromPrometheusValue("kube_pod_start_time")(rawMetric)
	assert.Equal(t, expectedFetchedValue, fetchedValue)
	assert.Nil(t, err)
}

func TestFromRawPrometheusValue_RawMetricNotFound(t *testing.T) {

	fetchedValue, err := FromPrometheusValue("foo")(rawMetric)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "raw metric not found with key foo")
}

func TestFromRawPrometheusValue_IncompatibleType(t *testing.T) {

	fetchedValue, err := FromPrometheusValue("kube_pod_start_time")(rawMetricIncompatibleType)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "incompatible metric type. Expected: prometheus.Metric. Got: string")
}

// --------------- FromPrometheusLabelValue ---------------
func TestFromRawPrometheusLabelValue_CorrectValue(t *testing.T) {
	expectedFetchedValue := "kube-system"

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "namespace")(rawMetric)
	assert.Equal(t, expectedFetchedValue, fetchedValue)
	assert.Nil(t, err)
}

func TestFromRawPrometheusLabelValue_RawMetricNotFound(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("foo", "namespace")(rawMetric)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "raw metric not found with key foo")
}

func TestFromRawPrometheusLabelValue_IncompatibleType(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "namespace")(rawMetricIncompatibleType)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "incompatible metric type. Expected: prometheus.Metric. Got: string")
}

func TestFromRawPrometheusLabelValue_LabelNotFoundInRawMetric(t *testing.T) {

	fetchedValue, err := FromPrometheusLabelValue("kube_pod_start_time", "foo")(rawMetric)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "label 'foo' not found in raw metrics")
}

// --------------- FromPrometheusMultipleLabels ---------------
func TestFromPrometheusMultipleLabels_CorrectValues(t *testing.T) {
	expectedFetchedValues := definition.FetchedValues{
		"label.namespace": "kube-system",
		"label.pod":       "fluentd-elasticsearch-jnqb7",
	}

	fetchedValues, err := FromPrometheusMultipleLabels("kube_pod_start_time")(rawMetric)
	assert.Equal(t, expectedFetchedValues, fetchedValues)
	assert.Nil(t, err)
}

func TestFromPrometheusMultipleLabels_RawMetricNotFound(t *testing.T) {

	fetchedValue, err := FromPrometheusMultipleLabels("foo")(rawMetric)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "raw metric not found with key foo")
}

func TestFromPrometheusMultipleLabels_IncompatibleType(t *testing.T) {

	fetchedValue, err := FromPrometheusMultipleLabels("kube_pod_start_time")(rawMetricIncompatibleType)
	assert.Nil(t, fetchedValue)
	assert.Error(t, err, "incompatible metric type. Expected: prometheus.Metric. Got: string")
}
