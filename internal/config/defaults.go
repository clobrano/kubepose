package config

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		KubectlBin: "kubectl",
		Pager:      "less",
		Keybindings: Keybindings{
			Quit:            "q",
			Help:            "?",
			Refresh:         "r",
			Search:          "/",
			Describe:        "d",
			Logs:            "l",
			LogsFollow:      "L",
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
				Name:     "Pods",
				Resource: "pods",
			},
		},
		CustomCommands: []CustomCommand{},
	}
}
