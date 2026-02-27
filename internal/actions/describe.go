package actions

import (
	"fmt"
	"strings"
)

// Describe executes kubectl describe for the given resources
// For multiple resources, outputs are concatenated with separators
func Describe(ctx *ActionContext) (string, error) {
	if len(ctx.Names) == 0 {
		return "", fmt.Errorf("no resources specified")
	}

	var outputs []string

	for _, name := range ctx.Names {
		args := []string{"describe", ctx.ResourceType, name}
		if ctx.Namespace != "" {
			args = append(args, "-n", ctx.Namespace)
		}

		stdout, _, err := ctx.Kubectl.Execute(args...)
		if err != nil {
			// Include error in output but continue with other resources
			outputs = append(outputs, fmt.Sprintf("=== Error describing %s/%s ===\n%v", ctx.ResourceType, name, err))
			continue
		}

		if len(ctx.Names) > 1 {
			outputs = append(outputs, fmt.Sprintf("=== %s/%s ===\n%s", ctx.ResourceType, name, stdout))
		} else {
			outputs = append(outputs, stdout)
		}
	}

	return strings.Join(outputs, "\n\n"), nil
}
