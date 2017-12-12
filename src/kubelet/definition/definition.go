package definition

import (
	sdk "github.com/newrelic/infra-integrations-sdk/metric"
)

// Metric means the definition of a metric, aka expectations.
type Metric struct {
	Name      string
	ValueFunc FetchFunc
	Type      sdk.SourceType
}

// Aggregation represents lists of definition metrics indexed by identity source name.
// It is the main definition model used for defining all the expected metrics.
type Aggregation map[string][]Metric
