package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// APIKeyGatekeeperModel represents the API key setup screen
type APIKeyGatekeeperModel struct {
	textInput  textinput.Model
	err        error
	validating bool
	state      apiKeyState
	userInfo   *APIValidationResponse
	config     *Config
}

type apiKeyState int

const (
	inputState apiKeyState = iota
	validatingState
	successState
	errorState
)

type validationResultMsg struct {
	result *APIValidationResponse
	err    error
}

// NewAPIKeyGatekeeperModel creates a new API key gatekeeper model
func NewAPIKeyGatekeeperModel(config *Config) APIKeyGatekeeperModel {
	ti := textinput.New()
	ti.Placeholder = "Paste your API key here..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	// If there's an existing API key, pre-fill it (user can modify/replace it)
	if config.APIKey != "" {
		ti.SetValue(config.APIKey)
	}

	return APIKeyGatekeeperModel{
		textInput: ti,
		state:     inputState,
		config:    config,
	}
}

func (m APIKeyGatekeeperModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m APIKeyGatekeeperModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if (m.state == inputState || m.state == errorState) && m.textInput.Value() != "" {
				// Start validation
				m.state = validatingState
				m.validating = true
				return m, m.validateAPIKey(m.textInput.Value())
			} else if m.state == successState {
				// Proceed to main menu with updated config
				mainMenu := NewMainMenuModelWithConfig(m.config)
				return mainMenu, mainMenu.Init()
			}
		}

	case validationResultMsg:
		m.validating = false
		if msg.err != nil {
			m.state = errorState
			m.err = msg.err
			// Reset the input for retry
			m.textInput.SetValue("")
			m.textInput.Focus()
		} else {
			m.state = successState
			m.userInfo = msg.result
			// Save the API key to config
			m.config.APIKey = m.textInput.Value()
			// Save the config to file
			SaveConfig(m.config)
		}
	}

	if m.state == inputState || m.state == errorState {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m APIKeyGatekeeperModel) View() string {
	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(1, 2)

	s.WriteString(headerStyle.Render("🔑 API Key Setup"))
	s.WriteString("\n\n")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 2)

	if m.config.APIKey != "" {
		s.WriteString(instructionStyle.Render("Your existing API key appears to be invalid or expired."))
		s.WriteString("\n")
		s.WriteString(instructionStyle.Render("Please enter a valid API key below."))
		s.WriteString("\n\n")
	} else {
		s.WriteString(instructionStyle.Render("To get started, you need an API key from the seofor.dev dashboard."))
		s.WriteString("\n")
		s.WriteString(instructionStyle.Render("1. Visit: https://seofor.dev/dashboard"))
		s.WriteString("\n")
		s.WriteString(instructionStyle.Render("2. Generate an API key"))
		s.WriteString("\n")
		s.WriteString(instructionStyle.Render("3. Copy and paste it below"))
		s.WriteString("\n\n")
	}

	// Content based on state
	switch m.state {
	case inputState:
		s.WriteString("Enter your API key:\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to validate • Ctrl+C to quit"))

	case validatingState:
		s.WriteString("🔄 Validating API key...")
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Please wait..."))

	case successState:
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

		s.WriteString(successStyle.Render("✅ API key validated successfully!"))
		s.WriteString("\n\n")

		// Show user info
		if m.userInfo != nil {
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Padding(0, 2)

			s.WriteString(infoStyle.Render(fmt.Sprintf("Email: %s", m.userInfo.User.Email)))
			s.WriteString("\n")

			if m.userInfo.User.Username != nil && *m.userInfo.User.Username != "" {
				s.WriteString(infoStyle.Render(fmt.Sprintf("Welcome, %s! 👋", *m.userInfo.User.Username)))
				s.WriteString("\n")
			}

			s.WriteString(infoStyle.Render(fmt.Sprintf("Credits: %d", m.userInfo.User.Credits)))
			s.WriteString("\n")

			if m.userInfo.User.HasPaidPlan {
				s.WriteString(infoStyle.Render("Plan: ✅ Paid"))
			} else {
				s.WriteString(infoStyle.Render("Plan: ❌ Free"))
			}
		}

		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to continue • Ctrl+C to quit"))

	case errorState:
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

		s.WriteString(errorStyle.Render("❌ API Key Validation Failed"))
		s.WriteString("\n\n")

		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Please check your API key and try again."))
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Visit https://seofor.dev/dashboard to generate a new key."))
		s.WriteString("\n\n")

		s.WriteString("Enter your API key:\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to validate • Ctrl+C to quit"))
	}

	// Wrap everything in a border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60)

	return borderStyle.Render(s.String())
}

// validateAPIKey returns a command that validates the API key
func (m APIKeyGatekeeperModel) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		baseURL := m.config.GetEffectiveBaseURL()
		result, err := ValidateAPIKeyWithServer(apiKey, baseURL)
		return validationResultMsg{result: result, err: err}
	}
}
