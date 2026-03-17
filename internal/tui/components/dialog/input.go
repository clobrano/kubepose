package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputResult represents the result of an input dialog
type InputResult int

const (
	InputPending InputResult = iota
	InputSubmitted
	InputCancelled
)

// InputModel represents an input dialog
type InputModel struct {
	title       string
	placeholder string
	hint        string
	textInput   textinput.Model
	result      InputResult
	width       int
	height      int
	styles      *InputStyles
	actionID    string
}

// InputStyles defines the styles for the input dialog
type InputStyles struct {
	Dialog  lipgloss.Style
	Title   lipgloss.Style
	Input   lipgloss.Style
	Hint    lipgloss.Style
}

// DefaultInputStyles returns the default input styles
func DefaultInputStyles() *InputStyles {
	return &InputStyles{
		Dialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// NewInput creates a new input dialog
func NewInput(title, placeholder string) *InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 30

	return &InputModel{
		title:       title,
		placeholder: placeholder,
		textInput:   ti,
		result:      InputPending,
		styles:      DefaultInputStyles(),
	}
}

// WithHint sets a descriptive hint shown below the title
func (m *InputModel) WithHint(hint string) *InputModel {
	m.hint = hint
	return m
}

// WithActionID sets an identifier for the action
func (m *InputModel) WithActionID(id string) *InputModel {
	m.actionID = id
	return m
}

// WithValue sets the initial value (builder pattern)
func (m *InputModel) WithValue(value string) *InputModel {
	m.textInput.SetValue(value)
	return m
}

// SetValue sets the input value and places the cursor at the end
func (m *InputModel) SetValue(value string) {
	m.textInput.SetValue(value)
	m.textInput.CursorEnd()
}

// SetSize sets the dialog dimensions and adjusts the input width to 70% of available space
func (m *InputModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Dialog is 70% of terminal width, minus border/padding (2 border + 4 padding = 6)
	dialogInner := width*70/100 - 6
	if dialogInner > 0 {
		m.textInput.Width = dialogInner
	}
}

// Result returns the current result
func (m *InputModel) Result() InputResult {
	return m.result
}

// Value returns the input value
func (m *InputModel) Value() string {
	return m.textInput.Value()
}

// ActionID returns the action identifier
func (m *InputModel) ActionID() string {
	return m.actionID
}

// Reset resets the dialog to pending state
func (m *InputModel) Reset() {
	m.result = InputPending
	m.textInput.SetValue("")
	m.textInput.Focus()
}

// Update handles messages
func (m *InputModel) Update(msg tea.Msg) (*InputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.result = InputSubmitted
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = InputCancelled
			return m, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the input dialog
func (m *InputModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render(m.title))
	b.WriteString("\n")

	// Hint (if set)
	if m.hint != "" {
		b.WriteString(m.styles.Hint.Render(m.hint))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Input field
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Key hints
	b.WriteString(m.styles.Hint.Render("[Enter] submit  [Esc] cancel"))

	// Use 70% of terminal width for the dialog
	dialogWidth := m.width * 70 / 100
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	dialogStyle := m.styles.Dialog.Width(dialogWidth)
	content := dialogStyle.Render(b.String())

	// Center the dialog
	if m.width > 0 && m.height > 0 {
		dialogWidth := lipgloss.Width(content)
		dialogHeight := lipgloss.Height(content)

		horizontalPadding := (m.width - dialogWidth) / 2
		verticalPadding := (m.height - dialogHeight) / 2

		if horizontalPadding < 0 {
			horizontalPadding = 0
		}
		if verticalPadding < 0 {
			verticalPadding = 0
		}

		var centered strings.Builder
		for i := 0; i < verticalPadding; i++ {
			centered.WriteString(strings.Repeat(" ", m.width))
			centered.WriteString("\n")
		}

		lines := strings.Split(content, "\n")
		for _, line := range lines {
			centered.WriteString(strings.Repeat(" ", horizontalPadding))
			centered.WriteString(line)
			centered.WriteString("\n")
		}

		return centered.String()
	}

	return content
}

// String implements fmt.Stringer
func (m *InputModel) String() string {
	return fmt.Sprintf("Input{title=%q, value=%q}", m.title, m.textInput.Value())
}
