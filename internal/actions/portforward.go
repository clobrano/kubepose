package actions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// PortForwardOptions contains options for port forwarding
type PortForwardOptions struct {
	LocalPort  int
	RemotePort int
	Address    string // Local address to bind to, defaults to localhost
}

// PortForward starts kubectl port-forward in the foreground
// This function blocks until the port-forward is terminated (Ctrl+C)
func PortForward(ctx *ActionContext, opts PortForwardOptions) error {
	if len(ctx.Names) == 0 {
		return fmt.Errorf("no resource specified")
	}

	name := ctx.Names[0]

	args := []string{"port-forward"}

	// Resource can be pod, service, or deployment
	args = append(args, fmt.Sprintf("%s/%s", ctx.ResourceType, name))

	if ctx.Namespace != "" {
		args = append(args, "-n", ctx.Namespace)
	}

	if opts.Address != "" {
		args = append(args, "--address", opts.Address)
	}

	// Port mapping
	portMapping := fmt.Sprintf("%d:%d", opts.LocalPort, opts.RemotePort)
	args = append(args, portMapping)

	kubectlBin := ctx.Kubectl.BinaryPath()
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// PortForwardBackground starts kubectl port-forward in the background
// Returns a cancel function to stop the port-forward
func PortForwardBackground(actionCtx *ActionContext, opts PortForwardOptions) (cancel context.CancelFunc, err error) {
	if len(actionCtx.Names) == 0 {
		return nil, fmt.Errorf("no resource specified")
	}

	name := actionCtx.Names[0]

	args := []string{"port-forward"}
	args = append(args, fmt.Sprintf("%s/%s", actionCtx.ResourceType, name))

	if actionCtx.Namespace != "" {
		args = append(args, "-n", actionCtx.Namespace)
	}

	if opts.Address != "" {
		args = append(args, "--address", opts.Address)
	}

	portMapping := fmt.Sprintf("%d:%d", opts.LocalPort, opts.RemotePort)
	args = append(args, portMapping)

	ctx, cancel := context.WithCancel(context.Background())

	kubectlBin := actionCtx.Kubectl.BinaryPath()
	cmd := exec.CommandContext(ctx, kubectlBin, args...)

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	// Start a goroutine to wait for the command and clean up
	go func() {
		_ = cmd.Wait()
	}()

	return cancel, nil
}
