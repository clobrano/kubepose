package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
	"github.com/clobrano/kubepose/internal/tui/components/header"
)

// ViewState represents the current view mode of the application
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewSearch
	ViewConfirm
	ViewHelp
)

// Model is the main application model for the TUI
type Model struct {
	config  *config.Config
	kubectl *kubectl.Kubectl

	// UI state
	viewState    ViewState
	currentTab   int
	width        int
	height       int

	// Components
	header *header.Model

	// Data state
	currentContext   string
	currentNamespace string
	resources        *kubectl.TableData
	selectedIndex    int

	// Error state
	lastError error
}

// NewModel creates a new application model
func NewModel(cfg *config.Config, k *kubectl.Kubectl) *Model {
	return &Model{
		config:    cfg,
		kubectl:   k,
		viewState: ViewList,
		header:    header.New("", "", 0),
	}
}

// Init returns the initial command to run when the program starts
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadContext(),
		m.loadResources(),
	)
}

// Update handles all messages and updates the model state
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header.SetWidth(msg.Width)

	case ContextLoadedMsg:
		m.currentContext = msg.Context
		m.currentNamespace = msg.Namespace
		m.header.SetContext(msg.Context)
		m.header.SetNamespace(msg.Namespace)

	case ResourcesLoadedMsg:
		m.resources = msg.Data
		m.lastError = nil

	case ErrorMsg:
		m.lastError = msg.Err
	}

	return m, nil
}

// View renders the current state
func (m *Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.header.View())
	b.WriteString("\n")

	// Placeholder content
	b.WriteString("\nKubePose TUI - Press 'q' to quit\n")

	// Show error if any
	if m.lastError != nil {
		b.WriteString("\nError: ")
		b.WriteString(m.lastError.Error())
		b.WriteString("\n")
	}

	// Show resource count if loaded
	if m.resources != nil && len(m.resources.Rows) > 0 {
		b.WriteString("\nResources loaded: ")
		b.WriteString(strings.Join(m.resources.Headers, " | "))
		b.WriteString("\n")
		for i, row := range m.resources.Rows {
			if i >= 5 {
				break
			}
			b.WriteString("  ")
			b.WriteString(strings.Join(row, " | "))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// loadContext returns a command to load the current kubectl context
func (m *Model) loadContext() tea.Cmd {
	return func() tea.Msg {
		ctx, err := m.kubectl.GetCurrentContext()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		ns, err := m.kubectl.GetCurrentNamespace()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return ContextLoadedMsg{
			Context:   ctx,
			Namespace: ns,
		}
	}
}

// loadResources returns a command to load resources for the current tab
func (m *Model) loadResources() tea.Cmd {
	return func() tea.Msg {
		if len(m.config.Tabs) == 0 {
			return ResourcesLoadedMsg{Data: &kubectl.TableData{}}
		}

		tab := m.config.Tabs[m.currentTab]
		output, err := m.kubectl.GetResources(
			tab.Resource,
			tab.Namespace,
			tab.AllNamespaces,
			tab.LabelSelector,
			tab.FieldSelector,
		)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		data := kubectl.ParseTableOutput(output)
		return ResourcesLoadedMsg{Data: data}
	}
}
