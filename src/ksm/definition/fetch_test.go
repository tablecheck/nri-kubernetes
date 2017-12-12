package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromRawFetchesProperly(t *testing.T) {
	raw := RawMetrics{
		"metric_name_1": "metric_value_1",
		"metric_name_2": "metric_value_2",
		"metric_name_3": "metric_value_3",
		"metric_name_4": "metric_value_4",
		"metric_name_5": "metric_value_5",
	}

	v, err := FromRaw("metric_name_2")(raw)
	assert.NoError(t, err)
	assert.Equal(t, "metric_value_2", v)
}

func TestFromRawErrorsOnNotFound(t *testing.T) {
	raw := RawMetrics{
		"metric_name_1": "metric_value_1",
	}

	v, err := FromRaw("nonexisting")(raw)
	assert.Error(t, err)
	assert.Nil(t, v)
}
