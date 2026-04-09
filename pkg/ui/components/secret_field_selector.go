package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SecretField represents a top-level field from a JSON secret.
type SecretField struct {
	Key       string
	CopyValue string
	Preview   string
}

// secretFieldItem is a list item for a secret field.
type secretFieldItem struct {
	field SecretField
}

// FilterValue implements list.Item.
func (i secretFieldItem) FilterValue() string {
	return i.field.Key + " " + i.field.Preview
}

// Title returns the list item title.
func (i secretFieldItem) Title() string {
	return i.field.Key
}

// Description returns the list item description.
func (i secretFieldItem) Description() string {
	return i.field.Preview
}

// SecretFieldSelector is a component for selecting a secret field to copy.
type SecretFieldSelector struct {
	list list.Model
}

// NewSecretFieldSelector creates a new secret field selector.
func NewSecretFieldSelector(fields []SecretField, width, height int) SecretFieldSelector {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(2)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		PaddingLeft(2)

	items := make([]list.Item, len(fields))
	for i, field := range fields {
		items[i] = secretFieldItem{field: field}
	}

	l := list.New(items, delegate, width, height)
	l.Title = "Copy Secret Field"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return SecretFieldSelector{
		list: l,
	}
}

// SelectedField returns the selected field, or nil if none is selected.
func (sfs *SecretFieldSelector) SelectedField() *SecretField {
	item := sfs.list.SelectedItem()
	if item == nil {
		return nil
	}
	fieldItem, ok := item.(secretFieldItem)
	if !ok {
		return nil
	}
	return &fieldItem.field
}

// Update updates the selector.
func (sfs *SecretFieldSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	sfs.list, cmd = sfs.list.Update(msg)
	return cmd
}

// View renders the selector.
func (sfs *SecretFieldSelector) View() string {
	return sfs.list.View()
}

// SetSize updates the selector dimensions.
func (sfs *SecretFieldSelector) SetSize(width, height int) {
	sfs.list.SetSize(width, height)
}
