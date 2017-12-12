package metric

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

// Populate populates integration (protocol2) setting Entity and Metrics objects.
// When at least one metric set was populated then true is returned.
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

// GroupPrometheusMetricsByLabel groups metrics coming from Prometheus by an specified prometheus metric label.
// Example: grouping by K8s pod or container.
func GroupPrometheusMetricsByLabel(label string, families []prometheus.MetricFamily) definition.MetricGroups {
	g := make(definition.MetricGroups)

	for _, f := range families {
		for _, m := range f.Metrics {
			if !m.Labels.Has(label) {
				continue
			}

			objectID := m.Labels[label]

			if _, ok := g[label]; !ok {
				g[label] = make(map[string]definition.RawMetrics)
			}

			if _, ok := g[label][objectID]; !ok {
				g[label][objectID] = make(definition.RawMetrics)
			}

			g[label][objectID][f.Name] = m
		}
	}

	return g
}

// FromPrometheusValue creates a FetchFunc that fetches values from prometheus metrics values.
func FromPrometheusValue(key string) definition.FetchFunc {
	return func(raw definition.RawMetrics) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(raw)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		return v.Value, nil
	}
}

// FromPrometheusLabelValue creates a FetchFunc that fetches values from prometheus metrics labels.
func FromPrometheusLabelValue(key, label string) definition.FetchFunc {
	return func(raw definition.RawMetrics) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(raw)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		l, ok := v.Labels[label]
		if !ok {
			return nil, fmt.Errorf("label '%v' not found in raw metrics", label)
		}

		return l, nil
	}
}

// FromPrometheusMultipleLabels creates a FetchFunc that fetches multiple values from the specified metric labels.
// It creates one value per each label found.
func FromPrometheusMultipleLabels(key string) definition.FetchFunc {
	return func(raw definition.RawMetrics) (definition.FetchedValue, error) {
		value, err := definition.FromRaw(key)(raw)
		if err != nil {
			return nil, err
		}

		v, ok := value.(prometheus.Metric)
		if !ok {
			return nil, fmt.Errorf("incompatible metric type. Expected: prometheus.Metric. Got: %T", value)
		}

		multiple := make(definition.FetchedValues)
		for k, v := range v.Labels {
			multiple[fmt.Sprintf("label.%v", k)] = v
		}

		return multiple, nil
	}
}
