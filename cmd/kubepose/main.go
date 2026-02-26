package main

import (
	"fmt"
	"os"

	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
)

func main() {
	fmt.Println("KubePose")

	// Ensure config exists (create default if needed)
	configPath, err := config.EnsureConfigExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ensuring config exists: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Print loaded tabs for verification
	fmt.Printf("Config loaded from: %s\n", configPath)
	fmt.Printf("Kubectl binary: %s\n", cfg.KubectlBin)
	fmt.Printf("Tabs:\n")
	for i, tab := range cfg.Tabs {
		fmt.Printf("  %d. %s (resource: %s)\n", i+1, tab.Name, tab.Resource)
	}

	// Test kubectl integration
	k := kubectl.NewKubectl(cfg.KubectlBin)

	// Get current context
	ctx, err := k.GetCurrentContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not get current context: %v\n", err)
	} else {
		fmt.Printf("\nCurrent context: %s\n", ctx)
	}

	// Get current namespace
	ns, err := k.GetCurrentNamespace()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not get current namespace: %v\n", err)
	} else {
		fmt.Printf("Current namespace: %s\n", ns)
	}

	// Fetch pods from first tab
	if len(cfg.Tabs) > 0 {
		tab := cfg.Tabs[0]
		fmt.Printf("\nFetching %s...\n", tab.Resource)

		output, err := k.GetResources(tab.Resource, tab.Namespace, tab.AllNamespaces, tab.LabelSelector, tab.FieldSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching resources: %v\n", err)
		} else {
			data := kubectl.ParseTableOutput(output)
			fmt.Printf("Headers: %v\n", data.Headers)
			fmt.Printf("Found %d resources\n", len(data.Rows))
			for i, row := range data.Rows {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(data.Rows)-5)
					break
				}
				fmt.Printf("  %v\n", row)
			}
		}
	}
}
