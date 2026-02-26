package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Format represents the display format for the detail view
type Format int

const (
	FormatTable Format = iota
	FormatYAML
	FormatJSON
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case FormatYAML:
		return "YAML"
	case FormatJSON:
		return "JSON"
	default:
		return "Table"
	}
}

// Model represents the detail view component
type Model struct {
	resourceName string
	content      string
	format       Format
	scrollOffset int
	width        int
	height       int
	styles       *Styles
}

// Styles defines the styles for the detail view component
type Styles struct {
	Header    lipgloss.Style
	Content   lipgloss.Style
	LineNum   lipgloss.Style
	KeyWord   lipgloss.Style
	String    lipgloss.Style
	Number    lipgloss.Style
	Hint      lipgloss.Style
}

// DefaultStyles returns the default detail view styles
func DefaultStyles() *Styles {
	return &Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("237")).
			Padding(0, 1),
		Content: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		LineNum: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(5).
			Align(lipgloss.Right),
		KeyWord: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")),
		String: lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")),
		Number: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// New creates a new detail view component
func New(resourceName, content string, format Format) *Model {
	return &Model{
		resourceName: resourceName,
		content:      content,
		format:       format,
		styles:       DefaultStyles(),
	}
}

// SetContent updates the content and format
func (m *Model) SetContent(resourceName, content string, format Format) {
	m.resourceName = resourceName
	m.content = content
	m.format = format
	m.scrollOffset = 0
}

// SetSize sets the dimensions for the detail view
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ensureScrollBounds()
}

// ScrollUp scrolls up by one line
func (m *Model) ScrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
}

// ScrollDown scrolls down by one line
func (m *Model) ScrollDown() {
	m.scrollOffset++
	m.ensureScrollBounds()
}

// ScrollToTop scrolls to the top
func (m *Model) ScrollToTop() {
	m.scrollOffset = 0
}

// ScrollToBottom scrolls to the bottom
func (m *Model) ScrollToBottom() {
	lines := strings.Split(m.content, "\n")
	m.scrollOffset = len(lines) - m.contentHeight()
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// PageUp scrolls up by one page
func (m *Model) PageUp() {
	m.scrollOffset -= m.contentHeight()
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// PageDown scrolls down by one page
func (m *Model) PageDown() {
	m.scrollOffset += m.contentHeight()
	m.ensureScrollBounds()
}

// contentHeight returns the available height for content
func (m *Model) contentHeight() int {
	// Reserve 2 lines for header and hint
	h := m.height - 2
	if h < 1 {
		h = 1
	}
	return h
}

// ensureScrollBounds ensures scroll offset is within valid bounds
func (m *Model) ensureScrollBounds() {
	lines := strings.Split(m.content, "\n")
	maxOffset := len(lines) - m.contentHeight()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// Content returns the current content
func (m *Model) Content() string {
	return m.content
}

// Format returns the current format
func (m *Model) Format() Format {
	return m.format
}

// ResourceName returns the resource name
func (m *Model) ResourceName() string {
	return m.resourceName
}

// View renders the detail view
func (m *Model) View() string {
	var b strings.Builder

	// Header with resource name and format
	headerText := fmt.Sprintf("%s - %s", m.resourceName, m.format.String())
	hint := m.styles.Hint.Render("[Esc] Back")
	headerPadding := m.width - lipgloss.Width(headerText) - lipgloss.Width(hint) - 4
	if headerPadding < 1 {
		headerPadding = 1
	}
	header := headerText + strings.Repeat(" ", headerPadding) + hint
	b.WriteString(m.styles.Header.Width(m.width).Render(header))
	b.WriteString("\n")

	// Content
	lines := strings.Split(m.content, "\n")
	contentHeight := m.contentHeight()

	endIdx := m.scrollOffset + contentHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		line := lines[i]
		// Truncate long lines
		if len(line) > m.width-2 {
			line = line[:m.width-5] + "..."
		}
		b.WriteString(m.styles.Content.Render(line))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Fill remaining space
	renderedLines := endIdx - m.scrollOffset
	for i := renderedLines; i < contentHeight; i++ {
		b.WriteString("\n")
	}

	// Scroll indicator
	b.WriteString("\n")
	scrollInfo := fmt.Sprintf("Line %d-%d of %d", m.scrollOffset+1, endIdx, len(lines))
	if m.scrollOffset > 0 {
		scrollInfo = "↑ " + scrollInfo
	}
	if endIdx < len(lines) {
		scrollInfo = scrollInfo + " ↓"
	}
	b.WriteString(m.styles.Hint.Render(scrollInfo))

	return b.String()
}

// String implements fmt.Stringer
func (m *Model) String() string {
	return fmt.Sprintf("Detail{resource=%q, format=%s, lines=%d}", m.resourceName, m.format, len(strings.Split(m.content, "\n")))
}
