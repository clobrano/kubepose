package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath returns the default config file path (~/.config/kubepose/config.yaml)
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kubepose", "config.yaml"), nil
}

// Config represents the main configuration for KubePose
type Config struct {
	KubectlBin      string          `yaml:"kubectl_bin"`
	Pager           string          `yaml:"pager"`
	RefreshInterval time.Duration   `yaml:"refresh_interval"`
	Keybindings     Keybindings     `yaml:"keybindings"`
	Tabs            []TabConfig     `yaml:"tabs"`
	CustomCommands  []CustomCommand `yaml:"custom_commands"`
}

// TabConfig represents the configuration for a single tab
type TabConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// CustomCommand represents a user-defined custom command
type CustomCommand struct {
	Name    string `yaml:"name"`
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

// Keybindings represents all configurable keyboard shortcuts
type Keybindings struct {
	Quit            string `yaml:"quit"`
	Help            string `yaml:"help"`
	Refresh         string `yaml:"refresh"`
	Search          string `yaml:"search"`
	Describe        string `yaml:"describe"`
	Logs            string `yaml:"logs"`
	LogsFollow      string `yaml:"logs_follow"`
	Delete          string `yaml:"delete"`
	Edit            string `yaml:"edit"`
	Exec            string `yaml:"exec"`
	PortForward     string `yaml:"port_forward"`
	Scale           string `yaml:"scale"`
	RolloutRestart  string `yaml:"rollout_restart"`
	YAMLView        string `yaml:"yaml_view"`
	JSONView        string `yaml:"json_view"`
	SwitchNamespace string `yaml:"switch_namespace"`
	SwitchContext   string `yaml:"switch_context"`
	MultiSelect     string `yaml:"multi_select"`
	Enter           string `yaml:"enter"`
	Escape          string `yaml:"escape"`
	Up              string `yaml:"up"`
	Down            string `yaml:"down"`
	TabNext         string `yaml:"tab_next"`
	TabPrev         string `yaml:"tab_prev"`
}

// LoadConfig reads and parses a YAML configuration file from the given path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// EnsureConfigExists creates a default config file if one doesn't exist
func EnsureConfigExists() (string, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return "", err
	}

	// Check if config already exists
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Write default config
	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return path, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.KubectlBin == "" {
		return errors.New("kubectl_bin cannot be empty")
	}

	if len(c.Tabs) == 0 {
		return errors.New("at least one tab must be configured")
	}

	for i, tab := range c.Tabs {
		if tab.Name == "" {
			return fmt.Errorf("tab %d: name cannot be empty", i)
		}
		if tab.Command == "" {
			return fmt.Errorf("tab %d (%s): command cannot be empty", i, tab.Name)
		}
	}

	for i, cmd := range c.CustomCommands {
		if cmd.Name == "" {
			return fmt.Errorf("custom_command %d: name cannot be empty", i)
		}
		if cmd.Key == "" {
			return fmt.Errorf("custom_command %d (%s): key cannot be empty", i, cmd.Name)
		}
		if cmd.Command == "" {
			return fmt.Errorf("custom_command %d (%s): command cannot be empty", i, cmd.Name)
		}
	}

	return nil
}
