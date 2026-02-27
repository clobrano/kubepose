package actions

import (
	"fmt"
	"strings"
)

// Scale scales a deployment, replicaset, or statefulset to the specified replica count
func Scale(ctx *ActionContext, replicas int) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resource specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"scale", ctx.ResourceType, name, fmt.Sprintf("--replicas=%d", replicas)}
		if ctx.Namespace != "" {
			args = append(args, "-n", ctx.Namespace)
		}

		stdout, stderr, err := ctx.Kubectl.Execute(args...)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("%s: %s", name, strings.TrimSpace(stderr)))
		} else {
			outputs = append(outputs, strings.TrimSpace(stdout))
		}
	}

	return strings.Join(outputs, "\n"), nil
}

// GetCurrentReplicas returns the current replica count for a resource
func GetCurrentReplicas(ctx *ActionContext) (int, error) {
	if len(ctx.Names) == 0 {
		return 0, fmt.Errorf("no resource specified")
	}

	name := ctx.Names[0]
	args := []string{"get", ctx.ResourceType, name, "-o", "jsonpath={.spec.replicas}"}
	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	stdout, stderr, err := ctx.Kubectl.Execute(args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", strings.TrimSpace(stderr), err)
	}

	var replicas int
	_, err = fmt.Sscanf(strings.TrimSpace(stdout), "%d", &replicas)
	if err != nil {
		return 0, fmt.Errorf("failed to parse replica count: %w", err)
	}

	return replicas, nil
}
