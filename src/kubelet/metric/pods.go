package metric

import (
	"net/http"
	"strings"

	"fmt"

	"encoding/json"

	"io/ioutil"

	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/client"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/data"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"k8s.io/api/core/v1"
)

// KubeletPodsPath is the path where kubelet serves information about pods.
const KubeletPodsPath = "/pods"

// PodsFetchFunc creates a FetchFunc that fetches data from the kubelet pods path.
func PodsFetchFunc(c client.HTTPClient) data.FetchFunc {
	return func() (definition.RawGroups, error) {
		r, err := c.Do(http.MethodGet, KubeletPodsPath)
		if err != nil {
			return nil, err
		}

		defer func() {
			r.Body.Close() // nolint: errcheck
		}()

		if r.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error calling kubelet %s path. Status code %d", KubeletPodsPath, r.StatusCode)
		}

		rawPods, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response from kubelet %s path. %s", KubeletPodsPath, err)
		}

		if len(rawPods) == 0 {
			return nil, fmt.Errorf("error reading response from kubelet %s path. Response is empty", KubeletPodsPath)
		}

		// v1.PodList comes from k8s api core library.
		var pods v1.PodList
		err = json.Unmarshal(rawPods, &pods)
		if err != nil {
			return nil, fmt.Errorf("error decoding response from kubelet %s path. %s", KubeletPodsPath, err)
		}

		raw := definition.RawGroups{
			"pod":       make(map[string]definition.RawMetrics),
			"container": make(map[string]definition.RawMetrics),
		}

		// If missing, we get the nodeIP from any other container in the node.
		// Due to Kubelet "Wrong Pending status" bug. See https://github.com/kubernetes/kubernetes/pull/57106
		var missingNodeIPContainerIDs []string
		var missingNodeIPPodIDs []string
		var nodeIP string

		for _, p := range pods.Items {
			podData, id := fetchPodData(&p)
			raw["pod"][id] = podData

			if _, ok := podData["nodeIP"]; ok && nodeIP == "" {
				nodeIP = podData["nodeIP"].(string)
			}

			if nodeIP == "" {
				missingNodeIPPodIDs = append(missingNodeIPPodIDs, id)
			} else {
				raw["pod"][id]["nodeIP"] = nodeIP
			}

			containers := fetchContainersData(&p)
			for id, c := range containers {
				raw["container"][id] = c

				if _, ok := c["nodeIP"]; ok && nodeIP == "" {
					nodeIP = c["nodeIP"].(string)
				}

				if nodeIP == "" {
					missingNodeIPContainerIDs = append(missingNodeIPContainerIDs, id)
				} else {
					raw["container"][id]["nodeIP"] = nodeIP
				}
			}
		}

		for _, id := range missingNodeIPPodIDs {
			raw["pod"][id]["nodeIP"] = nodeIP
		}

		for _, id := range missingNodeIPContainerIDs {
			raw["container"][id]["nodeIP"] = nodeIP
		}

		return raw, nil
	}
}

// TODO handle errors and missing data
func fetchContainersData(p *v1.Pod) map[string]definition.RawMetrics {
	// ContainerStatuses is sometimes missing.
	status := make(map[string]definition.RawMetrics)
	for _, c := range p.Status.ContainerStatuses {
		id := fmt.Sprintf("%v_%v_%v", p.GetObjectMeta().GetNamespace(), p.GetObjectMeta().GetName(), c.Name)

		status[id] = make(definition.RawMetrics)

		switch {
		case c.State.Running != nil:
			status[id]["status"] = "Running"
			status[id]["startedAt"] = c.State.Running.StartedAt.Time.In(time.UTC)
			status[id]["containerID"] = c.ContainerID
			status[id]["containerImageID"] = c.ImageID
			status[id]["restartCount"] = c.RestartCount
			status[id]["isReady"] = c.Ready
		case c.State.Waiting != nil:
			status[id]["status"] = "Waiting"
			status[id]["reason"] = c.State.Waiting.Reason
		case c.State.Terminated != nil:
			status[id]["containerID"] = c.State.Terminated.ContainerID
			status[id]["status"] = "Terminated"
			status[id]["startedAt"] = c.State.Terminated.StartedAt.Time.In(time.UTC)
		default:
			status[id]["status"] = "Unknown"
		}
	}

	specs := make(map[string]definition.RawMetrics)
	for _, c := range p.Spec.Containers {
		id := fmt.Sprintf("%v_%v_%v", p.GetObjectMeta().GetNamespace(), p.GetObjectMeta().GetName(), c.Name)

		specs[id] = definition.RawMetrics{
			"containerName":  c.Name,
			"containerImage": c.Image,
			"namespace":      p.GetObjectMeta().GetNamespace(),
			"podName":        p.GetObjectMeta().GetName(),
			"nodeName":       p.Spec.NodeName,
		}

		if v := p.Status.HostIP; v != "" {
			specs[id]["nodeIP"] = v
		}

		if v, ok := c.Resources.Requests[v1.ResourceCPU]; ok {
			specs[id]["cpuRequestedCores"] = v.MilliValue()
		}

		if v, ok := c.Resources.Limits[v1.ResourceCPU]; ok {
			specs[id]["cpuLimitCores"] = v.MilliValue()
		}

		if v, ok := c.Resources.Requests[v1.ResourceMemory]; ok {
			specs[id]["memoryRequestedBytes"] = v.Value()
		}

		if v, ok := c.Resources.Limits[v1.ResourceMemory]; ok {
			specs[id]["memoryLimitBytes"] = v.Value()
		}

		// TODO get from already fetched pod data
		if ref := p.GetOwnerReferences(); len(ref) > 0 {
			specs[id]["deploymentName"] = deploymentNameBasedOnCreator(ref[0].Kind, ref[0].Name)
		}

		// Assuming that the container is running. See https://github.com/kubernetes/kubernetes/pull/57106
		if _, ok := status[id]; !ok {
			specs[id]["status"] = "Running"
		}

		// merging status data
		for k, v := range status[id] {
			specs[id][k] = v
		}

	}

	return specs
}

// TODO handle errors and missing data
func fetchPodData(p *v1.Pod) (definition.RawMetrics, string) {
	var isReady, isScheduled bool
	for _, p := range p.Status.Conditions {
		switch p.Type {
		case "Ready":
			isReady = true
		case "PodScheduled":
			isScheduled = true
		}
	}

	r := definition.RawMetrics{
		"namespace":   p.GetObjectMeta().GetNamespace(),
		"podName":     p.GetObjectMeta().GetName(),
		"nodeName":    p.Spec.NodeName,
		"isReady":     isReady,
		"isScheduled": isScheduled,
	}

	if v := p.Status.HostIP; v != "" {
		r["nodeIP"] = v
	}

	switch p.Status.Phase {
	case v1.PodPending, v1.PodRunning:
		r["status"] = string(v1.PodRunning)
	default:
		r["status"] = string(p.Status.Phase)
	}

	if t := p.GetObjectMeta().GetCreationTimestamp(); !t.IsZero() {
		r["createdAt"] = t.In(time.UTC)
	}

	if ref := p.GetOwnerReferences(); len(ref) > 0 {
		r["createdKind"] = ref[0].Kind
		r["createdBy"] = ref[0].Name
		r["deploymentName"] = deploymentNameBasedOnCreator(ref[0].Kind, ref[0].Name)
	}

	if p.Status.StartTime != nil {
		r["startTime"] = p.Status.StartTime.Time.In(time.UTC)
	}

	if l := len(p.GetObjectMeta().GetLabels()); l > 0 {
		r["labels"] = make(map[string]string, l)
		for k, v := range p.GetObjectMeta().GetLabels() {
			r["labels"].(map[string]string)[k] = v
		}
	}

	rawEntityID := fmt.Sprintf("%v_%v", p.GetObjectMeta().GetNamespace(), p.GetObjectMeta().GetName())

	return r, rawEntityID
}

func deploymentNameBasedOnCreator(creatorKind, creatorName string) string {
	var deploymentName string
	if creatorKind == "ReplicaSet" {
		deploymentName = replicasetNameToDeploymentName(creatorName)
	}
	return deploymentName
}

func replicasetNameToDeploymentName(rsName string) string {
	s := strings.Split(rsName, "-")
	return strings.Join(s[:len(s)-1], "-")
}

// OneMetricPerLabel transforms a map of labels to FetchedValues type,
// which will be converted later to one metric per label.
// It also prefix the labels with 'label.'
func OneMetricPerLabel(rawLabels definition.FetchedValue) definition.FetchedValue {
	labels, ok := rawLabels.(map[string]string)
	if !ok {
		return rawLabels
	}

	modified := make(definition.FetchedValues)
	for k, v := range labels {
		modified[fmt.Sprintf("label.%v", k)] = v
	}

	return modified
}
