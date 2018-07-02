package helm

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// InstallRelease installs a chart release
func InstallRelease(ctx context.Context, path string, config ...string) ([]byte, error) {
	args := []string{
		"install",
		path,
		"--wait",
	}

	if len(config) > 0 {
		args = append(args, "--set", strings.Join(config, ","))
	}

	c := exec.CommandContext(ctx, "helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s - %s", err, o)
	}

	return o, nil
}

// DeleteRelease deletes a chart release
func DeleteRelease(ctx context.Context, release string) error {
	args := []string{
		"delete",
		release,
	}

	c := exec.CommandContext(ctx, "helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	return nil
}

// Init installs Tiller (the Helm server-side component) onto your cluster
func Init(arg ...string) error {
	args := append([]string{
		"init",
		"--wait",
	}, arg...)

	c := exec.Command("helm", args...)
	o, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err, o)
	}

	return nil
}
