package definition

import "fmt"

// RawValue is just any value from a raw metric.
type RawValue interface{}

// RawMetrics is a map of RawValue indexed by metric name.
type RawMetrics map[string]RawValue

// FetchedValue is just any value from an already fetched metric.
type FetchedValue interface{}

// FetchedValues is a map of FetchedValue indexed by metric name.
type FetchedValues map[string]FetchedValue

// FetchFunc fetches values or values from raw metrics.
// Return FetchedValues if you want to prototype metrics.
type FetchFunc func(raw RawMetrics) (FetchedValue, error)

// FromRaw fetches metrics from raw metrics. Is the most simple use case.
func FromRaw(key string) FetchFunc {
	return func(raw RawMetrics) (FetchedValue, error) {
		value, ok := raw[key]
		if !ok {
			return nil, fmt.Errorf("raw metric not found with key %v", key)
		}

		return value, nil
	}
}
