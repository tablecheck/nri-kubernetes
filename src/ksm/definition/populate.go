package definition

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

// RawGroups are grouped raw metrics.
type RawGroups map[string]map[string]RawMetrics

// GuessFunc guesses from data.
type GuessFunc func(groupLabel, entityID string, groups RawGroups) string

// FromGroupMetricSetTypeGuessFunc uses the groupLabel for creating the metric set type sample.
func FromGroupMetricSetTypeGuessFunc(groupLabel, _ string, _ RawGroups) string {
	return fmt.Sprintf("%vSample", strings.Title(groupLabel))
}

// FromGroupMetricSetEntitTypeGuessFunc uses the grouplabel as guess for the entity type.
func FromGroupMetricSetEntitTypeGuessFunc(groupLabel, _ string, _ RawGroups) string {
	return fmt.Sprintf("%v", groupLabel)
}

// PopulateFunc populates raw metric groups using your specs
type PopulateFunc func(RawGroups, Specs) (bool, []error)

// IntegrationProtocol2PopulateFunc populates an integration protocol v2 with the given metrics and definition.
func IntegrationProtocol2PopulateFunc(i *sdk.IntegrationProtocol2, msTypeGuesser, msEntityTypeGuesser GuessFunc) PopulateFunc {
	return func(groups RawGroups, specs Specs) (bool, []error) {
		var populated bool
		var errs []error
		for groupLabel, entities := range groups {
			for entityID := range entities {
				e, err := i.Entity(entityID, msEntityTypeGuesser(groupLabel, entityID, groups))
				if err != nil {
					errs = append(errs, err)
					continue
				}

				ms := metric.NewMetricSet(msTypeGuesser(groupLabel, entityID, groups))
				wasPopulated, populateErrs := metricSetPopulateFunc(ms, groupLabel, entityID)(groups, specs)
				if len(populateErrs) != 0 {
					for _, err := range populateErrs {
						errs = append(errs, fmt.Errorf("entity id: %s: %s", entityID, err))
					}
				}

				if wasPopulated {
					e.Metrics = append(e.Metrics, ms)
					populated = true
				}
			}
		}

		return populated, errs
	}
}

func metricSetPopulateFunc(ms metric.MetricSet, groupLabel, entityID string) PopulateFunc {
	return func(groups RawGroups, specs Specs) (populated bool, errs []error) {
		for _, ex := range specs[groupLabel] {
			val, err := ex.ValueFunc(groupLabel, entityID, groups)
			if err != nil {
				errs = append(errs, fmt.Errorf("error fetching value for metric %s. Error: %s", ex.Name, err))
				continue
			}

			if multiple, ok := val.(FetchedValues); ok {
				for k, v := range multiple {
					err := ms.SetMetric(k, v, ex.Type)
					if err != nil {
						errs = append(errs, fmt.Errorf("error setting metric %s with value %v in metric set. Error: %s", k, v, err))
						continue
					}

					populated = true
				}
			} else {
				err := ms.SetMetric(ex.Name, val, ex.Type)
				if err != nil {
					errs = append(errs, fmt.Errorf("error setting metric %s with value %v in metric set. Error: %s", ex.Name, val, err))
					continue
				}

				populated = true
			}
		}

		return
	}
}
