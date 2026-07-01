package ui

import (
	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/charmbracelet/lipgloss"
)

// Mode accent colours. Fluoro pink signals Secrets Manager; fluoro green
// signals SSM Parameter Store.
const (
	accentSecretsManager = lipgloss.Color("#FF10F0") // Fluoro pink
	accentSSM            = lipgloss.Color("#39FF14") // Fluoro green
)

// accentColor returns the accent colour for the given mode.
func accentColor(mode models.Kind) lipgloss.Color {
	if mode == models.KindParameter {
		return accentSSM
	}
	return accentSecretsManager
}

var (
	// Colors
	primaryColor   = lipgloss.Color("205") // Pink
	secondaryColor = lipgloss.Color("170") // Purple
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red
	subtleColor    = lipgloss.Color("241") // Gray

	// Header style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1)

	// Selected item style
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				PaddingLeft(2)

	// Normal item style
	NormalItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(1)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	// Detail view styles
	DetailKeyStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Width(20)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Underline(true)

	// Filter status style (for grid filtering)
	FilterStatusStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true).
				MarginBottom(1)
)
