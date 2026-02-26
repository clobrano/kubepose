package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
	"github.com/clobrano/kubepose/internal/tui/components/detail"
	"github.com/clobrano/kubepose/internal/tui/components/header"
	"github.com/clobrano/kubepose/internal/tui/components/list"
	"github.com/clobrano/kubepose/internal/tui/components/search"
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
	search *search.Model
	detail *detail.Model

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
		search:    search.New(),
		detail:    detail.New("", "", detail.FormatTable),
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
	// Handle detail view state
	if m.viewState == ViewDetail {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.viewState = ViewList
				return m, nil
			case "j", "down":
				m.detail.ScrollDown()
			case "k", "up":
				m.detail.ScrollUp()
			case "g":
				m.detail.ScrollToTop()
			case "G":
				m.detail.ScrollToBottom()
			case "ctrl+d", "pgdown":
				m.detail.PageDown()
			case "ctrl+u", "pgup":
				m.detail.PageUp()
			case "Y":
				return m, m.loadResourceDetail(detail.FormatYAML)
			case "J":
				return m, m.loadResourceDetail(detail.FormatJSON)
			}
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.detail.SetSize(msg.Width, msg.Height-2)
		case DetailLoadedMsg:
			m.detail.SetContent(msg.ResourceName, msg.Content, detail.Format(msg.Format))
		}
		return m, nil
	}

	// If search is active, forward messages to search component
	if m.search.IsActive() {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)

		// Update list with filtered results
		if m.resources != nil {
			filtered := m.search.FilteredItems()
			m.list.SetItems(m.resources.Headers, filtered)
		}

		// Check if search was deactivated (Esc pressed)
		if !m.search.IsActive() && m.resources != nil {
			m.list.SetItems(m.resources.Headers, m.resources.Rows)
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			// Activate search
			if m.resources != nil {
				m.search.SetItems(m.resources.Rows)
			}
			m.search.Activate()
			return m, nil
		case "enter":
			// Show detail view (table format)
			return m, m.loadResourceDetail(detail.FormatTable)
		case "Y":
			// Show detail view (YAML format)
			return m, m.loadResourceDetail(detail.FormatYAML)
		case "J":
			// Show detail view (JSON format)
			return m, m.loadResourceDetail(detail.FormatJSON)
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
		m.search.SetWidth(msg.Width)
		// List height = total height - header (1) - tabs (1) - search (1 if active) - footer area (3)
		listHeight := msg.Height - 5
		if m.search.IsActive() {
			listHeight--
		}
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
	// Detail view is full screen
	if m.viewState == ViewDetail {
		return m.detail.View()
	}

	var b strings.Builder

	// Header
	b.WriteString(m.header.View())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.tabs.View())
	b.WriteString("\n")

	// Search bar (if active)
	if m.search.IsActive() {
		b.WriteString(m.search.View())
		b.WriteString("\n")
	}

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
	if m.search.IsActive() {
		b.WriteString("[Enter] confirm  [Esc] cancel  [Type] to filter")
	} else {
		b.WriteString("[j/k] navigate  [Enter] details  [Y] yaml  [J] json  [/] search  [r] refresh  [q] quit")
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

// loadResourceDetail returns a command to load detail for the selected resource
func (m *Model) loadResourceDetail(format detail.Format) tea.Cmd {
	return func() tea.Msg {
		selected := m.list.SelectedItem()
		if selected == nil || len(selected) == 0 {
			return ErrorMsg{Err: nil}
		}

		resourceName := selected[0]
		tab := m.config.Tabs[m.currentTab]

		// Determine namespace from resource or current namespace
		namespace := m.currentNamespace
		if tab.AllNamespaces && len(selected) > 1 {
			// For all-namespaces view, namespace is usually in the first column
			namespaceIdx := m.resources.GetColumnIndex("NAMESPACE")
			if namespaceIdx >= 0 && namespaceIdx < len(selected) {
				namespace = selected[namespaceIdx]
			}
		} else if tab.Namespace != "" {
			namespace = tab.Namespace
		}

		var content string
		var err error

		switch format {
		case detail.FormatYAML:
			content, err = m.kubectl.GetResourceYAML(tab.Resource, resourceName, namespace)
		case detail.FormatJSON:
			content, err = m.kubectl.GetResourceJSON(tab.Resource, resourceName, namespace)
		default:
			// Table format - use describe
			content, _, err = m.kubectl.Execute("describe", tab.Resource, resourceName, "-n", namespace)
		}

		if err != nil {
			return ErrorMsg{Err: err}
		}

		m.viewState = ViewDetail
		m.detail.SetSize(m.width, m.height-2)

		return DetailLoadedMsg{
			ResourceName: resourceName,
			Content:      content,
			Format:       int(format),
		}
	}
}
