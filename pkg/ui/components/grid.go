package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

const (
	MinCellWidth      = 35  // Minimum cell width
	MaxCellWidth      = 60  // Maximum cell width
	DefaultCellHeight = 4
	CellSpacing       = 2 // Space between cells
)

// SecretGrid displays secrets in a 2D grid layout
type SecretGrid struct {
	secrets         []models.Secret  // All secrets
	filteredSecrets []models.Secret  // Filtered secrets (used for display)
	cursorRow       int              // Current cursor row (0-based)
	cursorCol       int              // Current cursor column (0-based)
	numCols         int              // Number of columns in grid
	numRows         int              // Number of rows visible on screen
	cellWidth       int              // Calculated cell width based on available space
	gridPageIndex   int              // Current screen page index
	totalGridPages  int              // Total screen pages for filtered secrets
	width           int              // Available width
	height          int              // Available height
	filterQuery     string           // Current filter text
	filtering       bool             // Whether filter mode is active
}

// NewSecretGrid creates a new secret grid component
func NewSecretGrid(width, height int) SecretGrid {
	g := SecretGrid{
		secrets:         []models.Secret{},
		filteredSecrets: []models.Secret{},
		cursorRow:       0,
		cursorCol:       0,
		width:           width,
		height:          height,
		filterQuery:     "",
		filtering:       false,
	}
	g.calculateGridDimensions()
	return g
}

// SetSecrets updates the grid with new secrets
func (g *SecretGrid) SetSecrets(secrets []models.Secret) {
	g.secrets = secrets
	g.applyFilter(g.filterQuery)
	g.cursorRow = 0
	g.cursorCol = 0
	g.gridPageIndex = 0
}

// SetSize updates the grid dimensions
func (g *SecretGrid) SetSize(width, height int) {
	g.width = width
	g.height = height
	g.calculateGridDimensions()
	g.validateCursorPosition()
}

// calculateGridDimensions calculates numCols, numRows, cellWidth, and totalGridPages
func (g *SecretGrid) calculateGridDimensions() {
	// Calculate rows based on available height
	cellHeight := DefaultCellHeight + 1
	g.numRows = max(1, g.height/cellHeight)

	// Calculate optimal number of columns and cell width
	// Strategy: fit as many columns as possible while keeping cells between min and max width

	// Calculate max columns that fit at minimum width
	maxPossibleCols := g.width / (MinCellWidth + CellSpacing)
	if maxPossibleCols < 1 {
		maxPossibleCols = 1
	}

	// Find the optimal number of columns
	// Start with max possible and work down to find best fit
	optimalCols := 1
	optimalWidth := MinCellWidth

	for cols := maxPossibleCols; cols >= 1; cols-- {
		// Calculate width each cell would get with this many columns
		availableWidth := g.width - (cols-1)*CellSpacing
		widthPerCell := availableWidth / cols

		// If cells fit within our constraints, use this configuration
		if widthPerCell >= MinCellWidth && widthPerCell <= MaxCellWidth {
			optimalCols = cols
			optimalWidth = widthPerCell
			break
		} else if widthPerCell > MaxCellWidth {
			// Cells would be too wide, try more columns
			continue
		}
	}

	// If we still haven't found a fit, just use 1 column at min width
	if optimalCols == 0 || optimalWidth < MinCellWidth {
		optimalCols = 1
		optimalWidth = max(MinCellWidth, g.width)
	}

	g.numCols = optimalCols
	g.cellWidth = optimalWidth

	// Calculate total grid pages
	secretsPerPage := g.numCols * g.numRows
	if secretsPerPage > 0 && len(g.filteredSecrets) > 0 {
		g.totalGridPages = (len(g.filteredSecrets) + secretsPerPage - 1) / secretsPerPage
	} else {
		g.totalGridPages = 1
	}

	// Ensure gridPageIndex is valid
	if g.gridPageIndex >= g.totalGridPages {
		g.gridPageIndex = max(0, g.totalGridPages-1)
	}
}

// validateCursorPosition ensures cursor is within valid bounds
func (g *SecretGrid) validateCursorPosition() {
	// Get current flat index
	idx := g.cursorIndex()
	visibleSecrets := g.getVisibleSecrets()

	if idx >= len(visibleSecrets) {
		// Reset to first position
		g.cursorRow = 0
		g.cursorCol = 0
	}
}

// cursorIndex returns the flat index of the current cursor position
func (g *SecretGrid) cursorIndex() int {
	return g.cursorRow*g.numCols + g.cursorCol
}

// SelectedSecret returns the currently selected secret
func (g *SecretGrid) SelectedSecret() *models.Secret {
	visibleSecrets := g.getVisibleSecrets()
	idx := g.cursorIndex()

	if idx >= 0 && idx < len(visibleSecrets) {
		return &visibleSecrets[idx]
	}

	return nil
}

// getVisibleSecrets returns the secrets visible on the current grid page
func (g *SecretGrid) getVisibleSecrets() []models.Secret {
	secretsPerPage := g.numCols * g.numRows
	startIdx := g.gridPageIndex * secretsPerPage
	endIdx := min(startIdx+secretsPerPage, len(g.filteredSecrets))

	if startIdx >= len(g.filteredSecrets) {
		return []models.Secret{}
	}

	return g.filteredSecrets[startIdx:endIdx]
}

// moveUp moves cursor up or to previous grid page
func (g *SecretGrid) moveUp() {
	if g.cursorRow > 0 {
		g.cursorRow--
		g.validateCursorPosition()
	} else if g.gridPageIndex > 0 {
		// Move to previous grid page
		g.prevGridPage()
		// Place cursor at bottom of new page
		g.cursorRow = g.numRows - 1
		g.validateCursorPosition()
	}
}

// moveDown moves cursor down or to next grid page
func (g *SecretGrid) moveDown() {
	newRow := g.cursorRow + 1

	// Check if we can move down in current page
	if newRow < g.numRows {
		idx := newRow*g.numCols + g.cursorCol
		visibleSecrets := g.getVisibleSecrets()
		if idx < len(visibleSecrets) {
			g.cursorRow = newRow
			return
		}
	}

	// Try to move to next grid page
	if g.gridPageIndex < g.totalGridPages-1 {
		g.nextGridPage()
		g.cursorRow = 0
	}
}

// moveLeft moves cursor left
func (g *SecretGrid) moveLeft() {
	if g.cursorCol > 0 {
		g.cursorCol--
	}
}

// moveRight moves cursor right
func (g *SecretGrid) moveRight() {
	newCol := g.cursorCol + 1

	// Check if we can move right in current row
	if newCol < g.numCols {
		idx := g.cursorRow*g.numCols + newCol
		visibleSecrets := g.getVisibleSecrets()
		if idx < len(visibleSecrets) {
			g.cursorCol = newCol
		}
	}
}

// nextGridPage advances to the next grid page
func (g *SecretGrid) nextGridPage() {
	if g.gridPageIndex < g.totalGridPages-1 {
		g.gridPageIndex++
		g.cursorRow = 0
		g.cursorCol = 0
	}
}

// prevGridPage goes to the previous grid page
func (g *SecretGrid) prevGridPage() {
	if g.gridPageIndex > 0 {
		g.gridPageIndex--
		g.cursorRow = 0
		g.cursorCol = 0
	}
}

// applyFilter filters secrets by query
func (g *SecretGrid) applyFilter(query string) {
	g.filterQuery = query

	if query == "" {
		g.filteredSecrets = g.secrets
	} else {
		filtered := []models.Secret{}
		lowerQuery := strings.ToLower(query)

		for _, secret := range g.secrets {
			if strings.Contains(strings.ToLower(secret.Name), lowerQuery) {
				filtered = append(filtered, secret)
			}
		}

		g.filteredSecrets = filtered
	}

	// Reset navigation state after filter
	g.cursorRow = 0
	g.cursorCol = 0
	g.gridPageIndex = 0
	g.calculateGridDimensions()
}

// clearFilter clears the current filter
func (g *SecretGrid) clearFilter() {
	g.filterQuery = ""
	g.filtering = false
	g.applyFilter("")
}

// IsFiltering returns whether filter mode is active
func (g *SecretGrid) IsFiltering() bool {
	return g.filtering
}

// GetFilterQuery returns the current filter query
func (g *SecretGrid) GetFilterQuery() string {
	return g.filterQuery
}

// Update handles keyboard input
func (g *SecretGrid) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle filter mode
		if g.filtering {
			switch msg.String() {
			case "esc":
				g.clearFilter()
				return nil
			case "backspace":
				if len(g.filterQuery) > 0 {
					g.applyFilter(g.filterQuery[:len(g.filterQuery)-1])
				}
				return nil
			case "enter":
				g.filtering = false
				return nil
			default:
				// Only handle single character input
				if len(msg.String()) == 1 {
					g.applyFilter(g.filterQuery + msg.String())
				}
				return nil
			}
		}

		// Handle normal navigation
		switch msg.String() {
		case "up", "k":
			g.moveUp()
		case "down", "j":
			g.moveDown()
		case "left", "h":
			g.moveLeft()
		case "right", "l":
			g.moveRight()
		case " ", "pgdown":
			g.nextGridPage()
		case "pgup":
			g.prevGridPage()
		case "/":
			g.filtering = true
			return nil
		}
	}

	return nil
}

// View renders the grid
func (g *SecretGrid) View() string {
	visibleSecrets := g.getVisibleSecrets()

	if len(visibleSecrets) == 0 {
		if g.filtering && g.filterQuery != "" {
			return lipgloss.NewStyle().
				Padding(2).
				Foreground(lipgloss.Color("241")).
				Render(fmt.Sprintf("No secrets match '%s'", g.filterQuery))
		}
		return lipgloss.NewStyle().
			Padding(2).
			Foreground(lipgloss.Color("241")).
			Render("No secrets found")
	}

	// Build grid
	rows := []string{}

	for row := 0; row < g.numRows; row++ {
		cellsInRow := []string{}

		for col := 0; col < g.numCols; col++ {
			idx := row*g.numCols + col

			if idx >= len(visibleSecrets) {
				// Empty cell
				break
			}

			secret := visibleSecrets[idx]
			isSelected := (row == g.cursorRow && col == g.cursorCol)

			cellsInRow = append(cellsInRow, g.renderCell(secret, isSelected))
		}

		if len(cellsInRow) > 0 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cellsInRow...))
		}
	}

	gridView := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Add pagination indicator if needed
	if g.totalGridPages > 1 {
		paginationInfo := fmt.Sprintf("Screen %d/%d", g.gridPageIndex+1, g.totalGridPages)
		paginationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
		gridView = lipgloss.JoinVertical(lipgloss.Left, gridView, paginationStyle.Render(paginationInfo))
	}

	return gridView
}

// renderCell renders a single grid cell
func (g *SecretGrid) renderCell(secret models.Secret, isSelected bool) string {
	// Wrap the secret name to fit width (account for padding)
	nameLines := g.wrapText(secret.Name, g.cellWidth-2)

	// Take only first 2 lines for the name (save room for date)
	if len(nameLines) > 2 {
		nameLines = nameLines[:2]
		// Add ellipsis to last line
		lastLine := nameLines[1]
		if len(lastLine) > 3 {
			nameLines[1] = lastLine[:len(lastLine)-3] + "..."
		}
	} else if len(nameLines) == 0 {
		nameLines = []string{"(unnamed)"}
	}

	// Format the last modified date
	dateStr := "Unknown"
	if secret.LastChangedDate != nil {
		dateStr = secret.LastChangedDate.Format("Jan 2, 2006")
	}

	// Style the name based on selection
	var nameStyle lipgloss.Style
	if isSelected {
		// Selected: pink text, bold (no background)
		nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
	} else {
		// Normal: light grey text
		nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	}

	// Style the date (always greyed out)
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Render styled parts
	styledName := nameStyle.Render(strings.Join(nameLines, "\n"))
	styledDate := dateStyle.Render(dateStr)

	// Combine content
	content := styledName + "\n" + styledDate

	// Apply cell style (sizing and padding only, no background color)
	// Use the dynamically calculated cell width
	cellStyle := lipgloss.NewStyle().
		Width(g.cellWidth).
		Height(DefaultCellHeight).
		Padding(0, 1)

	return cellStyle.Render(content)
}

// wrapText wraps text to fit within maxWidth
func (g *SecretGrid) wrapText(text string, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if lipgloss.Width(testLine) > maxWidth {
			if currentLine != "" {
				// Save current line and start new one
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Single word too long, truncate it
				truncated := word
				for lipgloss.Width(truncated) > maxWidth-3 && len(truncated) > 0 {
					truncated = truncated[:len(truncated)-1]
				}
				lines = append(lines, truncated+"...")
				currentLine = ""
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
