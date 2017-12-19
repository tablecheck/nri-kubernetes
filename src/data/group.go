package data

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type Grouper interface {
	Group(definition.SpecGroups) (definition.RawGroups, []error)
}

type Populator interface {
	Populate(definition.RawGroups, definition.SpecGroups, *sdk.IntegrationProtocol2) (bool, error)
}

type k8sPopulator struct {
	logger *logrus.Logger
}

func (p *k8sPopulator) Populate(groups definition.RawGroups, specGroups definition.SpecGroups, i *sdk.IntegrationProtocol2) (bool, error) {
	populatorFunc := definition.IntegrationProtocol2PopulateFunc(i, ksmMetric.K8sMetricSetTypeGuesser, ksmMetric.K8sMetricSetEntityTypeGuesser)
	ok, errs := populatorFunc(groups, specGroups)

	if len(errs) > 0 {
		for _, err := range errs {
			p.logger.Debug("%s", err)
		}
	}

	if !ok {
		// TODO better error
		return false, errors.New("no data was populated")
	}

	return true, nil
}

func NewK8sPopulator(logger *logrus.Logger) Populator {
	return &k8sPopulator{
		logger: logger,
	}
}

type ksmGrouper struct {
	kubeletURL *url.URL
	queries    []prometheus.Query
	logger     *logrus.Logger
}

func (r *ksmGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, []error) {
	r.logger.Debug("Endpoint %s called for getting data from kube-state-metrics service", r.kubeletURL)

	mFamily, err := prometheus.Do(r.kubeletURL.String(), r.queries)
	if err != nil {
		return nil, []error{err}
	}

	return ksmMetric.GroupPrometheusMetricsBySpec(specGroups, mFamily)
}

// NewKSM returns an executor for only KSM metrics.
func NewKSMGrouper(kubeletURL *url.URL, queries []prometheus.Query, logger *logrus.Logger) Grouper {
	return &ksmGrouper{
		kubeletURL: kubeletURL,
		queries:    queries,
		logger:     logger,
	}
}

type kubelet struct {
	metricsURL        *url.URL
	httpClient        *http.Client
	kubeletDiscoverer endpoints.Discoverer
	logger            *logrus.Logger
}

func (r *kubelet) Group(definition.SpecGroups) (definition.RawGroups, []error) {
	urlString := r.metricsURL.String()
	r.logger.Debug("Getting metrics data from: %v", urlString)
	response, err := kubeletMetric.GetMetricsData(r.httpClient, urlString)
	if err != nil {
		return nil, []error{err}
	}

	return kubeletMetric.GroupStatsSummary(response)
}

// NewKubelet returns an executor for only local kubelet metrics.
func NewKubeletGrouper(metricsURL *url.URL, httpClient *http.Client, logger *logrus.Logger) Grouper {
	return &kubelet{
		metricsURL: metricsURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

type kubeletKSMGrouper struct {
	ksmMetricsURL     *url.URL
	ksmPartialQueries []prometheus.Query
	ksmSpecGroups     definition.SpecGroups
	kubeletGrouper    Grouper
	logger            *logrus.Logger
}

func (r *kubeletKSMGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, []error) {
	var errs []error
	groups, groupErrs := r.kubeletGrouper.Group(specGroups)
	if len(groupErrs) > 0 {
		errs = append(errs, groupErrs...)
	}

	ksmGrouper := NewKSMGrouper(r.ksmMetricsURL, r.ksmPartialQueries, r.logger)
	ksmGroups, groupErrs := ksmGrouper.Group(r.ksmSpecGroups)
	if len(groupErrs) > 0 {
		errs = append(errs, groupErrs...)
	}

	fillGroups(groups, ksmGroups)

	return groups, errs
}

func NewKubeletKSMGrouper(kubeletURL, ksmMetricsURL *url.URL, c *http.Client, ksmPodAndContainerQueries []prometheus.Query, ksmSpecGroups definition.SpecGroups, logger *logrus.Logger) Grouper {
	return &kubeletKSMGrouper{
		ksmMetricsURL:     ksmMetricsURL,
		ksmPartialQueries: ksmPodAndContainerQueries,
		ksmSpecGroups:     ksmSpecGroups,
		kubeletGrouper:    NewKubeletGrouper(kubeletURL, c, logger),
		logger:            logger,
	}
}

type kubeletKSMAndRestGrouper struct {
	kubeletKSMRawGroups definition.RawGroups
	ksmMetricsURL       *url.URL
	restQueries         []prometheus.Query
	logger              *logrus.Logger
}

func (r *kubeletKSMAndRestGrouper) Group(specGroups definition.SpecGroups) (definition.RawGroups, []error) {
	var errs []error

	ksmGrouper := NewKSMGrouper(r.ksmMetricsURL, r.restQueries, r.logger)
	ksmGroups, groupErrs := ksmGrouper.Group(specGroups)
	if len(groupErrs) > 0 {
		errs = append(errs, groupErrs...)
	}

	mergeNonExistentGroups(ksmGroups, r.kubeletKSMRawGroups)

	return ksmGroups, errs
}

func NewKubeletKSMAndRestGrouper(kubeletKSMGroups definition.RawGroups, ksmMetricsURL *url.URL, ksmRestQueries []prometheus.Query, logger *logrus.Logger) Grouper {
	return &kubeletKSMAndRestGrouper{
		kubeletKSMRawGroups: kubeletKSMGroups,
		ksmMetricsURL:       ksmMetricsURL,
		restQueries:         ksmRestQueries,
		logger:              logger,
	}
}

func fillGroups(destination definition.RawGroups, from definition.RawGroups) {
	for l, g := range destination {
		if fromGroup, ok := from[l]; ok {
			for entityID, e := range fromGroup {
				if _, ok := g[entityID]; !ok {
					continue
				}

				for k, v := range e {
					g[entityID][k] = v
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
