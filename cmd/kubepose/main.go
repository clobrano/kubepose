package main

import (
	"fmt"
	"os"

	"github.com/clobrano/kubepose/internal/config"
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
}
