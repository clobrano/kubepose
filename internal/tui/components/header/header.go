package header

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Model represents the header component
type Model struct {
	context   string
	namespace string
	width     int
	styles    *Styles
}

// Styles defines the styles for the header component
type Styles struct {
	Container lipgloss.Style
	Context   lipgloss.Style
	Namespace lipgloss.Style
	Help      lipgloss.Style
	Separator lipgloss.Style
}

// DefaultStyles returns the default header styles
func DefaultStyles() *Styles {
	return &Styles{
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Padding(0, 1),
		Context: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),
		Namespace: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// New creates a new header component
func New(context, namespace string, width int) *Model {
	return &Model{
		context:   context,
		namespace: namespace,
		width:     width,
		styles:    DefaultStyles(),
	}
}

// SetContext updates the current context
func (m *Model) SetContext(context string) {
	m.context = context
}

// SetNamespace updates the current namespace
func (m *Model) SetNamespace(namespace string) {
	m.namespace = namespace
}

// SetWidth updates the header width for responsive layout
func (m *Model) SetWidth(width int) {
	m.width = width
}

// Context returns the current context
func (m *Model) Context() string {
	return m.context
}

// Namespace returns the current namespace
func (m *Model) Namespace() string {
	return m.namespace
}

// View renders the header
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Left side: Context and namespace info
	ctx := "Context: "
	if m.context != "" {
		ctx += m.styles.Context.Render(m.context)
	} else {
		ctx += m.styles.Context.Render("(none)")
	}

	sep := m.styles.Separator.Render(" | ")

	ns := "Namespace: "
	if m.namespace != "" {
		ns += m.styles.Namespace.Render(m.namespace)
	} else {
		ns += m.styles.Namespace.Render("default")
	}

	left := ctx + sep + ns

	// Right side: Help indicator
	right := m.styles.Help.Render("[?] Help")

	// Calculate padding to right-align help
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	padding := m.width - leftLen - rightLen - 2 // -2 for container padding

	if padding < 1 {
		padding = 1
	}

	content := left + strings.Repeat(" ", padding) + right

	return m.styles.Container.Width(m.width).Render(content)
}

// String implements fmt.Stringer
func (m *Model) String() string {
	return fmt.Sprintf("Header{context=%q, namespace=%q}", m.context, m.namespace)
}
