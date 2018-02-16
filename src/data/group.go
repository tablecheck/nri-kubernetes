package data

import (
	"errors"

	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
	"github.com/sirupsen/logrus"
)

// Grouper groups raw data by any desired label such object (pod, container...).
type Grouper interface {
	Group(definition.SpecGroups) (definition.RawGroups, *ErrorGroup)
}

// Populator populates a given integration with grouped raw data.
type Populator interface {
	Populate(definition.RawGroups, definition.SpecGroups, *sdk.IntegrationProtocol2, string) (bool, error)
}

type k8sPopulator struct {
	logger *logrus.Logger
}

// GroupPatcher performs programmatic patching of the destination RawGroups, as a function of the source RawGroups
type GroupPatcher func(destination definition.RawGroups, source definition.RawGroups)

func (p *k8sPopulator) Populate(groups definition.RawGroups, specGroups definition.SpecGroups, i *sdk.IntegrationProtocol2, clusterName string) (bool, error) {
	populatorFunc := definition.IntegrationProtocol2PopulateFunc(i, clusterName, metric.K8sMetricSetTypeGuesser, metric.K8sMetricSetEntityTypeGuesser, metric.K8sEntityMetricsManipulator, metric.K8sClusterMetricsManipulator)
	ok, errs := populatorFunc(groups, specGroups)

	if len(errs) > 0 {
		for _, err := range errs {
			p.logger.Debugf("%s", err)
		}
	}

	if !ok {
		// TODO better error
		return false, errors.New("no data was populated")
	}

	return true, nil
}

// NewK8sPopulator creates a Kubernetes aware populator.
func NewK8sPopulator(logger *logrus.Logger) Populator {
	return &k8sPopulator{
		logger: logger,
	}
}

type ksmGrouper struct {
	queries []prometheus.Query
	client  endpoints.Client
	logger  *logrus.Logger
}

func (r *ksmGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, *ErrorGroup) {
	mFamily, err := prometheus.Do(r.client, r.queries)
	if err != nil {
		return nil, &ErrorGroup{
			Recoverable: false,
			Errors:      []error{fmt.Errorf("error querying KSM. %s", err)},
		}
	}

	groups, errs := ksmMetric.GroupPrometheusMetricsBySpec(specGroups, mFamily)
	if len(errs) == 0 {
		return groups, nil
	}
	return groups, &ErrorGroup{Recoverable: true, Errors: errs}
}

// NewKSMGrouper creates a grouper aware of Kube State Metrics raw metrics.
func NewKSMGrouper(c endpoints.Client, queries []prometheus.Query, logger *logrus.Logger) Grouper {
	return &ksmGrouper{
		queries: queries,
		client:  c,
		logger:  logger,
	}
}

type kubelet struct {
	client endpoints.Client
	logger *logrus.Logger
}

func (r *kubelet) Group(definition.SpecGroups) (definition.RawGroups, *ErrorGroup) {
	response, err := kubeletMetric.GetMetricsData(r.client)
	if err != nil {
		return nil, &ErrorGroup{
			Recoverable: false,
			Errors:      []error{fmt.Errorf("error querying Kubelet. %s", err)},
		}
	}

	groups, errs := kubeletMetric.GroupStatsSummary(response)
	if len(errs) == 0 {
		return groups, nil
	}
	return groups, &ErrorGroup{Recoverable: true, Errors: errs}
}

// NewKubeletGrouper creates a grouper aware of Kubelet raw metrics.
func NewKubeletGrouper(c endpoints.Client, logger *logrus.Logger) Grouper {
	return &kubelet{
		client: c,
		logger: logger,
	}
}

type kubeletKSMGrouper struct {
	ksmClient         endpoints.Client
	ksmPartialQueries []prometheus.Query
	ksmSpecGroups     definition.SpecGroups
	kubeletGrouper    Grouper
	logger            *logrus.Logger
	groupPatcher      GroupPatcher
}

func (r *kubeletKSMGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, *ErrorGroup) {
	errs := ErrorGroup{Recoverable: true}

	groups, groupErrs := r.kubeletGrouper.Group(specGroups)
	if groupErrs != nil {
		if groupErrs.Recoverable {
			errs.Append(groupErrs.Errors...)
		} else {
			return nil, groupErrs
		}
	}

	ksmGrouper := NewKSMGrouper(r.ksmClient, r.ksmPartialQueries, r.logger)
	ksmGroups, groupErrs := ksmGrouper.Group(r.ksmSpecGroups)
	if groupErrs != nil {
		if groupErrs.Recoverable {
			errs.Append(groupErrs.Errors...)
		} else {
			return nil, groupErrs
		}
	}

	if r.groupPatcher != nil {
		r.groupPatcher(groups, ksmGroups)
	}

	fillGroupsAndMergeNonExistent(groups, ksmGroups)

	return groups, &errs
}

// NewKubeletKSMGrouper creates a grouper that merges groups provided by the
// kubelet and ksm groupers.
func NewKubeletKSMGrouper(kubeletClient, ksmClient endpoints.Client, ksmPodAndContainerQueries []prometheus.Query, ksmSpecGroups definition.SpecGroups, logger *logrus.Logger) Grouper {
	return NewKubeletKSMPatchedGrouper(kubeletClient, ksmClient, ksmPodAndContainerQueries, ksmSpecGroups, logger, nil)
}

// NewKubeletKSMPatchedGrouper creates a grouper that merges groups provided by
// the kubeletKSMAndRestGrouper plus some missing ksm raw metrics.
func NewKubeletKSMPatchedGrouper(kubeletClient, ksmClient endpoints.Client, ksmPodAndContainerQueries []prometheus.Query, ksmSpecGroups definition.SpecGroups, logger *logrus.Logger, patcher GroupPatcher) Grouper {
	return &kubeletKSMGrouper{
		ksmClient:         ksmClient,
		ksmPartialQueries: ksmPodAndContainerQueries,
		ksmSpecGroups:     ksmSpecGroups,
		kubeletGrouper:    NewKubeletGrouper(kubeletClient, logger),
		logger:            logger,
		groupPatcher:      patcher,
	}
}

type kubeletKSMAndRestGrouper struct {
	kubeletKSMRawGroups definition.RawGroups
	ksmClient           endpoints.Client
	restQueries         []prometheus.Query
	logger              *logrus.Logger
}

func (r *kubeletKSMAndRestGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, *ErrorGroup) {
	errs := ErrorGroup{Recoverable: true}

	ksmGrouper := NewKSMGrouper(r.ksmClient, r.restQueries, r.logger)
	ksmGroups, groupErrs := ksmGrouper.Group(specGroups)
	if groupErrs != nil {
		if groupErrs.Recoverable {
			errs.Append(groupErrs.Errors...)
		} else {
			return nil, groupErrs
		}
	}

	mergeNonExistentGroups(ksmGroups, r.kubeletKSMRawGroups)

	return ksmGroups, &errs
}

// NewKubeletKSMAndRestGrouper creates a grouper that merges groups provided by
// the kubelet and ksm groupers plus some additional ksm raw metrics.
func NewKubeletKSMAndRestGrouper(kubeletKSMGroups definition.RawGroups, ksmClient endpoints.Client, ksmRestQueries []prometheus.Query, logger *logrus.Logger) Grouper {
	return &kubeletKSMAndRestGrouper{
		kubeletKSMRawGroups: kubeletKSMGroups,
		ksmClient:           ksmClient,
		restQueries:         ksmRestQueries,
		logger:              logger,
	}
}

func fillGroupsAndMergeNonExistent(destination definition.RawGroups, from definition.RawGroups) {
	for l, g := range from {
		if _, ok := destination[l]; !ok {
			destination[l] = g
			continue
		}

		for entityID, e := range destination[l] {
			if _, ok := g[entityID]; !ok {
				continue
			}

			for k, v := range g[entityID] {
				if _, ok := e[k]; !ok {
					e[k] = v
				}
			}
		}
	}
}

func mergeNonExistentGroups(destination, from definition.RawGroups) {
	for g, e := range from {
		if _, ok := destination[g]; !ok {
			destination[g] = e
		}
	}
}
