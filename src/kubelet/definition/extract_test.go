package definition

import (
	"errors"
	"testing"

	sdk "github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/stretchr/testify/assert"
)

var metricGroup = MetricGroups{
	"entity_type_1": {
		"entity_id_1": RawMetrics{
			"metric_name_1": 1,
			"metric_name_2": "metric_value_2",
			"metric_name_3": map[string]interface{}{
				"foo": "bar",
			},
		},
		"entity_id_2": RawMetrics{
			"metric_name_1": 2,
			"metric_name_2": "metric_value_4",
			"metric_name_3": map[string]interface{}{
				"foo": "bar",
			},
		},
	},
}

var raw = RawMetrics{
	"metric_name_1": 1,
	"metric_name_2": "metric_value_2",
	"metric_name_3": map[string]interface{}{
		"foo_1": "bar_1",
		"foo_2": "bar_2",
	},
}

var def = []Metric{
	{"onHost_Name_1", FromRaw("metric_name_1"), sdk.GAUGE},
	{"onHost_Name_2", FromRaw("metric_name_2"), sdk.ATTRIBUTE},
	{
		"onHost_Name_3",
		fromMultiple(
			FetchedValues(
				map[string]FetchedValue{
					"foo_1": "bar_1",
					"foo_2": "bar_2",
				},
			),
		),
		sdk.ATTRIBUTE,
	},
}

func fromMultiple(values FetchedValues) FetchFunc {
	return func(raw RawMetrics) (FetchedValue, error) {
		return values, nil
	}
}

// --------------- OneMetricSetExtract ---------------

func TestOneMetricSetExtract_CorrectValues(t *testing.T) {
	expectedMetricSets := []sdk.MetricSet{
		{
			"onHost_Name_1": 1,
			"onHost_Name_2": "metric_value_2",
			"foo_1":         "bar_1",
			"foo_2":         "bar_2",
		},
	}

	metricSets, errs := OneMetricSetExtract(raw)(def)
	assert.Equal(t, expectedMetricSets, metricSets)
	assert.Empty(t, errs)
}

func TestOneMetricSetExtract_ErrorFromFetchedFunc(t *testing.T) {
	var def = []Metric{
		{"onHost_Name_4", FromRaw("metric_name_4"), sdk.GAUGE}, // raw metric not found
	}

	metricSets, errs := OneMetricSetExtract(raw)(def)

	assert.Empty(t, metricSets)
	assert.Equal(t, 1, len(errs))
	assert.Contains(t, errs, errors.New("error fetching value for metric onHost_Name_4. Error: raw metric not found with key metric_name_4"))
}

func TestOneMetricSetExtract_ErrorSettingValue(t *testing.T) {
	var def = []Metric{
		{"onHost_Name_1", FromRaw("metric_name_1"), sdk.ATTRIBUTE}, // source type invalid
		{"onHost_Name_2", FromRaw("metric_name_2"), sdk.GAUGE},     // source type invalid
	}

	metricSets, errs := OneMetricSetExtract(raw)(def)

	assert.Empty(t, metricSets)
	assert.Equal(t, 2, len(errs))
	assert.Contains(t, errs, errors.New("error setting metric onHost_Name_1 with value 1 in metric set. Error: Invalid data type for attribute onHost_Name_1"))
	assert.Contains(t, errs, errors.New("error setting metric onHost_Name_2 with value metric_value_2 in metric set. Error: Invalid (non-numeric) data type for metric onHost_Name_2"))
}

func TestOneMetricSetExtract_ErrorSettingMultipleValue(t *testing.T) {
	var def = []Metric{
		{
			"onHost_Name_3",
			fromMultiple(
				FetchedValues(
					map[string]FetchedValue{
						"foo_1": "bar_1", // source type invalid
						"foo_2": 2,
					},
				),
			),
			sdk.GAUGE,
		},
	}
	expectedMetricSets := []sdk.MetricSet{
		{
			"foo_2": 2,
		},
	}

	metricSets, errs := OneMetricSetExtract(raw)(def)
	assert.Equal(t, expectedMetricSets, metricSets)
	assert.Equal(t, 1, len(errs))
	assert.Contains(t, errs, errors.New("error setting metric foo_1 with value bar_1 in metric set. Error: Invalid (non-numeric) data type for metric foo_1"))

}

// --------------- OneMetricSetPerGroup ---------------
func TestOneMetricSetPerGroup_CorrectValues(t *testing.T) {
	expectedMetricSets := []sdk.MetricSet{
		{
			"onHost_Name_1": 1,
			"onHost_Name_2": "metric_value_2",
			"foo_1":         "bar_1",
			"foo_2":         "bar_2",
		},
		{
			"onHost_Name_1": 2,
			"onHost_Name_2": "metric_value_4",
			"foo_1":         "bar_1",
			"foo_2":         "bar_2",
		},
	}
	metricSets, errs := OneMetricSetPerGroup(metricGroup)(def)
	assert.Equal(t, 2, len(metricSets))
	assert.Contains(t, metricSets, expectedMetricSets[0])
	assert.Contains(t, metricSets, expectedMetricSets[1])
	assert.Empty(t, errs)
}

func TestOneMetricSetPerGroup_ErrorsFromOneMetricSetExtractHavingNoMetricSet(t *testing.T) {
	var def = []Metric{
		{"onHost_Name_1", FromRaw("metric_name_1"), sdk.ATTRIBUTE}, // source type invalid
		{"onHost_Name_2", FromRaw("metric_name_2"), sdk.GAUGE},     // source type invalid
	}

	metricSets, errs := OneMetricSetPerGroup(metricGroup)(def)

	assert.Empty(t, metricSets)
	assert.Equal(t, 4, len(errs))
	assert.Contains(t, errs, errors.New("entity id: entity_id_1: error setting metric onHost_Name_1 with value 1 in metric set. Error: Invalid data type for attribute onHost_Name_1"))
	assert.Contains(t, errs, errors.New("entity id: entity_id_1: error setting metric onHost_Name_2 with value metric_value_2 in metric set. Error: Invalid (non-numeric) data type for metric onHost_Name_2"))
	assert.Contains(t, errs, errors.New("entity id: entity_id_2: error setting metric onHost_Name_1 with value 2 in metric set. Error: Invalid data type for attribute onHost_Name_1"))
	assert.Contains(t, errs, errors.New("entity id: entity_id_2: error setting metric onHost_Name_2 with value metric_value_4 in metric set. Error: Invalid (non-numeric) data type for metric onHost_Name_2"))
}

func TestOneMetricSetPerGroup_ErrorsFromOneMetricSetExtractHavingPartialMetricSet(t *testing.T) {
	var def = []Metric{
		{"onHost_Name_1", FromRaw("metric_name_1"), sdk.ATTRIBUTE}, // source type invalid
		{"onHost_Name_2", FromRaw("metric_name_2"), sdk.ATTRIBUTE},
	}

	metricSets, errs := OneMetricSetPerGroup(metricGroup)(def)

	assert.Len(t, metricSets, 2) // We have partial results (onHost_Name_2 metrics)
	assert.Equal(t, 2, len(errs))
	assert.Contains(t, errs, errors.New("entity id: entity_id_1: error setting metric onHost_Name_1 with value 1 in metric set. Error: Invalid data type for attribute onHost_Name_1"))
	assert.Contains(t, errs, errors.New("entity id: entity_id_2: error setting metric onHost_Name_1 with value 2 in metric set. Error: Invalid data type for attribute onHost_Name_1"))
}
