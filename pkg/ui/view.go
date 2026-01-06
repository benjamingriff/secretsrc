package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var content string

	switch m.currentScreen {
	case ScreenSecretList:
		content = m.viewSecretList()
	case ScreenSecretDetail:
		content = m.viewSecretDetail()
	case ScreenProfileSelector:
		content = m.viewProfileSelector()
	case ScreenRegionSelector:
		content = m.viewRegionSelector()
	case ScreenMFAInput:
		content = m.viewMFAInput()
	default:
		content = "Unknown screen"
	}

	// Build the header and footer
	header := m.viewHeader()
	footer := m.viewFooter()

	// Calculate heights for each section
	// Border takes 2 lines, padding takes 0, header ~3 lines, footer varies (1-4 lines)
	footerHeight := lipgloss.Height(footer)
	headerHeight := lipgloss.Height(header)
	availableHeight := m.height - 4 - headerHeight - footerHeight // 4 for border and padding

	// Ensure content fills the available height to push footer to bottom
	contentStyle := lipgloss.NewStyle().
		Height(availableHeight)

	styledContent := contentStyle.Render(content)

	// Join header, content, and footer vertically
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledContent,
		footer,
	)

	// Create a bordered style that fills the terminal
	appStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Width(m.width - 2).   // Account for border width
		Height(m.height - 2). // Account for border height
		Padding(0, 1)         // Internal padding

	return appStyle.Render(fullContent)
}

// viewHeader renders the header
func (m Model) viewHeader() string {
	title := "Secret Src - AWS Secrets Manager TUI"
	info := fmt.Sprintf("Profile: %s | Region: %s", m.currentProfile, m.currentRegion)

	return fmt.Sprintf("%s\n%s",
		HeaderStyle.Render(title),
		StatusBarStyle.Render(info),
	)
}

// viewFooter renders the footer with help text and status
func (m Model) viewFooter() string {
	var parts []string

	// Show error if present
	if m.errorMessage != "" {
		parts = append(parts, ErrorStyle.Render(fmt.Sprintf("Error: %s", m.errorMessage)))
	}

	// Show status message if present
	if m.statusMessage != "" {
		parts = append(parts, SuccessStyle.Render(m.statusMessage))
	}

	// Show loading indicator
	if m.loading {
		parts = append(parts, "Loading...")
	}

	// Show help based on current screen
	var help string
	switch m.currentScreen {
	case ScreenSecretList:
		help = "enter: view | p: profile | g: region | r: refresh | ?: help | q: quit"
		if m.currentPage > 0 {
			help += " | b: prev page"
		}
		if m.hasMore {
			help += " | n: next page"
		}
	case ScreenSecretDetail:
		if m.secretValue == "" {
			help = "v: view value | esc: back | q: quit"
		} else {
			help = "c: copy plain | j: copy json | esc: back | q: quit"
		}
	case ScreenProfileSelector:
		help = "enter: select | esc: back | q: quit"
	case ScreenRegionSelector:
		help = "enter: select | esc: back | q: quit"
	case ScreenMFAInput:
		help = "enter: submit | esc: cancel"
	}

	if help != "" {
		parts = append(parts, HelpStyle.Render(help))
	}

	return strings.Join(parts, "\n")
}

// viewSecretList renders the secret list screen
func (m Model) viewSecretList() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if len(m.secrets) == 0 && !m.loading {
		return "\n  No secrets found in this region.\n\n  Try switching regions with 'g' or refreshing with 'r'."
	}

	return m.list.View()
}

// viewSecretDetail renders the secret detail screen
func (m Model) viewSecretDetail() string {
	secret := m.list.SelectedSecret()
	if secret == nil {
		return "No secret selected"
	}

	var b strings.Builder

	b.WriteString(TitleStyle.Render("Secret Details") + "\n\n")

	// Secret metadata
	b.WriteString(DetailKeyStyle.Render("Name:") + " " + DetailValueStyle.Render(secret.Name) + "\n")
	b.WriteString(DetailKeyStyle.Render("ARN:") + " " + DetailValueStyle.Render(secret.ARN) + "\n")

	if secret.Description != "" {
		b.WriteString(DetailKeyStyle.Render("Description:") + " " + DetailValueStyle.Render(secret.Description) + "\n")
	}

	if secret.LastChangedDate != nil {
		b.WriteString(DetailKeyStyle.Render("Last Modified:") + " " +
			DetailValueStyle.Render(secret.LastChangedDate.Format(time.RFC1123)) + "\n")
	}

	if len(secret.Tags) > 0 {
		b.WriteString("\n" + DetailKeyStyle.Render("Tags:") + "\n")
		for k, v := range secret.Tags {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	// Secret value section
	b.WriteString("\n" + strings.Repeat("─", 60) + "\n\n")

	if m.secretValue == "" {
		b.WriteString(HelpStyle.Render("Press 'v' to view the secret value\n"))
	} else {
		b.WriteString(DetailKeyStyle.Render("Secret Value:") + "\n\n")

		// Try to format as JSON if possible
		var formatted string
		var jsonData interface{}
		if err := json.Unmarshal([]byte(m.secretValue), &jsonData); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
			if err == nil {
				formatted = string(prettyJSON)
			} else {
				formatted = m.secretValue
			}
		} else {
			formatted = m.secretValue
		}

		valueStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("241")).
			Padding(1).
			MaxWidth(m.width - 10)

		b.WriteString(valueStyle.Render(formatted) + "\n")
	}

	return b.String()
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	help := `
AWS Secrets Manager TUI - Help

NAVIGATION
  ↑/k         Move up
  ↓/j         Move down
  enter       Select item / View secret details
  esc/q       Go back / Quit

ACTIONS
  v           View secret value (on detail screen)
  c           Copy secret value as plain text
  j           Copy secret value as JSON
  r           Refresh secret list
  p           Switch AWS profile
  g           Switch AWS region
  n           Next page (when available)
  b           Previous page (when available)

GLOBAL
  ?           Toggle this help
  ctrl+c      Force quit

SECURITY NOTE
  • Secret values are only fetched on-demand (when you press 'v')
  • Values are cleared from memory when you navigate away
  • Clipboard contents persist after app closes

Press '?' to close this help.
`
	return BorderStyle.Render(help)
}

// viewProfileSelector renders the profile selector screen
func (m Model) viewProfileSelector() string {
	return m.profileSelector.View()
}

// viewRegionSelector renders the region selector screen
func (m Model) viewRegionSelector() string {
	return m.regionSelector.View()
}

// viewMFAInput renders the MFA input screen
func (m Model) viewMFAInput() string {
	// Center the MFA input box
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width-6, m.height-10,
			lipgloss.Center, lipgloss.Center,
			m.mfaInput.View())
	}
	return m.mfaInput.View()
}
