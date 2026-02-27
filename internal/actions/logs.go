package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// LogsOptions contains options for the logs action
type LogsOptions struct {
	Follow    bool
	Container string
	TailLines int
	Previous  bool
}

// GetContainers returns the list of containers in a pod
func GetContainers(ctx *ActionContext) ([]string, error) {
	if len(ctx.Names) == 0 {
		return nil, fmt.Errorf("no pod specified")
	}

	podName := ctx.Names[0]
	args := []string{"get", "pod", podName, "-o", "json"}
	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	stdout, _, err := ctx.Kubectl.Execute(args...)
	if err != nil {
		return nil, err
	}

	// Parse JSON to extract container names
	var pod struct {
		Spec struct {
			Containers     []struct{ Name string } `json:"containers"`
			InitContainers []struct{ Name string } `json:"initContainers"`
		} `json:"spec"`
	}

	if err := json.Unmarshal([]byte(stdout), &pod); err != nil {
		return nil, fmt.Errorf("failed to parse pod spec: %w", err)
	}

	var containers []string
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c.Name)
	}
	// Also include init containers
	for _, c := range pod.Spec.InitContainers {
		containers = append(containers, fmt.Sprintf("%s (init)", c.Name))
	}

	return containers, nil
}

// Logs returns the logs for a pod (non-follow mode)
func Logs(ctx *ActionContext, opts LogsOptions) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no pod specified")
	}

	podName := ctx.Names[0]
	args := []string{"logs", podName}

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	if opts.Container != "" {
		args = append(args, "-c", opts.Container)
	}

	if opts.TailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.TailLines))
	}

	if opts.Previous {
		args = append(args, "--previous")
	}

	stdout, _, err := ctx.Kubectl.Execute(args...)
	if err != nil {
		return "", err
	}

	return stdout, nil
}

// LogsWithPager runs kubectl logs piped to a pager (less)
// This function blocks until the pager exits
func LogsWithPager(ctx *ActionContext, opts LogsOptions, pager string) error {
	if len(ctx.Names) == 0 {
		return fmt.Errorf("no pod specified")
	}

	if pager == "" {
		pager = "less"
	}

	podName := ctx.Names[0]
	args := []string{"logs", podName}

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	if opts.Container != "" {
		args = append(args, "-c", opts.Container)
	}

	if opts.TailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.TailLines))
	}

	if opts.Previous {
		args = append(args, "--previous")
	}

	// Get logs output
	stdout, _, err := ctx.Kubectl.Execute(args...)
	if err != nil {
		return err
	}

	// Pipe to pager
	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(stdout)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// LogsFollow runs kubectl logs -f, which streams logs until Ctrl+C
// This function takes over the terminal
func LogsFollow(ctx *ActionContext, opts LogsOptions) error {
	if len(ctx.Names) == 0 {
		return fmt.Errorf("no pod specified")
	}

	podName := ctx.Names[0]
	args := []string{"logs", "-f", podName}

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	if opts.Container != "" {
		args = append(args, "-c", opts.Container)
	}

	if opts.TailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.TailLines))
	}

	kubectlBin := ctx.Kubectl.BinaryPath()
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
