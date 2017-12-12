package definition

import (
	"fmt"

	sdk "github.com/newrelic/infra-integrations-sdk/metric"
)

// MetricGroups are grouped raw metrics.
type MetricGroups map[string]map[string]RawMetrics

// MetricSetsExtractFunc extracts MetricSets using definitions.
type MetricSetsExtractFunc func([]Metric) ([]sdk.MetricSet, []error)

// OneMetricSetPerGroup creates a MetricSetsExtractFunc that extracts one MetricSet by each group of raw metrics.
func OneMetricSetPerGroup(groups MetricGroups) MetricSetsExtractFunc {
	return func(definitions []Metric) ([]sdk.MetricSet, []error) {
		var metricSets []sdk.MetricSet
		var errs []error
		for _, entities := range groups {
			for entityID, r := range entities {
				oneMetricSet, err := OneMetricSetExtract(r)(definitions)
				if len(err) != 0 {
					for _, e := range err {
						errs = append(errs, fmt.Errorf("entity id: %s: %s", entityID, e))
					}
				}

				if len(oneMetricSet) > 0 {
					metricSets = append(metricSets, oneMetricSet[0])
				}
			}
		}

		return metricSets, errs
	}
}

// OneMetricSetExtract creates a MetricSetsExtractFunc that extracts just one MetricSet from raw metrics		.
func OneMetricSetExtract(raw RawMetrics) MetricSetsExtractFunc {
	return func(definitions []Metric) ([]sdk.MetricSet, []error) {
		ms := make(sdk.MetricSet)
		var errs []error
		for _, d := range definitions {
			val, err := d.ValueFunc(raw)
			if err != nil {
				errs = append(errs, fmt.Errorf("error fetching value for metric %s. Error: %s", d.Name, err))
				continue
			}

			if multiple, ok := val.(FetchedValues); ok {
				for k, v := range multiple {
					err := ms.SetMetric(k, v, d.Type)
					if err != nil {
						errs = append(errs, fmt.Errorf("error setting metric %s with value %v in metric set. Error: %s", k, v, err))
						continue
					}
				}
			} else {
				err := ms.SetMetric(d.Name, val, d.Type)
				if err != nil {
					errs = append(errs, fmt.Errorf("error setting metric %s with value %v in metric set. Error: %s", d.Name, val, err))
					continue
				}
			}

		}

		metricSets := make([]sdk.MetricSet, 0)
		if len(ms) > 0 {
			metricSets = append(metricSets, ms)
		}

		return metricSets, errs
	}
}
