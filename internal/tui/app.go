package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
	"github.com/clobrano/kubepose/internal/tui/components/header"
	"github.com/clobrano/kubepose/internal/tui/components/list"
	"github.com/clobrano/kubepose/internal/tui/components/tabs"
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
	tabs   *tabs.Model
	list   *list.Model

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
	// Extract tab names from config
	tabNames := make([]string, len(cfg.Tabs))
	for i, tab := range cfg.Tabs {
		tabNames[i] = tab.Name
	}

	return &Model{
		config:    cfg,
		kubectl:   k,
		viewState: ViewList,
		header:    header.New("", "", 0),
		tabs:      tabs.New(tabNames, 0),
		list:      list.New([]string{}, [][]string{}),
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
		case "tab":
			m.tabs.Next()
			m.currentTab = m.tabs.Active()
			return m, m.loadResources()
		case "shift+tab":
			m.tabs.Previous()
			m.currentTab = m.tabs.Active()
			return m, m.loadResources()
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < m.tabs.Count() {
				m.tabs.SetActive(idx)
				m.currentTab = idx
				return m, m.loadResources()
			}
		case "j", "down":
			m.list.MoveDown()
		case "k", "up":
			m.list.MoveUp()
		case "g":
			m.list.MoveToTop()
		case "G":
			m.list.MoveToBottom()
		case "r":
			return m, m.loadResources()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header.SetWidth(msg.Width)
		m.tabs.SetWidth(msg.Width)
		// List height = total height - header (1) - tabs (1) - footer area (3)
		listHeight := msg.Height - 5
		if listHeight < 3 {
			listHeight = 3
		}
		m.list.SetSize(msg.Width, listHeight)

	case ContextLoadedMsg:
		m.currentContext = msg.Context
		m.currentNamespace = msg.Namespace
		m.header.SetContext(msg.Context)
		m.header.SetNamespace(msg.Namespace)

	case ResourcesLoadedMsg:
		m.resources = msg.Data
		m.lastError = nil
		if msg.Data != nil {
			m.list.SetItems(msg.Data.Headers, msg.Data.Rows)
		}

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

	// Tabs
	b.WriteString(m.tabs.View())
	b.WriteString("\n")

	// Show error if any
	if m.lastError != nil {
		b.WriteString("\nError: ")
		b.WriteString(m.lastError.Error())
		b.WriteString("\n")
	} else {
		// Resource list
		b.WriteString(m.list.View())
	}

	// Footer/help
	b.WriteString("\n\n")
	b.WriteString("[j/k] navigate  [Tab] switch tab  [r] refresh  [q] quit")

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
