package actions

import (
	"fmt"
	"os"
	"os/exec"
)

// Edit opens the resource in the default editor via kubectl edit
// This function suspends the TUI and takes over the terminal
func Edit(ctx *ActionContext) error {
	if len(ctx.Names) == 0 {
		return fmt.Errorf("no resource specified")
	}

	// Only edit one resource at a time
	name := ctx.Names[0]

	args := []string{"edit", ctx.ResourceType, name}
	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	kubectlBin := ctx.Kubectl.BinaryPath()
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
