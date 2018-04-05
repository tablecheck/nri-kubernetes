package metric

import (
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/prometheus"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
)

// KSMSpecs are the metric specifications we want to collect from KSM.
var KSMSpecs = definition.SpecGroups{
	"replicaset": {
		IDGenerator:   prometheus.FromLabelValueEntityIDGenerator("kube_replicaset_created", "replicaset"),
		TypeGenerator: prometheus.FromLabelValueEntityTypeGenerator("kube_replicaset_created"),
		Specs: []definition.Spec{
			{"createdAt", prometheus.FromValue("kube_replicaset_created"), sdkMetric.GAUGE},
			{"podsDesired", prometheus.FromValue("kube_replicaset_spec_replicas"), sdkMetric.GAUGE},
			{"podsReady", prometheus.FromValue("kube_replicaset_status_ready_replicas"), sdkMetric.GAUGE},
			{"podsTotal", prometheus.FromValue("kube_replicaset_status_replicas"), sdkMetric.GAUGE},
			{"podsFullyLabeled", prometheus.FromValue("kube_replicaset_status_fully_labeled_replicas"), sdkMetric.GAUGE},
			{"observedGeneration", prometheus.FromValue("kube_replicaset_status_observed_generation"), sdkMetric.GAUGE},
			{"replicasetName", prometheus.FromLabelValue("kube_replicaset_created", "replicaset"), sdkMetric.ATTRIBUTE},
			{"namespace", prometheus.FromLabelValue("kube_replicaset_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", ksmMetric.GetDeploymentNameForReplicaSet(), sdkMetric.ATTRIBUTE},
		},
	},
	"namespace": {
		TypeGenerator: prometheus.FromLabelValueEntityTypeGenerator("kube_namespace_created"),
		Specs: []definition.Spec{
			{"createdAt", prometheus.FromValue("kube_namespace_created"), sdkMetric.GAUGE},
			{"namespace", prometheus.FromLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"status", prometheus.FromLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"label.*", prometheus.InheritAllLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"deployment": {
		IDGenerator:   prometheus.FromLabelValueEntityIDGenerator("kube_deployment_created", "deployment"),
		TypeGenerator: prometheus.FromLabelValueEntityTypeGenerator("kube_deployment_created"),
		Specs: []definition.Spec{
			{"podsDesired", prometheus.FromValue("kube_deployment_spec_replicas"), sdkMetric.GAUGE},
			{"createdAt", prometheus.FromValue("kube_deployment_created"), sdkMetric.GAUGE},
			{"podsTotal", prometheus.FromValue("kube_deployment_status_replicas"), sdkMetric.GAUGE},
			{"podsAvailable", prometheus.FromValue("kube_deployment_status_replicas_available"), sdkMetric.GAUGE},
			{"podsUnavailable", prometheus.FromValue("kube_deployment_status_replicas_unavailable"), sdkMetric.GAUGE},
			{"podsUpdated", prometheus.FromValue("kube_deployment_status_replicas_updated"), sdkMetric.GAUGE},
			{"podsMaxUnavailable", prometheus.FromValue("kube_deployment_spec_strategy_rollingupdate_max_unavailable"), sdkMetric.GAUGE},
			{"namespace", prometheus.FromLabelValue("kube_deployment_labels", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", prometheus.FromLabelValue("kube_deployment_labels", "deployment"), sdkMetric.ATTRIBUTE},
			// Important: The order of these lines is important: we could have the same label in different entities, and we would like to keep the value closer to deployment
			{"label.*", prometheus.InheritAllLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", prometheus.InheritAllLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	// We get Pod metrics from kube-state-metrics for those pods that aren't
	// running (ex: Pending pods). We can't get the data from kubelet because
	// they aren't running in any node and the information about them is only
	// present in the API.
	"pod": {
		IDGenerator:   prometheus.FromLabelValueEntityIDGenerator("kube_pod_status_phase", "pod"),
		TypeGenerator: prometheus.FromLabelValueEntityTypeGenerator("kube_pod_status_phase"),
		Specs: []definition.Spec{
			{"createdAt", prometheus.FromValue("kube_pod_created"), sdkMetric.GAUGE},
			{"startTime", prometheus.FromValue("kube_pod_start_time"), sdkMetric.GAUGE},
			{"createdKind", prometheus.FromLabelValue("kube_pod_info", "created_by_kind"), sdkMetric.ATTRIBUTE},
			{"createdBy", prometheus.FromLabelValue("kube_pod_info", "created_by_name"), sdkMetric.ATTRIBUTE},
			{"nodeIP", prometheus.FromLabelValue("kube_pod_info", "host_ip"), sdkMetric.ATTRIBUTE},
			{"namespace", prometheus.FromLabelValue("kube_pod_info", "namespace"), sdkMetric.ATTRIBUTE},
			{"nodeName", prometheus.FromLabelValue("kube_pod_info", "node"), sdkMetric.ATTRIBUTE},
			{"podName", prometheus.FromLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
			{"isReady", prometheus.FromLabelValue("kube_pod_status_ready", "condition"), sdkMetric.ATTRIBUTE},
			{"status", prometheus.FromLabelValue("kube_pod_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"isScheduled", prometheus.FromLabelValue("kube_pod_status_scheduled", "condition"), sdkMetric.ATTRIBUTE},
			{"deploymentName", ksmMetric.GetDeploymentNameForPod(), sdkMetric.ATTRIBUTE},
			{"label.*", prometheus.InheritAllLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
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
		Value: prometheus.QueryValue{
			Value: prometheus.GaugeValue(1),
		},
	},
	{
		MetricName: "kube_namespace_created",
	},
	{
		MetricName: "kube_namespace_status_phase",
		Value: prometheus.QueryValue{
			Value: prometheus.GaugeValue(1),
		},
	},
	{
		MetricName: "kube_namespace_created",
	},
	{
		MetricName: "kube_deployment_labels",
		Value: prometheus.QueryValue{
			Value: prometheus.GaugeValue(1),
		},
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
	{
		MetricName: "kube_pod_status_phase",
		Labels: prometheus.QueryLabels{
			Labels: prometheus.Labels{"phase": "Pending"},
		},
		Value: prometheus.QueryValue{
			Value: prometheus.GaugeValue(1),
		},
	},
	{
		MetricName: "kube_pod_info",
	},
	{
		MetricName: "kube_pod_created",
	},
	{
		MetricName: "kube_pod_labels",
	},
	{
		MetricName: "kube_pod_status_scheduled",
		Value: prometheus.QueryValue{
			Value: prometheus.GaugeValue(1),
		},
	},
}

// CadvisorQueries are the queries we will do to the kubelet metrics cadvisor endpoint in order to fetch all the raw metrics.
var CadvisorQueries = []prometheus.Query{
	{
		MetricName: "container_memory_usage_bytes",
		Labels: prometheus.QueryLabels{
			Operator: prometheus.QueryOpNor,
			Labels: prometheus.Labels{
				"container_name": "",
			},
		},
	},
}

// KubeletSpecs are the metric specifications we want to collect from Kubelet.
var KubeletSpecs = definition.SpecGroups{
	"pod": {
		IDGenerator:   kubeletMetric.FromRawEntityIDGroupEntityIDGenerator("namespace"),
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator,
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
		IDGenerator:   kubeletMetric.FromRawGroupsEntityIDGenerator("containerName"),
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator,
		Specs: []definition.Spec{
			// /stats/summary endpoint
			{"memoryUsedBytes", definition.FromRaw("usageBytes"), sdkMetric.GAUGE},
			{"cpuUsedCores", definition.Transform(definition.FromRaw("usageNanoCores"), fromNano), sdkMetric.GAUGE},

			// /metrics/cadvisor endpoint
			{"containerID", definition.FromRaw("containerID"), sdkMetric.ATTRIBUTE},
			{"containerImageID", definition.FromRaw("containerImageID"), sdkMetric.ATTRIBUTE},

			// /pods endpoint
			{"containerName", definition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
			{"containerImage", definition.FromRaw("containerImage"), sdkMetric.ATTRIBUTE},
			{"deploymentName", definition.FromRaw("deploymentName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
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
		TypeGenerator: kubeletMetric.FromRawGroupsEntityTypeGenerator,
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
