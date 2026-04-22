package search

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the search component
type Model struct {
	input         textinput.Model
	active        bool
	query         string
	originalRows  [][]string
	filteredRows  [][]string
	styles        *Styles

	// History navigation
	history      []string // oldest → newest
	historyIndex int      // -1 = editing draft, 0..n-1 = position in history
	draft        string   // saved draft text before navigating history
	suggestion   string   // autosuggestion suffix from history
}

// Styles defines the styles for the search component
type Styles struct {
	Container  lipgloss.Style
	Prompt     lipgloss.Style
	Input      lipgloss.Style
	Suggestion lipgloss.Style
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
		Suggestion: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// New creates a new search component
func New() *Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 100
	ti.Width = 40

	return &Model{
		input:        ti,
		styles:       DefaultStyles(),
		historyIndex: -1,
	}
}

// Activate enables search mode and focuses the input.
// If a filter is already applied, keep it so the user can edit it.
func (m *Model) Activate() {
	m.active = true
	m.historyIndex = -1
	m.draft = ""
	m.suggestion = ""
	m.input.Focus()
	if m.query != "" {
		m.input.SetValue(m.query)
	} else {
		m.input.SetValue("")
		m.filteredRows = m.originalRows
	}
}

// Deactivate disables search mode and clears the filter
func (m *Model) Deactivate() {
	m.active = false
	m.query = ""
	m.input.Blur()
	m.input.SetValue("")
	m.filteredRows = m.originalRows
	m.historyIndex = -1
	m.draft = ""
	m.suggestion = ""
}

// IsActive returns whether search mode is active (typing mode)
func (m *Model) IsActive() bool {
	return m.active
}

// IsFiltered returns true when a filter query is applied (even after Enter confirms it)
func (m *Model) IsFiltered() bool {
	return m.query != ""
}

// SetItems sets the searchable items
func (m *Model) SetItems(rows [][]string) {
	m.originalRows = rows
	if m.query != "" {
		m.filteredRows = FilterRows(m.query, rows)
	} else {
		m.filteredRows = rows
	}
}

// Filter filters items based on the current query
func (m *Model) Filter(query string) {
	m.query = query
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

// HasSuggestion returns true when an autosuggestion is available.
func (m *Model) HasSuggestion() bool {
	return m.suggestion != ""
}

// HistoryLen returns the number of entries in the filter history.
func (m *Model) HistoryLen() int {
	return len(m.history)
}

// addToHistory adds a confirmed query to the history, skipping empty or duplicate entries.
func (m *Model) addToHistory(query string) {
	if query == "" {
		return
	}
	if len(m.history) > 0 && m.history[len(m.history)-1] == query {
		return
	}
	m.history = append(m.history, query)
}

// computeSuggestion finds the most recent history entry that starts with the
// current input and stores the remaining suffix as the suggestion.
func (m *Model) computeSuggestion(input string) {
	if input == "" {
		m.suggestion = ""
		return
	}
	for i := len(m.history) - 1; i >= 0; i-- {
		if strings.HasPrefix(m.history[i], input) && m.history[i] != input {
			m.suggestion = m.history[i][len(input):]
			return
		}
	}
	m.suggestion = ""
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
			if m.input.Value() != "" {
				m.addToHistory(m.input.Value())
			}
			m.active = false
			m.input.Blur()
			m.historyIndex = -1
			m.draft = ""
			m.suggestion = ""
			return m, nil

		case "up":
			if len(m.history) == 0 {
				return m, nil
			}
			if m.historyIndex == -1 {
				m.draft = m.input.Value()
				m.historyIndex = len(m.history) - 1
			} else if m.historyIndex > 0 {
				m.historyIndex--
			}
			val := m.history[m.historyIndex]
			m.input.SetValue(val)
			m.Filter(val)
			m.suggestion = ""
			return m, nil

		case "down":
			if m.historyIndex == -1 {
				return m, nil
			}
			if m.historyIndex < len(m.history)-1 {
				m.historyIndex++
				val := m.history[m.historyIndex]
				m.input.SetValue(val)
				m.Filter(val)
			} else {
				m.historyIndex = -1
				m.input.SetValue(m.draft)
				m.Filter(m.draft)
			}
			m.suggestion = ""
			return m, nil

		case "tab":
			if m.suggestion != "" {
				full := m.input.Value() + m.suggestion
				m.input.SetValue(full)
				m.Filter(full)
				m.suggestion = ""
				m.historyIndex = -1
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.Filter(m.input.Value())
	if _, ok := msg.(tea.KeyMsg); ok {
		m.historyIndex = -1
		m.computeSuggestion(m.input.Value())
	}

	return m, cmd
}

// View renders the search input
func (m *Model) View() string {
	if !m.active && !m.IsFiltered() {
		return ""
	}

	prompt := m.styles.Prompt.Render("/ ")
	input := m.input.View()

	if m.active && m.suggestion != "" {
		suggestion := m.styles.Suggestion.Render(m.suggestion)
		return m.styles.Container.Render(prompt + input + suggestion)
	}

	return m.styles.Container.Render(prompt + input)
}

// RestoreFilter re-applies a previously saved filter query without entering
// active (typing) mode. Used when switching back to a tab that had a filter.
func (m *Model) RestoreFilter(query string) {
	if query == "" {
		return
	}
	m.query = query
	m.input.SetValue(query)
	m.filteredRows = FilterRows(query, m.originalRows)
}

// SetWidth sets the width for the search input
func (m *Model) SetWidth(width int) {
	m.input.Width = width - 10 // Account for prompt and padding
	if m.input.Width < 10 {
		m.input.Width = 10
	}
}
