package testdata

import (
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
)

// ExpectedRawData is the expectation for main fetch_test tests.
var ExpectedRawData = definition.RawGroups{
	"pod": {
		"kube-system_newrelic-infra-rz225": {
			"createdKind":    "DaemonSet",
			"createdBy":      "newrelic-infra",
			"nodeIP":         "192.168.99.100",
			"namespace":      "kube-system",
			"podName":        "newrelic-infra-rz225",
			"nodeName":       "minikube",
			"startTime":      parseTime("2018-02-14T16:26:33Z"),
			"status":         "Running",
			"isReady":        true,
			"isScheduled":    true,
			"createdAt":      parseTime("2018-02-14T16:26:33Z"),
			"deploymentName": "",
			"labels": map[string]string{
				"controller-revision-hash": "3887482659",
				"name": "newrelic-infra",
				"pod-template-generation": "1",
			},
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq": {
			"createdKind": "ReplicaSet",
			"createdBy":   "kube-state-metrics-57f4659995",
			"nodeIP":      "192.168.99.100",
			"namespace":   "kube-system",
			"podName":     "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":    "minikube",
			//"startTime":      parseTime("2018-02-14T16:27:38Z"), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"status":         "Running", // Note that even the status in the payload is Pending, we set it as Running. This is due to a bug in Kubelet. See https://github.com/kubernetes/kubernetes/pull/57106
			"isReady":        false,     // TODO investigate. See https://newrelic.atlassian.net/browse/IHOST-658
			"isScheduled":    true,
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
			"containerName":    "newrelic-infra",
			"containerID":      "docker://ef0b60ee1eea54af356847a5c99a3ec91a57bae627028e3efe41c1a29fe641e5",
			"containerImage":   "newrelic/ohaik:1.0.0-beta3",
			"containerImageID": "docker-pullable://newrelic/ohaik@sha256:115eb17a8242c02bf698259f6c883c9ad5e9e020517156881a4017fd88295444",
			"namespace":        "kube-system",
			"podName":          "newrelic-infra-rz225",
			"nodeName":         "minikube",
			"nodeIP":           "192.168.99.100",
			"restartCount":     int32(6),
			"isReady":          true,
			"status":           "Running",
			"deploymentName":   "",
			//"reason": "", // TODO
			"startedAt":            parseTime("2018-02-27T15:21:16Z"),
			"cpuRequestedCores":    int64(100),
			"memoryRequestedBytes": int64(104857600),
			"memoryLimitBytes":     int64(104857600),
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_kube-state-metrics": {
			"containerName": "kube-state-metrics",
			//"containerID":      "", // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"containerImage": "quay.io/coreos/kube-state-metrics:v1.1.0",
			//"containerImageID": "", // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"namespace": "kube-system",
			"podName":   "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":  "minikube",
			"nodeIP":    "192.168.99.100",
			//"restartCount": int32(7), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			//"isReady":              false, // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106 // TODO investigate. See https://newrelic.atlassian.net/browse/IHOST-658
			"status":         "Running", // The value does not exist but we force it to "Running" because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"deploymentName": "kube-state-metrics",
			//"startedAt":            parseTime("2018-02-27T15:21:37Z"), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"cpuRequestedCores":    int64(101),
			"cpuLimitCores":        int64(101),
			"memoryRequestedBytes": int64(106954752),
			"memoryLimitBytes":     int64(106954752),
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_addon-resizer": {
			"containerName": "addon-resizer",
			//"containerID":      "", //"containerID":      "", // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"containerImage": "gcr.io/google_containers/addon-resizer:1.0",
			//"containerImageID": "", // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"namespace": "kube-system",
			"podName":   "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":  "minikube",
			"nodeIP":    "192.168.99.100",
			//"restartCount": int32(7), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			//"isReady":        false, // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106 // TODO investigate. See https://newrelic.atlassian.net/browse/IHOST-658
			"status":         "Running", // The value does not exist but we force it to "Running" because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"deploymentName": "kube-state-metrics",
			//"reason": "", // TODO
			//"startedAt":            parseTime("2018-02-27T15:21:38Z"), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"cpuRequestedCores":    int64(100),
			"cpuLimitCores":        int64(100),
			"memoryRequestedBytes": int64(31457280),
			"memoryLimitBytes":     int64(31457280),
		},
	},
}

func parseTime(raw string) time.Time {
	t, _ := time.Parse(time.RFC3339, raw)

	return t
}
