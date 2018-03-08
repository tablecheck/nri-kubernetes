package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	v1 "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/config"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
)

// StatsSummaryPath is the path where kubelet serves a summary with several information.
const StatsSummaryPath = "/stats/summary"

// GetMetricsData calls kubelet /stats/summary endpoint and returns unmarshalled response
func GetMetricsData(c endpoints.Client) (v1.Summary, error) {
	resp, err := c.Do(http.MethodGet, StatsSummaryPath)
	if err != nil {
		return v1.Summary{}, err
	}
	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode != http.StatusOK {
		return v1.Summary{}, fmt.Errorf("error calling kubelet endpoint. Got status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return v1.Summary{}, fmt.Errorf("error reading the response body of kubelet endpoint. Got error: %v", err.Error())
	}

	var summary = new(v1.Summary)
	err = json.Unmarshal(body, summary)
	if err != nil {
		return v1.Summary{}, fmt.Errorf("error unmarshaling the response body. Got error: %v", err.Error())
	}

	return *summary, nil

}

func fetchNodeStats(n v1.NodeStats) (definition.RawMetrics, string, error) {
	r := make(definition.RawMetrics)

	nodeName := n.NodeName
	if nodeName == "" {
		return r, "", fmt.Errorf("empty node identifier found in %s response, fetching node data skipped", StatsSummaryPath)
	}

	r["nodeName"] = nodeName

	if n.CPU != nil {
		AddUintRawMetric(r, "usageNanoCores", n.CPU.UsageNanoCores)
		AddUintRawMetric(r, "usageCoreNanoSeconds", n.CPU.UsageCoreNanoSeconds)
	}

	if n.Memory != nil {
		AddUintRawMetric(r, "memoryUsageBytes", n.Memory.UsageBytes)
		AddUintRawMetric(r, "memoryAvailableBytes", n.Memory.AvailableBytes)
		AddUintRawMetric(r, "memoryWorkingSetBytes", n.Memory.WorkingSetBytes)
		AddUintRawMetric(r, "memoryRssBytes", n.Memory.RSSBytes)
		AddUintRawMetric(r, "memoryPageFaults", n.Memory.PageFaults)
		AddUintRawMetric(r, "memoryMajorPageFaults", n.Memory.MajorPageFaults)
	}

	if n.Network != nil {
		AddUintRawMetric(r, "rxBytes", n.Network.RxBytes)
		AddUintRawMetric(r, "txBytes", n.Network.TxBytes)
		if n.Network.RxErrors != nil && n.Network.TxErrors != nil {
			r["errors"] = *n.Network.RxErrors + *n.Network.TxErrors
		}
	}

	if n.Fs != nil {
		AddUintRawMetric(r, "fsAvailableBytes", n.Fs.AvailableBytes)
		AddUintRawMetric(r, "fsCapacityBytes", n.Fs.CapacityBytes)
		AddUintRawMetric(r, "fsUsedBytes", n.Fs.UsedBytes)
		AddUintRawMetric(r, "fsInodesFree", n.Fs.InodesFree)
		AddUintRawMetric(r, "fsInodes", n.Fs.Inodes)
		AddUintRawMetric(r, "fsInodesUsed", n.Fs.InodesUsed)
	}
	if n.Runtime != nil && n.Runtime.ImageFs != nil {
		AddUintRawMetric(r, "runtimeAvailableBytes", n.Runtime.ImageFs.AvailableBytes)
		AddUintRawMetric(r, "runtimeCapacityBytes", n.Runtime.ImageFs.CapacityBytes)
		AddUintRawMetric(r, "runtimeUsedBytes", n.Runtime.ImageFs.UsedBytes)
		AddUintRawMetric(r, "runtimeInodesFree", n.Runtime.ImageFs.InodesFree)
		AddUintRawMetric(r, "runtimeInodes", n.Runtime.ImageFs.Inodes)
		AddUintRawMetric(r, "runtimeInodesUsed", n.Runtime.ImageFs.InodesUsed)
	}

	rawEntityID := fmt.Sprintf("%v", nodeName)

	return r, rawEntityID, nil
}

func fetchPodStats(pod v1.PodStats) (definition.RawMetrics, string, error) {
	r := make(definition.RawMetrics)

	podName := pod.PodRef.Name
	namespace := pod.PodRef.Namespace

	if podName == "" || namespace == "" {
		return r, "", errors.New("empty pod identifiers")
	}

	r["podName"] = podName
	r["namespace"] = namespace

	if pod.Network != nil {
		AddUintRawMetric(r, "rxBytes", pod.Network.RxBytes)
		AddUintRawMetric(r, "txBytes", pod.Network.TxBytes)
		if pod.Network.RxErrors != nil && pod.Network.TxErrors != nil {
			r["errors"] = *pod.Network.RxErrors + *pod.Network.TxErrors
		}
	}

	rawEntityID := fmt.Sprintf("%s_%s", namespace, podName)

	return r, rawEntityID, nil
}

func fetchContainerStats(c v1.ContainerStats) (definition.RawMetrics, error) {
	r := make(definition.RawMetrics)

	if c.Name == "" {
		return r, errors.New("empty container name, fetching container data skipped")
	}
	r["containerName"] = c.Name

	if c.CPU != nil {
		AddUintRawMetric(r, "usageNanoCores", c.CPU.UsageNanoCores)
	}
	if c.Memory != nil {
		AddUintRawMetric(r, "usageBytes", c.Memory.UsageBytes)
	}

	return r, nil

}

// GroupStatsSummary groups specific data for pods, containers and node
func GroupStatsSummary(statsSummary v1.Summary) (definition.RawGroups, []error) {
	var errs []error
	var rawEntityID string
	g := definition.RawGroups{
		"pod":       {},
		"container": {},
		"node":      {},
	}

	rawNodeData, rawEntityID, err := fetchNodeStats(statsSummary.Node)
	if err != nil {
		errs = append(errs, err)
	} else {
		g["node"][rawEntityID] = rawNodeData
	}

	if statsSummary.Pods == nil {
		errs = append(errs, fmt.Errorf("pods data not found in %s response, fetching pod and container data skipped", StatsSummaryPath))
		return g, errs
	}

PodListLoop:
	for _, pod := range statsSummary.Pods {
		rawPodMetrics, rawEntityID, err := fetchPodStats(pod)
		if err != nil {
			errs = append(errs, err)
			continue PodListLoop
		}
		g["pod"][rawEntityID] = rawPodMetrics

		if pod.Containers == nil {
			errs = append(errs, fmt.Errorf("container data not found in %s response", StatsSummaryPath))
			continue PodListLoop
		}

	ContainerListLoop:
		for _, container := range pod.Containers {
			rawContainerMetrics, err := fetchContainerStats(container)
			if err != nil {
				errs = append(errs, err)
				continue ContainerListLoop
			}
			rawContainerMetrics["podName"] = rawPodMetrics["podName"]
			rawContainerMetrics["namespace"] = rawPodMetrics["namespace"]

			rawEntityID = fmt.Sprintf("%s_%s_%s", rawPodMetrics["namespace"], rawPodMetrics["podName"], rawContainerMetrics["containerName"])

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

// AddUintRawMetric adds a new metric to a RawMetrics if it exists
func AddUintRawMetric(r definition.RawMetrics, name string, valuePtr *uint64) {
	if valuePtr != nil {
		r[name] = *valuePtr
	}
}
