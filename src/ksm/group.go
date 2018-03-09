package ksm

import (
	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/data"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	"github.com/sirupsen/logrus"
)

type ksmGrouper struct {
	queries []prometheus.Query
	client  endpoints.Client
	logger  *logrus.Logger
}

func (r *ksmGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, *data.ErrorGroup) {
	mFamily, err := prometheus.Do(r.client, r.queries)
	if err != nil {
		return nil, &data.ErrorGroup{
			Recoverable: false,
			Errors:      []error{fmt.Errorf("error querying KSM. %s", err)},
		}
	}

	groups, errs := metric.GroupPrometheusMetricsBySpec(specGroups, mFamily)
	if len(errs) == 0 {
		return groups, nil
	}
	return groups, &data.ErrorGroup{Recoverable: true, Errors: errs}
}

// NewGrouper creates a grouper aware of Kube State Metrics raw metrics.
func NewGrouper(c endpoints.Client, queries []prometheus.Query, logger *logrus.Logger) data.Grouper {
	return &ksmGrouper{
		queries: queries,
		client:  c,
		logger:  logger,
	}
}
