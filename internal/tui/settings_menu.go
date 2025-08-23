package tui

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// isMacOS returns true if running on macOS
func isMacOS() bool {
	return runtime.GOOS == "darwin"
}

// getPasteShortcut returns the appropriate paste shortcut for the current OS
func getPasteShortcut() string {
	if isMacOS() {
		return "Cmd+V"
	}
	return "Ctrl+V"
}

// SettingsMenuModel provides configuration interface
type SettingsMenuModel struct {
	width   int
	height  int
	focused int

	// Working copy of config
	config *Config

	// Input fields - make them all editable
	inputs  []SettingInput
	editing int // -1 = not editing, index = which field is being edited
}

// SettingInput represents an editable setting field
type SettingInput struct {
	Label    string
	Value    string
	Key      string // Internal key for saving
	Required bool
}

// NewSettingsMenuModel creates a new settings menu
func NewSettingsMenuModel(config *Config) *SettingsMenuModel {
	// Work directly with the provided config (no copy)

	// Create input fields for all configurable settings
	inputs := []SettingInput{
		{
			Label:    "API Key",
			Value:    config.APIKey,
			Key:      "api_key",
			Required: true,
		},
		{
			Label:    "Default Port",
			Value:    strconv.Itoa(config.DefaultPort),
			Key:      "default_port",
			Required: true,
		},
		{
			Label:    "Default Max Pages",
			Value:    strconv.Itoa(config.DefaultMaxPages),
			Key:      "default_max_pages",
			Required: false,
		},
		{
			Label:    "Default Max Depth",
			Value:    strconv.Itoa(config.DefaultMaxDepth),
			Key:      "default_max_depth",
			Required: false,
		},
		{
			Label:    "Default Concurrency",
			Value:    strconv.Itoa(config.DefaultConcurrency),
			Key:      "default_concurrency",
			Required: true,
		},
		{
			Label:    "Ignore Patterns",
			Value:    strings.Join(config.DefaultIgnorePatterns, ", "),
			Key:      "default_ignore_patterns",
			Required: false,
		},
	}

	return &SettingsMenuModel{
		focused: 0,
		config:  config,
		inputs:  inputs,
		editing: -1,
	}
}

// Init implements tea.Model
func (m *SettingsMenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *SettingsMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.editing >= 0 {
			return m.handleEditingInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "up", "k":
			if m.focused > 0 {
				m.focused--
			}
			return m, nil

		case "down", "j":
			maxFocus := len(m.inputs) // Only Back button now
			if m.focused < maxFocus {
				m.focused++
			}
			return m, nil

		case "enter", " ":
			return m.handleSelection()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *SettingsMenuModel) View() string {
	// Title with consistent left alignment
	title := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Margin(0, 0, 1, 0). // Consistent margin
		Render("⚙️ Settings")

	// Settings form
	settingsSection := m.renderSettingsForm()

	// Action buttons
	buttonsSection := m.renderButtons()

	// Help section
	var help string
	if m.editing >= 0 {
		help = m.renderCompactHelp(map[string]string{
			"Enter":            "Save & apply",
			"Esc":              "Cancel edit",
			"Ctrl+U":           "Clear field",
			getPasteShortcut(): "Paste",
		})
	} else {
		help = m.renderCompactHelp(map[string]string{
			"↑↓":    "Navigate",
			"Enter": "Edit",
			"Esc":   "Back to menu",
		})
	}

	// Use consistent container style
	compactStyle := lipgloss.NewStyle().
		Padding(1, 2). // Consistent with AppStyle
		Width(100).
		MaxWidth(160)

	// Create consistent layout with standardized spacing
	return compactStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"", // Single empty line after title
			settingsSection,
			"", // Single empty line after settings
			buttonsSection,
			"", // Single empty line after buttons
			help,
		),
	)
}

// renderSettingsForm shows all editable settings
func (m *SettingsMenuModel) renderSettingsForm() string {
	var lines []string

	for i, input := range m.inputs {
		line := m.renderSettingInput(i, input)
		lines = append(lines, line)
	}

	// Use a more compact subtitle style
	subtitle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true).
		Margin(0, 0, 0, 0). // No margin
		Render("Configuration:")

	return lipgloss.JoinVertical(lipgloss.Left,
		subtitle,
		"", // Small gap
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// renderSettingInput renders a single setting input field
func (m *SettingsMenuModel) renderSettingInput(index int, input SettingInput) string {
	label := fmt.Sprintf("%-20s:", input.Label)

	var valueDisplay string
	var style lipgloss.Style

	if index == m.focused && m.editing < 0 {
		// Focused but not editing
		style = SelectedItemStyle
		if input.Key == "api_key" && input.Value != "" {
			valueDisplay = fmt.Sprintf("%s (press Enter to edit)", MaskAPIKey(input.Value))
		} else {
			valueDisplay = fmt.Sprintf("%s (press Enter to edit)", input.Value)
		}
	} else if m.editing == index {
		// Currently editing this field
		style = SelectedItemStyle
		valueDisplay = input.Value + "█" // Show cursor
	} else {
		// Not focused
		style = ListItemStyle
		if input.Key == "api_key" && input.Value != "" {
			valueDisplay = MaskAPIKey(input.Value)
		} else {
			valueDisplay = input.Value
		}
	}

	// Show placeholder hints for some fields
	var hint string
	switch input.Key {
	case "default_max_pages", "default_max_depth":
		hint = " (0 = unlimited)"
	case "default_port":
		hint = " (your dev server port)"
	case "default_ignore_patterns":
		hint = " (comma-separated, e.g., /api, /admin, /blog/*)"
	}

	return fmt.Sprintf("  %s %s%s",
		lipgloss.NewStyle().Foreground(TextColor).Render(label),
		style.Render(valueDisplay),
		lipgloss.NewStyle().Foreground(MutedColor).Render(hint),
	)
}

// renderButtons shows action buttons
func (m *SettingsMenuModel) renderButtons() string {
	backIndex := len(m.inputs)

	// Use the consistent button rendering helper
	backBtn := RenderButton("[ Back to Menu ]", m.focused == backIndex)

	return backBtn
}

// renderCompactHelp renders help text with minimal spacing
func (m *SettingsMenuModel) renderCompactHelp(bindings map[string]string) string {
	// Create a compact container with left alignment for consistency
	container := lipgloss.NewStyle().
		Align(lipgloss.Left). // Changed from Center to Left for consistency
		Margin(0, 0)          // No margin

	// Define a consistent order for common keys
	keyOrder := []string{"↑↓", "Enter", "Esc", "Tab", "Ctrl+C", "q", "Ctrl+U", getPasteShortcut()}

	// Create ordered pairs
	var pairs []string
	usedKeys := make(map[string]bool)

	// First, add keys in the defined order
	for _, key := range keyOrder {
		if desc, exists := bindings[key]; exists {
			keyText := lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Render(key)
			descText := lipgloss.NewStyle().
				Foreground(MutedColor).
				Render(desc)

			pair := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(12).Render(keyText),
				descText,
			)
			pairs = append(pairs, pair)
			usedKeys[key] = true
		}
	}

	// Then add any remaining keys
	for key, desc := range bindings {
		if !usedKeys[key] {
			keyText := lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Render(key)
			descText := lipgloss.NewStyle().
				Foreground(MutedColor).
				Render(desc)

			pair := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(12).Render(keyText),
				descText,
			)
			pairs = append(pairs, pair)
		}
	}

	// Join all pairs with proper spacing
	return container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			pairs...,
		),
	)
}

// handleEditingInput handles input while editing a field
func (m *SettingsMenuModel) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Finish editing, update and save the config
		m.updateConfigFromInputs()
		if err := SaveConfig(m.config); err != nil {
			// If save fails, revert and show error
			m.revertCurrentInput()
			m.editing = -1
			return m, tea.Printf("Error saving config: %s", err.Error())
		}
		m.editing = -1
		return m, nil

	case "esc":
		// Cancel editing - revert value
		m.revertCurrentInput()
		m.editing = -1
		return m, nil

	case "backspace":
		if len(m.inputs[m.editing].Value) > 0 {
			m.inputs[m.editing].Value = m.inputs[m.editing].Value[:len(m.inputs[m.editing].Value)-1]
		}
		return m, nil

	case "ctrl+u":
		// Clear the input
		m.inputs[m.editing].Value = ""
		return m, nil

	case "ctrl+v", "cmd+v":
		// Handle paste operations - the actual content will come as individual characters
		// We'll handle the paste content in the default case
		return m, nil

	default:
		// Handle character input and paste operations
		if len(msg.String()) == 1 {
			// Single character - accept if it's printable
			if msg.String() >= " " && msg.String() <= "~" {
				m.inputs[m.editing].Value += msg.String()
			}
		} else if len(msg.String()) > 1 {
			// Multi-character input - likely from paste
			// Filter out unwanted characters like square brackets and escape sequences
			var filtered string
			for _, r := range msg.String() {
				// Accept only printable characters, exclude common terminal escape sequences
				if r >= ' ' && r <= '~' && r != '[' && r != ']' {
					filtered += string(r)
				}
			}
			if filtered != "" {
				m.inputs[m.editing].Value += filtered
			}
		}
		return m, nil
	}
}

// handleSelection processes enter key on focused elements
func (m *SettingsMenuModel) handleSelection() (tea.Model, tea.Cmd) {
	if m.focused < len(m.inputs) {
		// Start editing this field
		m.editing = m.focused
		return m, nil
	}

	backIndex := len(m.inputs)

	if m.focused == backIndex {
		// Back button - return updated config to main menu
		return m, func() tea.Msg {
			return BackToMenuMsg{Data: m.config}
		}
	}

	return m, nil
}

// updateConfigFromInputs updates the config from current input values
func (m *SettingsMenuModel) updateConfigFromInputs() {
	for _, input := range m.inputs {
		switch input.Key {
		case "api_key":
			m.config.APIKey = strings.TrimSpace(input.Value)
		case "default_port":
			if val, err := strconv.Atoi(input.Value); err == nil {
				m.config.DefaultPort = val
			}
		case "default_max_pages":
			if val, err := strconv.Atoi(input.Value); err == nil {
				m.config.DefaultMaxPages = val
			}
		case "default_max_depth":
			if val, err := strconv.Atoi(input.Value); err == nil {
				m.config.DefaultMaxDepth = val
			}
		case "default_concurrency":
			if val, err := strconv.Atoi(input.Value); err == nil {
				m.config.DefaultConcurrency = val
			}
		case "default_ignore_patterns":
			// Split by comma and trim whitespace
			patterns := strings.Split(input.Value, ",")
			var trimmedPatterns []string
			for _, pattern := range patterns {
				trimmed := strings.TrimSpace(pattern)
				if trimmed != "" {
					trimmedPatterns = append(trimmedPatterns, trimmed)
				}
			}
			m.config.DefaultIgnorePatterns = trimmedPatterns
		}
	}
}

// revertCurrentInput reverts the currently editing input to its original value
func (m *SettingsMenuModel) revertCurrentInput() {
	if m.editing < 0 || m.editing >= len(m.inputs) {
		return
	}

	input := &m.inputs[m.editing]
	switch input.Key {
	case "api_key":
		input.Value = m.config.APIKey
	case "default_port":
		input.Value = strconv.Itoa(m.config.DefaultPort)
	case "default_max_pages":
		input.Value = strconv.Itoa(m.config.DefaultMaxPages)
	case "default_max_depth":
		input.Value = strconv.Itoa(m.config.DefaultMaxDepth)
	case "default_concurrency":
		input.Value = strconv.Itoa(m.config.DefaultConcurrency)
	case "default_ignore_patterns":
		input.Value = strings.Join(m.config.DefaultIgnorePatterns, ", ")
	}
}

// validateAndSave validates all inputs and saves the configuration
func (m *SettingsMenuModel) validateAndSave() error {
	// Update config from inputs first
	m.updateConfigFromInputs()

	// Validate API key
	if m.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if len(m.config.APIKey) < 10 {
		return fmt.Errorf("API key seems too short")
	}

	// Validate port
	if m.config.DefaultPort <= 0 || m.config.DefaultPort > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	// Validate concurrency
	if m.config.DefaultConcurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}

	// Validate max pages and depth (0 is ok for unlimited)
	if m.config.DefaultMaxPages < 0 {
		return fmt.Errorf("max pages cannot be negative")
	}

	if m.config.DefaultMaxDepth < 0 {
		return fmt.Errorf("max depth cannot be negative")
	}

	// Actually save the config to file
	if err := SaveConfig(m.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
