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

		for _, p := range pods.Items {
			podData, id := fetchPodData(&p)
			raw["pod"][id] = podData

			containers := fetchContainersData(&p)
			for id, c := range containers {
				raw["container"][id] = c
			}
		}

		return raw, nil
	}
}

// TODO handle errors and missing data
func fetchContainersData(p *v1.Pod) map[string]definition.RawMetrics {
	specs := make(map[string]definition.RawMetrics)
	for _, c := range p.Spec.Containers {
		id := fmt.Sprintf("%v_%v_%v", p.GetObjectMeta().GetNamespace(), p.GetObjectMeta().GetName(), c.Name)
		specs[id] = make(definition.RawMetrics)

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
	}

	// ContainerStatuses slice contains all the containers
	r := make(map[string]definition.RawMetrics)
	for _, c := range p.Status.ContainerStatuses {
		id := fmt.Sprintf("%v_%v_%v", p.GetObjectMeta().GetNamespace(), p.GetObjectMeta().GetName(), c.Name)

		r[id] = definition.RawMetrics{
			"containerName":    c.Name,
			"containerID":      c.ContainerID,
			"containerImage":   c.Image,
			"containerImageID": c.ImageID,
			"namespace":        p.GetObjectMeta().GetNamespace(),
			"podName":          p.GetObjectMeta().GetName(),
			"podIP":            p.Status.PodIP,
			"nodeName":         p.Spec.NodeName,
			"nodeIP":           p.Status.HostIP,
			"restartCount":     c.RestartCount,
			"isReady":          c.Ready,
		}

		if ref := p.GetOwnerReferences(); len(ref) > 0 {
			r[id]["deploymentName"] = deploymentNameBasedOnCreator(ref[0].Kind, ref[0].Name)
		}

		if v, ok := specs[id]["cpuRequestedCores"]; ok {
			r[id]["cpuRequestedCores"] = v
		}

		if v, ok := specs[id]["cpuLimitCores"]; ok {
			r[id]["cpuLimitCores"] = v
		}

		if v, ok := specs[id]["memoryRequestedBytes"]; ok {
			r[id]["memoryRequestedBytes"] = v
		}

		if v, ok := specs[id]["memoryLimitBytes"]; ok {
			r[id]["memoryLimitBytes"] = v
		}

		switch {
		case c.State.Running != nil:
			r[id]["status"] = "Running"
			r[id]["startedAt"] = c.State.Running.StartedAt.Time.In(time.UTC)
		case c.State.Waiting != nil:
			r[id]["status"] = "Waiting"
			r[id]["reason"] = c.State.Waiting.Reason
		case c.State.Terminated != nil:
			r[id]["status"] = "Terminated"
			r[id]["startedAt"] = c.State.Terminated.StartedAt.Time.In(time.UTC)
		default:
			r[id]["status"] = "Unknown"
		}
	}

	return r
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
		"nodeIP":      p.Status.HostIP,
		"namespace":   p.GetObjectMeta().GetNamespace(),
		"podName":     p.GetObjectMeta().GetName(),
		"nodeName":    p.Spec.NodeName,
		"podIP":       p.Status.PodIP,
		"status":      string(p.Status.Phase),
		"isReady":     isReady,
		"isScheduled": isScheduled,
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
