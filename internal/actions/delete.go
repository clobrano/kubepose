package actions

import (
	"fmt"
	"strings"
)

// DeleteResult contains the result of a delete operation
type DeleteResult struct {
	Name    string
	Success bool
	Message string
}

// Delete executes kubectl delete for the given resources
// Returns results for each resource
func Delete(ctx *ActionContext) ([]DeleteResult, error) {
	if len(ctx.Names) == 0 {
		return nil, fmt.Errorf("no resources specified")
	}

	var results []DeleteResult

	for _, name := range ctx.Names {
		args := []string{"delete", ctx.ResourceType, name}
		if ctx.Namespace != "" {
			args = append(args, "-n", ctx.Namespace)
		}

		stdout, stderr, err := ctx.Kubectl.Execute(args...)
		result := DeleteResult{Name: name}

		if err != nil {
			result.Success = false
			result.Message = strings.TrimSpace(stderr)
			if result.Message == "" {
				result.Message = err.Error()
			}
		} else {
			result.Success = true
			result.Message = strings.TrimSpace(stdout)
		}

		results = append(results, result)
	}

	return results, nil
}

// DeleteDryRun performs a dry-run delete to verify the operation
func DeleteDryRun(ctx *ActionContext) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resources specified")
	}

	args := []string{"delete", ctx.ResourceType}
	args = append(args, ctx.Names...)
	args = append(args, "--dry-run=client")

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	stdout, stderr, err := ctx.Kubectl.Execute(args...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", strings.TrimSpace(stderr), err)
	}

	return stdout, nil
}
