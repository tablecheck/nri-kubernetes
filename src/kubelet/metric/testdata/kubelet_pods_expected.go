package testdata

import (
	"time"

	"github.com/newrelic/nri-kubernetes/src/definition"
)

// ExpectedRawData is the expectation for main fetch_test tests.
var ExpectedRawData = definition.RawGroups{
	"pod": {
		"kube-system_newrelic-infra-rz225": {
			"createdKind": "DaemonSet",
			"createdBy":   "newrelic-infra",
			"nodeIP":      "192.168.99.100",
			"namespace":   "kube-system",
			"podName":     "newrelic-infra-rz225",
			"nodeName":    "minikube",
			"startTime":   parseTime("2018-02-14T16:26:33Z"),
			"status":      "Running",
			"isReady":     "True",
			"isScheduled": "True",
			"createdAt":   parseTime("2018-02-14T16:26:33Z"),
			"labels": map[string]string{
				"controller-revision-hash": "3887482659",
				"name": "newrelic-infra",
				"pod-template-generation": "1",
			},
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq": {
			"createdKind":    "ReplicaSet",
			"createdBy":      "kube-state-metrics-57f4659995",
			"nodeIP":         "192.168.99.100",
			"namespace":      "kube-system",
			"podName":        "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":       "minikube",
			"status":         "Running", // Running because is fake pending pod.
			"isReady":        "True",
			"isScheduled":    "True",
			"createdAt":      parseTime("2018-02-14T16:27:38Z"),
			"deploymentName": "kube-state-metrics",
			"labels": map[string]string{
				"k8s-app":           "kube-state-metrics",
				"pod-template-hash": "1390215551",
			},
		},
	},
	"container": {
		"kube-system_newrelic-infra-rz225_newrelic-infra": {
			"containerName":  "newrelic-infra",
			"containerImage": "newrelic/ohaik:1.0.0-beta3",
			"namespace":      "kube-system",
			"podName":        "newrelic-infra-rz225",
			"nodeName":       "minikube",
			"nodeIP":         "192.168.99.100",
			"restartCount":   int32(6),
			"isReady":        true,
			"status":         "Running",
			//"reason": "", // TODO
			"startedAt":            parseTime("2018-02-27T15:21:16Z"),
			"cpuRequestedCores":    int64(100),
			"memoryRequestedBytes": int64(104857600),
			"memoryLimitBytes":     int64(104857600),
			"labels": map[string]string{
				"controller-revision-hash": "3887482659",
				"name": "newrelic-infra",
				"pod-template-generation": "1",
			},
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_kube-state-metrics": {
			"containerName":  "kube-state-metrics",
			"containerImage": "quay.io/coreos/kube-state-metrics:v1.1.0",
			"namespace":      "kube-system",
			"podName":        "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":       "minikube",
			"nodeIP":         "192.168.99.100",
			//"restartCount": int32(7), // No restartCount since there is no restartCount in status field in the pod.
			//"isReady":              false, // No isReady since there is no isReady in status field in the pod.
			//"status":         "Running", // No Status since there is no ContainerStatuses field in the pod.
			"deploymentName": "kube-state-metrics",
			//"startedAt":            parseTime("2018-02-27T15:21:37Z"), // No startedAt since there is no startedAt in status field in the pod.
			"cpuRequestedCores":    int64(101),
			"cpuLimitCores":        int64(101),
			"memoryRequestedBytes": int64(106954752),
			"memoryLimitBytes":     int64(106954752),
			"labels": map[string]string{
				"k8s-app":           "kube-state-metrics",
				"pod-template-hash": "1390215551",
			},
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_addon-resizer": {
			"containerName":  "addon-resizer",
			"containerImage": "gcr.io/google_containers/addon-resizer:1.0",
			"namespace":      "kube-system",
			"podName":        "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":       "minikube",
			"nodeIP":         "192.168.99.100",
			//"restartCount": int32(7), // No restartCount since there is no restartCount in status field in the pod.
			//"isReady":        false, // No isReady since there is no isReady in status field in the pod.
			//"status":         "Running", // No Status since there is no ContainerStatuses field in the pod.
			"deploymentName": "kube-state-metrics",
			//"reason": "", // TODO
			//"startedAt":            parseTime("2018-02-27T15:21:38Z"), // No startedAt since there is no startedAt in status field in the pod.
			"cpuRequestedCores":    int64(100),
			"cpuLimitCores":        int64(100),
			"memoryRequestedBytes": int64(31457280),
			"memoryLimitBytes":     int64(31457280),
			"labels": map[string]string{
				"k8s-app":           "kube-state-metrics",
				"pod-template-hash": "1390215551",
			},
		},
	},
}

func parseTime(raw string) time.Time {
	t, _ := time.Parse(time.RFC3339, raw)

	return t
}
