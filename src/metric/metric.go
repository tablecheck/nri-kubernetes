package metric

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
)

const (
	UnknownNamespace = "_unknown"
)

// NamespaceFetcher fetches the namespace from the provided RawGroups, for the given groupLabel and entityId
type NamespaceFetcher func(groupLabel, entityID string, groups definition.RawGroups) (string, error)

// MultipleNamespaceFetcher asks multiple NamespaceFetcher instances for the namespace. It queries the NamespaceFetchers
// in the provided order and returns the namespace from the first invocation that does not return error. If all the
// invocations fail, this function returns "_unknown" namespace and an error instance.
func MultipleNamespaceFetcher(groupLabel, entityID string, groups definition.RawGroups, fetchers ...NamespaceFetcher) (string, error) {

	errors := make([]string, 0)

	for _, fetcher := range fetchers {
		ns, err := fetcher(groupLabel, entityID, groups)
		if err == nil {
			return ns, nil
		}
		errors = append(errors, err.Error())
	}

	return UnknownNamespace, fmt.Errorf("error fetching namespace: %s", strings.Join(errors, ", "))
}
