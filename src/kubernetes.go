package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/definition"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/prometheus"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/endpoints"
	kubeletMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	MetricsURL        string `default:"http://localhost:8080/metrics" help:"Kube-state-metrics URL."`
	KubeletURL        string `help:"overrides kubelet schema://host:port URL parts (if not set, it will be self-discovered)"`
	KubeletIgnoreCert bool   `default:"false" help:"disables kubelet HTTPS certificate verification"`
	Timeout           int    `default:"1000" help:"Timeout in milliseconds for calling kubelet /stats/summary URL."`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "0.1.0"
	statsSummaryPath   = "/stats/summary"
)

var args argumentList

func main() {
	defer log.Debug("Integration '%s' exited", integrationName)

	integration, err := sdk.NewIntegrationProtocol2(integrationName, integrationVersion, &args)
	log.Debug("Integration '%s' with version %s started", integrationName, integrationVersion)
	fatalIfErr(err)

	if args.All || args.Metrics {
		// Kube State Metrics
		populateKubeStateMetrics(args.MetricsURL, integration)

		// Kubelet Metrics
		netClient := &http.Client{
			Timeout: time.Millisecond * time.Duration(args.Timeout),
		}

		var kubeletURL url.URL
		if args.KubeletURL != "" {
			pURL, err := url.Parse(args.KubeletURL)
			fatalIfErr(err)
			kubeletURL = *pURL
		} else {
			kubelet, err := endpoints.NewKubeletDiscoverer()
			fatalIfErr(err)
			kubeletURL, err = kubelet.Discover()
			fatalIfErr(err)
		}

		kubeletURL.Path = statsSummaryPath

		if args.KubeletIgnoreCert && kubeletURL.Scheme == "https" {
			netClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		populateKubeletMetrics(kubeletURL, netClient, integration)
	}

	fatalIfErr(integration.Publish())
}

func populateKubeStateMetrics(ksmMetricsURL string, integration *sdk.IntegrationProtocol2) {
	mFamily, err := prometheus.Do(ksmMetricsURL, prometheusQueries)
	log.Debug("Endpoint %s called for getting data from kube-state-metrics service", ksmMetricsURL)
	fatalIfErr(err)

	groups, errs := metric.GroupPrometheusMetricsBySpec(ksmAggregation, mFamily)
	for _, err := range errs {
		log.Warn("%s", err)
	}

	if len(groups) == 0 {
		log.Fatal(errors.New("no data was fetched"))
	}

	populator := definition.IntegrationProtocol2PopulateFunc(integration, metric.K8sMetricSetTypeGuesser, metric.K8sMetricSetEntityTypeGuesser)
	ok, errs := populator(groups, ksmAggregation)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug("%s", err)
		}
	}

	if !ok {
		// TODO better error
		log.Fatal(errors.New("no data was populated"))
	}
}

func populateKubeletMetrics(kubeletURL url.URL, netClient *http.Client, integration *sdk.IntegrationProtocol2) {
	log.Debug("Getting metrics data from: %v", kubeletURL)
	response, err := kubeletMetric.GetMetricsData(netClient, kubeletURL.String())
	if err != nil {
		log.Fatal(err)
	}
	groups, errs := kubeletMetric.GroupStatsSummary(response)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug("%s", err)
		}
	}
	for entitySourceName, d := range kubeletAggregation {
		if len(groups) == 0 {
			log.Debug("No data found for %s object", entitySourceName)
			continue
		}

		populated, errs := kubeletMetric.Populate(integration, d, groups)
		if len(errs) > 0 {
			for _, err := range errs {
				log.Debug("%s", err)
			}
		}

		if !populated {
			log.Warn("empty metrics for %s", entitySourceName)
			continue
		}
	}
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
