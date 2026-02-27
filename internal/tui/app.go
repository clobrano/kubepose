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
		case "l":
			// View logs (for pods)
			return m, m.viewLogs(false)
		case "L":
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
		b.WriteString("[d]escribe [l]ogs [D]elete [e]dit [x]exec [R]estart  [c]ontext [n]amespace  [/]search [r]efresh [?]help")
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

// getSelectedResourceInfo returns resource names and namespace for selected items
func (m *Model) getSelectedResourceInfo() (names []string, namespace string) {
	tab := m.config.Tabs[m.currentTab]
	namespace = m.currentNamespace

	if tab.Namespace != "" {
		namespace = tab.Namespace
	}

	// Get selected items (or current item if none selected)
	items := m.list.SelectedItems()
	if items == nil {
		return nil, namespace
	}

	// Extract names and potentially namespace from items
	nameIdx := 0
	namespaceIdx := -1

	if tab.AllNamespaces && m.resources != nil {
		namespaceIdx = m.resources.GetColumnIndex("NAMESPACE")
		// NAME column might be after NAMESPACE
		nameColIdx := m.resources.GetColumnIndex("NAME")
		if nameColIdx >= 0 {
			nameIdx = nameColIdx
		}
	}

	for _, item := range items {
		if len(item) > nameIdx {
			names = append(names, item[nameIdx])
		}
		// Use namespace from first item if all-namespaces mode
		if namespaceIdx >= 0 && namespaceIdx < len(item) && namespace == m.currentNamespace {
			namespace = item[namespaceIdx]
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

		tab := m.config.Tabs[m.currentTab]
		ctx := actions.NewContext(m.kubectl, tab.Resource, names, namespace)

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

// viewLogs returns a command to view logs for the selected pod
func (m *Model) viewLogs(follow bool) tea.Cmd {
	return func() tea.Msg {
		tab := m.config.Tabs[m.currentTab]

		// Logs only work for pods
		if tab.Resource != "pods" && tab.Resource != "pod" && tab.Resource != "po" {
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

		// If multiple containers, we'd need a selector
		// For now, just use the first container or no specific container
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

		tab := m.config.Tabs[m.currentTab]
		var message string
		if len(names) == 1 {
			message = fmt.Sprintf("Delete %s/%s?\n\nThis action cannot be undone.", tab.Resource, names[0])
		} else {
			message = fmt.Sprintf("Delete %d %s resources?\n\nThis action cannot be undone.", len(names), tab.Resource)
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

		tab := m.config.Tabs[m.currentTab]
		ctx := actions.NewContext(m.kubectl, tab.Resource, m.pendingNames, m.pendingNs)

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

	tab := m.config.Tabs[m.currentTab]

	// Return a command that suspends the TUI
	return tea.ExecProcess(
		makeKubectlEditCmd(m.kubectl.BinaryPath(), tab.Resource, names[0], namespace),
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

	tab := m.config.Tabs[m.currentTab]
	// Exec only works for pods
	if tab.Resource != "pods" && tab.Resource != "pod" && tab.Resource != "po" {
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

		tab := m.config.Tabs[m.currentTab]
		ctx := actions.NewContext(m.kubectl, tab.Resource, names, namespace)

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
