package main

import (
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletDefinition "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/definition"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
)

var ksmAggregation = definition.Specs{
	"pod": {
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
		{"statusScheduled", metric.FromPrometheusLabelValue("kube_pod_status_scheduled", "condition"), sdkMetric.ATTRIBUTE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
		{"deploymentName", metric.GetDeploymentNameForPod(), sdkMetric.ATTRIBUTE},
	},

	"replicaset": {
		{"createdAt", metric.FromPrometheusValue("kube_replicaset_created"), sdkMetric.GAUGE},
		{"podsDesired", metric.FromPrometheusValue("kube_replicaset_spec_replicas"), sdkMetric.GAUGE},
		{"podsReady", metric.FromPrometheusValue("kube_replicaset_status_ready_replicas"), sdkMetric.GAUGE},
		{"podsAvailable", metric.FromPrometheusValue("kube_replicaset_status_replicas"), sdkMetric.GAUGE},
		{"podsFullyLabeled", metric.FromPrometheusValue("kube_replicaset_status_fully_labeled_replicas"), sdkMetric.GAUGE},
		{"replicasetName", metric.FromPrometheusLabelValue("kube_replicaset_created", "replicaset"), sdkMetric.ATTRIBUTE},
		{"namespace", metric.FromPrometheusLabelValue("kube_replicaset_created", "namespace"), sdkMetric.ATTRIBUTE},
		{"deploymentName", metric.GetDeploymentNameForReplicaSet(), sdkMetric.ATTRIBUTE},
	},

	"container": {
		{"containerName", metric.FromPrometheusLabelValue("kube_pod_container_info", "container"), sdkMetric.ATTRIBUTE},
		{"containerID", metric.FromPrometheusLabelValue("kube_pod_container_info", "container_id"), sdkMetric.ATTRIBUTE},
		{"containerImage", metric.FromPrometheusLabelValue("kube_pod_container_info", "image"), sdkMetric.ATTRIBUTE},
		{"containerImageID", metric.FromPrometheusLabelValue("kube_pod_container_info", "image_id"), sdkMetric.ATTRIBUTE},
		{"namespace", metric.FromPrometheusLabelValue("kube_pod_container_info", "namespace"), sdkMetric.ATTRIBUTE},
		{"podName", metric.FromPrometheusLabelValue("kube_pod_container_info", "pod"), sdkMetric.ATTRIBUTE},
		// Note that kube_pod_container_status_restarts will become kube_pod_container_status_restarts_total in a next version of KSM. Now we are compatible with 1.1.x.
		{"restartCount", metric.FromPrometheusValue("kube_pod_container_status_restarts"), sdkMetric.GAUGE},
		{"cpuRequestedCores", metric.FromPrometheusValue("kube_pod_container_resource_requests_cpu_cores"), sdkMetric.GAUGE},
		{"cpuLimitCores", metric.FromPrometheusValue("kube_pod_container_resource_limits_cpu_cores"), sdkMetric.GAUGE},
		{"memoryRequestedBytes", metric.FromPrometheusValue("kube_pod_container_resource_requests_memory_bytes"), sdkMetric.GAUGE},
		{"memoryLimitBytes", metric.FromPrometheusValue("kube_pod_container_resource_limits_memory_bytes"), sdkMetric.GAUGE},
		{"statusRunning", metric.FromPrometheusValue("kube_pod_container_status_running"), sdkMetric.GAUGE},
		{"statusTerminated", metric.FromPrometheusValue("kube_pod_container_status_terminated"), sdkMetric.GAUGE},
		{"statusReady", metric.FromPrometheusValue("kube_pod_container_status_ready"), sdkMetric.GAUGE},
		{"statusWaitingReason", metric.FromPrometheusLabelValue("kube_pod_container_status_waiting_reason", "reason"), sdkMetric.ATTRIBUTE},
		// Example of how to inherit labels from other metrics
		//{"related.label.*", metric.InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"podIP": "pod_ip"}), sdkMetric.ATTRIBUTE},
	},

	"namespace": {
		{"createdAt", metric.FromPrometheusValue("kube_namespace_created"), sdkMetric.GAUGE},
		{"namespace", metric.FromPrometheusLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
		{"namespaceStatus", metric.FromPrometheusLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
	},

	"deployment": {
		{"podsDesired", metric.FromPrometheusValue("kube_deployment_spec_replicas"), sdkMetric.GAUGE},
		{"createdAt", metric.FromPrometheusValue("kube_deployment_created"), sdkMetric.GAUGE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
		{"pods", metric.FromPrometheusValue("kube_deployment_status_replicas"), sdkMetric.GAUGE},
		{"podsAvailable", metric.FromPrometheusValue("kube_deployment_status_replicas_available"), sdkMetric.GAUGE},
		{"podsUnavailable", metric.FromPrometheusValue("kube_deployment_status_replicas_unavailable"), sdkMetric.GAUGE},
		{"updatedAt", metric.FromPrometheusValue("kube_deployment_status_replicas_updated"), sdkMetric.GAUGE},
		{"podsMaxUnavailable", metric.FromPrometheusValue("kube_deployment_spec_strategy_rollingupdate_max_unavailable"), sdkMetric.GAUGE},
		{"namespace", metric.FromPrometheusLabelValue("kube_deployment_labels", "namespace"), sdkMetric.ATTRIBUTE},
		{"deploymentName", metric.FromPrometheusLabelValue("kube_deployment_labels", "deployment"), sdkMetric.ATTRIBUTE},
	},
}

var prometheusQueries = []prometheus.Query{
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
		MetricName: "kube_replicaset_created",
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

var kubeletAggregation = kubeletDefinition.Aggregation{
	"pod": {
		{"cluster.pod.name", kubeletDefinition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
		{"cluster.namespace", kubeletDefinition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
		{"net.rxBytesPerSecond", kubeletDefinition.FromRaw("rxBytes"), sdkMetric.RATE},
		{"net.txBytesPerSecond", kubeletDefinition.FromRaw("txBytes"), sdkMetric.RATE},
		{"net.errorsPerSecond", kubeletDefinition.FromRaw("errors"), sdkMetric.RATE},
	},

	"container": {
		{"cluster.container.name", kubeletDefinition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
		{"cluster.container.memoryUsedBytes", kubeletDefinition.FromRaw("usageBytes"), sdkMetric.GAUGE},
		{"cluster.container.cpuCoresUsed", kubeletDefinition.FromRaw("usageNanoCores"), sdkMetric.GAUGE},
		{"cluster.pod.name", kubeletDefinition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
		{"cluster.namespace", kubeletDefinition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
	},
}
