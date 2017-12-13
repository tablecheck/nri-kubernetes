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
		{"podCreated", metric.FromPrometheusValue("kube_pod_created"), sdkMetric.GAUGE},
		{"podStartTime", metric.FromPrometheusValue("kube_pod_start_time"), sdkMetric.GAUGE},
		{"podInfo.createdByKind", metric.FromPrometheusLabelValue("kube_pod_info", "created_by_kind"), sdkMetric.ATTRIBUTE},
		{"podInfo.createdByName", metric.FromPrometheusLabelValue("kube_pod_info", "created_by_name"), sdkMetric.ATTRIBUTE},
		{"podInfo.hostIp", metric.FromPrometheusLabelValue("kube_pod_info", "host_ip"), sdkMetric.ATTRIBUTE},
		{"podInfo.podIp", metric.FromPrometheusLabelValue("kube_pod_info", "pod_ip"), sdkMetric.ATTRIBUTE},
		{"podInfo.namespace", metric.FromPrometheusLabelValue("kube_pod_info", "namespace"), sdkMetric.ATTRIBUTE},
		{"podInfo.node", metric.FromPrometheusLabelValue("kube_pod_info", "node"), sdkMetric.ATTRIBUTE},
		{"podInfo.pod", metric.FromPrometheusLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
		{"podStatusReady", metric.FromPrometheusLabelValue("kube_pod_status_ready", "condition"), sdkMetric.ATTRIBUTE},
		{"podStatusPhase", metric.FromPrometheusLabelValue("kube_pod_status_phase", "phase"), sdkMetric.ATTRIBUTE},
		{"podStatusScheduled", metric.FromPrometheusLabelValue("kube_pod_status_scheduled", "condition"), sdkMetric.ATTRIBUTE},
		{"podName", metric.FromPrometheusLabelValue("kube_pod_info", "pod"), sdkMetric.ATTRIBUTE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
	},

	"replicaset": {
		{"replicasetSpecReplicas", metric.FromPrometheusValue("kube_replicaset_spec_replicas"), sdkMetric.GAUGE},
		{"replicasetStatusReadyReplicas", metric.FromPrometheusValue("kube_replicaset_status_ready_replicas"), sdkMetric.GAUGE},
		{"replicasetStatusReplicas", metric.FromPrometheusValue("kube_replicaset_status_replicas"), sdkMetric.GAUGE},
		{"replicasetCreated", metric.FromPrometheusValue("kube_replicaset_created"), sdkMetric.GAUGE},
		{"replicasetName", metric.FromPrometheusLabelValue("kube_replicaset_created", "replicaset"), sdkMetric.ATTRIBUTE},
	},

	"container": {
		{"podContainerInfo.container", metric.FromPrometheusLabelValue("kube_pod_container_info", "container"), sdkMetric.ATTRIBUTE},
		{"podContainerInfo.container_id", metric.FromPrometheusLabelValue("kube_pod_container_info", "container_id"), sdkMetric.ATTRIBUTE},
		{"podContainerInfo.image", metric.FromPrometheusLabelValue("kube_pod_container_info", "image"), sdkMetric.ATTRIBUTE},
		{"podContainerInfo.image_id", metric.FromPrometheusLabelValue("kube_pod_container_info", "image_id"), sdkMetric.ATTRIBUTE},
		{"podContainerInfo.namespace", metric.FromPrometheusLabelValue("kube_pod_container_info", "namespace"), sdkMetric.ATTRIBUTE},
		{"podContainerInfo.pod", metric.FromPrometheusLabelValue("kube_pod_container_info", "pod"), sdkMetric.ATTRIBUTE},
		{"podContainerResourceCPURequest", metric.FromPrometheusValue("kube_pod_container_resource_requests_cpu_cores"), sdkMetric.GAUGE},
		{"podContainerResourceCPULimit", metric.FromPrometheusValue("kube_pod_container_resource_limits_cpu_cores"), sdkMetric.GAUGE},
		{"podContainerResourceMemoryRequestBytes", metric.FromPrometheusValue("kube_pod_container_resource_requests_memory_bytes"), sdkMetric.GAUGE},
		{"podContainerResourceMemoryLimitBytes", metric.FromPrometheusValue("kube_pod_container_resource_limits_memory_bytes"), sdkMetric.GAUGE},

		// Note that kube_pod_container_status_restarts will become kube_pod_container_status_restarts_total in a next version of KSM. Now we are compatible with 1.1.x.
		{"podContainerStatusRestartsPerSecond", metric.FromPrometheusValue("kube_pod_container_status_restarts"), sdkMetric.RATE},
		{"podContainerStatusRunning", metric.FromPrometheusValue("kube_pod_container_status_running"), sdkMetric.GAUGE},
		{"podContainerStatusTerminated", metric.FromPrometheusValue("kube_pod_container_status_terminated"), sdkMetric.GAUGE},
		{"podContainerStatusReady", metric.FromPrometheusValue("kube_pod_container_status_ready"), sdkMetric.GAUGE},
		{"podContainerName", metric.FromPrometheusLabelValue("kube_pod_container_info", "container"), sdkMetric.ATTRIBUTE},

		// Example of how to inherit labels from other metrics
		//{"related.label.*", metric.InheritSpecificPrometheusLabelValuesFrom("pod", "kube_pod_info", map[string]string{"podIP": "pod_ip"}), sdkMetric.ATTRIBUTE},
	},

	"namespace": {
		{"namespaceCreated", metric.FromPrometheusValue("kube_namespace_created"), sdkMetric.GAUGE},
		{"namespaceName", metric.FromPrometheusLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
		{"namespaceStatusPhase", metric.FromPrometheusLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
	},

	"deployment": {
		{"deploymentName", metric.FromPrometheusLabelValue("kube_deployment_labels", "deployment"), sdkMetric.ATTRIBUTE},
		{"deploymentCreated", metric.FromPrometheusValue("kube_deployment_created"), sdkMetric.GAUGE},
		{"deploymentSpecReplicas", metric.FromPrometheusValue("kube_deployment_spec_replicas"), sdkMetric.GAUGE},
		{"deploymentStatusReplicas", metric.FromPrometheusValue("kube_deployment_status_replicas"), sdkMetric.GAUGE},
		{"deploymentStatusReplicasAvailable", metric.FromPrometheusValue("kube_deployment_status_replicas_available"), sdkMetric.GAUGE},
		{"deploymentStatusReplicasUnavailable", metric.FromPrometheusValue("kube_deployment_status_replicas_unavailable"), sdkMetric.GAUGE},
		{"deploymentStatusReplicasUpdated", metric.FromPrometheusValue("kube_deployment_status_replicas_updated"), sdkMetric.GAUGE},
		{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
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
