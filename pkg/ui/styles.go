package ui

import (
	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/charmbracelet/lipgloss"
)

// Mode accent colours. Pink signals Secrets Manager; fluoro green signals SSM
// Parameter Store. The Secrets Manager pink matches primaryColor ("205").
const (
	accentSecretsManager = lipgloss.Color("205")     // Pink (xterm-256)
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

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1)

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

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1)

	// Filter status style (for grid filtering)
	FilterStatusStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true).
				MarginBottom(1)
)
