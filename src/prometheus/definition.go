package prometheus

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
)

// FromPrometheusLabelValueEntityTypeGenerator generates the entity type using the value of the specified label
// for the given metric key. If group label is different than "namespace" or "node", then entity type
// is composed of group label and specified label value (in case of error fetching the label,
// default value is used). Otherwise entity type is the same as group label.
func FromPrometheusLabelValueEntityTypeGenerator(key, label, defaultValue string) definition.EntityTypeGeneratorFunc {
	return func(groupLabel string, rawEntityID string, g definition.RawGroups, clusterName string) (string, error) {
		var actualGroupLabel string
		switch groupLabel {
		case "namespace", "node":
			return fmt.Sprintf("k8s:%s:%s", clusterName, groupLabel), nil
		case "container":
			actualGroupLabel = "pod"
		default:
			actualGroupLabel = groupLabel
		}

		v, err := FromPrometheusLabelValue(key, label)(groupLabel, rawEntityID, g)
		if err != nil {
			return fmt.Sprintf("k8s:%s:%s:%s", clusterName, defaultValue, actualGroupLabel), fmt.Errorf("error fetching %s for %q: %v", label, groupLabel, err.Error())
		}
		if v == nil {
			return fmt.Sprintf("k8s:%s:%s:%s", clusterName, defaultValue, actualGroupLabel), fmt.Errorf("%s not found for %q", label, groupLabel)

		}

		val, ok := v.(string)
		if !ok {
			return fmt.Sprintf("k8s:%s:%s:%s", clusterName, defaultValue, actualGroupLabel), fmt.Errorf("incorrect type of %s for %q", label, groupLabel)
		}

		return fmt.Sprintf("k8s:%s:%s:%s", clusterName, val, actualGroupLabel), nil
	}
}

// FromPrometheusLabelValueEntityIDGenerator generates an entityID using the value of the specified label
// for the given metric key.
func FromPrometheusLabelValueEntityIDGenerator(key, label string) definition.EntityIDGeneratorFunc {
	return func(groupLabel string, rawEntityID string, g definition.RawGroups) (string, error) {
		v, err := FromPrometheusLabelValue(key, label)(groupLabel, rawEntityID, g)
		if err != nil {

			return "", fmt.Errorf("error fetching %q: %v", label, err)
		}

		if v == nil {
			return "", fmt.Errorf("incorrect value of fetched data for %q", key)
		}

		val, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("incorrect type of fetched data for %q", key)
		}

		return val, err
	}
}

// GroupPrometheusMetricsBySpec groups metrics coming from Prometheus by a given metric spec.
// Example: grouping by K8s pod, container, etc.
func GroupPrometheusMetricsBySpec(specs definition.SpecGroups, families []MetricFamily) (g definition.RawGroups, errs []error) {
	g = make(definition.RawGroups)
	for groupLabel := range specs {
		for _, f := range families {
			for _, m := range f.Metrics {
				if !m.Labels.Has(groupLabel) {
					continue
				}

				var rawEntityID string
				switch groupLabel {
				case "namespace", "node":
					rawEntityID = m.Labels[groupLabel]
				case "container":
					rawEntityID = fmt.Sprintf("%v_%v_%v", m.Labels["namespace"], m.Labels["pod"], m.Labels[groupLabel])
				default:
					rawEntityID = fmt.Sprintf("%v_%v", m.Labels["namespace"], m.Labels[groupLabel])
				}

				if _, ok := g[groupLabel]; !ok {
					g[groupLabel] = make(map[string]definition.RawMetrics)
				}

				if _, ok := g[groupLabel][rawEntityID]; !ok {
					g[groupLabel][rawEntityID] = make(definition.RawMetrics)
				}

				g[groupLabel][rawEntityID][f.Name] = m
			}
		}

		if len(g[groupLabel]) == 0 {
			errs = append(errs, fmt.Errorf("no data found for %s object", groupLabel))
			continue
		}
	}

	return g, errs
}

// FromPrometheusValue creates a FetchFunc that fetches values from prometheus metrics values.
func FromPrometheusValue(key string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: Metric. Got: %T", value)
		}

		return v.Value, nil
	}
}

// FromPrometheusLabelValue creates a FetchFunc that fetches values from prometheus metrics labels.
func FromPrometheusLabelValue(key, label string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: Metric. Got: %T", value)
		}

		l, ok := v.Labels[label]
		if !ok {
			return nil, fmt.Errorf("label '%v' not found in prometheus metric", label)
		}

		return l, nil
	}
}

func getRandomPrometheusMetric(metrics definition.RawMetrics) (metricKey string, value definition.RawValue) {
	for metricKey, value = range metrics {
		if _, ok := value.(Metric); !ok {
			continue
		}
		// We just want 1.
		break
	}

	return
}

func fetchPrometheusMetric(metricKey string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {

		value, err := definition.FromRaw(metricKey)(groupLabel, entityID, groups)
		if err != nil {
			return nil, err
		}

		v, ok := value.(Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: Metric. Got: %T", value)
		}

		return v, nil
	}
}

// InheritSpecificPrometheusLabelValuesFrom gets the specified label values from a related metric.
// Related metric means any metric you can get with the info that you have in your own metric.
func InheritSpecificPrometheusLabelValuesFrom(parentGroupLabel, relatedMetricKey string, labelsToRetrieve map[string]string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		rawEntityID, err := getRawEntityID(parentGroupLabel, groupLabel, entityID, groups)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve the entity ID of metrics to inherit value from, got error: %v", err)
		}
		parent, err := definition.FromRaw(relatedMetricKey)(parentGroupLabel, rawEntityID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, parentGroupLabel, rawEntityID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(Metric).Labels {
			for n, l := range labelsToRetrieve {
				if l == k {
					multiple[n] = v
				}
			}
		}

		return multiple, nil
	}
}

// InheritAllPrometheusLabelsFrom gets all the label values from from a related metric.
// Related metric means any metric you can get with the info that you have in your own metric.
func InheritAllPrometheusLabelsFrom(parentGroupLabel, relatedMetricKey string) definition.FetchFunc {
	return func(groupLabel, entityID string, groups definition.RawGroups) (definition.FetchedValue, error) {
		rawEntityID, err := getRawEntityID(parentGroupLabel, groupLabel, entityID, groups)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve the entity ID of metrics to inherit labels from, got error: %v", err)
		}

		parent, err := fetchPrometheusMetric(relatedMetricKey)(parentGroupLabel, rawEntityID, groups)
		if err != nil {
			return nil, fmt.Errorf("related metric not found. Metric: %s %s:%s", relatedMetricKey, parentGroupLabel, rawEntityID)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range parent.(Metric).Labels {
			multiple[fmt.Sprintf("label.%v", strings.TrimPrefix(k, "label_"))] = v
		}

		return multiple, nil
	}
}

func getRawEntityID(parentGroupLabel, groupLabel, entityID string, groups definition.RawGroups) (string, error) {
	group, ok := groups[groupLabel][entityID]
	if !ok {
		return "", fmt.Errorf("metrics not found for %v with entity ID: %v", groupLabel, entityID)
	}
	metricKey, r := getRandomPrometheusMetric(group)
	m, ok := r.(Metric)

	if !ok {
		return "", fmt.Errorf("incompatible metric type. Expected: Metric. Got: %T", r)
	}

	var rawEntityID string
	switch parentGroupLabel {
	case "node", "namespace":
		rawEntityID, ok = m.Labels[parentGroupLabel]
		if !ok {
			return "", fmt.Errorf("label not found. Label: '%s', Metric: %s", parentGroupLabel, metricKey)
		}
	default:
		namespaceID, ok := m.Labels["namespace"]
		if !ok {
			return "", fmt.Errorf("label not found. Label: 'namespace', Metric: %s", metricKey)
		}
		relatedMetricID, ok := m.Labels[parentGroupLabel]
		if !ok {
			return "", fmt.Errorf("label not found. Label: %s, Metric: %s", parentGroupLabel, metricKey)
		}
		rawEntityID = fmt.Sprintf("%v_%v", namespaceID, relatedMetricID)
	}
	return rawEntityID, nil
}
