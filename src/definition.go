package main

import (
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
)

var ksmPodContainerNodeGroupSpecs = definition.SpecGroups{
	"pod": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_pod_info", "pod"),
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_pod_created"), sdkMetric.GAUGE},
			{"startTime", metric.FromPrometheusValue("kube_pod_start_time"), sdkMetric.GAUGE},
			{"createdKind", metric.FromPrometheusLabelValue("kube_pod_info", "created_by_kind"), sdkMetric.ATTRIBUTE},
			{"createdBy", metric.FromPrometheusLabelValue("kube_pod_info", "created_by_name"), sdkMetric.ATTRIBUTE},
			{"nodeIP", metric.FromPrometheusLabelValue("kube_pod_info", "host_ip"), sdkMetric.ATTRIBUTE},
			{"podIP", metric.FromPrometheusLabelValue("kube_pod_info", "pod_ip"), sdkMetric.ATTRIBUTE},
			{"namespace", metric.FromPrometheusLabelValue("kube_pod_info", "namespace"), sdkMetric.ATTRIBUTE},
			{"nodeName", metric.FromPrometheusLabelValue("kube_pod_info", "node"), sdkMetric.ATTRIBUTE},
			{"podName", metric.FromPrometheusLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
			{"isReady", metric.FromPrometheusLabelValue("kube_pod_status_ready", "condition"), sdkMetric.ATTRIBUTE},
			{"status", metric.FromPrometheusLabelValue("kube_pod_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"isScheduled", metric.FromPrometheusLabelValue("kube_pod_status_scheduled", "condition"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.GetDeploymentNameForPod(), sdkMetric.ATTRIBUTE},
			// Important: The order of these lines is important: we could have the same label in different entities, and we would like to keep the value closer to pod
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"container": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_pod_container_info", "pod"),
		Specs: []definition.Spec{
			{"containerName", metric.FromPrometheusLabelValue("kube_pod_container_info", "container"), sdkMetric.ATTRIBUTE},
			{"containerID", metric.FromPrometheusLabelValue("kube_pod_container_info", "container_id"), sdkMetric.ATTRIBUTE},
			{"containerImage", metric.FromPrometheusLabelValue("kube_pod_container_info", "image"), sdkMetric.ATTRIBUTE},
			{"containerImageID", metric.FromPrometheusLabelValue("kube_pod_container_info", "image_id"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.GetDeploymentNameForContainer(), sdkMetric.ATTRIBUTE},
			{"namespace", metric.FromPrometheusLabelValue("kube_pod_container_info", "namespace"), sdkMetric.ATTRIBUTE},
			{"podName", metric.FromPrometheusLabelValue("kube_pod_container_info", "pod"), sdkMetric.ATTRIBUTE},
			{"podID", metric.InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"podIP": "pod_ip"}), sdkMetric.ATTRIBUTE},
			{"nodeName", metric.InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"nodeName": "node"}), sdkMetric.ATTRIBUTE},
			{"nodeIP", metric.InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"nodeIP": "host_ip"}), sdkMetric.ATTRIBUTE},
			// Note that kube_pod_container_status_restarts will become kube_pod_container_status_restarts_total in a next version of KSM. Now we are compatible with 1.1.x.
			{"restartCount", metric.FromPrometheusValue("kube_pod_container_status_restarts"), sdkMetric.GAUGE},
			{"cpuRequestedCores", metric.FromPrometheusValue("kube_pod_container_resource_requests_cpu_cores"), sdkMetric.GAUGE},
			{"cpuLimitCores", metric.FromPrometheusValue("kube_pod_container_resource_limits_cpu_cores"), sdkMetric.GAUGE},
			{"memoryRequestedBytes", metric.FromPrometheusValue("kube_pod_container_resource_requests_memory_bytes"), sdkMetric.GAUGE},
			{"memoryLimitBytes", metric.FromPrometheusValue("kube_pod_container_resource_limits_memory_bytes"), sdkMetric.GAUGE},
			{"status", metric.GetStatusForContainer(), sdkMetric.ATTRIBUTE},
			{"isReady", metric.FromPrometheusValue("kube_pod_container_status_ready"), sdkMetric.GAUGE},
			{"statusWaitingReason", metric.FromPrometheusLabelValue("kube_pod_container_status_waiting_reason", "reason"), sdkMetric.ATTRIBUTE},
			// Important: The order of these lines is important: we could have the same label in different entities, and we would like to keep the value closer to container
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"node": {
		Specs: []definition.Spec{
			{"node", metric.FromPrometheusLabelValue("kube_node_info", "node"), sdkMetric.ATTRIBUTE},
			{"containerRuntimeVersion", metric.FromPrometheusLabelValue("kube_node_info", "container_runtime_version"), sdkMetric.ATTRIBUTE},
			{"kernelVersion", metric.FromPrometheusLabelValue("kube_node_info", "kernel_version"), sdkMetric.ATTRIBUTE},
			{"kubeletVersion", metric.FromPrometheusLabelValue("kube_node_info", "kubelet_version"), sdkMetric.ATTRIBUTE},
			{"kubeproxyVersion", metric.FromPrometheusLabelValue("kube_node_info", "kubeproxy_version"), sdkMetric.ATTRIBUTE},
			{"osImage", metric.FromPrometheusLabelValue("kube_node_info", "os_image"), sdkMetric.ATTRIBUTE},
			{"providerId", metric.FromPrometheusLabelValue("kube_node_info", "provider_id"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("node", "kube_node_labels"), sdkMetric.ATTRIBUTE},
			{"isUnschedulable", definition.Transform(metric.FromPrometheusValue("kube_node_spec_unschedulable"), toBoolean), sdkMetric.ATTRIBUTE},
			{"capacityCpuCores", metric.FromPrometheusValue("kube_node_status_capacity_cpu_cores"), sdkMetric.GAUGE},
			{"capacityNvidiaGpuCards", metric.FromPrometheusValue("kube_node_status_capacity_nvidia_gpu_cards"), sdkMetric.GAUGE},
			{"capacityMemoryBytes", metric.FromPrometheusValue("kube_node_status_capacity_memory_bytes"), sdkMetric.GAUGE},
			{"capacityPods", metric.FromPrometheusValue("kube_node_status_capacity_pods"), sdkMetric.GAUGE},
			{"allocatableCpuCores", metric.FromPrometheusValue("kube_node_status_allocatable_cpu_cores"), sdkMetric.GAUGE},
			{"allocatableNvidiaGpuCards", metric.FromPrometheusValue("kube_node_status_allocatable_nvidia_gpu_cards"), sdkMetric.GAUGE},
			{"allocatableMemoryBytes", metric.FromPrometheusValue("kube_node_status_allocatable_memory_bytes"), sdkMetric.GAUGE},
			{"allocatablePods", metric.FromPrometheusValue("kube_node_status_allocatable_pods"), sdkMetric.GAUGE},
			{"createdAt", metric.FromPrometheusValue("kube_node_created"), sdkMetric.GAUGE},
			{"isReady", metric.FromPrometheusLabelValue("isReady", "status"), sdkMetric.ATTRIBUTE},
			{"isDiskPressure", metric.FromPrometheusLabelValue("isDiskPressure", "status"), sdkMetric.ATTRIBUTE},
			{"isMemoryPressure", metric.FromPrometheusLabelValue("isMemoryPressure", "status"), sdkMetric.ATTRIBUTE},
			{"isOutOfDisk", metric.FromPrometheusLabelValue("isOutOfDisk", "status"), sdkMetric.ATTRIBUTE},
		},
	},
	"namespace": ksmRestSpecs["namespace"], // Needed for labels inheritance
}

var ksmRestSpecs = definition.SpecGroups{
	"replicaset": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_replicaset_created", "replicaset"),
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
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_namespace_created"), sdkMetric.GAUGE},
			{"namespace", metric.FromPrometheusLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"status", metric.FromPrometheusLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"deployment": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_deployment_labels", "deployment"),
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

var prometheusPodsContainerNodeQueries = []prometheus.Query{
	{
		MetricName: "kube_pod_info",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_start_time",
	},
	{
		MetricName: "kube_pod_status_phase",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_created",
	},
	{
		MetricName: "kube_pod_status_ready",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_status_scheduled",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_labels",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_container_info",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_container_resource_requests_cpu_cores",
	},
	{
		MetricName: "kube_pod_container_resource_limits_cpu_cores",
	},
	{
		MetricName: "kube_pod_container_resource_requests_memory_bytes",
	},
	{
		MetricName: "kube_pod_container_resource_limits_memory_bytes",
	},
	{
		MetricName: "kube_pod_container_status_restarts",
	},
	{
		MetricName: "kube_pod_container_status_running",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_container_status_terminated",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_container_status_ready",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_pod_container_status_waiting_reason",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_namespace_labels", // Needed for labels inheritance
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_node_info",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_node_labels",
		Value:      prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_node_spec_unschedulable",
	},
	{
		MetricName: "kube_node_status_capacity_cpu_cores",
	},
	{
		MetricName: "kube_node_status_capacity_nvidia_gpu_cards",
	},
	{
		MetricName: "kube_node_status_capacity_memory_bytes",
	},
	{
		MetricName: "kube_node_status_capacity_pods",
	},
	{
		MetricName: "kube_node_status_allocatable_cpu_cores",
	},
	{
		MetricName: "kube_node_status_allocatable_nvidia_gpu_cards",
	},
	{
		MetricName: "kube_node_status_allocatable_memory_bytes",
	},
	{
		MetricName: "kube_node_status_allocatable_pods",
	},
	{
		CustomName: "isReady",
		MetricName: "kube_node_status_condition",
		Labels: prometheus.Labels{
			"condition": "Ready",
		},
		Value: prometheus.GaugeValue(1),
	},
	{
		CustomName: "isDiskPressure",
		MetricName: "kube_node_status_condition",
		Labels: prometheus.Labels{
			"condition": "DiskPressure",
		},
		Value: prometheus.GaugeValue(1),
	},
	{
		CustomName: "isMemoryPressure",
		MetricName: "kube_node_status_condition",
		Labels: prometheus.Labels{
			"condition": "MemoryPressure",
		},
		Value: prometheus.GaugeValue(1),
	},
	{
		CustomName: "isOutOfDisk",
		MetricName: "kube_node_status_condition",
		Labels: prometheus.Labels{
			"condition": "OutOfDisk",
		},
		Value: prometheus.GaugeValue(1),
	},
	{
		MetricName: "kube_node_created",
	},
}

var prometheusRestQueries = []prometheus.Query{
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

// Used to transform from usageNanoCores to cpuUsedCores
var fromNano = func(value definition.FetchedValue) definition.FetchedValue {
	return float64(value.(int)) / 1000000000
}

var toBoolean = func(value definition.FetchedValue) definition.FetchedValue {
	if value == prometheus.GaugeValue(1) {
		return "true"
	}
	return "false"
}

var kubeletSpecs = definition.SpecGroups{
	"pod": {
		IDGenerator: kubeletMetric.FromRawEntityIDGroupEntityIDGenerator("namespace"),
		Specs: []definition.Spec{
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
			{"net.rxBytesPerSecond", definition.FromRaw("rxBytes"), sdkMetric.RATE},
			{"net.txBytesPerSecond", definition.FromRaw("txBytes"), sdkMetric.RATE},
			{"net.errorCount", definition.FromRaw("errors"), sdkMetric.GAUGE},
		},
	},
	"container": {
		IDGenerator: kubeletMetric.FromRawGroupsEntityIDGenerator("podName"),
		Specs: []definition.Spec{
			{"containerName", definition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
			{"memoryUsedBytes", definition.FromRaw("usageBytes"), sdkMetric.GAUGE},
			{"cpuUsedCores", definition.Transform(definition.FromRaw("usageNanoCores"), fromNano), sdkMetric.GAUGE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
		},
	},
	"node": {
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

var kubeletKSMPopulateSpecs = definition.SpecGroups{
	"pod": {
		IDGenerator: kubeletSpecs["pod"].IDGenerator,
		Specs:       append(kubeletSpecs["pod"].Specs, ksmPodContainerNodeGroupSpecs["pod"].Specs...),
	},
	"container": {
		IDGenerator: kubeletSpecs["container"].IDGenerator,
		Specs:       append(kubeletSpecs["container"].Specs, ksmPodContainerNodeGroupSpecs["container"].Specs...),
	},
	"node": {
		Specs: append(kubeletSpecs["node"].Specs, ksmPodContainerNodeGroupSpecs["node"].Specs...),
	},
}

var kubeletKSMAndRestPopulateSpecs = definition.SpecGroups{
	"pod":        kubeletKSMPopulateSpecs["pod"],
	"container":  kubeletKSMPopulateSpecs["container"],
	"node":       kubeletKSMPopulateSpecs["node"],
	"replicaset": ksmRestSpecs["replicaset"],
	"namespace":  ksmRestSpecs["namespace"],
	"deployment": ksmRestSpecs["deployment"],
}
