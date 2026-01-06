package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

// SecretListItem wraps a Secret for the list component
type SecretListItem struct {
	Secret models.Secret
}

// FilterValue implements list.Item
func (i SecretListItem) FilterValue() string {
	return i.Secret.Name
}

// Title returns the title for the list item
func (i SecretListItem) Title() string {
	return i.Secret.Name
}

// Description returns the description for the list item
func (i SecretListItem) Description() string {
	if i.Secret.LastChangedDate != nil {
		return fmt.Sprintf("Last Modified: %s", i.Secret.LastChangedDate.Format(time.RFC1123))
	}
	return "Last Modified: Unknown"
}

// SecretList wraps the bubbles list component
type SecretList struct {
	list list.Model
}

// NewSecretList creates a new secret list component
func NewSecretList(width, height int) SecretList {
	delegate := list.NewDefaultDelegate()

	// Customize delegate styles
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(2)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		PaddingLeft(2)

	l := list.New([]list.Item{}, delegate, width, height)
	l.Title = "AWS Secrets Manager"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return SecretList{
		list: l,
	}
}

// SetSecrets updates the list with new secrets
func (sl *SecretList) SetSecrets(secrets []models.Secret) {
	items := make([]list.Item, len(secrets))
	for i, secret := range secrets {
		items[i] = SecretListItem{Secret: secret}
	}
	sl.list.SetItems(items)
}

// SelectedSecret returns the currently selected secret, or nil if none selected
func (sl *SecretList) SelectedSecret() *models.Secret {
	item := sl.list.SelectedItem()
	if item == nil {
		return nil
	}
	secretItem, ok := item.(SecretListItem)
	if !ok {
		return nil
	}
	return &secretItem.Secret
}

// Update updates the list component
func (sl *SecretList) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	sl.list, cmd = sl.list.Update(msg)
	return cmd
}

// View renders the list component
func (sl *SecretList) View() string {
	return sl.list.View()
}

// SetSize updates the list dimensions
func (sl *SecretList) SetSize(width, height int) {
	sl.list.SetSize(width, height)
}
