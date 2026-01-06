package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MFAInput is a component for entering MFA codes
type MFAInput struct {
	textInput textinput.Model
	err       error
}

// NewMFAInput creates a new MFA input component
func NewMFAInput() MFAInput {
	ti := textinput.New()
	ti.Placeholder = "000000"
	ti.Focus()
	ti.CharLimit = 6
	ti.Width = 20
	ti.Prompt = "> "

	return MFAInput{
		textInput: ti,
	}
}

// Value returns the current input value
func (m *MFAInput) Value() string {
	return m.textInput.Value()
}

// Update updates the MFA input component
func (m *MFAInput) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return cmd
}

// View renders the MFA input component
func (m *MFAInput) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(50)

	content := titleStyle.Render("MFA Authentication Required") + "\n\n" +
		instructionStyle.Render("Enter your 6-digit MFA code:") + "\n\n" +
		m.textInput.View() + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press Enter to submit | Esc to cancel")

	return boxStyle.Render(content)
}

// Reset clears the input
func (m *MFAInput) Reset() {
	m.textInput.SetValue("")
	m.textInput.Focus()
}
