package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/definition"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

// Summary represents list of required data from /stats/summary endpoint
type Summary struct {
	Pods []Pod `json:"pods"`
}

// Pod represents all required pod and container data for kubelet integration
type Pod struct {
	PodRef     PodRef      `json:"podRef"`
	Network    Network     `json:"network"`
	Containers []Container `json:"containers"`
}

// PodRef represents name and namespace information of pod
type PodRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Network represents network data of pod
type Network struct {
	RxBytes  int `json:"rxBytes"`
	TxBytes  int `json:"txBytes"`
	RxErrors int `json:"rxErrors"`
	TxErrors int `json:"txErrors"`
}

// Container represents all required container data for kubelet integration
type Container struct {
	Name   string `json:"name"`
	CPU    CPU    `json:"cpu"`
	Memory Memory `json:"memory"`
}

// CPU represents core CPU usage data of container
type CPU struct {
	UsageNanoCores int `json:"usageNanoCores"`
}

// Memory represents memory usage data of container
type Memory struct {
	UsageBytes int `json:"usageBytes"`
}

// GetMetricsData calls kubelet /stats/summary endpoint and returns unmarshalled response
func GetMetricsData(netClient *http.Client, URL string) (*Summary, error) {
	resp, err := netClient.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("Error trying to connect to '%s'. Got error: %v", URL, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading the response body of '%s'. Got error: %v", URL, err.Error())
	}

	var message = new(Summary)
	err = json.Unmarshal(body, message)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling the response body. Got error: %v", err.Error())
	}

	return message, nil

}

// GroupStatsSummary groups specific data for pods and containers
func GroupStatsSummary(statsSummary *Summary) (definition.MetricGroups, []error) {
	var errs []error
	g := definition.MetricGroups{
		"pod":       {},
		"container": {},
	}

	for _, pod := range statsSummary.Pods {
		if pod.PodRef.Name == "" || pod.PodRef.Namespace == "" {
			errs = append(errs, fmt.Errorf("empty pod identifier, fetching pod data skipped"))
			continue
		}
		podData := definition.RawMetrics{
			"podName":   pod.PodRef.Name,
			"namespace": pod.PodRef.Namespace,
			"rxBytes":   pod.Network.RxBytes,
			"txBytes":   pod.Network.TxBytes,
			"errors":    pod.Network.RxErrors + pod.Network.TxErrors,
		}

		g["pod"][pod.PodRef.Name] = podData

		for _, container := range pod.Containers {
			if container.Name == "" {
				errs = append(errs, fmt.Errorf("empty container name, fetching container data skipped"))
				continue
			}
			containerData := definition.RawMetrics{
				"usageBytes":     container.Memory.UsageBytes,
				"usageNanoCores": container.CPU.UsageNanoCores,
				"podName":        podData["podName"],
				"namespace":      podData["namespace"],
				"containerName":  container.Name,
			}

			g["container"][container.Name] = containerData
		}
	}

	return g, errs
}

// Populate populates an integration with the given metrics and definition.
func Populate(i *sdk.IntegrationProtocol2, definitions []definition.Metric, groups definition.MetricGroups) (bool, []error) {
	var populated bool
	var errs []error
	for entitySourceName, entities := range groups {
		for entityID, r := range entities {
			e, err := i.Entity(entityID, fmt.Sprintf("k8s/%s", entitySourceName))
			if err != nil {
				errs = append(errs, err)
				continue
			}

			oneMetricSet, extractErrs := definition.OneMetricSetExtract(r)(definitions)
			if len(extractErrs) != 0 {
				for _, err := range extractErrs {
					errs = append(errs, fmt.Errorf("entity id: %s: %s", entityID, err))
				}
			}

			if len(oneMetricSet) > 0 {
				ms := e.NewMetricSet(fmt.Sprintf("K8s%vSample", strings.Title(entitySourceName)))
				for k, v := range oneMetricSet[0] {
					ms[k] = v
				}

				populated = true
			}
		}
	}

	return populated, errs
}
