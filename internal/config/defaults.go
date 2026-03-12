package config

import "time"

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		KubectlBin:      "kubectl",
		Pager:           "less",
		RefreshInterval: 5 * time.Second,
		Keybindings: Keybindings{
			Quit:            "q",
			Help:            "?",
			Refresh:         "r",
			Search:          "/",
			Describe:        "d",
			Logs:            "L",
			LogsFollow:      "ctrl+l",
			Delete:          "D",
			Edit:            "e",
			Exec:            "x",
			PortForward:     "p",
			Scale:           "s",
			RolloutRestart:  "R",
			YAMLView:        "Y",
			JSONView:        "J",
			SwitchNamespace: "n",
			SwitchContext:   "c",
			MultiSelect:     " ",
			Enter:           "enter",
			Escape:          "esc",
			Up:              "up",
			Down:            "down",
			TabNext:         "tab",
			TabPrev:         "shift+tab",
		},
		Tabs: []TabConfig{
			{
				Name:    "Pods",
				Command: "get pods -A",
			},
			{
				Name:    "Deployments",
				Command: "get deployments -A",
			},
			{
				Name:    "Services",
				Command: "get services -A",
			},
		},
		CustomCommands: []CustomCommand{},
	}
}
