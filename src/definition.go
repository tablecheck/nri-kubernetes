package main

import (
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
)

var ksmAggregation = definition.SpecGroups{
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
			{"statusScheduled", metric.FromPrometheusLabelValue("kube_pod_status_scheduled", "condition"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.GetDeploymentNameForPod(), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"replicaset": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_replicaset_created", "replicaset"),
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_replicaset_created"), sdkMetric.GAUGE},
			{"podsDesired", metric.FromPrometheusValue("kube_replicaset_spec_replicas"), sdkMetric.GAUGE},
			{"podsReady", metric.FromPrometheusValue("kube_replicaset_status_ready_replicas"), sdkMetric.GAUGE},
			{"podsAvailable", metric.FromPrometheusValue("kube_replicaset_status_replicas"), sdkMetric.GAUGE},
			{"podsFullyLabeled", metric.FromPrometheusValue("kube_replicaset_status_fully_labeled_replicas"), sdkMetric.GAUGE},
			{"replicasetName", metric.FromPrometheusLabelValue("kube_replicaset_created", "replicaset"), sdkMetric.ATTRIBUTE},
			{"namespace", metric.FromPrometheusLabelValue("kube_replicaset_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.GetDeploymentNameForReplicaSet(), sdkMetric.ATTRIBUTE},
		},
	},
	"container": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_pod_container_info", "pod"),
		Specs: []definition.Spec{
			{"containerName", metric.FromPrometheusLabelValue("kube_pod_container_info", "container"), sdkMetric.ATTRIBUTE},
			{"containerID", metric.FromPrometheusLabelValue("kube_pod_container_info", "container_id"), sdkMetric.ATTRIBUTE},
			{"containerImage", metric.FromPrometheusLabelValue("kube_pod_container_info", "image"), sdkMetric.ATTRIBUTE},
			{"containerImageID", metric.FromPrometheusLabelValue("kube_pod_container_info", "image_id"), sdkMetric.ATTRIBUTE},
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
			{"statusRunning", metric.FromPrometheusValue("kube_pod_container_status_running"), sdkMetric.GAUGE},
			{"statusTerminated", metric.FromPrometheusValue("kube_pod_container_status_terminated"), sdkMetric.GAUGE},
			{"statusReady", metric.FromPrometheusValue("kube_pod_container_status_ready"), sdkMetric.GAUGE},
			{"statusWaitingReason", metric.FromPrometheusLabelValue("kube_pod_container_status_waiting_reason", "reason"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("pod", "kube_pod_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"namespace": {
		Specs: []definition.Spec{
			{"createdAt", metric.FromPrometheusValue("kube_namespace_created"), sdkMetric.GAUGE},
			{"namespace", metric.FromPrometheusLabelValue("kube_namespace_created", "namespace"), sdkMetric.ATTRIBUTE},
			{"namespaceStatus", metric.FromPrometheusLabelValue("kube_namespace_status_phase", "phase"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
		},
	},
	"deployment": {
		IDGenerator: metric.FromPrometheusLabelValueEntityIDGenerator("kube_deployment_labels", "deployment"),
		Specs: []definition.Spec{
			{"podsDesired", metric.FromPrometheusValue("kube_deployment_spec_replicas"), sdkMetric.GAUGE},
			{"createdAt", metric.FromPrometheusValue("kube_deployment_created"), sdkMetric.GAUGE},
			{"pods", metric.FromPrometheusValue("kube_deployment_status_replicas"), sdkMetric.GAUGE},
			{"podsAvailable", metric.FromPrometheusValue("kube_deployment_status_replicas_available"), sdkMetric.GAUGE},
			{"podsUnavailable", metric.FromPrometheusValue("kube_deployment_status_replicas_unavailable"), sdkMetric.GAUGE},
			{"updatedAt", metric.FromPrometheusValue("kube_deployment_status_replicas_updated"), sdkMetric.GAUGE},
			{"podsMaxUnavailable", metric.FromPrometheusValue("kube_deployment_spec_strategy_rollingupdate_max_unavailable"), sdkMetric.GAUGE},
			{"namespace", metric.FromPrometheusLabelValue("kube_deployment_labels", "namespace"), sdkMetric.ATTRIBUTE},
			{"deploymentName", metric.FromPrometheusLabelValue("kube_deployment_labels", "deployment"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("namespace", "kube_namespace_labels"), sdkMetric.ATTRIBUTE},
			{"label.*", metric.InheritAllPrometheusLabelsFrom("deployment", "kube_deployment_labels"), sdkMetric.ATTRIBUTE},
		},
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

var kubeletAggregation = definition.SpecGroups{
	"pod": {
		IDGenerator: kubeletMetric.FromRawEntityIDGroupEntityIDGenerator("namespace"),
		Specs: []definition.Spec{
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
			{"net.rxBytesPerSecond", definition.FromRaw("rxBytes"), sdkMetric.RATE},
			{"net.txBytesPerSecond", definition.FromRaw("txBytes"), sdkMetric.RATE},
			{"net.errorsPerSecond", definition.FromRaw("errors"), sdkMetric.RATE},
		},
	},
	"container": {
		IDGenerator: kubeletMetric.FromRawGroupsEntityIDGenerator("podName"),
		Specs: []definition.Spec{
			{"containerName", definition.FromRaw("containerName"), sdkMetric.ATTRIBUTE},
			{"memoryUsedBytes", definition.FromRaw("usageBytes"), sdkMetric.GAUGE},
			{"cpuUsedCores", definition.FromRaw("usageNanoCores"), sdkMetric.GAUGE},
			{"podName", definition.FromRaw("podName"), sdkMetric.ATTRIBUTE},
			{"namespace", definition.FromRaw("namespace"), sdkMetric.ATTRIBUTE},
		},
	},
}
