package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderInputView renders the input collection view
func (m Model) renderInputView() string {
	if m.selectedFunc == nil {
		return "No function selected"
	}

	var s strings.Builder

	// Title
	title := fmt.Sprintf("✏️  Function: %s", m.selectedFunc.DisplayName)
	s.WriteString(TitleStyle.Render(title))
	s.WriteString("\n\n")

	// Input fields
	for i, input := range m.selectedFunc.Inputs {
		// Label
		s.WriteString(InputLabelStyle.Render(input.DisplayName + ":"))
		s.WriteString("\n")

		// Value
		value := input.Value
		if value == "" {
			value = "  " // Placeholder space
		}

		// Add cursor if active
		if i == m.activeInputField {
			s.WriteString(ActiveInputValueStyle.Render(value + "│"))
		} else {
			s.WriteString(InputValueStyle.Render(value))
		}
		s.WriteString("\n\n")
	}

	// Buttons
	var buttons []string

	executeBtn := "[Execute]"
	cancelBtn := "[Cancel]"

	// Button navigation (after inputs)
	executeIndex := len(m.selectedFunc.Inputs)
	cancelIndex := len(m.selectedFunc.Inputs) + 1

	if m.activeInputField == executeIndex {
		buttons = append(buttons, ActiveButtonStyle.Render(executeBtn))
	} else {
		buttons = append(buttons, ButtonStyle.Render(executeBtn))
	}

	if m.activeInputField == cancelIndex {
		buttons = append(buttons, ActiveButtonStyle.Render(cancelBtn))
	} else {
		buttons = append(buttons, ButtonStyle.Render(cancelBtn))
	}

	s.WriteString(strings.Join(buttons, " "))
	s.WriteString("\n\n")

	// Help text
	helpText := "↑/↓: Navigate • Type: Edit • Enter: Confirm/Execute • Tab: Next Field • Esc: Cancel"
	s.WriteString(HelpStyle.Render(helpText))

	return ActiveBoxStyle.Render(s.String())
}

// handleInputKeys handles key presses in the input view
func (m Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedFunc == nil {
		return m, nil
	}

	maxField := len(m.selectedFunc.Inputs) + 1 // inputs + execute + cancel buttons

	switch msg.String() {
	case "up":
		if m.activeInputField > 0 {
			m.activeInputField--
		}

	case "down":
		if m.activeInputField < maxField {
			m.activeInputField++
		}

	case "tab":
		m.activeInputField = (m.activeInputField + 1) % (maxField + 1)

	case "esc":
		m.goBack()

	case "enter":
		return m.handleInputEnter()

	case "backspace":
		// Only edit if we're on an input field (not on buttons)
		if m.activeInputField < len(m.selectedFunc.Inputs) {
			if len(m.selectedFunc.Inputs[m.activeInputField].Value) > 0 {
				m.selectedFunc.Inputs[m.activeInputField].Value =
					m.selectedFunc.Inputs[m.activeInputField].Value[:len(m.selectedFunc.Inputs[m.activeInputField].Value)-1]
			}
		}

	default:
		// Type into active input field
		if m.activeInputField < len(m.selectedFunc.Inputs) {
			// Filter out special keys
			if len(msg.String()) == 1 {
				m.selectedFunc.Inputs[m.activeInputField].Value += msg.String()
			} else if msg.String() == "space" {
				m.selectedFunc.Inputs[m.activeInputField].Value += " "
			}
		}
	}

	return m, nil
}

// handleInputEnter handles Enter key in input view
func (m Model) handleInputEnter() (tea.Model, tea.Cmd) {
	if m.selectedFunc == nil {
		return m, nil
	}

	executeIndex := len(m.selectedFunc.Inputs)
	cancelIndex := len(m.selectedFunc.Inputs) + 1

	switch m.activeInputField {
	case executeIndex:
		// Execute button pressed
		return m.handleExecuteButton()

	case cancelIndex:
		// Cancel button pressed
		m.goBack()

	default:
		// On an input field - move to next field
		if m.activeInputField < len(m.selectedFunc.Inputs) {
			m.activeInputField++
		}
	}

	return m, nil
}

// handleExecuteButton handles Execute button press
func (m Model) handleExecuteButton() (tea.Model, tea.Cmd) {
	// Apply default values for common cases
	for i := range m.selectedFunc.Inputs {
		input := &m.selectedFunc.Inputs[i]
		if input.Value == "" {
			// Set sensible defaults
			switch input.Name {
			case "rootDir":
				cwd, err := os.Getwd()
				if err == nil {
					input.Value = cwd
				}
			case "path":
				input.Value = "."
			case "key":
				input.Value = "HOME"
			}
		}
	}

	return m, m.executeFunction()
}
