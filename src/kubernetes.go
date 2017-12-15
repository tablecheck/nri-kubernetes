package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/data"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	ksmEndpoints "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/endpoints"
	kubeletEndpoints "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/endpoints"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	KubeStateMetricsURL string `help:"overrides Kube State Metrics schema://host:port URL parts (if not set, it will be self-discovered)."`
	KubeletURL          string `help:"overrides kubelet schema://host:port URL parts (if not set, it will be self-discovered)"`
	IgnoreCerts         bool   `default:"false" help:"disables HTTPS certificate verification for metrics sources"`
	Ksm                 string `default:"auto" help:"whether the Kube State Metrics must be reported or not (accepted values: true, false, auto)"`
	Timeout             int    `default:"1000" help:"Timeout in milliseconds for calling metrics sources"`
	Role                string `help:"For debugging purpose. Sets the role of the integration (accepted values: kubelet-ksm-rest, kubelet-ksm"`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "0.1.0"
	statsSummaryPath   = "/stats/summary"
	metricsPath        = "/metrics"
)

var args argumentList

func kubeletKSM(kubeletKSMGrouper data.Grouper, i *sdk.IntegrationProtocol2, logger *logrus.Logger) {
	groups, errs := kubeletKSMGrouper.Group(kubeletSpecs)
	for _, err := range errs {
		logger.Warn("%s", err)
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, kubeletKSMPopulateSpecs, i)
	fatalIfErr(err)

	if !ok {
		// TODO better error
		log.Fatal(errors.New("no data was populated"))
	}
}

func kubeletKSMAndRest(kubeletKSMGrouper data.Grouper, ksmMetricsURL *url.URL, i *sdk.IntegrationProtocol2, logger *logrus.Logger) {
	kubeletKSMGroups, errs := kubeletKSMGrouper.Group(kubeletSpecs)
	g := data.NewKubeletKSMAndRestGrouper(kubeletKSMGroups, ksmMetricsURL, prometheusRestQueries, logger)
	groups, errs := g.Group(ksmRestSpecs)
	for _, err := range errs {
		logger.Warn("%s", err)
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, kubeletKSMAndRestPopulateSpecs, i)
	fatalIfErr(err)

	if !ok {
		// TODO better error
		log.Fatal(errors.New("no data was populated"))
	}
}

func main() {
	defer log.Debug("Integration '%s' exited", integrationName)

	integration, err := sdk.NewIntegrationProtocol2(integrationName, integrationVersion, &args)
	log.Debug("Integration '%s' with version %s started", integrationName, integrationVersion)
	fatalIfErr(err)

	var ksmDiscoverer endpoints.Discoverer

	if args.All || args.Metrics {
		// Kube State Metrics Discovery
		//var ksmURL url.URL
		//var ksmNode string

		//if args.KubeStateMetricsURL != "" {
		//	pURL, err := url.Parse(args.KubeStateMetricsURL)
		//	fatalIfErr(err)
		//	ksmURL = *pURL
		//} else if args.Ksm != "false" {
		//	ksmDiscoverer, err = ksmEndpoints.NewKSMDiscoverer()
		//	if err == nil {
		//		ksmNode, err = ksmDiscoverer.NodeIP()
		//	}
		//	if err == nil {
		//		log.Debug("KSM Node = %s", ksmNode)
		//	} else {
		//		log.Debug("can't get Kube State Metrics node: %q", err.Error())
		//	}
		//}
		//ksmURL.Path = metricsPath

		//log.Debug("KSM URL = %s", ksmURL.String())
		//log.Debug("KSM Node = %s", ksmNode)

		// Kubelet Discovery

		logger := logrus.New()
		// TODO decide role using autodiscovery mechanism.
		//
		//if args.KubeletURL != "" {
		//	pURL, err := url.Parse(args.KubeletURL)
		//	fatalIfErr(err)
		//	kubeletURL = *pURL
		//	kubeletNode = kubeletURL.Hostname()
		//} else {
		//	kubelet, err := kubeletEndpoints.NewKubeletDiscoverer()
		//	fatalIfErr(err)
		//	kubeletURL, err = kubelet.Discover()
		//	fatalIfErr(err)
		//	kubeletNode, err = kubelet.NodeIP()
		//	fatalIfErr(err)
		//}
		//kubeletURL.Path = statsSummaryPath
		//log.Debug("Kubelet URL = %s", kubeletURL.String())
		//log.Debug("Kubelet Node = %s", kubeletNode)

		// We populate KSM metrics in the next cases
		// - If "ksm==true", metrics are always populated
		// - If "ksm==false", metrics are never populated
		// - If "ksm==auto", metrics are populated if:
		//       . The user sets the MetricsURL argument
		//       . The discovery mechanisms shows that Kubelet and KSM are in the same node
		//if args.Ksm == "true" ||
		//	(args.Ksm != "false" && args.MetricsURL != "") ||
		//	(args.Ksm == "auto" && kubeletNode == ksmNode) {
		//
		//	ksmURL, err = ksmDiscoverer.Discover()
		//	ksmURL.Path = metricsPath
		//
		//	log.Debug("KSM URL = %s", ksmURL.String())
		//
		//	fatalIfErr(err)
		//	if err == nil {
		//		populateKubeStateMetrics(ksmURL.String(), integration)
		//	}
		//}

		var kubeletURL url.URL
		var ksmURL url.URL

		role := args.Role
		if role == "" {
			// autodiscover

			kubeletDiscoverer, err := kubeletEndpoints.NewKubeletDiscoverer()
			fatalIfErr(err)

			kubeletURL, err = kubeletDiscoverer.Discover()
			fatalIfErr(err)
			log.Debug("Kubelet URL = %s", kubeletURL.String())

			kubeletNodeIP, err := kubeletDiscoverer.NodeIP()
			fatalIfErr(err)
			log.Debug("Kubelet Node = %s", kubeletNodeIP)

			ksmDiscoverer, err = ksmEndpoints.NewKSMDiscoverer()
			fatalIfErr(err)

			ksmNodeIP, err := ksmDiscoverer.NodeIP()
			fatalIfErr(err)

			log.Debug("KSM Node = %s", ksmNodeIP)

			discoveredKubeletURL, err := ksmDiscoverer.Discover()
			fatalIfErr(err)
			log.Debug("KSM URL = %s", ksmNodeIP)

			kubeletURL = discoveredKubeletURL

			// setting role by auto discovery
			if kubeletNodeIP == ksmNodeIP {
				role = "kubelet-ksm-rest"
			} else {
				role = "kubelet-ksm"
			}
		}

		if ksmURL.String() == "" {
			log.Fatal(errors.New("kube_state_metrics_url should be provided"))
		}

		if kubeletURL.String() == "" {
			log.Fatal(errors.New("kubelet_url should be provided"))
		}

		netClient := &http.Client{
			Timeout: time.Millisecond * time.Duration(args.Timeout),
		}

		if args.IgnoreCerts && kubeletURL.Scheme == "https" {
			netClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		kubeletURL.Path = statsSummaryPath
		ksmURL.Path = metricsPath

		// todo fix pointers indirection stuff
		kubeletKSMGrouper := data.NewKubeletKSMGrouper(
			&kubeletURL,
			&ksmURL,
			netClient,
			prometheusPodsAndContainerQueries,
			ksmPodAndContainerSpecs,
			logger,
		)

		switch role {
		case "kubelet-ksm-rest":
			// todo fix pointers indirection stuff
			kubeletKSMAndRest(kubeletKSMGrouper, &ksmURL, integration, logger)
		case "kubelet-ksm":
			kubeletKSM(kubeletKSMGrouper, integration, logger)
		}

	}

	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
