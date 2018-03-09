package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
)

// StatsSummaryPath is the path where kubelet serves a summary with several information.
const StatsSummaryPath = "/stats/summary"

// Summary represents list of required data from /stats/summary endpoint
type Summary struct {
	Pods []Pod `json:"pods"`
	Node *Node `json:"node"`
}

// Node represents all required node data from kubelet endpoint
type Node struct {
	Name    *string  `json:"nodeName"`
	CPU     *CPU     `json:"cpu"`
	Memory  *Memory  `json:"memory"`
	Network *Network `json:"network"`
	Fs      *Fs      `json:"fs"`
	Runtime *Runtime `json:"runtime"`
}

// Pod represents all required pod and container data for kubelet integration
type Pod struct {
	PodRef     *PodRef     `json:"podRef"`
	Network    *Network    `json:"network"`
	Containers []Container `json:"containers"`
}

// PodRef represents name and namespace information of pod
type PodRef struct {
	Name      *string `json:"name"`
	Namespace *string `json:"namespace"`
}

// Network represents network data of pod or node
type Network struct {
	RxBytes  *int `json:"rxBytes"`
	TxBytes  *int `json:"txBytes"`
	RxErrors *int `json:"rxErrors"`
	TxErrors *int `json:"txErrors"`
}

// Container represents all required container data for kubelet integration
type Container struct {
	Name   *string `json:"name"`
	CPU    *CPU    `json:"cpu"`
	Memory *Memory `json:"memory"`
}

// CPU represents core CPU usage data of container or node
type CPU struct {
	UsageNanoCores       *int `json:"usageNanoCores"`
	UsageCoreNanoSeconds *int `json:"usageCoreNanoSeconds"`
}

// Memory represents memory usage data of container or node
type Memory struct {
	UsageBytes      *int `json:"usageBytes"`
	AvailableBytes  *int `json:"availableBytes"`
	WorkingSetBytes *int `json:"workingSetBytes"`
	RssBytes        *int `json:"rssBytes"`
	PageFaults      *int `json:"pageFaults"`
	MajorPageFaults *int `json:"majorPageFaults"`
}

// Fs represents filesystem data of a node
type Fs struct {
	AvailableBytes *int `json:"availableBytes"`
	CapacityBytes  *int `json:"capacityBytes"`
	UsedBytes      *int `json:"usedBytes"`
	InodesFree     *int `json:"inodesFree"`
	Inodes         *int `json:"inodes"`
	InodesUsed     *int `json:"inodesUsed"`
}

// Runtime represents runtime image filesystem data of a node
type Runtime struct {
	ImageFS *ImageFS `json:"imageFs"`
}

// ImageFS represents image filesystem usage data of a node
type ImageFS struct {
	AvailableBytes *int `json:"availableBytes"`
	CapacityBytes  *int `json:"capacityBytes"`
	UsedBytes      *int `json:"usedBytes"`
	InodesFree     *int `json:"inodesFree"`
	Inodes         *int `json:"inodes"`
	InodesUsed     *int `json:"inodesUsed"`
}

// GetMetricsData calls kubelet /stats/summary endpoint and returns unmarshalled response
func GetMetricsData(c endpoints.Client) (*Summary, error) {
	resp, err := c.Do(http.MethodGet, StatsSummaryPath)
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

func fetchNodeStats(n *Node) (definition.RawMetrics, string, error) {

	r := make(definition.RawMetrics)

	if n == nil {
		// TODO: check if better return nil or empty map
		return r, "", fmt.Errorf("node data not found in %s response, fetching node data skipped", StatsSummaryPath)
	}

	if n.Name == nil || *n.Name == "" {
		return r, "", fmt.Errorf("node identifier not found in %s response, fetching node data skipped", StatsSummaryPath)
	}

	AddStringRawMetric(r, "nodeName", n.Name)

	if n.CPU != nil {
		AddIntRawMetric(r, "usageNanoCores", n.CPU.UsageNanoCores)
		AddIntRawMetric(r, "usageCoreNanoSeconds", n.CPU.UsageCoreNanoSeconds)
	}

	if n.Memory != nil {
		AddIntRawMetric(r, "memoryUsageBytes", n.Memory.UsageBytes)
		AddIntRawMetric(r, "memoryAvailableBytes", n.Memory.AvailableBytes)
		AddIntRawMetric(r, "memoryWorkingSetBytes", n.Memory.WorkingSetBytes)
		AddIntRawMetric(r, "memoryRssBytes", n.Memory.RssBytes)
		AddIntRawMetric(r, "memoryPageFaults", n.Memory.PageFaults)
		AddIntRawMetric(r, "memoryMajorPageFaults", n.Memory.MajorPageFaults)
	}

	if n.Network != nil {
		AddIntRawMetric(r, "rxBytes", n.Network.RxBytes)
		AddIntRawMetric(r, "txBytes", n.Network.TxBytes)
		if n.Network.RxErrors != nil && n.Network.TxErrors != nil {
			r["errors"] = *n.Network.RxErrors + *n.Network.TxErrors
		}
	}

	if n.Fs != nil {
		AddIntRawMetric(r, "fsAvailableBytes", n.Fs.AvailableBytes)
		AddIntRawMetric(r, "fsCapacityBytes", n.Fs.CapacityBytes)
		AddIntRawMetric(r, "fsUsedBytes", n.Fs.UsedBytes)
		AddIntRawMetric(r, "fsInodesFree", n.Fs.InodesFree)
		AddIntRawMetric(r, "fsInodes", n.Fs.Inodes)
		AddIntRawMetric(r, "fsInodesUsed", n.Fs.InodesUsed)
	}
	if n.Runtime != nil {
		AddIntRawMetric(r, "runtimeAvailableBytes", n.Runtime.ImageFS.AvailableBytes)
		AddIntRawMetric(r, "runtimeCapacityBytes", n.Runtime.ImageFS.CapacityBytes)
		AddIntRawMetric(r, "runtimeUsedBytes", n.Runtime.ImageFS.UsedBytes)
		AddIntRawMetric(r, "runtimeInodesFree", n.Runtime.ImageFS.InodesFree)
		AddIntRawMetric(r, "runtimeInodes", n.Runtime.ImageFS.Inodes)
		AddIntRawMetric(r, "runtimeInodesUsed", n.Runtime.ImageFS.InodesUsed)
	}

	rawEntityID := fmt.Sprintf("%v", *n.Name)

	return r, rawEntityID, nil
}

func fetchPodStats(pod *Pod) (definition.RawMetrics, string, error) {
	r := make(definition.RawMetrics)

	if pod.PodRef == nil {
		return r, "", errors.New("pod ref data not found")
	}
	if pod.PodRef.Name == nil || pod.PodRef.Namespace == nil {
		return r, "", errors.New("pod identifier not found")

	}
	if *pod.PodRef.Name == "" || *pod.PodRef.Namespace == "" {
		return r, "", errors.New("empty pod identifier")
	}
	AddStringRawMetric(r, "podName", pod.PodRef.Name)
	AddStringRawMetric(r, "namespace", pod.PodRef.Namespace)

	if pod.Network != nil {
		AddIntRawMetric(r, "rxBytes", pod.Network.RxBytes)
		AddIntRawMetric(r, "txBytes", pod.Network.TxBytes)
		if pod.Network.RxErrors != nil && pod.Network.TxErrors != nil {
			r["errors"] = *pod.Network.RxErrors + *pod.Network.TxErrors
		}
	}

	rawEntityID := fmt.Sprintf("%s_%s", *pod.PodRef.Namespace, *pod.PodRef.Name)

	return r, rawEntityID, nil
}

func fetchContainerStats(c *Container) (definition.RawMetrics, error) {
	r := make(definition.RawMetrics)
	if c.Name == nil || *c.Name == "" {
		return r, errors.New("empty container name, fetching container data skipped")
	}

	AddStringRawMetric(r, "containerName", c.Name)

	if c.CPU != nil {
		AddIntRawMetric(r, "usageNanoCores", c.CPU.UsageNanoCores)
	}
	if c.Memory != nil {
		AddIntRawMetric(r, "usageBytes", c.Memory.UsageBytes)
	}

	return r, nil

}

// GroupStatsSummary groups specific data for pods, containers and node
func GroupStatsSummary(statsSummary *Summary) (definition.RawGroups, []error) {
	var errs []error
	var rawEntityID string
	g := definition.RawGroups{
		"pod":       {},
		"container": {},
		"node":      {},
	}

	if statsSummary == nil {
		errs = append(errs, fmt.Errorf("data not found in %s response", StatsSummaryPath))
		return g, errs
	}

	rawNodeData, rawEntityID, err := fetchNodeStats(statsSummary.Node)
	if err != nil {
		errs = append(errs, err)
	}

	if rawEntityID != "" && len(rawNodeData) > 0 {
		g["node"][rawEntityID] = rawNodeData
	}

	if statsSummary.Pods == nil {
		errs = append(errs, fmt.Errorf("pods data not found in %s response, fetching pod and container data skipped", StatsSummaryPath))
		return g, errs
	}

PodListLoop:
	for _, pod := range statsSummary.Pods {
		rawPodMetrics, rawEntityID, err := fetchPodStats(&pod)
		if err != nil {
			errs = append(errs, err)
			continue PodListLoop
		}
		// TODO: check for empty rawEntity and len needed?
		g["pod"][rawEntityID] = rawPodMetrics

		if pod.Containers == nil {
			errs = append(errs, fmt.Errorf("container data not found in %s response", StatsSummaryPath))
			continue PodListLoop
		}

	ContainerListLoop:
		for _, container := range pod.Containers {
			rawContainerMetrics, err := fetchContainerStats(&container)
			if err != nil {
				errs = append(errs, err)
				continue ContainerListLoop
			}
			rawContainerMetrics["podName"] = rawPodMetrics["podName"]
			rawContainerMetrics["namespace"] = rawPodMetrics["namespace"]

			rawEntityID = fmt.Sprintf("%s_%s_%s", rawPodMetrics["namespace"], rawPodMetrics["podName"], rawContainerMetrics["containerName"])

			// TODO: check for empty rawENtity and len needed?
			g["container"][rawEntityID] = rawContainerMetrics
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

// AddIntRawMetric adds a new metric to a RawMetrics if it exists
func AddIntRawMetric(r definition.RawMetrics, name string, valuePtr *int) {
	if valuePtr != nil {
		r[name] = *valuePtr
	}
}

// AddStringRawMetric adds a new metric to a RawMetrics if it exists
func AddStringRawMetric(r definition.RawMetrics, name string, valuePtr *string) {
	if valuePtr != nil {
		r[name] = *valuePtr
	}
}
