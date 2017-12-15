package definition

import (
	sdk "github.com/newrelic/infra-integrations-sdk/metric"
)

// MetricSetEntityIDGeneratorFunc generates an entity ID that will be used as metric set entity ID.
type MetricSetEntityIDGeneratorFunc func(groupLabel, rawEntityID string, g RawGroups) (string, error)

// Spec is a metric specification.
type Spec struct {
	Name      string
	ValueFunc FetchFunc
	Type      sdk.SourceType
}

// SpecGroup represents a bunch of specs that share logic.
type SpecGroup struct {
	IDGenerator MetricSetEntityIDGeneratorFunc
	Specs       []Spec
}

// SpecGroups is a map of groups indexed by group name.
type SpecGroups map[string]SpecGroup
