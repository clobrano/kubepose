package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/clobrano/kubepose/internal/actions"
	"github.com/clobrano/kubepose/internal/config"
	"github.com/clobrano/kubepose/internal/kubectl"
	"github.com/clobrano/kubepose/internal/tui/components/detail"
	"github.com/clobrano/kubepose/internal/tui/components/dialog"
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
	ViewSelector
	ViewInput
	ViewHelp
	ViewSearchTab // Search tab command input mode
)

// SearchTabIndex is the index of the special Search tab (always first)
const SearchTabIndex = 0

// getNamespaceFromCommand extracts the explicit namespace from a kubectl command string.
// Returns "" if -A/--all-namespaces is used or no namespace flag is present.
func getNamespaceFromCommand(command string) string {
	parts := strings.Fields(command)
	for i, part := range parts {
		if part == "-A" || part == "--all-namespaces" {
			return ""
		}
		if (part == "-n" || part == "--namespace") && i+1 < len(parts) {
			return parts[i+1]
		}
		if strings.HasPrefix(part, "--namespace=") {
			return strings.TrimPrefix(part, "--namespace=")
		}
	}
	return ""
}

// getResourceTypeFromCommand extracts the resource type from a kubectl command string
// e.g., "get pods -A" returns "pods", "get deployments -n default" returns "deployments"
func getResourceTypeFromCommand(command string) string {
	parts := strings.Fields(command)
	// Look for the resource type after "get" or similar commands
	for i, part := range parts {
		if part == "get" || part == "describe" || part == "delete" {
			if i+1 < len(parts) {
				// Return the next part as the resource type, ignoring flags
				next := parts[i+1]
				if !strings.HasPrefix(next, "-") {
					return next
				}
			}
		}
	}
	// If no recognized command pattern, return empty
	return ""
}

// getCurrentTabCommand returns the command for the current tab
func (m *Model) getCurrentTabCommand() string {
	if m.currentTab == SearchTabIndex {
		return m.searchCommand
	}
	configTabIndex := m.currentTab - 1
	if configTabIndex < 0 || configTabIndex >= len(m.config.Tabs) {
		return ""
	}
	return m.config.Tabs[configTabIndex].Command
}

// getCurrentResourceType returns the resource type for the current tab
func (m *Model) getCurrentResourceType() string {
	return getResourceTypeFromCommand(m.getCurrentTabCommand())
}

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
	header   *header.Model
	tabs     *tabs.Model
	list     *list.Model
	search   *search.Model
	detail   *detail.Model
	confirm  *dialog.ConfirmModel
	selector *dialog.SelectorModel
	input    *dialog.InputModel

	// Pending action state
	pendingAction string
	pendingNames  []string
	pendingNs     string

	// Data state
	currentContext   string
	currentNamespace string
	resources        *kubectl.TableData
	selectedIndex    int

	// Search tab state
	searchInput   *dialog.InputModel
	searchCommand string // Last executed search command

	// Per-tab search state: saves the last confirmed filter for each tab index
	tabSearchStates map[int]string

	// Error state
	lastError error
}

// NewModel creates a new application model
func NewModel(cfg *config.Config, k *kubectl.Kubectl) *Model {
	// Extract tab names from config, prepending Search tab
	tabNames := make([]string, len(cfg.Tabs)+1)
	tabNames[0] = "Search"
	for i, tab := range cfg.Tabs {
		tabNames[i+1] = tab.Name
	}

	// Create search input for the Search tab
	searchInput := dialog.NewInput("kubectl", "get pods -A")
	searchInput.WithValue("")

	return &Model{
		config:          cfg,
		kubectl:         k,
		viewState:       ViewList,
		currentTab:      1, // Start on first configured tab, not Search
		header:          header.New("", "", 0),
		tabs:            tabs.New(tabNames, 1),
		list:            list.New([]string{}, [][]string{}),
		search:          search.New(),
		detail:          detail.New("", "", detail.FormatTable),
		searchInput:     searchInput,
		tabSearchStates: make(map[int]string),
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
	// Handle confirm dialog state
	if m.viewState == ViewConfirm && m.confirm != nil {
		m.confirm, _ = m.confirm.Update(msg)
		switch m.confirm.Result() {
		case dialog.ConfirmYes:
			m.viewState = ViewList
			return m, m.executeDelete()
		case dialog.ConfirmNo:
			m.viewState = ViewList
			m.pendingAction = ""
			m.pendingNames = nil
			m.pendingNs = ""
		}
		return m, nil
	}

	// Handle selector dialog state
	if m.viewState == ViewSelector && m.selector != nil {
		m.selector, _ = m.selector.Update(msg)
		switch m.selector.Result() {
		case dialog.SelectorSelected:
			m.viewState = ViewList
			return m, m.handleSelectorResult()
		case dialog.SelectorCancelled:
			m.viewState = ViewList
			m.pendingAction = ""
		}
		return m, nil
	}

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

	// Handle Search tab input mode
	if m.viewState == ViewSearchTab {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)

		switch m.searchInput.Result() {
		case dialog.InputSubmitted:
			m.viewState = ViewList
			m.searchCommand = m.searchInput.Value()
			m.searchInput.Reset()
			return m, m.executeSearchCommand()
		case dialog.InputCancelled:
			m.viewState = ViewList
			m.searchInput.Reset()
		}

		// Handle window resize in search input mode
		if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = wmsg.Width
			m.height = wmsg.Height
			m.searchInput.SetSize(wmsg.Width, wmsg.Height)
		}

		return m, cmd
	}

	// If search is active (typing mode), forward messages to search component
	if m.search.IsActive() {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)

		// Update list with filtered results
		if m.resources != nil {
			filtered := m.search.FilteredItems()
			m.list.SetItems(m.resources.Headers, filtered)
		}

		// Check if search was fully cleared (Esc pressed, no filter remaining)
		if !m.search.IsActive() && !m.search.IsFiltered() && m.resources != nil {
			m.list.SetItems(m.resources.Headers, m.resources.Rows)
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Clear an active filter (confirmed via Enter)
			if m.search.IsFiltered() {
				m.search.Deactivate()
				delete(m.tabSearchStates, m.currentTab)
				if m.resources != nil {
					m.list.SetItems(m.resources.Headers, m.resources.Rows)
				}
			}
			return m, nil
		case "/":
			// Activate search
			if m.resources != nil {
				m.search.SetItems(m.resources.Rows)
			}
			m.search.Activate()
			return m, nil
		case "enter":
			// On Search tab, activate command input
			if m.currentTab == SearchTabIndex {
				m.viewState = ViewSearchTab
				m.searchInput.SetSize(m.width, m.height)
				return m, nil
			}
			// Show detail view (table format)
			return m, m.loadResourceDetail(detail.FormatTable)
		case "Y":
			// Show detail view (YAML format)
			return m, m.loadResourceDetail(detail.FormatYAML)
		case "J":
			// Show detail view (JSON format)
			return m, m.loadResourceDetail(detail.FormatJSON)
		case "tab", "right", "l":
			m.saveCurrentTabSearch()
			m.tabs.Next()
			m.currentTab = m.tabs.Active()
			return m, m.handleTabChange()
		case "shift+tab", "left", "h":
			m.saveCurrentTabSearch()
			m.tabs.Previous()
			m.currentTab = m.tabs.Active()
			return m, m.handleTabChange()
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < m.tabs.Count() {
				m.saveCurrentTabSearch()
				m.tabs.SetActive(idx)
				m.currentTab = idx
				return m, m.handleTabChange()
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
		case " ":
			// Toggle selection of current item
			m.list.ToggleSelect()
		case "a":
			// Select all items
			m.list.SelectAll()
		case "A":
			// Deselect all items
			m.list.DeselectAll()
		case "d":
			// Describe selected resource(s)
			return m, m.describeResources()
		case "L":
			// View logs (for pods)
			return m, m.viewLogs(false)
		case "ctrl+l":
			// Follow logs (for pods)
			return m, m.viewLogs(true)
		case "D":
			// Delete selected resource(s) - show confirmation
			return m, m.confirmDelete()
		case "e":
			// Edit resource
			return m, m.editResource()
		case "x":
			// Exec into pod
			return m, m.execIntoPod()
		case "R":
			// Rollout restart
			return m, m.rolloutRestart()
		case "c":
			// Switch context
			return m, m.showContextSelector()
		case "n":
			// Switch namespace
			return m, m.showNamespaceSelector()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header.SetWidth(msg.Width)
		m.tabs.SetWidth(msg.Width)
		m.search.SetWidth(msg.Width)
		// List height = total height - header (1) - tabs (1) - search (1 if active or filtered) - footer area (3)
		listHeight := msg.Height - 5
		if m.search.IsActive() || m.search.IsFiltered() {
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
			if m.search.IsFiltered() {
				// Refresh on same tab: re-apply the active filter to the new data
				m.search.SetItems(msg.Data.Rows)
				m.list.SetItems(msg.Data.Headers, m.search.FilteredItems())
			} else if savedQuery := m.tabSearchStates[m.currentTab]; savedQuery != "" {
				// Returning to a tab that had a saved filter: restore it
				m.search.SetItems(msg.Data.Rows)
				m.search.RestoreFilter(savedQuery)
				m.list.SetItems(msg.Data.Headers, m.search.FilteredItems())
			} else {
				m.list.SetItems(msg.Data.Headers, msg.Data.Rows)
			}
		}

	case ErrorMsg:
		m.lastError = msg.Err

	case DescribeLoadedMsg:
		m.viewState = ViewDetail
		title := "Describe"
		if len(msg.ResourceNames) == 1 {
			title = msg.ResourceNames[0]
		} else if len(msg.ResourceNames) > 1 {
			title = "Multiple Resources"
		}
		m.detail.SetContent(title, msg.Content, detail.FormatTable)
		m.detail.SetSize(m.width, m.height-2)

	case LogsLoadedMsg:
		m.viewState = ViewDetail
		title := msg.PodName
		if msg.Container != "" {
			title = msg.PodName + "/" + msg.Container
		}
		m.detail.SetContent(title+" (logs)", msg.Content, detail.FormatTable)
		m.detail.SetSize(m.width, m.height-2)

	case ContainersLoadedMsg:
		m.pendingNames = []string{msg.PodName}
		m.pendingNs = msg.Namespace
		if msg.Follow {
			m.pendingAction = "container-logs-follow"
		} else {
			m.pendingAction = "container-logs"
		}
		m.selector = dialog.NewSelector("Select Container", msg.Containers)
		m.selector.SetSize(m.width, m.height)
		m.viewState = ViewSelector

	case RefreshMsg:
		return m, tea.Batch(m.loadContext(), m.loadResources())
	}

	return m, nil
}

// View renders the current state
func (m *Model) View() string {
	// Detail view is full screen
	if m.viewState == ViewDetail {
		return m.detail.View()
	}

	// Confirm dialog is full screen overlay
	if m.viewState == ViewConfirm && m.confirm != nil {
		return m.confirm.View()
	}

	// Selector dialog is full screen overlay
	if m.viewState == ViewSelector && m.selector != nil {
		return m.selector.View()
	}

	// Search tab input mode
	if m.viewState == ViewSearchTab {
		return m.searchInput.View()
	}

	var b strings.Builder

	// Header
	b.WriteString(m.header.View())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.tabs.View())
	b.WriteString("\n")

	// Search tab: show command prompt info when on Search tab
	if m.currentTab == SearchTabIndex {
		if m.searchCommand != "" {
			b.WriteString(fmt.Sprintf("kubectl %s", m.searchCommand))
		} else {
			b.WriteString("Press [Enter] to enter a kubectl command")
		}
		b.WriteString("\n")
	}

	// Search bar (if active or filter confirmed)
	if m.search.IsActive() || m.search.IsFiltered() {
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
	} else if m.search.IsFiltered() {
		b.WriteString("[Esc] clear filter  [/] modify filter  [d]escribe [L]ogs [D]elete [e]dit [x]exec")
	} else if m.currentTab == SearchTabIndex {
		b.WriteString("[Enter] enter command  [/]filter results  [r]efresh  [q]uit")
	} else {
		b.WriteString("[d]escribe [L]ogs [ctrl+l]follow [D]elete [e]dit [x]exec [R]estart  [c]ontext [n]amespace  [/]search [r]efresh [?]help")
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

// handleTabChange handles switching to a new tab
// saveCurrentTabSearch saves the active search filter for the current tab and
// deactivates the search component so the next tab starts clean.
func (m *Model) saveCurrentTabSearch() {
	m.tabSearchStates[m.currentTab] = m.search.Query()
	m.search.Deactivate()
}

func (m *Model) handleTabChange() tea.Cmd {
	// Search tab shows empty list until command is executed
	if m.currentTab == SearchTabIndex {
		m.resources = &kubectl.TableData{}
		m.list.SetItems([]string{}, [][]string{})
		return nil
	}
	return m.loadResources()
}

// loadResources returns a command to load resources for the current tab
func (m *Model) loadResources() tea.Cmd {
	return func() tea.Msg {
		// Search tab doesn't load automatically
		if m.currentTab == SearchTabIndex {
			return ResourcesLoadedMsg{Data: &kubectl.TableData{}}
		}

		// Config tabs are offset by 1 due to Search tab at index 0
		configTabIndex := m.currentTab - 1
		if configTabIndex < 0 || configTabIndex >= len(m.config.Tabs) {
			return ResourcesLoadedMsg{Data: &kubectl.TableData{}}
		}

		tab := m.config.Tabs[configTabIndex]
		output, err := m.kubectl.ExecuteRaw(tab.Command)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		data := kubectl.ParseTableOutput(output)
		return ResourcesLoadedMsg{Data: data}
	}
}

// executeSearchCommand executes the search tab command
func (m *Model) executeSearchCommand() tea.Cmd {
	return func() tea.Msg {
		if m.searchCommand == "" {
			return ResourcesLoadedMsg{Data: &kubectl.TableData{}}
		}

		output, err := m.kubectl.ExecuteRaw(m.searchCommand)
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

		resourceType := m.getCurrentResourceType()
		if resourceType == "" {
			return ErrorMsg{Err: nil}
		}

		// Determine namespace and resource name from the selected row
		namespace := m.currentNamespace
		resourceName := selected[0]
		namespaceFromColumn := false

		// Check if there's a NAMESPACE column (for -A commands)
		if m.resources != nil {
			namespaceIdx := m.resources.GetColumnIndex("NAMESPACE")
			if namespaceIdx >= 0 && namespaceIdx < len(selected) {
				namespace = selected[namespaceIdx]
				namespaceFromColumn = true
			}
			// NAME column might be after NAMESPACE
			nameIdx := m.resources.GetColumnIndex("NAME")
			if nameIdx >= 0 && nameIdx < len(selected) {
				resourceName = selected[nameIdx]
			}
		}

		// No NAMESPACE column: fall back to the namespace in the tab command
		if !namespaceFromColumn {
			if cmdNs := getNamespaceFromCommand(m.getCurrentTabCommand()); cmdNs != "" {
				namespace = cmdNs
			}
		}

		var content string
		var err error

		switch format {
		case detail.FormatYAML:
			content, err = m.kubectl.GetResourceYAML(resourceType, resourceName, namespace)
		case detail.FormatJSON:
			content, err = m.kubectl.GetResourceJSON(resourceType, resourceName, namespace)
		default:
			// Table format - use describe
			content, _, err = m.kubectl.Execute("describe", resourceType, resourceName, "-n", namespace)
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

// getSelectedResourceInfo returns resource names and namespace for selected items
func (m *Model) getSelectedResourceInfo() (names []string, namespace string) {
	namespace = m.currentNamespace

	// Get selected items (or current item if none selected)
	items := m.list.SelectedItems()
	if items == nil {
		return nil, namespace
	}

	// Extract names and potentially namespace from items
	nameIdx := 0
	namespaceIdx := -1

	if m.resources != nil {
		// Check for NAMESPACE column (present in -A commands)
		namespaceIdx = m.resources.GetColumnIndex("NAMESPACE")
		// NAME column might be after NAMESPACE
		nameColIdx := m.resources.GetColumnIndex("NAME")
		if nameColIdx >= 0 {
			nameIdx = nameColIdx
		}
	}

	namespaceExtracted := false
	for _, item := range items {
		if len(item) > nameIdx {
			names = append(names, item[nameIdx])
		}
		// Use namespace from the first item only when a NAMESPACE column is present
		if !namespaceExtracted && namespaceIdx >= 0 && namespaceIdx < len(item) {
			namespace = item[namespaceIdx]
			namespaceExtracted = true
		}
	}

	// If no NAMESPACE column in the data, fall back to the namespace specified
	// in the current tab command (e.g. "get pods -n rhwa" → "rhwa")
	if !namespaceExtracted {
		if cmdNs := getNamespaceFromCommand(m.getCurrentTabCommand()); cmdNs != "" {
			namespace = cmdNs
		}
	}

	return names, namespace
}

// describeResources returns a command to describe selected resources
func (m *Model) describeResources() tea.Cmd {
	return func() tea.Msg {
		names, namespace := m.getSelectedResourceInfo()
		if len(names) == 0 {
			return ErrorMsg{Err: nil}
		}

		resourceType := m.getCurrentResourceType()
		if resourceType == "" {
			return ErrorMsg{Err: nil}
		}

		ctx := actions.NewContext(m.kubectl, resourceType, names, namespace)

		content, err := actions.Describe(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return DescribeLoadedMsg{
			ResourceNames: names,
			Content:       content,
		}
	}
}

// isPodResource returns true if the resource type is a pod
func isPodResource(resourceType string) bool {
	return resourceType == "pods" || resourceType == "pod" || resourceType == "po"
}

// viewLogs returns a command to view logs for the selected pod
func (m *Model) viewLogs(follow bool) tea.Cmd {
	return func() tea.Msg {
		resourceType := m.getCurrentResourceType()

		// Logs only work for pods
		if !isPodResource(resourceType) {
			return ErrorMsg{Err: nil}
		}

		names, namespace := m.getSelectedResourceInfo()
		if len(names) == 0 {
			return ErrorMsg{Err: nil}
		}

		// For logs, we only use the first selected pod
		podName := names[0]
		ctx := actions.NewContext(m.kubectl, "pod", []string{podName}, namespace)

		// Get container list to check if we need selection
		containers, err := actions.GetContainers(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// If multiple containers, show a selector dialog
		if len(containers) > 1 {
			return ContainersLoadedMsg{
				PodName:    podName,
				Namespace:  namespace,
				Containers: containers,
				Follow:     follow,
			}
		}

		container := ""
		if len(containers) == 1 {
			container = containers[0]
		}

		opts := actions.LogsOptions{
			Container: container,
			TailLines: 500,
		}

		content, err := actions.Logs(ctx, opts)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return LogsLoadedMsg{
			PodName:   podName,
			Container: container,
			Content:   content,
		}
	}
}

// confirmDelete shows a confirmation dialog for delete
func (m *Model) confirmDelete() tea.Cmd {
	return func() tea.Msg {
		names, namespace := m.getSelectedResourceInfo()
		if len(names) == 0 {
			return nil
		}

		resourceType := m.getCurrentResourceType()
		if resourceType == "" {
			return nil
		}

		var message string
		if len(names) == 1 {
			message = fmt.Sprintf("Delete %s/%s?\n\nThis action cannot be undone.", resourceType, names[0])
		} else {
			message = fmt.Sprintf("Delete %d %s resources?\n\nThis action cannot be undone.", len(names), resourceType)
		}

		m.confirm = dialog.NewConfirm("Confirm Delete", message)
		m.confirm.SetSize(m.width, m.height)
		m.pendingAction = "delete"
		m.pendingNames = names
		m.pendingNs = namespace
		m.viewState = ViewConfirm

		return nil
	}
}

// executeDelete performs the actual delete operation
func (m *Model) executeDelete() tea.Cmd {
	return func() tea.Msg {
		if len(m.pendingNames) == 0 {
			return nil
		}

		resourceType := m.getCurrentResourceType()
		if resourceType == "" {
			return nil
		}

		ctx := actions.NewContext(m.kubectl, resourceType, m.pendingNames, m.pendingNs)

		results, err := actions.Delete(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Clear pending state
		m.pendingAction = ""
		m.pendingNames = nil
		m.pendingNs = ""

		// Build result message
		var messages []string
		for _, r := range results {
			messages = append(messages, r.Message)
		}

		// Show result in detail view briefly, then refresh
		m.detail.SetContent("Delete Result", strings.Join(messages, "\n"), detail.FormatTable)
		m.detail.SetSize(m.width, m.height-2)
		m.viewState = ViewDetail

		return nil
	}
}

// editResource opens the resource in an editor
func (m *Model) editResource() tea.Cmd {
	names, namespace := m.getSelectedResourceInfo()
	if len(names) == 0 {
		return nil
	}

	resourceType := m.getCurrentResourceType()
	if resourceType == "" {
		return nil
	}

	// Return a command that suspends the TUI
	return tea.ExecProcess(
		makeKubectlEditCmd(m.kubectl.BinaryPath(), resourceType, names[0], namespace),
		func(err error) tea.Msg {
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return RefreshMsg{}
		},
	)
}

// makeKubectlEditCmd creates an exec.Cmd for kubectl edit
func makeKubectlEditCmd(kubectlBin, resource, name, namespace string) *exec.Cmd {
	args := []string{"edit", resource, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// execIntoPod starts an exec session in the pod
func (m *Model) execIntoPod() tea.Cmd {
	names, namespace := m.getSelectedResourceInfo()
	if len(names) == 0 {
		return nil
	}

	resourceType := m.getCurrentResourceType()
	// Exec only works for pods
	if !isPodResource(resourceType) {
		return nil
	}

	// Return a command that suspends the TUI
	return tea.ExecProcess(
		makeKubectlExecCmd(m.kubectl.BinaryPath(), names[0], namespace, ""),
		func(err error) tea.Msg {
			return nil // Don't show error, just return to TUI
		},
	)
}

// makeKubectlExecCmd creates an exec.Cmd for kubectl exec
func makeKubectlExecCmd(kubectlBin, podName, namespace, container string) *exec.Cmd {
	args := []string{"exec", "-it", podName}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, "--", "/bin/sh")
	cmd := exec.Command(kubectlBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// rolloutRestart triggers a rollout restart
func (m *Model) rolloutRestart() tea.Cmd {
	return func() tea.Msg {
		names, namespace := m.getSelectedResourceInfo()
		if len(names) == 0 {
			return nil
		}

		resourceType := m.getCurrentResourceType()
		if resourceType == "" {
			return nil
		}

		ctx := actions.NewContext(m.kubectl, resourceType, names, namespace)

		output, err := actions.RolloutRestart(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Show result
		m.detail.SetContent("Rollout Restart", output, detail.FormatTable)
		m.detail.SetSize(m.width, m.height-2)
		m.viewState = ViewDetail

		return nil
	}
}

// handleSelectorResult processes the selector dialog result
func (m *Model) handleSelectorResult() tea.Cmd {
	if m.selector == nil {
		return nil
	}

	selected := m.selector.SelectedItem()
	action := m.pendingAction

	m.pendingAction = ""
	m.selector = nil

	switch action {
	case "container-logs":
		// Use selected container for logs
		return m.viewLogsWithContainer(selected, false)
	case "container-logs-follow":
		// Use selected container for follow logs
		return m.viewLogsWithContainer(selected, true)
	case "container-exec":
		// Use selected container for exec
		return m.execWithContainer(selected)
	case "context":
		// Switch context
		return m.switchContext(selected)
	case "namespace":
		// Switch namespace
		return m.switchNamespace(selected)
	}

	return nil
}

// viewLogsWithContainer views logs for a specific container
func (m *Model) viewLogsWithContainer(container string, follow bool) tea.Cmd {
	return func() tea.Msg {
		if len(m.pendingNames) == 0 {
			return nil
		}

		podName := m.pendingNames[0]
		ctx := actions.NewContext(m.kubectl, "pod", []string{podName}, m.pendingNs)

		opts := actions.LogsOptions{
			Container: container,
			TailLines: 500,
		}

		content, err := actions.Logs(ctx, opts)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		m.pendingNames = nil
		m.pendingNs = ""

		return LogsLoadedMsg{
			PodName:   podName,
			Container: container,
			Content:   content,
		}
	}
}

// execWithContainer starts exec with a specific container
func (m *Model) execWithContainer(container string) tea.Cmd {
	if len(m.pendingNames) == 0 {
		return nil
	}

	podName := m.pendingNames[0]
	namespace := m.pendingNs
	kubectlBin := m.kubectl.BinaryPath()

	m.pendingNames = nil
	m.pendingNs = ""

	return tea.ExecProcess(
		makeKubectlExecCmd(kubectlBin, podName, namespace, container),
		func(err error) tea.Msg {
			return nil
		},
	)
}

// switchContext switches to a different kubectl context
func (m *Model) switchContext(contextName string) tea.Cmd {
	return func() tea.Msg {
		_, _, err := m.kubectl.Execute("config", "use-context", contextName)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Reload context and resources
		return RefreshMsg{}
	}
}

// switchNamespace updates the current namespace
func (m *Model) switchNamespace(namespace string) tea.Cmd {
	return func() tea.Msg {
		m.currentNamespace = namespace
		m.header.SetNamespace(namespace)
		// Reload resources with new namespace
		return RefreshMsg{}
	}
}

// showContextSelector displays a selector for available contexts
func (m *Model) showContextSelector() tea.Cmd {
	return func() tea.Msg {
		contexts, err := m.kubectl.GetContexts()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		if len(contexts) == 0 {
			return nil
		}

		m.selector = dialog.NewSelector("Select Context", contexts)
		m.selector.SetSize(m.width, m.height)
		m.pendingAction = "context"
		m.viewState = ViewSelector

		return nil
	}
}

// showNamespaceSelector displays a selector for available namespaces
func (m *Model) showNamespaceSelector() tea.Cmd {
	return func() tea.Msg {
		namespaces, err := m.kubectl.GetNamespaces()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		if len(namespaces) == 0 {
			return nil
		}

		m.selector = dialog.NewSelector("Select Namespace", namespaces)
		m.selector.SetSize(m.width, m.height)
		m.pendingAction = "namespace"
		m.viewState = ViewSelector

		return nil
	}
}
