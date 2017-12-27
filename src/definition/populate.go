package definition

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

// GuessFunc guesses from data.
type GuessFunc func(clusterName, groupLabel, entityID string, groups RawGroups) (string, error)

// FromGroupMetricSetTypeGuessFunc uses the groupLabel for creating the metric set type sample.
func FromGroupMetricSetTypeGuessFunc(_, groupLabel, _ string, _ RawGroups) (string, error) {
	return fmt.Sprintf("%vSample", strings.Title(groupLabel)), nil
}

// FromGroupMetricSetEntitTypeGuessFunc uses the grouplabel as guess for the entity type.
func FromGroupMetricSetEntitTypeGuessFunc(_, groupLabel, _ string, _ RawGroups) (string, error) {
	return fmt.Sprintf("%v", groupLabel), nil
}

// PopulateFunc populates raw metric groups using your specs
type PopulateFunc func(RawGroups, SpecGroups) (bool, []error)

// MetricSetManipulator manipulates the MetricSet for a given entity and clusterName
type MetricSetManipulator func(ms metric.MetricSet, entity sdk.Entity, clusterName string)

// IntegrationProtocol2PopulateFunc populates an integration protocol v2 with the given metrics and definition.
func IntegrationProtocol2PopulateFunc(i *sdk.IntegrationProtocol2, clusterName string, msTypeGuesser, msEntityTypeGuesser GuessFunc, msManipulators ...MetricSetManipulator) PopulateFunc {
	return func(groups RawGroups, specs SpecGroups) (bool, []error) {
		var populated bool
		var errs []error
		for groupLabel, entities := range groups {
			for entityID := range entities {

				// Only populate specified groups.
				if _, ok := specs[groupLabel]; !ok {
					continue
				}

				msEntityID := entityID
				if generator := specs[groupLabel].IDGenerator; generator != nil {
					generatedEntityID, err := generator(groupLabel, entityID, groups)
					if err != nil {
						errs = append(errs, fmt.Errorf("error generating entity ID for: %s: %s", entityID, err))
					}

					msEntityID = generatedEntityID
				}

				msEntityType, err := msEntityTypeGuesser(clusterName, groupLabel, entityID, groups)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				e, err := i.Entity(msEntityID, msEntityType)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				msType, err := msTypeGuesser(clusterName, groupLabel, entityID, groups)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				ms := metric.NewMetricSet(msType)
				for _, m := range msManipulators {
					m(ms, e.Entity, clusterName)
				}

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
	return func(groups RawGroups, specs SpecGroups) (populated bool, errs []error) {
		for _, ex := range specs[groupLabel].Specs {
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
