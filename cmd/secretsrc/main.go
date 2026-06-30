package main

import (
	"fmt"
	"os"

	secretsrcaws "github.com/benjamingriff/secretsrc/pkg/aws"
	"github.com/benjamingriff/secretsrc/pkg/config"
	"github.com/benjamingriff/secretsrc/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	profile := secretsrcaws.GetDefaultProfile()
	region := secretsrcaws.GetDefaultRegion()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	if profile == "" {
		profile = cfg.LastProfile
	}
	if region == "" {
		region = cfg.LastRegion
	}

	program := tea.NewProgram(ui.NewModel(profile, region), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run secretsrc: %v\n", err)
		os.Exit(1)
	}
}
