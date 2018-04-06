package testdata

import (
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
)

// ExpectedGroupData is the expectation for main group_test tests.
var ExpectedGroupData = definition.RawGroups{
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
			"errors":  uint64(0),
			"rxBytes": uint64(106175985),
			"txBytes": uint64(35714359),
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
			"errors":  uint64(0),
			"rxBytes": uint64(32575098),
			"txBytes": uint64(27840584),
		},
	},
	"container": {
		"kube-system_newrelic-infra-rz225_newrelic-infra": {
			"containerName":    "newrelic-infra",
			"containerID":      "ef0b60ee1eea54af356847a5c99a3ec91a57bae627028e3efe41c1a29fe641e5",
			"containerImage":   "newrelic/ohaik:1.0.0-beta3",
			"containerImageID": "newrelic/ohaik@sha256:115eb17a8242c02bf698259f6c883c9ad5e9e020517156881a4017fd88295444",
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
			"usageBytes":           uint64(18083840),
			"usageNanoCores":       uint64(17428240),
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_kube-state-metrics": {
			"containerName":    "kube-state-metrics",
			"containerID":      "c452821fcf6c5f594d4f98a1426e7a2c51febb65d5d50d92903f9dfb367bfba7",
			"containerImage":   "quay.io/coreos/kube-state-metrics:v1.1.0",
			"containerImageID": "quay.io/coreos/kube-state-metrics@sha256:52a2c47355c873709bb4e37e990d417e9188c2a778a0c38ed4c09776ddc54efb",
			"namespace":        "kube-system",
			"podName":          "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":         "minikube",
			"nodeIP":           "192.168.99.100",
			//"restartCount": int32(7), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			//"isReady":              false, // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106 // TODO investigate. See https://newrelic.atlassian.net/browse/IHOST-658
			"status":         "Running", // The value does not exist but we force it to "Running" because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"deploymentName": "kube-state-metrics",
			//"startedAt":            parseTime("2018-02-27T15:21:37Z"), // Missing because Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
			"cpuRequestedCores":    int64(101),
			"cpuLimitCores":        int64(101),
			"memoryRequestedBytes": int64(106954752),
			"memoryLimitBytes":     int64(106954752),
			"usageBytes":           uint64(15568896),
			"usageNanoCores":       uint64(941138),
		},
		"kube-system_kube-state-metrics-57f4659995-6n2qq_addon-resizer": {
			"containerName":    "addon-resizer",
			"containerID":      "3328c17bfd22f1a82fcdf8707c2f8f040c462e548c24780079bba95d276d93e1",
			"containerImage":   "gcr.io/google_containers/addon-resizer:1.0",
			"containerImageID": "gcr.io/google_containers/addon-resizer@sha256:e77acf80697a70386c04ae3ab494a7b13917cb30de2326dcf1a10a5118eddabe",
			"namespace":        "kube-system",
			"podName":          "kube-state-metrics-57f4659995-6n2qq",
			"nodeName":         "minikube",
			"nodeIP":           "192.168.99.100",
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
			"usageBytes":           uint64(6373376),
			"usageNanoCores":       uint64(131742),
		},
	},
	"node": {
		"minikube": {
			"nodeName":              "minikube",
			"errors":                uint64(0),
			"fsAvailableBytes":      uint64(14924988416),
			"fsCapacityBytes":       uint64(17293533184),
			"fsInodes":              uint64(9732096),
			"fsInodesFree":          uint64(9713372),
			"fsInodesUsed":          uint64(18724),
			"fsUsedBytes":           uint64(1355673600),
			"memoryAvailableBytes":  uint64(791736320),
			"memoryMajorPageFaults": uint64(0),
			"memoryPageFaults":      uint64(113947),
			"memoryRssBytes":        uint64(660684800),
			"memoryUsageBytes":      uint64(1843650560),
			"memoryWorkingSetBytes": uint64(1305468928),
			"runtimeAvailableBytes": uint64(14924988416),
			"runtimeCapacityBytes":  uint64(17293533184),
			"runtimeInodes":         uint64(9732096),
			"runtimeInodesFree":     uint64(9713372),
			"runtimeInodesUsed":     uint64(18724),
			"runtimeUsedBytes":      uint64(969241979),
			"rxBytes":               uint64(1507694406),
			"txBytes":               uint64(120789968),
			"usageCoreNanoSeconds":  uint64(22332102208229),
			"usageNanoCores":        uint64(228759290),
		},
	},
}
