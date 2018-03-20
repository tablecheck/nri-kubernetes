package metric

import (
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
)

// KSMSpecs are the metric specifications we want to collect from KSM.
var KSMSpecs = definition.SpecGroups{
	"replicaset": {
		IDGenerator:   metric.FromPrometheusLabelValueEntityIDGenerator("kube_replicaset_created", "replicaset"),
		TypeGenerator: metric.FromPrometheusLabelValueEntityTypeGenerator("kube_replicaset_created", "namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_replicaset_created"), sdkMetric.GAUGE},
			{"podsDesired", metric.FromPrometheusValue("kube_replicaset_spec_replicas"), sdkMetric.GAUGE},
			{"podsReady", metric.FromPrometheusValue("kube_replicaset_status_ready_replicas"), sdkMetric.GAUGE},
			{"podsTotal", metric.FromPrometheusValue("kube_replicaset_status_replicas"), sdkMetric.GAUGE},
			{"podsFullyLabeled", metric.FromPrometheusValue("kube_replicaset_status_fully_labeled_replicas"), sdkMetric.GAUGE},
			{"observedGeneration", metric.FromPrometheusValue("kube_replicaset_status_observed_generation"), sdkMetric.GAUGE},
			{"replicasetName", metric.FromPrometheusLabelValue("kube_replicaset_created", "replicaset"), sdkMetric.ATTRIBUTE},
			{"namespace", metric.FromPrometheusLabelValue("kube_replicaset_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.GetDeploymentNameForReplicaSet(), sdkMetric.ATTRIBUTE},
		},
	},
	"namespace": {
		TypeGenerator: metric.FromPrometheusLabelValueEntityTypeGenerator("kube_namespace_created", "namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_namespace_created"), sdkMetric.GAUGE},
			{"namespace", metric.FromPrometheusLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"status", metric.FromPrometheusLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"deployment": {
		IDGenerator:   metric.FromPrometheusLabelValueEntityIDGenerator("kube_deployment_created", "deployment"),
		TypeGenerator: metric.FromPrometheusLabelValueEntityTypeGenerator("kube_deployment_created", "namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			{"podsDesired", metric.FromPrometheusValue("kube_deployment_spec_replicas"), sdkMetric.GAUGE},
			{"createdAt", metric.FromPrometheusValue("kube_deployment_created"), sdkMetric.GAUGE},
			{"podsTotal", metric.FromPrometheusValue("kube_deployment_status_replicas"), sdkMetric.GAUGE},
			{"podsAvailable", metric.FromPrometheusValue("kube_deployment_status_replicas_available"), sdkMetric.GAUGE},
			{"podsUnavailable", metric.FromPrometheusValue("kube_deployment_status_replicas_unavailable"), sdkMetric.GAUGE},
			{"podsUpdated", metric.FromPrometheusValue("kube_deployment_status_replicas_updated"), sdkMetric.GAUGE},
			{"podsMaxUnavailable", metric.FromPrometheusValue("kube_deployment_spec_strategy_rollingupdate_max_unavailable"), sdkMetric.GAUGE},
			{"namespace", metric.FromPrometheusLabelValue("kube_deployment_labels", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.FromPrometheusLabelValue("kube_deployment_labels", "deployment"), sdkMetric.ATTRIBUTE},
			// Important: The order of these lines is important: we could have the same label in different entities, and we would like to keep the value closer to deployment
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
		},
	},
}

// KSMQueries are the queries we will do to KSM in order to fetch all the raw metrics.
var KSMQueries = []prometheus.Query{
	{
		MetricName: "kube_replicaset_spec_replicas",
	},
	{
		MetricName: "kube_replicaset_status_ready_replicas",
	},
	{
		MetricName: "kube_replicaset_status_replicas",
	},
	{
		MetricName: "kube_replicaset_status_fully_labeled_replicas",
	},
	{
		MetricName: "kube_replicaset_status_observed_generation",
	},
	{
		MetricName: "kube_replicaset_created",
	},
	{
		MetricName: "kube_namespace_labels",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_namespace_created",
	},
	{
		MetricName: "kube_namespace_status_phase",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_namespace_created",
	},
	{
		MetricName: "kube_deployment_labels",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_deployment_created",
	},
	{
		MetricName: "kube_deployment_spec_replicas",
	},
	{
		MetricName: "kube_deployment_status_replicas",
	},
	{
		MetricName: "kube_deployment_status_replicas_available",
	},
	{
		MetricName: "kube_deployment_status_replicas_unavailable",
	},
	{
		MetricName: "kube_deployment_status_replicas_updated",
	},
}

// KubeletSpecs are the metric specifications we want to collect from Kubelet.
var KubeletSpecs = definition.SpecGroups{
	"pod": {
		IDGenerator:   kubeletMetric.FromRawEntityIDGroupEntityIDGenerator("namespace"),
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator("namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			// /stats/summary endpoint
			{"net.rxBytesPerSecond", definition.FromRaw("rxBytes"), sdkMetric.RATE},
			{"net.txBytesPerSecond", definition.FromRaw("txBytes"), sdkMetric.RATE},
			{"net.errorCount", definition.FromRaw("errors"), sdkMetric.GAUGE},

			// /pods endpoint
			{"createdAt", definition.Transform(definition.FromRaw("createdAt"), toTimestamp), sdkMetric.GAUGE},
			{"startTime", definition.Transform(definition.FromRaw("startTime"), toTimestamp), sdkMetric.GAUGE},
			{"createdKind", definition.FromRaw("createdKind"), sdkMetric.ATTRIBUTE},
			{"createdBy", definition.FromRaw("createdBy"), sdkMetric.ATTRIBUTE},
			{"nodeIP", definition.FromRaw("nodeIP"), sdkMetric.ATTRIBUTE},
			{"podIP", definition.FromRaw("podIP"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
			{"nodeName", definition.FromRaw("nodeName"), sdkMetric.ATTRIBUTE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"isReady", definition.Transform(definition.FromRaw("isReady"), toStringBoolean), sdkMetric.ATTRIBUTE},
			{"status", definition.FromRaw("status"), sdkMetric.ATTRIBUTE},
			{"isScheduled", definition.Transform(definition.FromRaw("isScheduled"), toStringBoolean), sdkMetric.ATTRIBUTE},
			{"deploymentName", definition.FromRaw("deploymentName"), sdkMetric.ATTRIBUTE},
			{"label.*", definition.Transform(definition.FromRaw("labels"), kubeletMetric.OneMetricPerLabel), sdkMetric.ATTRIBUTE},
		},
	},
	"container": {
		IDGenerator:   kubeletMetric.FromRawGroupsEntityIDGenerator("podName"),
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator("namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			// /stats/summary endpoint
			{"containerName", definition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
			{"memoryUsedBytes", definition.FromRaw("usageBytes"), sdkMetric.GAUGE},
			{"cpuUsedCores", definition.Transform(definition.FromRaw("usageNanoCores"), fromNano), sdkMetric.GAUGE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},

			// /pods endpoint
			{"containerName", definition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
			{"containerID", definition.FromRaw("containerID"), sdkMetric.ATTRIBUTE},
			{"containerImage", definition.FromRaw("containerImage"), sdkMetric.ATTRIBUTE},
			{"containerImageID", definition.FromRaw("containerImageID"), sdkMetric.ATTRIBUTE},
			{"deploymentName", definition.FromRaw("deploymentName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"podIP", definition.FromRaw("podIP"), sdkMetric.ATTRIBUTE},
			{"nodeName", definition.FromRaw("nodeName"), sdkMetric.ATTRIBUTE},
			{"nodeIP", definition.FromRaw("nodeIP"), sdkMetric.ATTRIBUTE},
			{"restartCount", definition.FromRaw("restartCount"), sdkMetric.GAUGE},
			{"cpuRequestedCores", definition.Transform(definition.FromRaw("cpuRequestedCores"), toCores), sdkMetric.GAUGE},
			{"cpuLimitCores", definition.Transform(definition.FromRaw("cpuLimitCores"), toCores), sdkMetric.GAUGE},
			{"memoryRequestedBytes", definition.FromRaw("memoryRequestedBytes"), sdkMetric.GAUGE},
			{"memoryLimitBytes", definition.FromRaw("memoryLimitBytes"), sdkMetric.GAUGE},
			{"status", definition.FromRaw("status"), sdkMetric.ATTRIBUTE},
			{"isReady", definition.Transform(definition.FromRaw("isReady"), toStringBoolean), sdkMetric.ATTRIBUTE},
			{"reason", definition.FromRaw("reason"), sdkMetric.ATTRIBUTE}, // Previously called statusWaitingReason
		},
	},
	"node": {
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator("namespace", config.UnknownNamespace),
		Specs: []definition.Spec{
			{"nodeName", definition.FromRaw("nodeName"), sdkMetric.ATTRIBUTE},
			{"cpuUsedCores", definition.Transform(definition.FromRaw("usageNanoCores"), fromNano), sdkMetric.GAUGE},
			{"usageCoreSeconds", definition.Transform(definition.FromRaw("usageCoreNanoSeconds"), fromNano), sdkMetric.GAUGE},
			{"memoryUsedBytes", definition.FromRaw("memoryUsageBytes"), sdkMetric.GAUGE},
			{"memoryAvailableBytes", definition.FromRaw("memoryAvailableBytes"), sdkMetric.GAUGE},
			{"memoryWorkingSetBytes", definition.FromRaw("memoryWorkingSetBytes"), sdkMetric.GAUGE},
			{"memoryRssBytes", definition.FromRaw("memoryRssBytes"), sdkMetric.GAUGE},
			{"memoryPageFaults", definition.FromRaw("memoryPageFaults"), sdkMetric.GAUGE},
			{"memoryMajorPageFaults", definition.FromRaw("memoryMajorPageFaults"), sdkMetric.GAUGE},
			{"net.rxBytesPerSecond", definition.FromRaw("rxBytes"), sdkMetric.RATE},
			{"net.txBytesPerSecond", definition.FromRaw("txBytes"), sdkMetric.RATE},
			{"net.errorCount", definition.FromRaw("errors"), sdkMetric.GAUGE},
			{"fsAvailableBytes", definition.FromRaw("fsAvailableBytes"), sdkMetric.GAUGE},
			{"fsCapacityBytes", definition.FromRaw("fsCapacityBytes"), sdkMetric.GAUGE},
			{"fsUsedBytes", definition.FromRaw("fsUsedBytes"), sdkMetric.GAUGE},
			{"fsInodesFree", definition.FromRaw("fsInodesFree"), sdkMetric.GAUGE},
			{"fsInodes", definition.FromRaw("fsInodes"), sdkMetric.GAUGE},
			{"fsInodesUsed", definition.FromRaw("fsInodesUsed"), sdkMetric.GAUGE},
			{"runtimeAvailableBytes", definition.FromRaw("runtimeAvailableBytes"), sdkMetric.GAUGE},
			{"runtimeCapacityBytes", definition.FromRaw("runtimeCapacityBytes"), sdkMetric.GAUGE},
			{"runtimeUsedBytes", definition.FromRaw("runtimeUsedBytes"), sdkMetric.GAUGE},
			{"runtimeInodesFree", definition.FromRaw("runtimeInodesFree"), sdkMetric.GAUGE},
			{"runtimeInodes", definition.FromRaw("runtimeInodes"), sdkMetric.GAUGE},
			{"runtimeInodesUsed", definition.FromRaw("runtimeInodesUsed"), sdkMetric.GAUGE},
		},
	},
}

// Used to transform from usageNanoCores to cpuUsedCores
func fromNano(value definition.FetchedValue) definition.FetchedValue {
	v, ok := value.(uint64)
	if !ok {
		// TODO: is it correct to return original value?
		return value
	}

	return float64(v) / 1000000000
}

func toTimestamp(value definition.FetchedValue) definition.FetchedValue {
	v, ok := value.(time.Time)
	if !ok {
		return value
	}

	return v.Unix()
}

func toStringBoolean(value definition.FetchedValue) definition.FetchedValue {
	if value == true || value == 1 {
		return "true"
	}
	return "false"
}

func toCores(value definition.FetchedValue) definition.FetchedValue {
	switch v := value.(type) {
	case int:
		return float64(v) / 1000
	case int64:
		return float64(v) / 1000
	default:
		return value
	}
}
