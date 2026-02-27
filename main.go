package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
	"github.com/clobrano/kubepose/internal/tui"
)

// Version information set by ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("kubepose %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}
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

	// Create kubectl client
	k := kubectl.NewKubectl(cfg.KubectlBin)

	// Create and run the TUI
	model := tui.NewModel(cfg, k)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
