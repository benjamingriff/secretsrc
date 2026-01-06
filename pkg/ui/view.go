package ui

import (
	"encoding/json"
	"fmt"
	"strings"

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
		help = "hjkl/arrows: navigate | enter: view | /: filter | p: profile | g: region | r: refresh | ?: help | q: quit"
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

	// Show filter status if filtering
	if m.grid.IsFiltering() {
		filterStatus := fmt.Sprintf("Filter: %s_", m.grid.GetFilterQuery())
		return fmt.Sprintf("%s\n%s", FilterStatusStyle.Render(filterStatus), m.grid.View())
	}

	return m.grid.View()
}

// viewSecretDetail renders the secret detail screen
func (m Model) viewSecretDetail() string {
	secret := m.grid.SelectedSecret()
	if secret == nil {
		return "No secret selected"
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Secret Details") + "\n\n")

	// Secret metadata with compact key-value styling
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Truncate name if too long
	displayName := secret.Name
	if len(displayName) > 60 {
		displayName = displayName[:57] + "..."
	}
	b.WriteString(keyStyle.Render("Name: ") + valueStyle.Render(displayName) + "\n")

	// Truncate ARN if too long
	displayARN := secret.ARN
	if len(displayARN) > 60 {
		displayARN = "..." + displayARN[len(displayARN)-57:]
	}
	b.WriteString(keyStyle.Render("ARN: ") + valueStyle.Render(displayARN) + "\n")

	if secret.Description != "" {
		displayDesc := secret.Description
		if len(displayDesc) > 60 {
			displayDesc = displayDesc[:57] + "..."
		}
		b.WriteString(keyStyle.Render("Description: ") + valueStyle.Render(displayDesc) + "\n")
	}

	if secret.LastChangedDate != nil {
		b.WriteString(keyStyle.Render("Last Modified: ") +
			valueStyle.Render(secret.LastChangedDate.Format("Jan 2, 2006 3:04 PM")) + "\n")
	}

	if len(secret.Tags) > 0 {
		b.WriteString("\n" + keyStyle.Render("Tags:") + "\n")
		for k, v := range secret.Tags {
			tagStr := fmt.Sprintf("  %s: %s", k, v)
			if len(tagStr) > 62 {
				tagStr = tagStr[:59] + "..."
			}
			b.WriteString(valueStyle.Render(tagStr) + "\n")
		}
	}

	// Secret value section
	b.WriteString("\n" + strings.Repeat("─", 70) + "\n\n")

	if m.secretValue == "" {
		instructionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		b.WriteString(instructionStyle.Render("Press 'v' to view the secret value") + "\n")
	} else {
		b.WriteString(keyStyle.Render("Secret Value:") + "\n\n")

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

		// Limit the displayed value to reasonable size
		lines := strings.Split(formatted, "\n")
		maxLines := 15
		if len(lines) > maxLines {
			formatted = strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
		}

		valueBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("241")).
			Padding(1).
			Width(66)

		b.WriteString(valueBoxStyle.Render(formatted) + "\n\n")

		// Copy instructions
		copyHelpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		b.WriteString(copyHelpStyle.Render("Press 'c' to copy as plain text | 'j' to copy as JSON"))
	}

	// Wrap in a bordered box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(76)

	boxContent := boxStyle.Render(b.String())

	// Center the box on screen
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width-6, m.height-10,
			lipgloss.Center, lipgloss.Center,
			boxContent)
	}

	return boxContent
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	help := `
AWS Secrets Manager TUI - Help

GRID NAVIGATION
  ↑/k         Move up
  ↓/j         Move down
  ←/h         Move left
  →/l         Move right
  enter       View secret details
  esc/q       Go back / Quit
  space       Next screen (within current page)
  pgup        Previous screen (within current page)

FILTERING
  /           Enter filter mode
  type        Filter secrets by name
  esc         Exit filter mode

ACTIONS
  v           View secret value (on detail screen)
  c           Copy secret value as plain text
  j           Copy secret value as JSON (on detail screen)
  r           Refresh secret list
  p           Switch AWS profile
  g           Switch AWS region
  n           Next AWS page (load 50 more secrets)
  b           Previous AWS page

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
