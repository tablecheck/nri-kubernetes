package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
)

const statsSummaryPath = "/stats/summary"

// Summary represents list of required data from /stats/summary endpoint
type Summary struct {
	Pods []Pod `json:"pods"`
	Node Node  `json:"node"`
}

// Node represents all required node data from kubelet endpoint
type Node struct {
	Name    string  `json:"nodeName"`
	CPU     CPU     `json:"cpu"`
	Memory  Memory  `json:"memory"`
	Network Network `json:"network"`
	Fs      Fs      `json:"fs"`
	Runtime Runtime `json:"runtime"`
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

// Network represents network data of pod or node
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

// CPU represents core CPU usage data of container or node
type CPU struct {
	UsageNanoCores       int `json:"usageNanoCores"`
	UsageCoreNanoSeconds int `json:"usageCoreNanoSeconds"`
}

// Memory represents memory usage data of container or node
type Memory struct {
	UsageBytes      int `json:"usageBytes"`
	AvailableBytes  int `json:"availableBytes"`
	WorkingSetBytes int `json:"workingSetBytes"`
	RssBytes        int `json:"rssBytes"`
	PageFaults      int `json:"pageFaults"`
	MajorPageFaults int `json:"majorPageFaults"`
}

// Fs represents filesystem data of a node
type Fs struct {
	AvailableBytes int `json:"availableBytes"`
	CapacityBytes  int `json:"capacityBytes"`
	UsedBytes      int `json:"usedBytes"`
	InodesFree     int `json:"inodesFree"`
	Inodes         int `json:"inodes"`
	InodesUsed     int `json:"inodesUsed"`
}

// Runtime represents runtime image filesystem data of a node
type Runtime struct {
	ImageFS ImageFS `json:"imageFs"`
}

// ImageFS represents image filesystem usage data of a node
type ImageFS struct {
	AvailableBytes int `json:"availableBytes"`
	CapacityBytes  int `json:"capacityBytes"`
	UsedBytes      int `json:"usedBytes"`
	InodesFree     int `json:"inodesFree"`
	Inodes         int `json:"inodes"`
	InodesUsed     int `json:"inodesUsed"`
}

// GetMetricsData calls kubelet /stats/summary endpoint and returns unmarshalled response
func GetMetricsData(c endpoints.Client) (*Summary, error) {
	resp, err := c.Do(http.MethodGet, statsSummaryPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error calling kubelet endpoint. Got status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response body of kubelet endpoint. Got error: %v", err.Error())
	}

	var summary = new(Summary)
	err = json.Unmarshal(body, summary)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling the response body. Got error: %v", err.Error())
	}

	return summary, nil

}

// GroupStatsSummary groups specific data for pods, containers and node
func GroupStatsSummary(statsSummary *Summary) (definition.RawGroups, []error) {
	var errs []error
	g := definition.RawGroups{
		"pod":       {},
		"container": {},
		"node":      {},
	}

	node := statsSummary.Node
	if node.Name == "" {
		errs = append(errs, fmt.Errorf("empty node identifier, fetching node data skipped"))
	} else {
		rawEntityID := fmt.Sprintf("%v", statsSummary.Node.Name)
		g["node"][rawEntityID] = definition.RawMetrics{
			"nodeName": node.Name,
			// CPU
			"usageNanoCores":       node.CPU.UsageNanoCores,
			"usageCoreNanoSeconds": node.CPU.UsageCoreNanoSeconds,
			// Memory
			"memoryUsageBytes":      node.Memory.UsageBytes,
			"memoryAvailableBytes":  node.Memory.AvailableBytes,
			"memoryWorkingSetBytes": node.Memory.WorkingSetBytes,
			"memoryRssBytes":        node.Memory.RssBytes,
			"memoryPageFaults":      node.Memory.PageFaults,
			"memoryMajorPageFaults": node.Memory.MajorPageFaults,
			// Network
			"rxBytes": node.Network.RxBytes,
			"txBytes": node.Network.TxBytes,
			"errors":  node.Network.RxErrors + node.Network.TxErrors,
			// Fs
			"fsAvailableBytes": node.Fs.AvailableBytes,
			"fsCapacityBytes":  node.Fs.CapacityBytes,
			"fsUsedBytes":      node.Fs.UsedBytes,
			"fsInodesFree":     node.Fs.InodesFree,
			"fsInodes":         node.Fs.Inodes,
			"fsInodesUsed":     node.Fs.InodesUsed,
			// Runtime
			"runtimeAvailableBytes": node.Runtime.ImageFS.AvailableBytes,
			"runtimeCapacityBytes":  node.Runtime.ImageFS.CapacityBytes,
			"runtimeUsedBytes":      node.Runtime.ImageFS.UsedBytes,
			"runtimeInodesFree":     node.Runtime.ImageFS.InodesFree,
			"runtimeInodes":         node.Runtime.ImageFS.Inodes,
			"runtimeInodesUsed":     node.Runtime.ImageFS.InodesUsed,
		}
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

// KubeletNamespaceFetcher fetches the namespace from a Kubelet RawGroups information
func KubeletNamespaceFetcher(groupLabel, entityID string, groups definition.RawGroups) (string, error) {
	gl, ok := groups[groupLabel]
	if !ok {
		return config.UnknownNamespace, fmt.Errorf("no grouplabel %q found", groupLabel)
	}
	en, ok := gl[entityID]
	if !ok {
		return config.UnknownNamespace, fmt.Errorf("no entityID %q found for grouplabel %q", entityID, groupLabel)
	}

	ns, ok := en["namespace"]
	if !ok {
		return config.UnknownNamespace, fmt.Errorf("no namespace found for groupLabel %q and entityID %q", groupLabel, entityID)
	}
	return ns.(string), nil
}
