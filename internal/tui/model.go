package tui

import (
	"fmt"

	"github.com/bmj2728/hst/shared/pkg/hostserve"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/go-hclog"
)

// Model is the root Bubble Tea model for the TUI
type Model struct {
	// State management
	currentView    ViewMode
	selectedPlugin *PluginInfo
	selectedFunc   *PluginFunction

	// Plugin management
	plugins      []PluginInfo
	hostServices *hostserve.HostServices
	logger       hclog.Logger

	// Menu state
	mainMenuCursor   int
	pluginMenuCursor int

	// Input state
	inputCursor      int
	activeInputField int

	// Output state
	outputContent string
	outputScroll  int
	outputHeight  int

	// UI state
	width  int
	height int
	error  error
}

// NewModel creates a new TUI model
func NewModel(plugins []PluginInfo, hostServices *hostserve.HostServices, logger hclog.Logger) Model {
	return Model{
		currentView:  ViewMainMenu,
		plugins:      plugins,
		hostServices: hostServices,
		logger:       logger,
		outputHeight: 15,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ExecutionResult:
		return m.handleExecutionResult(msg)
	}

	return m, nil
}

// View renders the current view
func (m Model) View() string {
	switch m.currentView {
	case ViewMainMenu:
		return m.renderMainMenu()
	case ViewPluginMenu:
		return m.renderPluginMenu()
	case ViewInputCollection:
		return m.renderInputView()
	case ViewOutput:
		return m.renderOutputView()
	default:
		return "Unknown view"
	}
}

// handleKeyPress routes key presses based on current view
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// View-specific handling
	switch m.currentView {
	case ViewMainMenu:
		return m.handleMainMenuKeys(msg)
	case ViewPluginMenu:
		return m.handlePluginMenuKeys(msg)
	case ViewInputCollection:
		return m.handleInputKeys(msg)
	case ViewOutput:
		return m.handleOutputKeys(msg)
	}

	return m, nil
}

// handleExecutionResult processes plugin execution results
func (m Model) handleExecutionResult(result ExecutionResult) (tea.Model, tea.Cmd) {
	if result.Error != nil {
		m.outputContent = ErrorStyle.Render(fmt.Sprintf("Error: %v", result.Error))
	} else {
		m.outputContent = result.Output
	}
	m.currentView = ViewOutput
	m.outputScroll = 0
	return m, nil
}

// executeFunction runs the selected plugin function
func (m *Model) executeFunction() tea.Cmd {
	if m.selectedPlugin == nil || m.selectedFunc == nil {
		return nil
	}

	return func() tea.Msg {
		return ExecutePluginFunction(m.selectedPlugin, *m.selectedFunc)
	}
}

// resetState resets demo state
func (m *Model) resetState() {
	// This would call the reset function from utils
	// For now, just clear any errors
	m.error = nil
	m.outputContent = SuccessStyle.Render("State reset successfully")
	m.currentView = ViewOutput
}

// Helper methods for navigation
func (m *Model) goToMainMenu() {
	m.currentView = ViewMainMenu
	m.selectedPlugin = nil
	m.selectedFunc = nil
	m.error = nil
}

func (m *Model) goToPluginMenu(pluginIndex int) {
	if pluginIndex < 0 || pluginIndex >= len(m.plugins) {
		return
	}
	m.selectedPlugin = &m.plugins[pluginIndex]
	m.currentView = ViewPluginMenu
	m.pluginMenuCursor = 0
}

func (m *Model) goToInputView(funcIndex int) {
	if m.selectedPlugin == nil || funcIndex < 0 || funcIndex >= len(m.selectedPlugin.Functions) {
		return
	}
	m.selectedFunc = &m.selectedPlugin.Functions[funcIndex]
	m.currentView = ViewInputCollection
	m.activeInputField = 0
}

func (m *Model) goBack() {
	switch m.currentView {
	case ViewPluginMenu:
		m.goToMainMenu()
	case ViewInputCollection:
		m.currentView = ViewPluginMenu
		m.selectedFunc = nil
	case ViewOutput:
		if m.selectedPlugin != nil {
			m.currentView = ViewPluginMenu
		} else {
			m.goToMainMenu()
		}
	}
}
