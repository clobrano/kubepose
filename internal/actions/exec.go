package actions

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecOptions contains options for the exec action
type ExecOptions struct {
	Container string
	Command   []string // If empty, defaults to /bin/sh
	TTY       bool
	Stdin     bool
}

// DefaultExecOptions returns default options for interactive shell
func DefaultExecOptions() ExecOptions {
	return ExecOptions{
		Command: []string{"/bin/sh"},
		TTY:     true,
		Stdin:   true,
	}
}

// Exec runs kubectl exec to get a shell in the pod
// This function suspends the TUI and takes over the terminal
func Exec(ctx *ActionContext, opts ExecOptions) error {
	if len(ctx.Names) == 0 {
		return fmt.Errorf("no pod specified")
	}

	podName := ctx.Names[0]

	args := []string{"exec"}

	if opts.Stdin {
		args = append(args, "-i")
	}
	if opts.TTY {
		args = append(args, "-t")
	}

	args = append(args, podName)

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	if opts.Container != "" {
		args = append(args, "-c", opts.Container)
	}

	// Add command separator and command
	args = append(args, "--")
	if len(opts.Command) > 0 {
		args = append(args, opts.Command...)
	} else {
		args = append(args, "/bin/sh")
	}

	kubectlBin := ctx.Kubectl.BinaryPath()
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
