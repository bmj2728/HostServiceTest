package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderOutputView renders the output display view
func (m Model) renderOutputView() string {
	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render("📊 Output"))
	s.WriteString("\n\n")

	// Output content with scrolling
	lines := strings.Split(m.outputContent, "\n")
	visibleLines := m.outputHeight

	// Calculate scroll bounds
	maxScroll := len(lines) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// Clamp scroll position
	if m.outputScroll > maxScroll {
		m.outputScroll = maxScroll
	}
	if m.outputScroll < 0 {
		m.outputScroll = 0
	}

	// Show scroll hint at top if scrolled down
	if m.outputScroll > 0 {
		s.WriteString(ScrollHintStyle.Render("↑ Scroll up to see more ↑"))
		s.WriteString("\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n")
	}

	// Display visible lines
	start := m.outputScroll
	end := start + visibleLines
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i < end; i++ {
		s.WriteString(lines[i])
		s.WriteString("\n")
	}

	// Show scroll hint at bottom if more content below
	if m.outputScroll < maxScroll {
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n")
		s.WriteString(ScrollHintStyle.Render("↓ Scroll down to see more ↓"))
	}

	s.WriteString("\n\n")

	// Navigation help
	help := "[B]ack to Plugin • [M]ain Menu • [R]eset State • [Q]uit"
	s.WriteString(HelpStyle.Render(help))

	return BoxStyle.Render(s.String())
}

// handleOutputKeys handles key presses in the output view
func (m Model) handleOutputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.outputContent, "\n")
	maxScroll := len(lines) - m.outputHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case "up", "k":
		if m.outputScroll > 0 {
			m.outputScroll--
		}

	case "down", "j":
		if m.outputScroll < maxScroll {
			m.outputScroll++
		}

	case "pgup":
		m.outputScroll -= m.outputHeight
		if m.outputScroll < 0 {
			m.outputScroll = 0
		}

	case "pgdown":
		m.outputScroll += m.outputHeight
		if m.outputScroll > maxScroll {
			m.outputScroll = maxScroll
		}

	case "home":
		m.outputScroll = 0

	case "end":
		m.outputScroll = maxScroll

	case "b":
		m.goBack()

	case "m":
		m.goToMainMenu()

	case "r":
		m.resetState()

	case "q":
		return m, tea.Quit
	}

	return m, nil
}
