package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RegionItem represents a region in the list
type RegionItem struct {
	code      string
	name      string
	isCurrent bool
}

// FilterValue implements list.Item
func (i RegionItem) FilterValue() string {
	return i.code + " " + i.name
}

// Title returns the title for the list item
func (i RegionItem) Title() string {
	if i.isCurrent {
		return "• " + i.code + " (current)"
	}
	return "  " + i.code
}

// Description returns the description for the list item
func (i RegionItem) Description() string {
	return i.name
}

// RegionSelector is a component for selecting AWS regions
type RegionSelector struct {
	list          list.Model
	regions       map[string]string
	currentRegion string
}

// Common AWS regions with their descriptions
var regionDescriptions = map[string]string{
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ca-central-1":   "Canada (Central)",
	"eu-west-1":      "Europe (Ireland)",
	"eu-west-2":      "Europe (London)",
	"eu-west-3":      "Europe (Paris)",
	"eu-central-1":   "Europe (Frankfurt)",
	"eu-north-1":     "Europe (Stockholm)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-northeast-3": "Asia Pacific (Osaka)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"sa-east-1":      "South America (São Paulo)",
}

// NewRegionSelector creates a new region selector
func NewRegionSelector(regions []string, currentRegion string, width, height int) RegionSelector {
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
	items := make([]list.Item, len(regions))
	for i, region := range regions {
		name := regionDescriptions[region]
		if name == "" {
			name = "AWS Region"
		}
		items[i] = RegionItem{
			code:      region,
			name:      name,
			isCurrent: region == currentRegion,
		}
	}

	l := list.New(items, delegate, width, height)
	l.Title = "Select AWS Region"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return RegionSelector{
		list:          l,
		regions:       regionDescriptions,
		currentRegion: currentRegion,
	}
}

// SelectedRegion returns the selected region code, or empty string if none selected
func (rs *RegionSelector) SelectedRegion() string {
	item := rs.list.SelectedItem()
	if item == nil {
		return ""
	}
	regionItem, ok := item.(RegionItem)
	if !ok {
		return ""
	}
	return regionItem.code
}

// Update updates the region selector
func (rs *RegionSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	rs.list, cmd = rs.list.Update(msg)
	return cmd
}

// View renders the region selector
func (rs *RegionSelector) View() string {
	return rs.list.View()
}

// SetSize updates the selector dimensions
func (rs *RegionSelector) SetSize(width, height int) {
	rs.list.SetSize(width, height)
}
