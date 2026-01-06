package ui

import "github.com/charmbracelet/lipgloss"

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
)
