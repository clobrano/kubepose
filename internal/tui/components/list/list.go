package list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Model represents the resource list component
type Model struct {
	headers        []string
	rows           [][]string
	cursor         int
	viewportOffset int
	width          int
	height         int
	columnWidths   []int
	styles         *Styles
}

// Styles defines the styles for the list component
type Styles struct {
	Header       lipgloss.Style
	Item         lipgloss.Style
	Selected     lipgloss.Style
	Cursor       lipgloss.Style
	Empty        lipgloss.Style
}

// DefaultStyles returns the default list styles
func DefaultStyles() *Styles {
	return &Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("39")).
			Bold(true),
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),
		Empty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
	}
}

// New creates a new list component
func New(headers []string, rows [][]string) *Model {
	m := &Model{
		styles: DefaultStyles(),
	}
	m.SetItems(headers, rows)
	return m
}

// SetItems updates the list content
func (m *Model) SetItems(headers []string, rows [][]string) {
	m.headers = headers
	m.rows = rows
	m.cursor = 0
	m.viewportOffset = 0
	m.calculateColumnWidths()
}

// SetSize sets the dimensions for the list
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ensureCursorVisible()
}

// calculateColumnWidths calculates optimal column widths based on content
func (m *Model) calculateColumnWidths() {
	if len(m.headers) == 0 {
		m.columnWidths = nil
		return
	}

	m.columnWidths = make([]int, len(m.headers))

	// Start with header widths
	for i, h := range m.headers {
		m.columnWidths[i] = len(h)
	}

	// Check row widths
	for _, row := range m.rows {
		for i, cell := range row {
			if i < len(m.columnWidths) && len(cell) > m.columnWidths[i] {
				m.columnWidths[i] = len(cell)
			}
		}
	}

	// Add padding
	for i := range m.columnWidths {
		m.columnWidths[i] += 2
	}
}

// MoveUp moves the cursor up
func (m *Model) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.ensureCursorVisible()
	}
}

// MoveDown moves the cursor down
func (m *Model) MoveDown() {
	if m.cursor < len(m.rows)-1 {
		m.cursor++
		m.ensureCursorVisible()
	}
}

// MoveToTop moves the cursor to the first item
func (m *Model) MoveToTop() {
	m.cursor = 0
	m.viewportOffset = 0
}

// MoveToBottom moves the cursor to the last item
func (m *Model) MoveToBottom() {
	if len(m.rows) > 0 {
		m.cursor = len(m.rows) - 1
		m.ensureCursorVisible()
	}
}

// ensureCursorVisible adjusts viewport to keep cursor visible
func (m *Model) ensureCursorVisible() {
	visibleRows := m.visibleRowCount()
	if visibleRows <= 0 {
		return
	}

	// Cursor above viewport
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}

	// Cursor below viewport
	if m.cursor >= m.viewportOffset+visibleRows {
		m.viewportOffset = m.cursor - visibleRows + 1
	}
}

// visibleRowCount returns the number of rows that can be displayed
func (m *Model) visibleRowCount() int {
	// Reserve 1 line for header
	if m.height <= 1 {
		return 0
	}
	return m.height - 1
}

// SelectedItem returns the currently selected row
func (m *Model) SelectedItem() []string {
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		return m.rows[m.cursor]
	}
	return nil
}

// SelectedIndex returns the cursor position
func (m *Model) SelectedIndex() int {
	return m.cursor
}

// RowCount returns the total number of rows
func (m *Model) RowCount() int {
	return len(m.rows)
}

// Headers returns the column headers
func (m *Model) Headers() []string {
	return m.headers
}

// View renders the list
func (m *Model) View() string {
	if len(m.headers) == 0 {
		return m.styles.Empty.Render("No data")
	}

	var b strings.Builder

	// Render header
	b.WriteString(m.renderRow(m.headers, -1))
	b.WriteString("\n")

	// Render visible rows
	visibleRows := m.visibleRowCount()
	if visibleRows <= 0 {
		return b.String()
	}

	endIdx := m.viewportOffset + visibleRows
	if endIdx > len(m.rows) {
		endIdx = len(m.rows)
	}

	for i := m.viewportOffset; i < endIdx; i++ {
		b.WriteString(m.renderRow(m.rows[i], i))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Fill remaining space if needed
	renderedRows := endIdx - m.viewportOffset
	for i := renderedRows; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

// renderRow renders a single row
func (m *Model) renderRow(cells []string, rowIndex int) string {
	var parts []string

	// Add cursor indicator
	cursorStr := "  "
	if rowIndex == m.cursor && rowIndex >= 0 {
		cursorStr = m.styles.Cursor.Render("> ")
	}
	parts = append(parts, cursorStr)

	// Render each cell
	for i, cell := range cells {
		width := 10 // default width
		if i < len(m.columnWidths) {
			width = m.columnWidths[i]
		}

		// Truncate if needed
		if len(cell) > width {
			cell = cell[:width-1] + "…"
		}

		// Pad to width
		cell = fmt.Sprintf("%-*s", width, cell)

		parts = append(parts, cell)
	}

	content := strings.Join(parts, "")

	// Apply styling
	if rowIndex == -1 {
		// Header row
		return m.styles.Header.Width(m.width).Render(content)
	} else if rowIndex == m.cursor {
		// Selected row
		return m.styles.Selected.Width(m.width).Render(content)
	}

	return m.styles.Item.Render(content)
}

// String implements fmt.Stringer
func (m *Model) String() string {
	return fmt.Sprintf("List{rows=%d, cursor=%d}", len(m.rows), m.cursor)
}
