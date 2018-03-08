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
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/storage"
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
	CacheDir    string `default:"/var/cache/nr-kubernetes" help:"The location of the integration cached values"`
	CacheTTL    string `default:"1h" help:"The duration since the cached entries are stored until they expire"`
}

const (
	integrationName    = "com.newrelic.kubernetes"
	integrationVersion = "1.0.0-beta4"
)

var args argumentList

func follower(kubeletKSMGrouper data.Grouper, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) error {
	groups, errs := kubeletKSMGrouper.Group(kubeletMergeableSpecs)
	if errs != nil && len(errs.Errors) > 0 {
		if !errs.Recoverable {
			return errors.New(errs.String())
		}
		logger.Warnf("%s", errs.String())
	}

	ok, err := data.NewK8sPopulator(logger).Populate(groups, mergeableObjectsPopulateSpecs, i, clusterName)
	if err != nil {
		logger.Panic(err)
	}

	e, err := i.Entity("nr-errors", "error")
	if err != nil {
		logger.Panic(err)
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

func leader(kubeletKSMGrouper data.Grouper, ksmClient endpoints.Client, i *sdk.IntegrationProtocol2, clusterName string, logger *logrus.Logger) error {
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
		logger.Panic(err)
	}

	e, err := i.Entity("nr-errors", "error")
	if err != nil {
		logger.Panic(err)
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
	defer func() {
		if r := recover(); r != nil {
			recErr, ok := r.(*logrus.Entry)
			if ok {
				recErr.Fatal(recErr.Message)
			} else {
				panic(r)
			}
		}
	}()

	defer logger.Debug(exitLog)
	logger.Debugf("Integration %q with version %s started", integrationName, integrationVersion)
	if args.ClusterName == "" {
		logger.Panic(errors.New("cluster_name argument is mandatory"))
	}

	if args.All || args.Metrics {
		ttl, err := time.ParseDuration(args.CacheTTL)
		if err != nil {
			logger.WithError(err).Error("while parsing the cache TTL value. Defaulting to 1h")
			ttl = time.Hour
		}

		timeout := time.Millisecond * time.Duration(args.Timeout)

		innerKubeletDiscoverer, err := kubeletEndpoints.NewKubeletDiscoverer(logger)
		if err != nil {
			logger.Panic(err)
		}
		cacheStorage := storage.NewJSONDiskStorage(args.CacheDir)
		kubeletDiscoverer := kubeletEndpoints.NewKubeletDiscoveryCacher(innerKubeletDiscoverer, cacheStorage, ttl, logger)

		kubeletClient, err := kubeletDiscoverer.Discover(timeout)
		if err != nil {
			logger.Panic(err)
		}
		kubeletNodeIP := kubeletClient.NodeIP()
		logger.Debugf("Kubelet Node = %s", kubeletNodeIP)

		innerKSMDiscoverer, err := ksmEndpoints.NewKSMDiscoverer(logger)
		if err != nil {
			logger.Panic(err)
		}
		ksmDiscoverer := ksmEndpoints.NewKSMDiscoveryCacher(innerKSMDiscoverer, cacheStorage, ttl, logger)
		ksmClient, err := ksmDiscoverer.Discover(timeout)
		if err != nil {
			logger.Panic(err)
		}
		ksmNodeIP := ksmClient.NodeIP()
		logger.Debugf("KSM Node = %s", ksmNodeIP)

		// setting role by auto discovery
		var role string
		if kubeletNodeIP == ksmNodeIP {
			role = "leader"
		} else {
			role = "follower"
		}
		logger.Debugf("Auto-discovered role = %s", role)

		switch role {
		case "leader":
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
			err = leader(kubeletKSMGrouper, ksmClient, integration, args.ClusterName, logger)
			if err != nil {
				logger.Panic(err)
			}
		case "follower":
			// todo fix pointers indirection stuff
			kubeletKSMGrouper := data.NewKubeletKSMGrouper(
				kubeletClient,
				ksmClient,
				mergeableObjectsQueries,
				ksmMergeableSpecs,
				logger,
			)

			err = follower(kubeletKSMGrouper, integration, args.ClusterName, logger)
			if err != nil {
				logger.Panic(err)
			}
		}
	}

	err = integration.Publish()
	if err != nil {
		logger.Panic(err)
	}
}
