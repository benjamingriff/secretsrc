package ui

import (
	"testing"

	"github.com/benjamingriff/secretsrc/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleSecretListKeysEscClearsFilterWithoutQuit(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.grid.SetSecrets([]models.Secret{
		{Name: "alpha"},
		{Name: "beta"},
	})

	model.grid.Update(keyRunes("/"))
	model.grid.Update(keyRunes("a"))

	if !model.grid.IsFiltering() {
		t.Fatal("expected filter mode to be active")
	}
	if got := model.grid.GetFilterQuery(); got != "a" {
		t.Fatalf("expected filter query %q, got %q", "a", got)
	}

	updatedModel, cmd := model.handleSecretListKeys(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("expected esc in filter mode to avoid quitting the app")
	}

	updated := updatedModel.(Model)
	if updated.grid.IsFiltering() {
		t.Fatal("expected esc to exit filter mode")
	}
	if got := updated.grid.GetFilterQuery(); got != "" {
		t.Fatalf("expected esc to clear filter query, got %q", got)
	}
}

func TestContentViewportSizeUsesRenderedHeaderAndFooterHeights(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.width = 100
	model.height = 30
	model.loading = true
	model.statusMessage = "Copied to clipboard!"

	width, height := model.contentViewportSize()

	if width != 96 {
		t.Fatalf("expected content width 96, got %d", width)
	}

	headerHeight := lipglossHeight(model.viewHeader())
	footerHeight := lipglossHeight(model.viewFooter())
	expectedHeight := 28 - headerHeight - footerHeight
	if expectedHeight < 0 {
		expectedHeight = 0
	}

	if height != expectedHeight {
		t.Fatalf("expected content height %d, got %d", expectedHeight, height)
	}
}

func keyRunes(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func lipglossHeight(value string) int {
	lines := 1
	for _, r := range value {
		if r == '\n' {
			lines++
		}
	}
	if value == "" {
		return 0
	}
	return lines
}
