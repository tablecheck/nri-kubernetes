package data

import "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"

// FetchFunc fetches data from a source.
// TODO Pass definition.SpecGroups and just retrieve requested data. See IHOST-332.
type FetchFunc func() (definition.RawGroups, error)
