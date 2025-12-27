package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/bmj2728/hst/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

// renderMainMenu renders the main menu view
func (m Model) renderMainMenu() string {
	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render("🚀 Plugin System Demo - Main Menu"))
	s.WriteString("\n\n")

	// Plugin options
	menuItems := []string{}
	for i, plugin := range m.plugins {
		item := fmt.Sprintf("%d. %s", i+1, plugin.Name)
		if m.mainMenuCursor == i {
			menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
		} else {
			menuItems = append(menuItems, MenuItemStyle.Render(item))
		}
	}

	// Additional options
	concurrencyIndex := len(m.plugins)
	resetIndex := len(m.plugins) + 1
	shutdownIndex := len(m.plugins) + 2

	// Concurrency demo
	item := fmt.Sprintf("%d. Run Concurrency Demo", concurrencyIndex+1)
	if m.mainMenuCursor == concurrencyIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	// Reset state
	item = fmt.Sprintf("%d. Reset State", resetIndex+1)
	if m.mainMenuCursor == resetIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	// Shutdown
	item = fmt.Sprintf("%d. Shutdown", shutdownIndex+1)
	if m.mainMenuCursor == shutdownIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	s.WriteString(strings.Join(menuItems, "\n"))
	s.WriteString("\n\n")

	// Help text
	s.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • Ctrl+C: Quit"))

	return BoxStyle.Render(s.String())
}

// handleMainMenuKeys handles key presses in the main menu
func (m Model) handleMainMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := len(m.plugins) + 2 // plugins + concurrency + reset + shutdown

	switch msg.String() {
	case "up", "k":
		if m.mainMenuCursor > 0 {
			m.mainMenuCursor--
		}

	case "down", "j":
		if m.mainMenuCursor < maxCursor {
			m.mainMenuCursor++
		}

	case "enter":
		return m.handleMainMenuSelection()

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// handleMainMenuSelection handles Enter key press in main menu
func (m Model) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	concurrencyIndex := len(m.plugins)
	resetIndex := len(m.plugins) + 1
	shutdownIndex := len(m.plugins) + 2

	switch m.mainMenuCursor {
	case shutdownIndex:
		return m, m.shutdown()

	case resetIndex:
		m.resetState()
		return m, nil

	case concurrencyIndex:
		return m, m.runConcurrencyDemo()

	default:
		// Plugin selection
		if m.mainMenuCursor < len(m.plugins) {
			m.goToPluginMenu(m.mainMenuCursor)
		}
	}

	return m, nil
}

// shutdown performs cleanup and exits
func (m Model) shutdown() tea.Cmd {
	return func() tea.Msg {
		// Run the utils.Shutdown function
		go utils.Shutdown(0, m.hostServices, m.logger)
		return tea.Quit()
	}
}

// runConcurrencyDemo executes all plugins concurrently
func (m Model) runConcurrencyDemo() tea.Cmd {
	return func() tea.Msg {
		results := []string{}

		// Get current working directory for file listers
		cwd, err := os.Getwd()
		if err != nil {
			return ExecutionResult{
				PluginName: "Concurrency Demo",
				Function:   "All",
				Error:      err,
			}
		}

		// Execute each plugin's first function concurrently
		resultChan := make(chan ExecutionResult, len(m.plugins))

		for _, plugin := range m.plugins {
			go func(p PluginInfo) {
				if len(p.Functions) == 0 {
					return
				}

				// Set default values for file listers
				function := p.Functions[0]
				if p.Type == PluginFileLister || p.Type == PluginColorLister || p.Type == PluginPyLister {
					function.Inputs[0].Value = cwd
					function.Inputs[1].Value = "."
				} else if p.Type == PluginHostDemo {
					function.Inputs[0].Value = "GOPATH"
				}

				result := ExecutePluginFunction(&p, function)
				resultChan <- result
			}(plugin)
		}

		// Collect results
		for i := 0; i < len(m.plugins); i++ {
			result := <-resultChan
			if result.Error != nil {
				results = append(results, ErrorStyle.Render(fmt.Sprintf("[%s] Error: %v", result.PluginName, result.Error)))
			} else {
				results = append(results, fmt.Sprintf("[%s]\n%s", SuccessStyle.Render(result.PluginName), result.Output))
			}
		}

		return ExecutionResult{
			PluginName: "Concurrency Demo",
			Function:   "All",
			Output:     strings.Join(results, "\n\n"+strings.Repeat("─", 50)+"\n\n"),
		}
	}
}
