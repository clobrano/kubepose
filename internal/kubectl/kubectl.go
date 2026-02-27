package kubectl

import (
	"bytes"
	"os/exec"
	"strings"
)

// Kubectl wraps kubectl binary execution
type Kubectl struct {
	binaryPath string
}

// NewKubectl creates a new Kubectl instance with the specified binary path
func NewKubectl(binaryPath string) *Kubectl {
	if binaryPath == "" {
		binaryPath = "kubectl"
	}
	return &Kubectl{
		binaryPath: binaryPath,
	}
}

// BinaryPath returns the path to the kubectl binary
func (k *Kubectl) BinaryPath() string {
	return k.binaryPath
}

// Execute runs a kubectl command with the given arguments and returns stdout, stderr, and error
func (k *Kubectl) Execute(args ...string) (string, string, error) {
	cmd := exec.Command(k.binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ExecuteRaw runs a kubectl command from a raw command string and returns the output
func (k *Kubectl) ExecuteRaw(command string) (string, error) {
	args := parseCommandArgs(command)
	stdout, stderr, err := k.Execute(args...)
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	return stdout, nil
}

// parseCommandArgs splits a command string into arguments, handling quoted strings
func parseCommandArgs(command string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if inQuote {
			if ch == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(ch)
			}
		} else {
			if ch == '"' || ch == '\'' {
				inQuote = true
				quoteChar = ch
			} else if ch == ' ' || ch == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// GetResources fetches resources of the specified type with optional filters
func (k *Kubectl) GetResources(resource, namespace string, allNamespaces bool, labelSelector, fieldSelector string) (string, error) {
	args := []string{"get", resource}

	if allNamespaces {
		args = append(args, "--all-namespaces")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}

	if fieldSelector != "" {
		args = append(args, "--field-selector", fieldSelector)
	}

	stdout, stderr, err := k.Execute(args...)
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	return stdout, nil
}

// GetResourceYAML fetches a specific resource in YAML format
func (k *Kubectl) GetResourceYAML(resource, name, namespace string) (string, error) {
	args := []string{"get", resource, name, "-o", "yaml"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	stdout, stderr, err := k.Execute(args...)
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	return stdout, nil
}

// GetResourceJSON fetches a specific resource in JSON format
func (k *Kubectl) GetResourceJSON(resource, name, namespace string) (string, error) {
	args := []string{"get", resource, name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	stdout, stderr, err := k.Execute(args...)
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	return stdout, nil
}

// GetCurrentContext returns the current kubectl context name
func (k *Kubectl) GetCurrentContext() (string, error) {
	stdout, stderr, err := k.Execute("config", "current-context")
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	return strings.TrimSpace(stdout), nil
}

// GetCurrentNamespace returns the current namespace from the kubectl context
func (k *Kubectl) GetCurrentNamespace() (string, error) {
	stdout, stderr, err := k.Execute("config", "view", "--minify", "-o", "jsonpath={..namespace}")
	if err != nil {
		return "", &KubectlError{Stderr: stderr, Err: err}
	}
	ns := strings.TrimSpace(stdout)
	if ns == "" {
		return "default", nil
	}
	return ns, nil
}

// GetContexts returns a list of available kubectl contexts
func (k *Kubectl) GetContexts() ([]string, error) {
	stdout, stderr, err := k.Execute("config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, &KubectlError{Stderr: stderr, Err: err}
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	var contexts []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			contexts = append(contexts, line)
		}
	}
	return contexts, nil
}

// GetNamespaces returns a list of available namespaces
func (k *Kubectl) GetNamespaces() ([]string, error) {
	stdout, stderr, err := k.Execute("get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, &KubectlError{Stderr: stderr, Err: err}
	}

	names := strings.Fields(strings.TrimSpace(stdout))
	return names, nil
}

// KubectlError wraps kubectl execution errors with stderr output
type KubectlError struct {
	Stderr string
	Err    error
}

func (e *KubectlError) Error() string {
	if e.Stderr != "" {
		return strings.TrimSpace(e.Stderr)
	}
	return e.Err.Error()
}

func (e *KubectlError) Unwrap() error {
	return e.Err
}
