package tui

import (
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/go-hclog"
)

// Launch starts the TUI application
func Launch(plugins []PluginInfo, hostServices *hostserve.HostServices, logger hclog.Logger) error {
	// Create the model
	model := NewModel(plugins, hostServices, logger)

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	return err
}
