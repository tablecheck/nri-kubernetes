package metric

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestFromNano(t *testing.T) {
	v, err := fromNano(uint64(123456789))
	assert.Equal(t, 0.123456789, v)
	assert.NoError(t, err)

	v, err = fromNano(123456789)
	assert.Nil(t, v)
	assert.Error(t, err)

	v, err = fromNano("not-valid")
	assert.Nil(t, v)
	assert.Error(t, err)
}

func TestToTimestap(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2018-02-14T16:26:33Z")
	v, err := toTimestamp(t1)
	assert.Equal(t, int64(1518625593), v)
	assert.NoError(t, err)

	t2, _ := time.Parse(time.RFC3339, "2016-10-21T00:45:12Z")
	v, err = toTimestamp(t2)
	assert.Equal(t, int64(1477010712), v)
	assert.NoError(t, err)
}

func TestToStringBoolean(t *testing.T) {
	v, err := toStringBoolean(1)
	assert.Equal(t, "true", v)
	assert.NoError(t, err)

	v, err = toStringBoolean(0)
	assert.Equal(t, "false", v)
	assert.NoError(t, err)

	v, err = toStringBoolean(true)
	assert.Equal(t, "true", v)
	assert.NoError(t, err)

	v, err = toStringBoolean(false)
	assert.Equal(t, "false", v)
	assert.NoError(t, err)
}

func TestToCores(t *testing.T) {
	v, err := toCores(100)
	assert.Equal(t, float64(0.1), v)
	assert.NoError(t, err)

	v, err = toCores(int64(1000))
	assert.Equal(t, float64(1), v)
	assert.NoError(t, err)
}
