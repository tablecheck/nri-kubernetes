package definition

import (
	"errors"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"

	"fmt"

	"github.com/stretchr/testify/assert"
)

var rawGroupsSample = RawGroups{
	"test": {
		"entity_id_1": RawMetrics{
			"raw_metric_name_1": 1,
			"raw_metric_name_2": "metric_value_2",
			"raw_metric_name_3": map[string]interface{}{
				"foo": "bar",
			},
		},
		"entity_id_2": RawMetrics{
			"raw_metric_name_1": 2,
			"raw_metric_name_2": "metric_value_4",
			"raw_metric_name_3": map[string]interface{}{
				"foo": "bar",
			},
		},
	},
}

var specs = SpecGroups{
	"test": SpecGroup{
		Specs: []Spec{

			{"metric_1", FromRaw("raw_metric_name_1"), metric.GAUGE},
			{"metric_2", FromRaw("raw_metric_name_2"), metric.ATTRIBUTE},
			{
				"metric_3",
				fromMultiple(
					FetchedValues(
						map[string]FetchedValue{
							"multiple_1": "one",
							"multiple_2": "two",
						},
					),
				),
				metric.ATTRIBUTE,
			},
		},
	},
}

func fromMultiple(values FetchedValues) FetchFunc {
	return func(groupLabel, entityID string, groups RawGroups) (FetchedValue, error) {
		return values, nil
	}
}

func TestIntegrationProtocol2PopulateFunc_CorrectValue(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("entity_id_1", "test")
	if err != nil {
		t.Fatal()
	}

	expectedMetricSet1 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   1,
		"metric_2":   "metric_value_2",
		"multiple_1": "one",
		"multiple_2": "two",
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("entity_id_2", "test")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   2,
		"metric_2":   "metric_value_4",
		"multiple_1": "one",
		"multiple_2": "two",
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(rawGroupsSample, specs)
	assert.True(t, populated)
	assert.Empty(t, errs)

	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestIntegrationProtocol2PopulateFunc_PartialResult(t *testing.T) {
	metricSpecsWithIncompatibleType := SpecGroups{
		"test": SpecGroup{
			Specs: []Spec{
				{"metric_1", FromRaw("raw_metric_name_1"), metric.GAUGE},
				{"metric_2", FromRaw("raw_metric_name_2"), metric.GAUGE}, // Source type not correct
			},
		},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedEntityData1, err := sdk.NewEntityData("entity_id_1", "test")
	if err != nil {
		t.Fatal()
	}

	expectedMetricSet1 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   1,
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("entity_id_2", "test")
	if err != nil {
		t.Fatal()
	}
	expectedMetricSet2 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   2,
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(rawGroupsSample, metricSpecsWithIncompatibleType)
	assert.True(t, populated)
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)

	assert.Len(t, errs, 2)
}

func TestIntegrationProtocol2PopulateFunc_EntitiesDataNotPopulated_EmptyMetricGroups(t *testing.T) {
	var metricGroupEmpty = RawGroups{}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	expectedData := make([]*sdk.EntityData, 0)

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(metricGroupEmpty, specs)
	assert.False(t, populated)
	assert.Nil(t, errs)
	assert.Equal(t, expectedData, integration.Data)
}

func TestIntegrationProtocol2PopulateFunc_EntitiesDataNotPopulated_ErrorSettingEntities(t *testing.T) {
	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	metricGroupEmptyEntityID := RawGroups{
		"test": {
			"": RawMetrics{
				"raw_metric_name_1": 1,
				"raw_metric_name_2": "metric_value_2",
				"raw_metric_name_3": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}
	expectedData := []*sdk.EntityData{}

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(metricGroupEmptyEntityID, specs)
	assert.False(t, populated)
	assert.EqualError(t, errs[0], "entity name and type are required when defining one")
	assert.Equal(t, expectedData, integration.Data)
}

func TestIntegrationProtocol2PopulateFunc_MetricsSetsNotPopulated_OnlyEntity(t *testing.T) {
	var metricSpecsIncorrect = SpecGroups{
		"test": SpecGroup{
			Specs: []Spec{
				{"useless", FromRaw("nonExistentMetric"), metric.GAUGE},
			},
		},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	expectedEntityData1, err := sdk.NewEntityData("entity_id_1", "test")
	if err != nil {
		t.Fatal()
	}
	expectedEntityData2, err := sdk.NewEntityData("entity_id_2", "test")
	if err != nil {
		t.Fatal()
	}

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(rawGroupsSample, metricSpecsIncorrect)
	assert.False(t, populated)
	assert.Len(t, errs, 2)

	assert.Contains(t, errs, errors.New("entity id: entity_id_1: error fetching value for metric useless. Error: FromRaw: metric not found. SpecGroup: test, EntityID: entity_id_1, Metric: nonExistentMetric"))
	assert.Contains(t, errs, errors.New("entity id: entity_id_2: error fetching value for metric useless. Error: FromRaw: metric not found. SpecGroup: test, EntityID: entity_id_2, Metric: nonExistentMetric"))
	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}

func TestIntegrationProtocol2PopulateFunc_EntityIDGenerator(t *testing.T) {

	generator := func(groupLabel, rawEntityID string, g RawGroups) (string, error) {
		return fmt.Sprintf("%v-generated", rawEntityID), nil
	}

	withGeneratorSpec := SpecGroups{
		"test": SpecGroup{
			IDGenerator: generator,
			Specs: []Spec{
				{"metric_1", FromRaw("raw_metric_name_1"), metric.GAUGE},
				{"metric_2", FromRaw("raw_metric_name_2"), metric.GAUGE},
			},
		},
	}

	integration, err := sdk.NewIntegrationProtocol2("nr.test", "1.0.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	raw := RawGroups{
		"test": {
			"testEntity1": {
				"raw_metric_name_1": 1,
				"raw_metric_name_2": 2,
			},
			"testEntity2": {
				"raw_metric_name_1": 3,
				"raw_metric_name_2": 4,
			},
		},
	}

	expectedEntityData1, err := sdk.NewEntityData("testEntity1-generated", "test")
	if err != nil {
		t.Fatal()
	}

	expectedMetricSet1 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   1,
		"metric_2":   2,
	}
	expectedEntityData1.Metrics = []metric.MetricSet{expectedMetricSet1}

	expectedEntityData2, err := sdk.NewEntityData("testEntity2-generated", "test")
	if err != nil {
		t.Fatal()
	}

	expectedMetricSet2 := metric.MetricSet{
		"event_type": "TestSample",
		"metric_1":   3,
		"metric_2":   4,
	}
	expectedEntityData2.Metrics = []metric.MetricSet{expectedMetricSet2}

	populated, errs := IntegrationProtocol2PopulateFunc(integration, FromGroupMetricSetTypeGuessFunc, FromGroupMetricSetEntitTypeGuessFunc)(raw, withGeneratorSpec)

	assert.True(t, populated)
	assert.Empty(t, errs)

	assert.Contains(t, integration.Data, &expectedEntityData1)
	assert.Contains(t, integration.Data, &expectedEntityData2)
}
