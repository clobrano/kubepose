package search

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the search component
type Model struct {
	input         textinput.Model
	active        bool
	originalRows  [][]string
	filteredRows  [][]string
	styles        *Styles
}

// Styles defines the styles for the search component
type Styles struct {
	Container lipgloss.Style
	Prompt    lipgloss.Style
	Input     lipgloss.Style
}

// DefaultStyles returns the default search styles
func DefaultStyles() *Styles {
	return &Styles{
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Padding(0, 1),
		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),
		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
	}
}

// New creates a new search component
func New() *Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 100
	ti.Width = 40

	return &Model{
		input:  ti,
		styles: DefaultStyles(),
	}
}

// Activate enables search mode and focuses the input
func (m *Model) Activate() {
	m.active = true
	m.input.Focus()
	m.input.SetValue("")
	m.filteredRows = m.originalRows
}

// Deactivate disables search mode
func (m *Model) Deactivate() {
	m.active = false
	m.input.Blur()
	m.input.SetValue("")
	m.filteredRows = m.originalRows
}

// IsActive returns whether search mode is active
func (m *Model) IsActive() bool {
	return m.active
}

// SetItems sets the searchable items
func (m *Model) SetItems(rows [][]string) {
	m.originalRows = rows
	if m.active && m.input.Value() != "" {
		m.filteredRows = FilterRows(m.input.Value(), rows)
	} else {
		m.filteredRows = rows
	}
}

// Filter filters items based on the current query
func (m *Model) Filter(query string) {
	m.filteredRows = FilterRows(query, m.originalRows)
}

// FilteredItems returns the current filtered results
func (m *Model) FilteredItems() [][]string {
	return m.filteredRows
}

// Query returns the current search query
func (m *Model) Query() string {
	return m.input.Value()
}

// Update handles input messages
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Deactivate()
			return m, nil
		case "enter":
			// Keep filter active but stop typing
			m.input.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Update filter on every keystroke
	m.Filter(m.input.Value())

	return m, cmd
}

// View renders the search input
func (m *Model) View() string {
	if !m.active {
		return ""
	}

	prompt := m.styles.Prompt.Render("/ ")
	input := m.input.View()

	return m.styles.Container.Render(prompt + input)
}

// SetWidth sets the width for the search input
func (m *Model) SetWidth(width int) {
	m.input.Width = width - 10 // Account for prompt and padding
	if m.input.Width < 10 {
		m.input.Width = 10
	}
}
