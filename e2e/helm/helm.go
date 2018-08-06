package helm

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// InstallRelease installs a chart release
func InstallRelease(path, context string, config ...string) ([]byte, error) {
	args := []string{
		"install",
		path,
		"--wait",
	}

	if len(config) > 0 {
		args = append(args, "--set", strings.Join(config, ","))
	}

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s - %s", err, o)
	}

	return o, nil
}

// DeleteRelease deletes a chart release
func DeleteRelease(release, context string) error {
	args := []string{
		"delete",
		release,
	}

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	return nil
}

// DeleteAllReleases deletes all chart releases
func DeleteAllReleases(context string) error {
	args := []string{
		"list",
		"--short",
	}

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	scanner := bufio.NewScanner(bytes.NewReader(o))
	for scanner.Scan() {
		err := DeleteRelease(scanner.Text(), context)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

// Init installs Tiller (the Helm server-side component) onto your cluster
func Init(context string, arg ...string) error {
	args := append([]string{
		"init",
		"--wait",
	}, arg...)

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	return nil
}

// DependencyBuild builds the dependencies for the e2e chart
func DependencyBuild(context, chart string) error {
	args := []string{
		"dependency",
		"build",
		chart,
	}

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	return nil
}
