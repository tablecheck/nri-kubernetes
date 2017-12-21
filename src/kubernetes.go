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
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	kubeletEndpoints "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/endpoints"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	KubeStateMetricsURL string `help:"overrides Kube State Metrics schema://host:port URL parts (if not set, it will be self-discovered)."`
	DebugKubeletURL     string `help:"for debugging purposes. Overrides kubelet schema://host:port URL parts (if not set, it will be self-discovered)"`
	DebugRole           string `help:"for debugging purposes. Sets the role of the integration (accepted values: kubelet-ksm-rest, kubelet-ksm. If not set, it will be self-discovered)"`
	IgnoreCerts         bool   `default:"false" help:"disables HTTPS certificate verification for metrics sources"`
	Timeout             int    `default:"1000" help:"timeout in milliseconds for calling metrics sources"`
	ClusterName         string `help:"Identifier of your cluster. You could use it later to filter data in your New Relic account"`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "1.0.0-beta"
	statsSummaryPath   = "/stats/summary"
	metricsPath        = "/metrics"
)

var args argumentList

func kubeletKSM(kubeletKSMGrouper data.Grouper, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) {
	groups, errs := kubeletKSMGrouper.Group(kubeletSpecs)
	for _, err := range errs {
		logger.Warn("%s", err)
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, kubeletKSMPopulateSpecs, i, clusterName)
	fatalIfErr(err)

	e, err := i.Entity("nr-errors", "error")
	fatalIfErr(err)
	for _, err := range errs {
		ms := e.NewMetricSet("K8sDebugErrors")
		ms.SetMetric("error", err.Error(), metric.ATTRIBUTE)
		ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
	}

	if !ok {
		// TODO better error
		log.Fatal(errors.New("no data was populated"))
	}
}

func kubeletKSMAndRest(kubeletKSMGrouper data.Grouper, ksmMetricsURL *url.URL, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) {
	kubeletKSMGroups, errs := kubeletKSMGrouper.Group(kubeletSpecs)
	g := data.NewKubeletKSMAndRestGrouper(kubeletKSMGroups, ksmMetricsURL, prometheusRestQueries, logger)
	groups, errs := g.Group(ksmRestSpecs)
	for _, err := range errs {
		logger.Warn("%s", err)
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, kubeletKSMAndRestPopulateSpecs, i, clusterName)
	fatalIfErr(err)

	e, err := i.Entity("nr-errors", "error")
	fatalIfErr(err)
	for _, err := range errs {
		ms := e.NewMetricSet("K8sDebugErrors")
		ms.SetMetric("error", err.Error(), metric.ATTRIBUTE)
		ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
	}

	if !ok {
		// TODO better error
		log.Fatal(errors.New("no data was populated"))
	}
}

func main() {
	defer log.Debug("Integration '%s' exited", integrationName)
	log.Debug("Integration '%s' with version %s started", integrationName, integrationVersion)
	integration, err := sdk.NewIntegrationProtocol2(integrationName, integrationVersion, &args)
	if args.ClusterName == "" {
		log.Fatal(errors.New("cluster_name argument is mandatory"))
	}
	fatalIfErr(err)

	var ksmDiscoverer endpoints.Discoverer

	if args.All || args.Metrics {
		if args.DebugKubeletURL != "" {
			log.Warn("using argument aimed for debugging purposes. debug_kubelet_url=%q", args.DebugKubeletURL)
		}
		kubeletURL, err := url.Parse(args.DebugKubeletURL)
		fatalIfErr(err)

		ksmURL, err := url.Parse(args.KubeStateMetricsURL)
		fatalIfErr(err)

		if args.DebugRole != "" {
			log.Warn("using argument aimed for debugging purposes. debug_role=%q", args.DebugRole)
		}
		role := args.DebugRole
		if role == "" {
			// autodiscover

			kubeletDiscoverer, err := kubeletEndpoints.NewKubeletDiscoverer()
			fatalIfErr(err)

			discoveredKubeletURL, err := kubeletDiscoverer.Discover()
			fatalIfErr(err)
			kubeletURL = &discoveredKubeletURL

			kubeletNodeIP, err := kubeletDiscoverer.NodeIP()
			fatalIfErr(err)
			log.Debug("Kubelet Node = %s", kubeletNodeIP)

			ksmDiscoverer, err = ksmEndpoints.NewKSMDiscoverer()
			fatalIfErr(err)

			ksmNodeIP, err := ksmDiscoverer.NodeIP()
			fatalIfErr(err)

			log.Debug("KSM Node = %s", ksmNodeIP)

			discoveredKSMURL, err := ksmDiscoverer.Discover()
			fatalIfErr(err)
			ksmURL = &discoveredKSMURL

			// setting role by auto discovery
			if kubeletNodeIP == ksmNodeIP {
				role = "kubelet-ksm-rest"
			} else {
				role = "kubelet-ksm"
			}
			log.Debug("Auto-discovered role = %s", role)
		}

		if ksmURL.String() == "" {
			log.Fatal(errors.New("kube_state_metrics_url should be provided"))
		}

		if kubeletURL.String() == "" {
			log.Fatal(errors.New("debug_kubelet_url should be provided"))
		}

		kubeletURL.Path = statsSummaryPath
		ksmURL.Path = metricsPath

		log.Debug("Kubelet URL = %s", kubeletURL)
		log.Debug("KSM URL = %s", ksmURL)

		netClient := &http.Client{
			Timeout: time.Millisecond * time.Duration(args.Timeout),
		}

		if args.IgnoreCerts && kubeletURL.Scheme == "https" {
			netClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		logger := logrus.New()

		switch role {
		case "kubelet-ksm-rest":
			// todo fix pointers indirection stuff
			kubeletKSMGrouper := data.NewKubeletKSMPatchedGrouper(
				kubeletURL,
				ksmURL,
				netClient,
				prometheusPodsAndContainerQueries,
				ksmPodAndContainerSpecs,
				logger,
				ksmMetric.UnscheduledItemsPatcher,
			)

			// todo fix pointers indirection stuff
			kubeletKSMAndRest(kubeletKSMGrouper, ksmURL, integration, args.ClusterName, logger)
		case "kubelet-ksm":
			// todo fix pointers indirection stuff
			kubeletKSMGrouper := data.NewKubeletKSMGrouper(
				kubeletURL,
				ksmURL,
				netClient,
				prometheusPodsAndContainerQueries,
				ksmPodAndContainerSpecs,
				logger,
			)

			kubeletKSM(kubeletKSMGrouper, integration, args.ClusterName, logger)
		}
	}

	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
