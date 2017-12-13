package definition

import (
	sdk "github.com/newrelic/infra-integrations-sdk/metric"
)

// Spec is the metric specification.
type Spec struct {
	Name      string
	ValueFunc FetchFunc
	Type      sdk.SourceType
}

// Specs represents lists of metric specifications indexed by identity source name.
// It is the main definition model used for defining all the expected metrics.
type Specs map[string][]Spec
