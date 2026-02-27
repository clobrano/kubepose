package actions

import (
	"fmt"
	"strings"
)

// RolloutRestart triggers a rollout restart for a deployment, daemonset, or statefulset
func RolloutRestart(ctx *ActionContext) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resource specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"rollout", "restart", fmt.Sprintf("%s/%s", ctx.ResourceType, name)}
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

// RolloutStatus returns the rollout status for a resource
func RolloutStatus(ctx *ActionContext) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resource specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"rollout", "status", fmt.Sprintf("%s/%s", ctx.ResourceType, name)}
		if ctx.Namespace != "" {
			args = append(args, "-n", ctx.Namespace)
		}

		stdout, stderr, err := ctx.Kubectl.Execute(args...)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("=== %s ===\n%s", name, strings.TrimSpace(stderr)))
		} else {
			if len(ctx.Names) > 1 {
				outputs = append(outputs, fmt.Sprintf("=== %s ===\n%s", name, strings.TrimSpace(stdout)))
			} else {
				outputs = append(outputs, strings.TrimSpace(stdout))
			}
		}
	}

	return strings.Join(outputs, "\n\n"), nil
}

// RolloutHistory returns the rollout history for a resource
func RolloutHistory(ctx *ActionContext) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resource specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"rollout", "history", fmt.Sprintf("%s/%s", ctx.ResourceType, name)}
		if ctx.Namespace != "" {
			args = append(args, "-n", ctx.Namespace)
		}

		stdout, stderr, err := ctx.Kubectl.Execute(args...)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("=== %s ===\n%s", name, strings.TrimSpace(stderr)))
		} else {
			if len(ctx.Names) > 1 {
				outputs = append(outputs, fmt.Sprintf("=== %s ===\n%s", name, strings.TrimSpace(stdout)))
			} else {
				outputs = append(outputs, strings.TrimSpace(stdout))
			}
		}
	}

	return strings.Join(outputs, "\n\n"), nil
}

// RolloutUndo rolls back to a previous revision
func RolloutUndo(ctx *ActionContext, revision int) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resource specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"rollout", "undo", fmt.Sprintf("%s/%s", ctx.ResourceType, name)}
		if revision > 0 {
			args = append(args, fmt.Sprintf("--to-revision=%d", revision))
		}
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
