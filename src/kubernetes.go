package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/definition"
	endpoints2 "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/endpoints"
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
	MetricsURL  string `help:"overrides Kube State Metrics schema://host:port URL parts (if not set, it will be self-discovered)."`
	KubeletURL  string `help:"overrides kubelet schema://host:port URL parts (if not set, it will be self-discovered)"`
	IgnoreCerts bool   `default:"false" help:"disables HTTPS certificate verification for metrics sources"`
	Ksm         string `default:"auto" help:"whether the Kube State Metrics must be reported or not (accepted values: true, false, auto)"`
	Timeout     int    `default:"1000" help:"Timeout in milliseconds for calling metrics sources"`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "0.1.0"
	statsSummaryPath   = "/stats/summary"
	metricsPath        = "/metrics"
)

var args argumentList

func main() {
	defer log.Debug("Integration '%s' exited", integrationName)

	integration, err := sdk.NewIntegrationProtocol2(integrationName, integrationVersion, &args)
	log.Debug("Integration '%s' with version %s started", integrationName, integrationVersion)
	fatalIfErr(err)

	if args.All || args.Metrics {
		// Kube State Metrics Discovery
		var ksmURL url.URL
		var ksmNode string

		if args.MetricsURL != "" {
			pURL, err := url.Parse(args.MetricsURL)
			fatalIfErr(err)
			ksmURL = *pURL
		} else if args.Ksm != "false" {
			ksm, err := endpoints2.NewKSMDiscoverer()
			fatalIfErr(err)
			ksmURL, err = ksm.Discover()
			fatalIfErr(err)
			ksmNode, err = ksm.NodeIP()
			fatalIfErr(err)
		}
		ksmURL.Path = metricsPath

		log.Debug("KSM URL = %s", ksmURL.String())
		log.Debug("KSM Node = %s", ksmNode)

		// Kubelet Discovery
		var kubeletURL url.URL
		var kubeletNode string

		if args.KubeletURL != "" {
			pURL, err := url.Parse(args.KubeletURL)
			fatalIfErr(err)
			kubeletURL = *pURL
			kubeletNode = kubeletURL.Hostname()
		} else {
			kubelet, err := endpoints.NewKubeletDiscoverer()
			fatalIfErr(err)
			kubeletURL, err = kubelet.Discover()
			fatalIfErr(err)
			kubeletNode, err = kubelet.NodeIP()
			fatalIfErr(err)
		}
		kubeletURL.Path = statsSummaryPath
		log.Debug("Kubelet URL = %s", kubeletURL.String())
		log.Debug("Kubelet Node = %s", kubeletNode)

		netClient := &http.Client{
			Timeout: time.Millisecond * time.Duration(args.Timeout),
		}

		if args.IgnoreCerts && kubeletURL.Scheme == "https" {
			netClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		// We populate KSM metrics in the next cases
		// - If "ksm==true", metrics are always populated
		// - If "ksm==false", metrics are never populated
		// - If "ksm==auto", metrics are populated if:
		//       . The user sets the MetricsURL argument
		//       . The discovery mechanisms shows that Kubelet and KSM are in the same node
		if args.Ksm == "true" ||
			(args.Ksm != "false" && args.MetricsURL != "") ||
			(args.Ksm == "auto" && kubeletNode == ksmNode) {
			populateKubeStateMetrics(ksmURL.String(), integration)
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

	populator := definition.IntegrationProtocol2PopulateFunc(integration, metric.K8sMetricSetTypeGuesser, metric.MetricSetEntityTypeGuesser(metric.KSMNamespaceFetcher))
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
	urlString := kubeletURL.String()
	log.Debug("Getting metrics data from: %v", urlString)
	response, err := kubeletMetric.GetMetricsData(netClient, urlString)
	if err != nil {
		log.Fatal(err)
	}
	groups, errs := kubeletMetric.GroupStatsSummary(response)
	for _, err := range errs {
		log.Warn("%s", err)
	}

	if len(groups) == 0 {
		log.Fatal(errors.New("no data was fetched"))
	}

	populator := definition.IntegrationProtocol2PopulateFunc(integration, metric.K8sMetricSetTypeGuesser, metric.MetricSetEntityTypeGuesser(metric.KubeletNamespaceFetcher))
	ok, errs := populator(groups, kubeletAggregation)
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

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
