package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	primaryColor   = lipgloss.Color("#00ADD8") // Go blue
	secondaryColor = lipgloss.Color("#7B42BC") // HashiCorp purple
	accentColor    = lipgloss.Color("#5AF78E") // Success green
	warningColor   = lipgloss.Color("#F1FA8C") // Warning yellow
	errorColor     = lipgloss.Color("#FF5555") // Error red
	textColor      = lipgloss.Color("#F8F8F2") // Light text
	dimTextColor   = lipgloss.Color("#6272A4") // Dim text
	borderColor    = lipgloss.Color("#44475A") // Border

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	// Menu item styles
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(accentColor)

	// Info styles
	InfoStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			Italic(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Input styles
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	InputValueStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(borderColor).
			Padding(0, 1)

	ActiveInputValueStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(accentColor).
				Padding(0, 1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(borderColor).
			Padding(0, 2).
			Margin(0, 1)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(accentColor).
				Bold(true).
				Padding(0, 2).
				Margin(0, 1)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			Italic(true).
			Padding(1, 0)

	// Output styles
	OutputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Height(20)

	ScrollHintStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			Italic(true).
			Align(lipgloss.Center)
)
