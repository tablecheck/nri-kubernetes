package metric

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestFromNano(t *testing.T) {
	assert.Equal(t, 0.123456789, fromNano(uint64(123456789)))
	assert.Equal(t, 123456789, fromNano(123456789))
	assert.Equal(t, "not-valid", fromNano("not-valid"))
}

func TestToTimestap(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2018-02-14T16:26:33Z")
	t2, _ := time.Parse(time.RFC3339, "2016-10-21T00:45:12Z")
	assert.Equal(t, int64(1518625593), toTimestamp(t1))
	assert.Equal(t, int64(1477010712), toTimestamp(t2))
}

func TestToStringBoolean(t *testing.T) {
	assert.Equal(t, "true", toStringBoolean(1))
	assert.Equal(t, "false", toStringBoolean(0))
	assert.Equal(t, "true", toStringBoolean(true))
	assert.Equal(t, "false", toStringBoolean(false))
}

func TestToCores(t *testing.T) {
	assert.Equal(t, float64(0.1), toCores(100))
	assert.Equal(t, float64(1), toCores(int64(1000)))
}
