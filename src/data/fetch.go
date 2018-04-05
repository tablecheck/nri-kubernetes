package data

import "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"

// FetchFunc fetches data from a source.
type FetchFunc func() (definition.RawGroups, error)
