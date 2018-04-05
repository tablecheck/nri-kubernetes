package metric

import (
	"testing"

	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric/testdata"
	sdkMetric "github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func parseTime(raw string) time.Time {
	t, _ := time.Parse(time.RFC3339, raw)

	return t
}

var expectedMetrics = []*sdk.EntityData{
	{
		Entity: sdk.Entity{
			Name: "newrelic-infra-rz225",
			Type: "k8s:test-cluster:kube-system:pod",
		},
		Metrics: []sdkMetric.MetricSet{
			{
				"entityName":                     "k8s:test-cluster:kube-system:pod:newrelic-infra-rz225",
				"event_type":                     "K8sPodSample",
				"net.rxBytesPerSecond":           0., // 106175985, but is RATE
				"net.txBytesPerSecond":           0., // 35714359, but is RATE
				"net.errorCount":                 uint64(0),
				"createdAt":                      parseTime("2018-02-14T16:26:33Z").Unix(),
				"startTime":                      parseTime("2018-02-14T16:26:33Z").Unix(),
				"createdKind":                    "DaemonSet",
				"createdBy":                      "newrelic-infra",
				"nodeIP":                         "192.168.99.100",
				"namespace":                      "kube-system",
				"nodeName":                       "minikube",
				"podName":                        "newrelic-infra-rz225",
				"isReady":                        "true",
				"status":                         "Running",
				"isScheduled":                    "true",
				"deploymentName":                 "",
				"label.controller-revision-hash": "3887482659",
				"label.name":                     "newrelic-infra",
				"label.pod-template-generation":  "1",
				"displayName":                    "newrelic-infra-rz225", // From manipulator
				"clusterName":                    "test-cluster",         // From manipulator
			},
		},
		Inventory: sdk.Inventory{},
		Events:    []sdk.Event{},
	},
	{
		Entity: sdk.Entity{
			Name: "newrelic-infra",
			Type: "k8s:test-cluster:kube-system:newrelic-infra-rz225:container",
		},
		Metrics: []sdkMetric.MetricSet{
			{
				"entityName":           "k8s:test-cluster:kube-system:newrelic-infra-rz225:container:newrelic-infra",
				"event_type":           "K8sContainerSample",
				"memoryUsedBytes":      uint64(18083840),
				"cpuUsedCores":         0.01742824,
				"containerName":        "newrelic-infra",
				"containerID":          "docker://ef0b60ee1eea54af356847a5c99a3ec91a57bae627028e3efe41c1a29fe641e5",
				"containerImage":       "newrelic/ohaik:1.0.0-beta3",
				"containerImageID":     "docker-pullable://newrelic/ohaik@sha256:115eb17a8242c02bf698259f6c883c9ad5e9e020517156881a4017fd88295444",
				"deploymentName":       "",
				"namespace":            "kube-system",
				"podName":              "newrelic-infra-rz225",
				"nodeName":             "minikube",
				"nodeIP":               "192.168.99.100",
				"restartCount":         int32(6),
				"cpuRequestedCores":    0.1,
				"memoryRequestedBytes": int64(104857600),
				"memoryLimitBytes":     int64(104857600),
				"status":               "Running",
				"isReady":              "true",
				//"reason":               "",      // TODO ?
				"displayName": "newrelic-infra", // From manipulator
				"clusterName": "test-cluster",   // From manipulator
			},
		},
		Inventory: sdk.Inventory{},
		Events:    []sdk.Event{},
	},
}

// We reduce the test fixtures in order to simplify testing.
var kubeletSpecs = definition.SpecGroups{
	"pod":       KubeletSpecs["pod"],
	"container": KubeletSpecs["container"],
}

func TestPopulateK8s(t *testing.T) {
	p := NewK8sPopulator()

	i, err := sdk.NewIntegrationProtocol2("test", "test", new(struct{}))
	assert.NoError(t, err)
	i.Clear()

	// We reduce the test fixtures in order to simplify testing.
	foo := definition.RawGroups{
		"pod": {
			"kube-system_newrelic-infra-rz225": testdata.ExpectedGroupData["pod"]["kube-system_newrelic-infra-rz225"],
		},
		"container": {
			"kube-system_newrelic-infra-rz225_newrelic-infra": testdata.ExpectedGroupData["container"]["kube-system_newrelic-infra-rz225_newrelic-infra"],
		},
	}
	ok, err := p.Populate(foo, kubeletSpecs, i, "test-cluster")

	// Expected errs (missing data)
	expectedErr := MultipleErrs{
		true,
		[]error{
			errors.New("entity id: kube-system_newrelic-infra-rz225_newrelic-infra: error fetching value for metric cpuLimitCores. Error: FromRaw: metric not found. SpecGroup: container, EntityID: kube-system_newrelic-infra-rz225_newrelic-infra, Metric: cpuLimitCores"),
			errors.New("entity id: kube-system_newrelic-infra-rz225_newrelic-infra: error fetching value for metric reason. Error: FromRaw: metric not found. SpecGroup: container, EntityID: kube-system_newrelic-infra-rz225_newrelic-infra, Metric: reason"),
		},
	}

	assert.EqualError(t, err, expectedErr.Error())
	assert.True(t, ok)
	assert.ElementsMatch(t, expectedMetrics, i.Data)
}
