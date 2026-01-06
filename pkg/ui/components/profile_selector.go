package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfileItem represents a profile in the list
type ProfileItem struct {
	name      string
	isCurrent bool
}

// FilterValue implements list.Item
func (i ProfileItem) FilterValue() string {
	return i.name
}

// Title returns the title for the list item
func (i ProfileItem) Title() string {
	if i.isCurrent {
		return "• " + i.name + " (current)"
	}
	return "  " + i.name
}

// Description returns the description for the list item
func (i ProfileItem) Description() string {
	if i.isCurrent {
		return "Currently active profile"
	}
	return "AWS profile"
}

// ProfileSelector is a component for selecting AWS profiles
type ProfileSelector struct {
	list          list.Model
	profiles      []string
	currentProfile string
}

// NewProfileSelector creates a new profile selector
func NewProfileSelector(profiles []string, currentProfile string, width, height int) ProfileSelector {
	delegate := list.NewDefaultDelegate()

	// Customize delegate styles
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(2)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		PaddingLeft(2)

	// Create list items
	items := make([]list.Item, len(profiles))
	for i, profile := range profiles {
		items[i] = ProfileItem{
			name:      profile,
			isCurrent: profile == currentProfile,
		}
	}

	l := list.New(items, delegate, width, height)
	l.Title = "Select AWS Profile"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return ProfileSelector{
		list:          l,
		profiles:      profiles,
		currentProfile: currentProfile,
	}
}

// SelectedProfile returns the selected profile name, or empty string if none selected
func (ps *ProfileSelector) SelectedProfile() string {
	item := ps.list.SelectedItem()
	if item == nil {
		return ""
	}
	profileItem, ok := item.(ProfileItem)
	if !ok {
		return ""
	}
	return profileItem.name
}

// Update updates the profile selector
func (ps *ProfileSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	ps.list, cmd = ps.list.Update(msg)
	return cmd
}

// View renders the profile selector
func (ps *ProfileSelector) View() string {
	return ps.list.View()
}

// SetSize updates the selector dimensions
func (ps *ProfileSelector) SetSize(width, height int) {
	ps.list.SetSize(width, height)
}
