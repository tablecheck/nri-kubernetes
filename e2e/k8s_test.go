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

	"time"

	"strings"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/helm"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/jsonschema"
	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var cliArgs = struct {
	NrChartPath  string `default:"../deploy/helm/newrelic-infrastructure-k8s-e2e",help:"Path to the newrelic-infrastructure-k8s-e2e chart"`
	ClusterName  string `help:"Identifier of your cluster. You could use it later to filter data in your New Relic account"`
	NrLicenseKey string `help:"New Relic account license key"`
	Verbose      int    `default:"0",help:"When enabled, more detailed output will be printed"`
	CollectorURL string `default:"https://staging-infra-api.newrelic.com",help:"New Relic backend collector url"`
}{}

const (
	nrLabelKey   = "name"
	nrLabelValue = "newrelic-infra"
	namespace    = "default"
	nrContainer  = "newrelic-infra"
)

var scenarios = []string{
	s(false, "v1.1.0", false),
	s(false, "v1.1.0", true),
	s(false, "v1.2.0", false),
	s(false, "v1.2.0", true),
	s(false, "v1.3.0", false),
	s(false, "v1.3.0", true),
}

func s(rbac bool, ksmVersion string, twoKSMInstances bool) string {
	str := fmt.Sprintf("rbac=%v,ksm-instance-one.rbac.create=%v,ksm-instance-one.image.tag=%s", rbac, rbac, ksmVersion)
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

func getNRPods(clientset *kubernetes.Clientset, config *rest.Config) ([]string, error) {
	sv, err := clientset.ServerVersion()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Executing our integration in %q cluster. K8s version: %s\n", config.Host, sv.String())
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", nrLabelKey, nrLabelValue),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving pod list by label %s = %s: %v", nrLabelKey, nrLabelValue, err)
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("pods not found by label: %s=%s", nrLabelKey, nrLabelValue)
	}
	podsName := make([]string, 0)
	for i := 0; i < len(pods.Items); i++ {
		podsName = append(podsName, pods.Items[i].Name)
	}
	return podsName, nil
}

type execOutput struct {
	execOut bytes.Buffer
	execErr bytes.Buffer
}

func podExec(clientset *kubernetes.Clientset, config *rest.Config, podName string, command ...string) (execOutput, error) {
	execReq := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", nrContainer).
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false")

	for _, c := range command {
		execReq.Param("command", c)
	}

	var output execOutput

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execReq.URL())
	if err != nil {
		return output, fmt.Errorf("failed to init executor for pod %s: %v", podName, err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &output.execOut,
		Stderr: &output.execErr,
	})

	if err != nil {
		return output, fmt.Errorf("could not execute command inside pod %s: %v. Output:\n\n%v", podName, err, output.execErr.String())
	}

	return output, nil
}

func execIntegration(clientset *kubernetes.Clientset, config *rest.Config, podName string, dataChannel chan integrationData, wg *sync.WaitGroup) {
	defer wg.Done()
	d := integrationData{
		podName: podName,
	}

	output, err := podExec(clientset, config, podName, "/var/db/newrelic-infra/newrelic-integrations/bin/nr-kubernetes", "-timeout=15000", "-verbose")
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

	matches := re.FindStringSubmatch(output.execErr.String())
	role := matches[1]
	if role == "" {
		d.err = fmt.Errorf("cannot find a role for pod %s", podName)
		dataChannel <- d
		return
	}

	d.role = role
	d.stdOut = output.execOut.Bytes()
	d.stdErr = output.execErr.Bytes()

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

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	// TODO
	ctx := context.TODO()
	for _, s := range scenarios {
		fmt.Printf("Executing scenario %q\n", s)
		err := executeScenario(ctx, config, s)
		assert.NoError(t, err)
	}
}

func executeScenario(ctx context.Context, config *rest.Config, scenario string) error {
	releaseName, err := installRelease(ctx, scenario)
	if err != nil {
		return err
	}

	defer helm.DeleteRelease(ctx, releaseName) // nolint: errcheck

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	podsName, err := getNRPods(clientset, config)
	if err != nil {
		return err
	}

	output := make(map[string]integrationData)

	dataChannel := make(chan integrationData)

	var wg sync.WaitGroup
	wg.Add(len(podsName))
	go func() {
		wg.Wait()
		close(dataChannel)
	}()

	for _, pName := range podsName {
		fmt.Printf("Scenario: %s. Executing integration inside pod: %s\n", scenario, pName)
		go execIntegration(clientset, config, pName, dataChannel, &wg)
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
			if cliArgs.Verbose == 1 {
				errStr = errStr + fmt.Sprintf("\nStdErr:\n%s\nStdOut:\n%s", string(o.stdErr), string(o.stdOut))
			}

			execErr.errs = append(execErr.errs, errors.New(errStr))
		}
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		execErr.errs = append(execErr.errs, err)
	}

	if lcount+fcount != len(nodes.Items) {
		execErr.errs = append(execErr.errs, fmt.Errorf("%d nodes were found, but got: %d", lcount+fcount, len(nodes.Items)))
	}

	if lcount != 1 {
		execErr.errs = append(execErr.errs, fmt.Errorf("%d pod leaders were found, but only 1 was expected", lcount))
	}

	return execErr
}

func installRelease(ctx context.Context, scenario string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// ensuring helm (tiller pod) is installed
	err = helm.Init(ctx)
	if err != nil {
		return "", err
	}

	options := strings.Split(scenario, ",")
	options = append(options,
		fmt.Sprintf("integration.k8sClusterName=%s", cliArgs.ClusterName),
		fmt.Sprintf("integration.newRelicLicenseKey=%s", cliArgs.NrLicenseKey),
		fmt.Sprintf("integration.verbose=%d", cliArgs.Verbose),
		fmt.Sprintf("integration.collectorURL=%s", cliArgs.CollectorURL),
	)

	o, err := helm.InstallRelease(ctx, filepath.Join(dir, cliArgs.NrChartPath), options...)
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
