package metric

import (
	"errors"
	"testing"

	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
	"github.com/stretchr/testify/assert"
)

var defaultNS = "playground"

var rawGroups = definition.RawGroups{
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
					"created_by_kind": "ReplicaSet",
					"created_by_name": "fluentd-elasticsearch-fafnoa",
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

var rawGroupWithReplicaSet = definition.RawGroups{
	"replicaset": {
		"kube-state-metrics-4044341274": definition.RawMetrics{
			"kube_replicaset_created": prometheus.Metric{
				Value: prometheus.GaugeValue(1507117436),
				Labels: map[string]string{
					"namespace":  "kube-system",
					"replicaset": "kube-state-metrics-4044341274",
				},
			},
		},
	},
}

var spec = []definition.Spec{
	{"podStartTime", ksmMetric.FromPrometheusValue("kube_pod_start_time"), metric.GAUGE},
	{"podInfo.namespace", ksmMetric.FromPrometheusLabelValue("kube_pod_info", "namespace"), metric.ATTRIBUTE},
	{"podInfo.pod", ksmMetric.FromPrometheusLabelValue("kube_pod_info", "pod"), metric.ATTRIBUTE},
}

var specs = definition.SpecGroups{
	"pod": definition.SpecGroup{
		Specs: spec,
	},
}

func validNamespaceFetcher(expected string) NamespaceFetcher {
	return func(groupLabel, entityID string, groups definition.RawGroups) (string, error) {
		return expected, nil
	}
}

func errorNamespaceFetcher(err error) NamespaceFetcher {
	return func(groupLabel, entityID string, groups definition.RawGroups) (string, error) {
		return config.UnknownNamespace, err
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

	assert.Equal(t, config.UnknownNamespace, namespace)
	assert.Equal(t, "error fetching namespace: error1, error2, error3", err.Error())
}

func TestK8sMetricSetEntityTypeGuesser_Pod(t *testing.T) {
	var rawGroups = definition.RawGroups{
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
						"created_by_kind": "ReplicaSet",
						"created_by_name": "fluentd-elasticsearch-fafnoa",
						"namespace":       "kube-system",
						"node":            "minikube",
						"pod":             "fluentd-elasticsearch-jnqb7",
					},
				},
			},
		},
	}

	guess, err := K8sMetricSetEntityTypeGuesser("playground", "pod", "fluentd-elasticsearch-jnqb7", rawGroups)

	assert.Equal(t, "k8s:playground:kube-system:pod", guess)
	assert.Nil(t, err)
}

func TestK8sMetricSetEntityTypeGuesser_Container(t *testing.T) {
	var rawGroups = definition.RawGroups{
		"replicaset": {
			"kube-state-metrics-4044341274": definition.RawMetrics{
				"kube_replicaset_created": prometheus.Metric{
					Value: prometheus.GaugeValue(1507117436),
					Labels: map[string]string{
						"namespace":  "kube-system",
						"replicaset": "kube-state-metrics-4044341274",
					},
				},
			},
		},
	}

	guess, err := K8sMetricSetEntityTypeGuesser("playground", "replicaset", "kube-state-metrics-4044341274", rawGroups)

	assert.Equal(t, "k8s:playground:kube-system:replicaset", guess)
	assert.Nil(t, err)
}

// TODO: K8sClusterMetricsManipulator
func TestK8sClusterMetricsManipulator(t *testing.T) {
	entityData, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	metricSet := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":       "playground",
	}

	err = K8sClusterMetricsManipulator(metricSet, entityData.Entity, "modifiedClusterName")
	assert.Nil(t, err)

	expectedMetricSet := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":       "modifiedClusterName",
	}
	assert.Equal(t, expectedMetricSet, metricSet)
}

func TestK8sMetricSetTypeGuesser(t *testing.T) {
	guess, _ := K8sMetricSetTypeGuesser("", "replicaset", "", nil)
	assert.Equal(t, "K8sReplicasetSample", guess)
}

func TestK8sEntityMetricsManipulator(t *testing.T) {
	entityData, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	metricSet := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "fluentd-elasticsearch-jnqb7",
		"clusterName":       "playground",
	}

	err = K8sEntityMetricsManipulator(metricSet, entityData.Entity, "")
	assert.Nil(t, err)

	expectedMetricSet := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":       "playground",
	}
	assert.Equal(t, expectedMetricSet, metricSet)
}

// --------------- IntegrationProtocol2PopulateFunc ---------------

func TestIntegrationProtocol2PopulateFunc_CorrectValue(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet1 := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":       "playground",
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1510579152),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "newrelic-infra-monitoring-cglrn",
		"displayName":       "newrelic-infra-monitoring-cglrn",
		"entityName":        "k8s:playground:kube-system:pod:newrelic-infra-monitoring-cglrn",
		"clusterName":       "playground",
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(rawGroups, specs)
	assert.True(t, populated)
	assert.Empty(t, errs)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestIntegrationProtocol2PopulateFunc_PopulateOnlySpecifiedGroups(t *testing.T) {
	groups := rawGroups

	// We don't want to populate replicaset
	groups["replicaset"] = rawGroupWithReplicaSet["replicaset"]

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet1 := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1507117436),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "fluentd-elasticsearch-jnqb7",
		"displayName":       "fluentd-elasticsearch-jnqb7",
		"entityName":        "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":       "playground",
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := metric.MetricSet{
		"event_type":        "K8sPodSample",
		"podStartTime":      prometheus.GaugeValue(1510579152),
		"podInfo.namespace": "kube-system",
		"podInfo.pod":       "newrelic-infra-monitoring-cglrn",
		"displayName":       "newrelic-infra-monitoring-cglrn",
		"entityName":        "k8s:playground:kube-system:pod:newrelic-infra-monitoring-cglrn",
		"clusterName":       "playground",
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(groups, specs)
	assert.True(t, populated)
	assert.Empty(t, errs)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
	assert.Len(t, integration.Data, 2)
}

func TestIntegrationProtocol2PopulateFunc_PartialResult(t *testing.T) {
	var metricDefWithIncompatibleType = definition.SpecGroups{
		"pod": {
			Specs: []definition.Spec{
				{"podStartTime", ksmMetric.FromPrometheusValue("kube_pod_start_time"), metric.GAUGE},
				{"podInfo.namespace", ksmMetric.FromPrometheusLabelValue("kube_pod_info", "namespace"), metric.GAUGE}, // Source type not correct
				{"podInfo.pod", ksmMetric.FromPrometheusLabelValue("kube_pod_info", "pod"), metric.ATTRIBUTE},
			},
		},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet1 := metric.MetricSet{
		"event_type":   "K8sPodSample",
		"podStartTime": prometheus.GaugeValue(1507117436),
		"podInfo.pod":  "fluentd-elasticsearch-jnqb7",
		"displayName":  "fluentd-elasticsearch-jnqb7",
		"entityName":   "k8s:playground:kube-system:pod:fluentd-elasticsearch-jnqb7",
		"clusterName":  "playground",
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := metric.MetricSet{
		"event_type":   "K8sPodSample",
		"podStartTime": prometheus.GaugeValue(1510579152),
		"podInfo.pod":  "newrelic-infra-monitoring-cglrn",
		"displayName":  "newrelic-infra-monitoring-cglrn",
		"entityName":   "k8s:playground:kube-system:pod:newrelic-infra-monitoring-cglrn",
		"clusterName":  "playground",
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(rawGroups, metricDefWithIncompatibleType)
	assert.True(t, populated)
	assert.Len(t, errs, 2)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestIntegrationProtocol2PopulateFunc_EntitiesDataNotPopulated_EmptyMetricGroups(t *testing.T) {
	var metricGroupEmpty = definition.RawGroups{}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedData := []*sdk.EntityData{}

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(metricGroupEmpty, specs)
	assert.False(t, populated)
	assert.Nil(t, errs)
	assert.Equal(t, expectedData, integration.Data)
}

func TestIntegrationProtocol2PopulateFunc_EntitiesDataNotPopulated_ErrorSettingEntities(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	var metricGroupEmptyEntityID = definition.RawGroups{
		"pod": {
			"": definition.RawMetrics{
				"kube_pod_info": prometheus.Metric{
					Value: prometheus.GaugeValue(1),
					Labels: map[string]string{
						"namespace": "kube-system",
						"pod":       "fluentd-elasticsearch-jnqb7",
					},
				},
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

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(metricGroupEmptyEntityID, specs)
	assert.False(t, populated)
	assert.EqualError(t, errs[0], "entity name and type are required when defining one")
	assert.Equal(t, expectedData, integration.Data)
}

func TestIntegrationProtocol2PopulateFunc_MetricsSetsNotPopulated_OnlyEntity(t *testing.T) {
	var metricDefIncorrect = definition.SpecGroups{
		"pod": {
			Specs: []definition.Spec{
				{"podStartTime", ksmMetric.FromPrometheusValue("foo"), metric.GAUGE},
			},
		},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("fluentd-elasticsearch-jnqb7", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}
	expectedEntityData2, err := sdk.NewEntityData("newrelic-infra-monitoring-cglrn", "k8s:playground:kube-system:pod")
	if err != nil {
		t.Fatal()
	}

	populated, errs := definition.IntegrationProtocol2PopulateFunc(integration, defaultNS, K8sMetricSetTypeGuesser, K8sMetricSetEntityTypeGuesser, K8sEntityMetricsManipulator, K8sClusterMetricsManipulator)(rawGroups, metricDefIncorrect)
	assert.False(t, populated)
	assert.Len(t, errs, 2)
	assert.Contains(t, errs, errors.New("entity id: fluentd-elasticsearch-jnqb7: error fetching value for metric podStartTime. Error: FromRaw: metric not found. SpecGroup: pod, EntityID: fluentd-elasticsearch-jnqb7, Metric: foo"))
	assert.Contains(t, errs, errors.New("entity id: newrelic-infra-monitoring-cglrn: error fetching value for metric podStartTime. Error: FromRaw: metric not found. SpecGroup: pod, EntityID: newrelic-infra-monitoring-cglrn, Metric: foo"))
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)

}
