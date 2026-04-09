package ui

import (
	"testing"

	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/benjamingriff/secretsrc/pkg/ui/components"
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

func TestParseSecretFieldsTopLevelObjectOnly(t *testing.T) {
	fields := parseSecretFields(`{"password":"secret","port":5432,"meta":{"team":"platform"},"flags":["a"],"enabled":true,"empty":"","nothing":null}`)
	if len(fields) != 7 {
		t.Fatalf("expected 7 fields, got %d", len(fields))
	}

	got := make(map[string]struct {
		copyValue string
		preview   string
	}, len(fields))
	for _, field := range fields {
		got[field.Key] = struct {
			copyValue string
			preview   string
		}{
			copyValue: field.CopyValue,
			preview:   field.Preview,
		}
	}

	if got["password"].copyValue != "secret" {
		t.Fatalf("expected password field to copy plain string, got %q", got["password"].copyValue)
	}
	if got["port"].copyValue != "5432" {
		t.Fatalf("expected port field to copy compact JSON number, got %q", got["port"].copyValue)
	}
	if got["meta"].copyValue != `{"team":"platform"}` {
		t.Fatalf("expected object field to copy compact JSON, got %q", got["meta"].copyValue)
	}
	if got["flags"].copyValue != `["a"]` {
		t.Fatalf("expected array field to copy compact JSON, got %q", got["flags"].copyValue)
	}
	if got["enabled"].copyValue != "true" {
		t.Fatalf("expected boolean field to copy compact JSON, got %q", got["enabled"].copyValue)
	}
	if got["empty"].copyValue != "" || got["empty"].preview != "(empty)" {
		t.Fatalf("expected empty string field to remain empty and preview as (empty), got value=%q preview=%q", got["empty"].copyValue, got["empty"].preview)
	}
	if got["nothing"].copyValue != "null" {
		t.Fatalf("expected null field to copy as literal null, got %q", got["nothing"].copyValue)
	}
}

func TestParseSecretFieldsRejectsNonObjects(t *testing.T) {
	if fields := parseSecretFields(`["a","b"]`); fields != nil {
		t.Fatalf("expected array secret to have no field list, got %d fields", len(fields))
	}
	if fields := parseSecretFields(`not-json`); fields != nil {
		t.Fatalf("expected invalid JSON secret to have no field list, got %d fields", len(fields))
	}
}

func TestHandleSecretDetailKeysOpensFieldSelector(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.width = 100
	model.height = 30
	model.currentScreen = ScreenSecretDetail
	model.secretValue = `{"password":"secret"}`
	model.secretFields = parseSecretFields(model.secretValue)

	updatedModel, cmd := model.handleSecretDetailKeys(keyRunes("k"))
	if cmd != nil {
		t.Fatal("expected opening the field selector to avoid immediate side effects")
	}

	updated := updatedModel.(Model)
	if updated.currentScreen != ScreenSecretFieldSelector {
		t.Fatalf("expected current screen %v, got %v", ScreenSecretFieldSelector, updated.currentScreen)
	}

	selected := updated.fieldSelector.SelectedField()
	if selected == nil || selected.Key != "password" {
		t.Fatalf("expected password to be selected in field picker, got %+v", selected)
	}
}

func TestHandleSecretFieldSelectorEscReturnsToDetail(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.currentScreen = ScreenSecretFieldSelector
	model.fieldSelector = fieldSelectorForTests([]string{"password"})

	updatedModel, cmd := model.handleSecretFieldSelectorKeys(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("expected esc to return without side effects")
	}

	updated := updatedModel.(Model)
	if updated.currentScreen != ScreenSecretDetail {
		t.Fatalf("expected current screen %v, got %v", ScreenSecretDetail, updated.currentScreen)
	}
}

func TestHandleSecretFieldSelectorEnterReturnsCopyCommand(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.currentScreen = ScreenSecretFieldSelector
	model.fieldSelector = fieldSelectorForTests([]string{"password"})

	updatedModel, cmd := model.handleSecretFieldSelectorKeys(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a clipboard copy command")
	}

	updated := updatedModel.(Model)
	if updated.currentScreen != ScreenSecretDetail {
		t.Fatalf("expected current screen %v, got %v", ScreenSecretDetail, updated.currentScreen)
	}
}

func TestHandleSecretListKeysEnterClearsSecretValueState(t *testing.T) {
	model := NewModel("default", "eu-west-2")
	model.grid.SetSecrets([]models.Secret{
		{Name: "alpha"},
		{Name: "beta"},
	})
	model.secretValue = `{"password":"secret"}`
	model.secretFields = parseSecretFields(model.secretValue)
	model.currentScreen = ScreenSecretList

	updatedModel, cmd := model.handleSecretListKeys(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected enter on the list screen to only change state")
	}

	updated := updatedModel.(Model)
	if updated.currentScreen != ScreenSecretDetail {
		t.Fatalf("expected current screen %v, got %v", ScreenSecretDetail, updated.currentScreen)
	}
	if updated.secretValue != "" {
		t.Fatalf("expected secret value to be cleared, got %q", updated.secretValue)
	}
	if updated.secretFields != nil {
		t.Fatalf("expected parsed secret fields to be cleared, got %d fields", len(updated.secretFields))
	}
}

func keyRunes(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func fieldSelectorForTests(keys []string) components.SecretFieldSelector {
	fields := make([]components.SecretField, len(keys))
	for i, key := range keys {
		fields[i] = components.SecretField{
			Key:       key,
			CopyValue: "value-for-" + key,
			Preview:   "preview-for-" + key,
		}
	}

	return components.NewSecretFieldSelector(fields, 80, 20)
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
