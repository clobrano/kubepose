package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("39")  // Light blue
	secondaryColor = lipgloss.Color("240") // Gray
	accentColor    = lipgloss.Color("212") // Pink
	errorColor     = lipgloss.Color("196") // Red
	successColor   = lipgloss.Color("46")  // Green
	warningColor   = lipgloss.Color("214") // Orange
)

// Styles holds all the application styles
type Styles struct {
	// Header styles
	Header          lipgloss.Style
	HeaderContext   lipgloss.Style
	HeaderNamespace lipgloss.Style
	HeaderHelp      lipgloss.Style

	// Tab styles
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabBar      lipgloss.Style

	// List styles
	ListHeader   lipgloss.Style
	ListItem     lipgloss.Style
	ListSelected lipgloss.Style

	// Detail view styles
	DetailHeader lipgloss.Style
	DetailBody   lipgloss.Style

	// Error styles
	Error lipgloss.Style

	// Footer styles
	Footer    lipgloss.Style
	FooterKey lipgloss.Style
}

// DefaultStyles returns the default application styles
func DefaultStyles() *Styles {
	return &Styles{
		// Header
		Header: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),

		HeaderContext: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true),

		HeaderNamespace: lipgloss.NewStyle().
			Foreground(accentColor),

		HeaderHelp: lipgloss.NewStyle().
			Foreground(secondaryColor),

		// Tabs
		TabActive: lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("0")).
			Padding(0, 2).
			Bold(true),

		TabInactive: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 2),

		TabBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")),

		// List
		ListHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")),

		ListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		ListSelected: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(primaryColor).
			Bold(true),

		// Detail view
		DetailHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1),

		DetailBody: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		// Error
		Error: lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true),

		// Footer
		Footer: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),

		FooterKey: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true),
	}
}
