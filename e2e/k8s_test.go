package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"bufio"

	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/helm"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/jsonschema"
	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	// This package includes the GKE auth provider automatically by import the package (init function does the job)
	"time"

	_ "github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/gcp"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/k8s"
)

var cliArgs = struct {
	NrChartPath                string `default:"../deploy/helm/newrelic-infrastructure-k8s-e2e",help:"Path to the newrelic-infrastructure-k8s-e2e chart"`
	IntegrationImageTag        string `default:"1.0.0",help:"Integration image tag"`
	IntegrationImageRepository string `default:"newrelic/infrastructure-k8s",help:"Integration image repository"`
	Rbac                       bool   `default:"false",help:"Enable rbac"`
	ClusterName                string `help:"Identifier of your cluster. You could use it later to filter data in your New Relic account"`
	NrLicenseKey               string `help:"New Relic account license key"`
	Verbose                    bool   `default:"false",help:"When enabled, more detailed output will be printed"`
	CollectorURL               string `default:"https://staging-infra-api.newrelic.com",help:"New Relic backend collector url"`
	Context                    string `default:"",help:"Kubernetes context"`
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

func execIntegration(podName string, dataChannel chan integrationData, wg *sync.WaitGroup, c *k8s.Client) {
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

func TestBasic(t *testing.T) {
	if os.Getenv("RUN_TESTS") == "" {
		t.Skip("Flag RUN_TESTS is not specified, skipping tests")
	}

	err := args.SetupArgs(&cliArgs)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if cliArgs.NrLicenseKey == "" || cliArgs.ClusterName == "" {
		assert.FailNow(t, "license key and cluster name are required args")
	}

	c, err := k8s.NewClient(cliArgs.Context)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = initHelm(c, cliArgs.Rbac)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	fmt.Printf("Executing tests in %q cluster. K8s version: %s\n", c.Config.Host, c.ServerVersion())

	// TODO
	ctx := context.TODO()
	for _, s := range scenarios(cliArgs.IntegrationImageRepository, cliArgs.IntegrationImageTag, cliArgs.Rbac) {
		fmt.Printf("Scenario %q\n", s)
		err := executeScenario(ctx, s, c)
		assert.NoError(t, err)
	}
}

func initHelm(c *k8s.Client, rbac bool) error {
	if !rbac {
		return helm.Init(cliArgs.Context)
	}
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
	return helm.Init(
		cliArgs.Context,
		[]string{"--service-account", n}...,
	)
}

func executeScenario(ctx context.Context, scenario string, c *k8s.Client) error {
	releaseName, err := installRelease(ctx, scenario)
	if err != nil {
		return err
	}

	defer helm.DeleteRelease(releaseName, cliArgs.Context) // nolint: errcheck

	// Waiting until all pods have consumed cpu, memory enough and are scheduled. Otherwise some metrics will be missing.
	// TODO Find a better way for generating load on all the pods rather than this time sleep.
	time.Sleep(2 * time.Minute)

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
		fmt.Printf("Executing integration inside pod: %s\n", p.Name)
		go execIntegration(p.Name, dataChannel, &wg, c)
	}

	for d := range dataChannel {
		if d.err != nil {
			return fmt.Errorf("scenario: %s. %s", scenario, d.err.Error())
		}
		output[d.podName] = d
	}

	leaderMap := jsonschema.EventTypeToSchemaFilepath{
		"K8sReplicasetSample": "schema/replicaset.json",
		"K8sNamespaceSample":  "schema/namespace.json",
		"K8sDeploymentSample": "schema/deployment.json",
		"K8sPodSample":        "schema/pod.json",
		"K8sContainerSample":  "schema/container.json",
		"K8sNodeSample":       "schema/node.json",
	}

	followerMap := jsonschema.EventTypeToSchemaFilepath{
		"K8sPodSample":       leaderMap["K8sPodSample"],
		"K8sContainerSample": leaderMap["K8sContainerSample"],
		"K8sNodeSample":      leaderMap["K8sNodeSample"],
	}

	var execErr executionErr
	var lcount int
	var fcount int

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

		err := jsonschema.Match(o.stdOut, m)
		if err != nil {
			errStr := fmt.Sprintf("\n------ scenario: %s. %s pod %s ------\n%s", scenario, o.role, podName, err)
			if cliArgs.Verbose {
				errStr = errStr + fmt.Sprintf("\nStdErr:\n%s\nStdOut:\n%s", string(o.stdErr), string(o.stdOut))
			}

			execErr.errs = append(execErr.errs, errors.New(errStr))
		}
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

func installRelease(ctx context.Context, scenario string) (string, error) {
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

	o, err := helm.InstallRelease(filepath.Join(dir, cliArgs.NrChartPath), cliArgs.Context, options...)
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
