package helm

import (
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
