package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/charmbracelet/lipgloss"
)

const (
	appBorderWidth       = 2
	appHorizontalPadding = 2
)

// accent returns the accent colour for the current mode.
func (m Model) accent() lipgloss.Color {
	return accentColor(m.mode)
}

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
	case ScreenSecretFieldSelector:
		content = m.viewSecretFieldSelector()
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

	contentWidth, availableHeight := m.contentViewportSize()

	// Ensure content fills the available height to push footer to bottom
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
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
		BorderForeground(m.accent()).
		Width(contentWidth).
		Height(maxInt(availableHeight+lipgloss.Height(header)+lipgloss.Height(footer), 0)).
		Padding(0, 1)

	return appStyle.Render(fullContent)
}

func (m Model) contentViewportSize() (int, int) {
	innerWidth := maxInt(m.width-appBorderWidth-appHorizontalPadding, 0)
	innerHeight := maxInt(m.height-appBorderWidth, 0)

	headerHeight := lipgloss.Height(m.viewHeader())
	footerHeight := lipgloss.Height(m.viewFooter())
	contentHeight := maxInt(innerHeight-headerHeight-footerHeight, 0)

	return innerWidth, contentHeight
}

// viewHeader renders the header
func (m Model) viewHeader() string {
	title := "Secret Src - AWS Secrets Manager"
	if m.mode == models.KindParameter {
		title = "Secret Src - AWS SSM Parameter Store"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.accent()).
		MarginBottom(1)

	info := fmt.Sprintf("Profile: %s | Region: %s", m.currentProfile, m.currentRegion)

	return fmt.Sprintf("%s\n%s",
		headerStyle.Render(title),
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
		help = "hjkl/arrows: navigate | enter: view | /: filter | tab: switch source | p: profile | g: region | r: refresh | ?: help | q: quit"
		if m.currentPage > 0 {
			help += " | b: prev page"
		}
		if m.hasMore {
			help += " | n: next page"
		}
	case ScreenSecretDetail:
		if m.entryValue == "" {
			help = "v: view value | esc: back | q: quit"
		} else if m.mode == models.KindParameter {
			// Parameters are plain strings: copy plain only.
			help = "c: copy plain | esc: back | q: quit"
		} else {
			help = "c: copy plain | j: copy json | esc: back | q: quit"
			if len(m.secretFields) > 0 {
				help = "c: copy plain | j: copy json | k: copy field | esc: back | q: quit"
			}
		}
	case ScreenSecretFieldSelector:
		help = "enter: copy field | esc: back | q: quit"
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// viewSecretList renders the secret list screen
func (m Model) viewSecretList() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if len(m.entries) == 0 && !m.loading {
		return fmt.Sprintf("\n  No %s found in this region.\n\n  Try switching regions with 'g', refreshing with 'r', or 'tab' to switch source.", m.sourceNoun(true))
	}

	// Show filter status if filtering
	if m.grid.IsFiltering() {
		filterStatus := fmt.Sprintf("Filter: %s_", m.grid.GetFilterQuery())
		return fmt.Sprintf("%s\n%s", FilterStatusStyle.Render(filterStatus), m.grid.View())
	}

	return m.grid.View()
}

// viewSecretDetail renders the entry detail screen
func (m Model) viewSecretDetail() string {
	entry := m.grid.SelectedEntry()
	if entry == nil {
		return fmt.Sprintf("No %s selected", m.sourceNoun(false))
	}

	isParameter := m.mode == models.KindParameter

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.accent()).
		MarginBottom(1)

	title := "Secret Details"
	if isParameter {
		title = "Parameter Details"
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Metadata with compact key-value styling
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Truncate name if too long
	displayName := entry.Name
	if len(displayName) > 60 {
		displayName = displayName[:57] + "..."
	}
	b.WriteString(keyStyle.Render("Name: ") + valueStyle.Render(displayName) + "\n")

	// Truncate ARN if too long
	displayARN := entry.ARN
	if len(displayARN) > 60 {
		displayARN = "..." + displayARN[len(displayARN)-57:]
	}
	b.WriteString(keyStyle.Render("ARN: ") + valueStyle.Render(displayARN) + "\n")

	// SSM-only metadata
	if isParameter {
		if entry.Type != "" {
			b.WriteString(keyStyle.Render("Type: ") + valueStyle.Render(entry.Type) + "\n")
		}
		if entry.Version != "" {
			b.WriteString(keyStyle.Render("Version: ") + valueStyle.Render(entry.Version) + "\n")
		}
	}

	if entry.Description != "" {
		displayDesc := entry.Description
		if len(displayDesc) > 60 {
			displayDesc = displayDesc[:57] + "..."
		}
		b.WriteString(keyStyle.Render("Description: ") + valueStyle.Render(displayDesc) + "\n")
	}

	if entry.LastModifiedDate != nil {
		b.WriteString(keyStyle.Render("Last Modified: ") +
			valueStyle.Render(entry.LastModifiedDate.Format("Jan 2, 2006 3:04 PM")) + "\n")
	}

	// Tags are only shown for secrets (Secrets Manager returns them inline;
	// they are intentionally not fetched for parameters).
	if len(entry.Tags) > 0 {
		b.WriteString("\n" + keyStyle.Render("Tags:") + "\n")
		for k, v := range entry.Tags {
			tagStr := fmt.Sprintf("  %s: %s", k, v)
			if len(tagStr) > 62 {
				tagStr = tagStr[:59] + "..."
			}
			b.WriteString(valueStyle.Render(tagStr) + "\n")
		}
	}

	// Value section
	b.WriteString("\n" + strings.Repeat("─", 70) + "\n\n")

	if m.entryValue == "" {
		instructionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		b.WriteString(instructionStyle.Render(fmt.Sprintf("Press 'v' to view the %s value", m.sourceNoun(false))) + "\n")
	} else {
		valueLabel := "Secret Value:"
		if isParameter {
			valueLabel = "Parameter Value:"
		}
		b.WriteString(keyStyle.Render(valueLabel) + "\n\n")

		// Secrets are pretty-printed if they are JSON; parameters are plain strings.
		formatted := m.entryValue
		if !isParameter {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(m.entryValue), &jsonData); err == nil {
				if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					formatted = string(prettyJSON)
				}
			}
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
		copyHelp := "Press 'c' to copy as plain text"
		if !isParameter {
			copyHelp += " | 'j' to copy as JSON"
			if len(m.secretFields) > 0 {
				copyHelp += fmt.Sprintf(" | 'k' to copy a field (%d keys)", len(m.secretFields))
			}
		}
		b.WriteString(copyHelpStyle.Render(copyHelp))
	}

	// Wrap in a bordered box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.accent()).
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
  k           Copy one top-level JSON field (on eligible detail screens)
  r           Refresh secret list
  p           Switch AWS profile
  g           Switch AWS region
  n           Next AWS page (load 50 more secrets)
  b           Previous AWS page

SOURCE
  tab         Switch between Secrets Manager (pink) and
              SSM Parameter Store (green)

GLOBAL
  ?           Toggle this help
  ctrl+c      Force quit

SECURITY NOTE
  • Values are only fetched on-demand (when you press 'v')
  • SecureString parameters are decrypted on fetch (needs kms:Decrypt)
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

// viewSecretFieldSelector renders the secret field selector screen.
func (m Model) viewSecretFieldSelector() string {
	return m.fieldSelector.View()
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
