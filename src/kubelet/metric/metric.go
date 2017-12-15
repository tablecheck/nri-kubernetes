package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
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
func GroupStatsSummary(statsSummary *Summary) (definition.RawGroups, []error) {
	var errs []error
	g := definition.RawGroups{
		"pod":       {},
		"container": {},
	}

	for _, pod := range statsSummary.Pods {
		if pod.PodRef.Name == "" || pod.PodRef.Namespace == "" {
			errs = append(errs, fmt.Errorf("empty pod identifier, fetching pod data skipped"))
			continue
		}
		rawEntityID := fmt.Sprintf("%v_%v", pod.PodRef.Namespace, pod.PodRef.Name)
		podData := definition.RawMetrics{
			"podName":   pod.PodRef.Name,
			"namespace": pod.PodRef.Namespace,
			"rxBytes":   pod.Network.RxBytes,
			"txBytes":   pod.Network.TxBytes,
			"errors":    pod.Network.RxErrors + pod.Network.TxErrors,
		}

		g["pod"][rawEntityID] = podData

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
			rawEntityID = fmt.Sprintf("%v_%v_%v", pod.PodRef.Namespace, pod.PodRef.Name, container.Name)
			g["container"][rawEntityID] = containerData
		}
	}

	return g, errs
}

// FromRawGroupsEntityIDGenerator generates an entityID from the pod name from kubelet. It's only used for k8s containers.
func FromRawGroupsEntityIDGenerator(key string) definition.MetricSetEntityIDGeneratorFunc {
	return func(groupLabel string, rawEntityID string, g definition.RawGroups) (string, error) {
		v, ok := g[groupLabel][rawEntityID][key]
		if !ok {
			return "", fmt.Errorf("error generating metric set entity id from kubelet raw data. Key: %v", key)
		}
		return v.(string), nil
	}
}

// FromRawEntityIDGroupEntityIDGenerator generates an entityID from the raw entity ID
// which is composed of namespace and pod name. It's used only for k8s pods.
func FromRawEntityIDGroupEntityIDGenerator(key string) definition.MetricSetEntityIDGeneratorFunc {
	return func(groupLabel string, rawEntityID string, g definition.RawGroups) (string, error) {
		toRemove := g[groupLabel][rawEntityID][key]
		v := strings.TrimPrefix(rawEntityID, fmt.Sprintf("%s_", toRemove))

		if v == "" {
			return "", fmt.Errorf("error generating metric set entity id from kubelet raw data")
		}

		return v, nil
	}
}
