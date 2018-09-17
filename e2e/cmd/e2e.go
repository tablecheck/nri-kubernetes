package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/gcp"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/helm"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/jsonschema"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/k8s"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/timer"
	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/sirupsen/logrus"
)

var cliArgs = struct {
	NrChartPath                string `default:"e2e/charts/newrelic-infrastructure-k8s-e2e" help:"Path to the newrelic-infrastructure-k8s-e2e chart"`
	SchemasDirectory           string `default:"e2e/schema" help:"Directory where JSON schema files are defined"`
	IntegrationImageTag        string `default:"1.0.0" help:"Integration image tag"`
	IntegrationImageRepository string `default:"newrelic/infrastructure-k8s" help:"Integration image repository"`
	Rbac                       bool   `default:"false" help:"Enable rbac"`
	ClusterName                string `help:"Identifier of your cluster. You could use it later to filter data in your New Relic account"`
	NrLicenseKey               string `help:"New Relic account license key"`
	Verbose                    bool   `default:"false" help:"When enabled, more detailed output will be printed"`
	CollectorURL               string `default:"https://staging-infra-api.newrelic.com" help:"New Relic backend collector url"`
	Context                    string `default:"" help:"Kubernetes context"`
	CleanBeforeRun             bool   `default:"true" help:"Clean the cluster before running the tests"`
	FailFast                   bool   `default:"false" help:"Fail the whole suit on the first failure"`
}{}

const (
	nrLabel     = "name=newrelic-infra"
	namespace   = "default"
	nrContainer = "newrelic-infra"
)

func scenarios(integrationImageRepository string, integrationImageTag string, rbac bool) []string {
	return []string{
		s(rbac, integrationImageRepository, integrationImageTag, "v1.1.0", false),
		s(rbac, integrationImageRepository, integrationImageTag, "v1.1.0", true),
		s(rbac, integrationImageRepository, integrationImageTag, "v1.2.0", false),
		s(rbac, integrationImageRepository, integrationImageTag, "v1.2.0", true),
		s(rbac, integrationImageRepository, integrationImageTag, "v1.3.0", false),
		s(rbac, integrationImageRepository, integrationImageTag, "v1.3.0", true),
	}
}

func s(rbac bool, integrationImageRepository, integrationImageTag, ksmVersion string, twoKSMInstances bool) string {
	str := fmt.Sprintf("rbac=%v,ksm-instance-one.rbac.create=%v,ksm-instance-one.image.tag=%s,daemonset.image.repository=%s,daemonset.image.tag=%s", rbac, rbac, ksmVersion, integrationImageRepository, integrationImageTag)
	if twoKSMInstances {
		return fmt.Sprintf("%s,ksm-instance-two.rbac.create=%v,ksm-instance-two.image.tag=%s,two-ksm-instances=true", str, rbac, ksmVersion)
	}

	return str
}

type integrationData struct {
	role    string
	podName string
	stdOut  []byte
	stdErr  []byte
	err     error
}

type executionErr struct {
	errs []error
}

// Error implements Error interface
func (err executionErr) Error() string {
	var errsStr string
	for _, e := range err.errs {
		errsStr += fmt.Sprintf("%s\n", e)
	}

	return errsStr
}

func execIntegration(podName string, dataChannel chan integrationData, wg *sync.WaitGroup, c *k8s.Client, logger *logrus.Logger) {
	defer timer.Track(time.Now(), fmt.Sprintf("execIntegration func for pod %s", podName), logger)
	defer wg.Done()
	d := integrationData{
		podName: podName,
	}

	output, err := c.PodExec(namespace, podName, nrContainer, "/var/db/newrelic-infra/newrelic-integrations/bin/nr-kubernetes", "-timeout=15000", "-verbose")
	if err != nil {
		d.err = err
		dataChannel <- d
		return
	}

	re, err := regexp.Compile("Auto-discovered role = (\\w*)")
	if err != nil {
		d.err = fmt.Errorf("cannot compile regex and determine role for pod %s, err: %v", podName, err)
		dataChannel <- d
		return
	}

	matches := re.FindStringSubmatch(output.Stderr.String())
	role := matches[1]
	if role == "" {
		d.err = fmt.Errorf("cannot find a role for pod %s", podName)
		dataChannel <- d
		return
	}

	d.role = role
	d.stdOut = output.Stdout.Bytes()
	d.stdErr = output.Stderr.Bytes()

	dataChannel <- d
}

func main() {
	err := args.SetupArgs(&cliArgs)
	if err != nil {
		panic(err.Error())
	}

	if cliArgs.NrLicenseKey == "" || cliArgs.ClusterName == "" {
		panic("license key and cluster name are required args")
	}
	logger := log.New(cliArgs.Verbose)

	c, err := k8s.NewClient(cliArgs.Context)
	if err != nil {
		panic(err.Error())
	}

	err = initHelm(c, cliArgs.Rbac, logger)
	if err != nil {
		panic(err.Error())
	}

	logger.Infof("Executing tests in %q cluster. K8s version: %s", c.Config.Host, c.ServerVersion())

	if cliArgs.CleanBeforeRun {
		logger.Infof("Cleaning cluster")
		err := helm.DeleteAllReleases(cliArgs.Context, logger)
		if err != nil {
			panic(err.Error())
		}
	}

	// TODO
	var errs []error
	ctx := context.TODO()
	for _, s := range scenarios(cliArgs.IntegrationImageRepository, cliArgs.IntegrationImageTag, cliArgs.Rbac) {
		logger.Infof("Scenario: %q", s)
		err := executeScenario(ctx, s, c, logger)
		if err != nil {
			if cliArgs.FailFast {
				logger.Info("Finishing execution because 'FailFast' is true")
				logger.Fatal(err.Error())
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		logger.Debugf("errors collected from all scenarios")
		for _, err := range errs {
			logger.Errorf(err.Error())
		}
	} else {
		logger.Infof("OK")
	}
}

func initHelm(c *k8s.Client, rbac bool, logger *logrus.Logger) error {
	var initArgs []string
	if rbac {
		ns := "kube-system"
		n := "tiller"
		sa, err := c.ServiceAccount(ns, n)
		if err != nil {
			sa, err = c.CreateServiceAccount(ns, n)
			if err != nil {
				return err
			}
		}
		_, err = c.ClusterRoleBinding(n)
		if err != nil {
			cr, err := c.ClusterRole("cluster-admin")
			if err != nil {
				return err
			}
			_, err = c.CreateClusterRoleBinding(n, sa, cr)
			if err != nil {
				return err
			}
		}
		initArgs = []string{"--service-account", n}
	}

	err := helm.Init(
		cliArgs.Context,
		logger,
		initArgs...,
	)

	if err != nil {
		return err
	}

	return helm.DependencyBuild(cliArgs.Context, cliArgs.NrChartPath, logger)
}

func executeScenario(ctx context.Context, scenario string, c *k8s.Client, logger *logrus.Logger) error {
	defer timer.Track(time.Now(), fmt.Sprintf("executeScenario func for %s", scenario), logger)

	releaseName, err := installRelease(ctx, scenario, logger)
	if err != nil {
		return err
	}

	defer helm.DeleteRelease(releaseName, cliArgs.Context, logger) // nolint: errcheck

	// At least one of kube-state-metrics pods needs to be ready to enter to the newrelic-infra pod and execute the integration.
	// If the kube-state-metrics pod is not ready, then metrics from replicaset, namespace and deployment will not be populate and JSON schemas will fail.
	timeStartKSMWaiting := time.Now()
	tickerRetry := time.NewTicker(2 * time.Second)
	tickerTimeout := time.NewTicker(2 * time.Minute)
KSMLoop:
	for {
		select {
		case <-tickerRetry.C:
			ksmPodList, err := c.PodsListByLabels(namespace, []string{"app=kube-state-metrics"})
			if err != nil {
				return err
			}
			if len(ksmPodList.Items) != 0 && ksmPodList.Items[0].Status.Phase == "Running" {
				for _, con := range ksmPodList.Items[0].Status.Conditions {
					logger.Debugf("Waiting for kube-state-metrics pod to be ready, current condition: %s - %s", con.Type, con.Status)

					if con.Type == "Ready" && con.Status == "True" {
						break KSMLoop
					}
				}
			}
		case <-tickerTimeout.C:
			tickerRetry.Stop()
			tickerTimeout.Stop()
			return errors.New("kube-state-metrics pod is not ready, reaching timeout")
		}
	}
	elapsed := time.Since(timeStartKSMWaiting)
	logger.Debugf("Waiting for KSM to be ready took %s", elapsed)

	var execErr executionErr
	var lcount int
	var fcount int
	var retriesNR int
NRLoop:
	for {
		podsList, err := c.PodsListByLabels(namespace, []string{nrLabel})
		if err != nil {
			return err
		}

		output := make(map[string]integrationData)
		dataChannel := make(chan integrationData)

		var wg sync.WaitGroup
		wg.Add(len(podsList.Items))
		go func() {
			wg.Wait()
			close(dataChannel)
		}()

		for _, p := range podsList.Items {
			logger.Debugf("Executing integration inside pod: %s", p.Name)
			go execIntegration(p.Name, dataChannel, &wg, c, logger)
		}

		for d := range dataChannel {
			if d.err != nil {
				return fmt.Errorf("scenario: %s. %s", scenario, d.err.Error())
			}
			output[d.podName] = d
		}

		lcount = 0
		fcount = 0
		leaderMap := jsonschema.EventTypeToSchemaFilepath{
			"K8sReplicasetSample": "replicaset.json",
			"K8sNamespaceSample":  "namespace.json",
			"K8sDeploymentSample": "deployment.json",
			"K8sPodSample":        "pod.json",
			"K8sContainerSample":  "container.json",
			"K8sNodeSample":       "node.json",
			"K8sVolumeSample":     "volume.json",
		}

		followerMap := jsonschema.EventTypeToSchemaFilepath{
			"K8sPodSample":       leaderMap["K8sPodSample"],
			"K8sContainerSample": leaderMap["K8sContainerSample"],
			"K8sNodeSample":      leaderMap["K8sNodeSample"],
			"K8sVolumeSample":    leaderMap["K8sVolumeSample"],
		}
	OutputLoop:
		for podName, o := range output {
			var m jsonschema.EventTypeToSchemaFilepath
			switch o.role {
			case "leader":
				lcount++
				m = leaderMap
			case "follower":
				fcount++
				m = followerMap
			}
			err := jsonschema.Match(o.stdOut, m, cliArgs.SchemasDirectory)
			if err != nil {
				errStr := fmt.Sprintf("received error during execution of scenario %q for pod %s with role %s:\n%s", scenario, podName, o.role, err)
				select {
				case <-tickerRetry.C:
					logger.Debugf("------ Retrying due to %s ------ ", errStr)
					retriesNR++
					continue NRLoop
				case <-tickerTimeout.C:
					tickerRetry.Stop()
					tickerTimeout.Stop()
					if cliArgs.Verbose {
						errStr = errStr + fmt.Sprintf("\nStdErr:\n%s\nStdOut:\n%s", string(o.stdErr), string(o.stdOut))
					}
					execErr.errs = append(execErr.errs, errors.New(errStr))
					break OutputLoop
				}
			}
		}
		if len(execErr.errs) == 0 {
			logger.Info("output of the integration is valid with all JSON schemas")
			break
		}
		return fmt.Errorf("failure during JSON schema validation, retries limit reached, number of retries: %d,\nlast error: %s", retriesNR, execErr)
	}

	nodes, err := c.NodesList()
	if err != nil {
		execErr.errs = append(execErr.errs, err)
	}

	if lcount+fcount != len(nodes.Items) {
		execErr.errs = append(execErr.errs, fmt.Errorf("%d nodes were found, but got: %d", lcount+fcount, len(nodes.Items)))
	}

	if lcount != 1 {
		execErr.errs = append(execErr.errs, fmt.Errorf("%d pod leaders were found, but only 1 was expected", lcount))
	}

	if len(execErr.errs) > 0 {
		return execErr
	}

	return nil
}

func installRelease(ctx context.Context, scenario string, logger *logrus.Logger) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	options := strings.Split(scenario, ",")
	options = append(options,
		fmt.Sprintf("integration.k8sClusterName=%s", cliArgs.ClusterName),
		fmt.Sprintf("integration.newRelicLicenseKey=%s", cliArgs.NrLicenseKey),
		"integration.verbose=true",
		fmt.Sprintf("integration.collectorURL=%s", cliArgs.CollectorURL),
	)

	o, err := helm.InstallRelease(filepath.Join(dir, cliArgs.NrChartPath), cliArgs.Context, logger, options...)
	if err != nil {
		return "", err
	}

	r := bufio.NewReader(bytes.NewReader(o))
	v, _, err := r.ReadLine()
	if err != nil {
		return "", err
	}

	releaseName := bytes.TrimPrefix(v, []byte("NAME:   "))

	return string(releaseName), nil
}
