package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderPluginMenu renders the plugin submenu view
func (m Model) renderPluginMenu() string {
	if m.selectedPlugin == nil {
		return "No plugin selected"
	}

	var s strings.Builder

	// Title
	title := fmt.Sprintf("📦 Plugin: %s", m.selectedPlugin.Name)
	s.WriteString(TitleStyle.Render(title))
	s.WriteString("\n\n")

	// Function list
	menuItems := []string{}
	for i, function := range m.selectedPlugin.Functions {
		item := fmt.Sprintf("%d. %s", i+1, function.DisplayName)
		if m.pluginMenuCursor == i {
			menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
		} else {
			menuItems = append(menuItems, MenuItemStyle.Render(item))
		}
	}

	// Navigation options
	backIndex := len(m.selectedPlugin.Functions)
	resetIndex := len(m.selectedPlugin.Functions) + 1
	shutdownIndex := len(m.selectedPlugin.Functions) + 2

	item := fmt.Sprintf("%d. ← Back to Main Menu", backIndex+1)
	if m.pluginMenuCursor == backIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	item = fmt.Sprintf("%d. Reset State", resetIndex+1)
	if m.pluginMenuCursor == resetIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	item = fmt.Sprintf("%d. Shutdown", shutdownIndex+1)
	if m.pluginMenuCursor == shutdownIndex {
		menuItems = append(menuItems, SelectedMenuItemStyle.Render(item))
	} else {
		menuItems = append(menuItems, MenuItemStyle.Render(item))
	}

	s.WriteString(strings.Join(menuItems, "\n"))
	s.WriteString("\n\n")

	// Help text
	s.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • B: Back • Q: Quit"))

	return BoxStyle.Render(s.String())
}

// handlePluginMenuKeys handles key presses in the plugin menu
func (m Model) handlePluginMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedPlugin == nil {
		return m, nil
	}

	maxCursor := len(m.selectedPlugin.Functions) + 2 // functions + back + reset + shutdown

	switch msg.String() {
	case "up", "k":
		if m.pluginMenuCursor > 0 {
			m.pluginMenuCursor--
		}

	case "down", "j":
		if m.pluginMenuCursor < maxCursor {
			m.pluginMenuCursor++
		}

	case "enter":
		return m.handlePluginMenuSelection()

	case "b":
		m.goBack()

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// handlePluginMenuSelection handles Enter key press in plugin menu
func (m Model) handlePluginMenuSelection() (tea.Model, tea.Cmd) {
	if m.selectedPlugin == nil {
		return m, nil
	}

	backIndex := len(m.selectedPlugin.Functions)
	resetIndex := len(m.selectedPlugin.Functions) + 1
	shutdownIndex := len(m.selectedPlugin.Functions) + 2

	switch m.pluginMenuCursor {
	case shutdownIndex:
		return m, m.shutdown()

	case resetIndex:
		m.resetState()
		return m, nil

	case backIndex:
		m.goBack()

	default:
		// Function selection
		if m.pluginMenuCursor < len(m.selectedPlugin.Functions) {
			selectedFunc := &m.selectedPlugin.Functions[m.pluginMenuCursor]
			// If function has no inputs, execute immediately
			if len(selectedFunc.Inputs) == 0 {
				m.selectedFunc = selectedFunc
				return m, m.executeFunction()
			}
			// Otherwise, go to input view
			m.goToInputView(m.pluginMenuCursor)
		}
	}

	return m, nil
}
