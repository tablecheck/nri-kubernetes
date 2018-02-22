package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/data"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/endpoints"
	ksmEndpoints "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/endpoints"
	ksmMetric "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/ksm/metric"
	kubeletEndpoints "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/endpoints"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
	"github.com/sirupsen/logrus"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Timeout     int    `default:"5000" help:"timeout in milliseconds for calling metrics sources"`
	ClusterName string `help:"Identifier of your cluster. You could use it later to filter data in your New Relic account"`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "1.0.0-beta4"
)

var args argumentList

func kubeletKSM(kubeletKSMGrouper data.Grouper, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) error {
	groups, errs := kubeletKSMGrouper.Group(kubeletMergeableSpecs)
	if errs != nil && len(errs.Errors) > 0 {
		if !errs.Recoverable {
			return errors.New(errs.String())
		}
		logger.Warnf("%s", errs.String())
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, mergeableObjectsPopulateSpecs, i, clusterName)
	if err != nil {
		logger.Fatal(err)
	}

	e, err := i.Entity("nr-errors", "error")
	if err != nil {
		logger.Fatal(err)
	}
	if errs != nil {
		for _, err := range errs.Errors {
			ms := e.NewMetricSet("K8sDebugErrors")
			mserr := ms.SetMetric("error", err.Error(), metric.ATTRIBUTE)
			if mserr != nil {
				logger.Debugf("error setting a value in '%s' in metric set '%s': %v", "error", "K8sDebugErrors", mserr)
			}
			mserr = ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
			if mserr != nil {
				logger.Debugf("error setting a value in '%s' in metric set '%s': %v", "clusterName", "K8sDebugErrors", mserr)
			}
		}
	}

	if !ok {
		// TODO better error
		return errors.New("no data was populated")
	}
	return nil
}

func kubeletKSMAndRest(kubeletKSMGrouper data.Grouper, ksmClient endpoints.Client, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) error {
	kubeletKSMGroups, errs := kubeletKSMGrouper.Group(kubeletMergeableSpecs)
	if errs != nil && len(errs.Errors) > 0 {
		if !errs.Recoverable {
			return errors.New(errs.String())
		}
		logger.Warnf("%s", errs.String())
	}

	g := data.NewKubeletKSMAndRestGrouper(kubeletKSMGroups, ksmClient, unmergeableObjectsQueries, logger)
	groups, errs := g.Group(ksmUnmergeableSpecs)
	if errs != nil && len(errs.Errors) > 0 {
		if !errs.Recoverable {
			return errors.New(errs.String())
		}
		logger.Warnf("%s", errs.String())
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, allObjectsPopulateSpecs, i, clusterName)
	if err != nil {
		logger.Fatal(err)
	}

	e, err := i.Entity("nr-errors", "error")
	if err != nil {
		logger.Fatal(err)
	}
	if errs != nil {
		for _, err := range errs.Errors {
			ms := e.NewMetricSet("K8sDebugErrors")
			mserr := ms.SetMetric("error", err.Error(), metric.ATTRIBUTE)
			if mserr != nil {
				logger.Debugf("error setting a value in '%s' in metric set '%s': %v", "error", "K8sDebugErrors", mserr)
			}
			mserr = ms.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
			if mserr != nil {
				logger.Debugf("error setting a value in '%s' in metric set '%s': %v", "clusterName", "K8sDebugErrors", mserr)
			}
		}
	}

	if !ok {
		// TODO better error
		return errors.New("no data was populated")
	}
	return nil
}

func main() {
	integration, err := sdk.NewIntegrationProtocol2(integrationName, integrationVersion, &args)
	exitLog := fmt.Sprintf("Integration %q exited", integrationName)
	if err != nil {
		defer log.Debug(exitLog)
		log.Fatal(err) // Global logs used as args processed inside NewIntegrationProtocol2
	}

	logger := log.New(args.Verbose)

	defer logger.Debug(exitLog)
	logger.Debugf("Integration %q with version %s started", integrationName, integrationVersion)
	if args.ClusterName == "" {
		logger.Fatal(errors.New("cluster_name argument is mandatory"))
	}

	if args.All || args.Metrics {
		timeout := time.Millisecond * time.Duration(args.Timeout)

		kubeletDiscoverer, err := kubeletEndpoints.NewKubeletDiscoverer(logger)
		if err != nil {
			logger.Fatal(err)
		}
		kubeletClient, err := kubeletDiscoverer.Discover(timeout)
		if err != nil {
			logger.Fatal(err)
		}
		kubeletNodeIP := kubeletClient.NodeIP()
		logger.Debugf("Kubelet Node = %s", kubeletNodeIP)

		ksmDiscoverer, err := ksmEndpoints.NewKSMDiscoverer(logger)
		if err != nil {
			logger.Fatal(err)
		}
		ksmClient, err := ksmDiscoverer.Discover(timeout)
		if err != nil {
			logger.Fatal(err)
		}
		ksmNodeIP := ksmClient.NodeIP()
		logger.Debugf("KSM Node = %s", ksmNodeIP)

		// setting role by auto discovery
		var role string
		if kubeletNodeIP == ksmNodeIP {
			role = "kubelet-ksm-rest"
		} else {
			role = "kubelet-ksm"
		}
		logger.Debugf("Auto-discovered role = %s", role)

		switch role {
		case "kubelet-ksm-rest":
			// todo fix pointers indirection stuff
			kubeletKSMGrouper := data.NewKubeletKSMPatchedGrouper(
				kubeletClient,
				ksmClient,
				mergeableObjectsQueries,
				ksmMergeableSpecs,
				logger,
				ksmMetric.UnscheduledItemsPatcher,
			)

			// todo fix pointers indirection stuff
			err = kubeletKSMAndRest(kubeletKSMGrouper, ksmClient, integration, args.ClusterName, logger)
			if err != nil {
				logger.Fatal(err)
			}
		case "kubelet-ksm":
			// todo fix pointers indirection stuff
			kubeletKSMGrouper := data.NewKubeletKSMGrouper(
				kubeletClient,
				ksmClient,
				mergeableObjectsQueries,
				ksmMergeableSpecs,
				logger,
			)

			err = kubeletKSM(kubeletKSMGrouper, integration, args.ClusterName, logger)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	err = integration.Publish()
	if err != nil {
		logger.Fatal(err)
	}
}
