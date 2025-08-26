package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewAuditModel shows audit settings and starts the audit
type NewAuditModel struct {
	width   int
	height  int
	focused int

	// Configuration
	config *Config

	// Starting the audit
	starting bool
}

// NewNewAuditModel creates a new audit configuration model
func NewNewAuditModel(config *Config) *NewAuditModel {
	return &NewAuditModel{
		focused: 0, // Focus on "Start Audit" button
		config:  config,
	}
}

// Init implements tea.Model
func (m *NewAuditModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *NewAuditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "enter", " ":
			return m.handleSelection()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *NewAuditModel) View() string {
	if m.starting {
		return AppStyle.Render(
			TitleStyle.Render("🚀 Starting Audit...") + "\n\n" +
				ContentStyle.Render("Initializing audit with your settings..."),
		)
	}

	// Title with consistent left alignment
	title := TitleStyle.Render("🆕 New Audit")

	// Target URL section
	targetURL := fmt.Sprintf("http://localhost:%d", m.config.DefaultPort)
	urlSection := lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("Target:"),
		"",
		ContentStyle.Render(fmt.Sprintf("URL: %s", targetURL)),
	)

	// Audit settings section
	settingsSection := m.renderAuditSettings()

	// Action buttons section
	buttonsSection := m.renderButtons()

	// Help section
	help := RenderKeyHelp(map[string]string{
		"Enter": "Start audit",
		"Esc":   "Back",
	})

	// Create consistent layout with standardized spacing
	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"", // Single empty line after title
			urlSection,
			"", // Single empty line after URL
			settingsSection,
			"", // Single empty line after settings
			buttonsSection,
			"", // Single empty line after buttons
			help,
		),
	)
}

// renderAuditSettings shows the audit configuration
func (m *NewAuditModel) renderAuditSettings() string {
	// Format ignore patterns for display
	var ignoreText string
	if len(m.config.DefaultIgnorePatterns) > 0 {
		ignoreText = strings.Join(m.config.DefaultIgnorePatterns, ", ")
	} else {
		ignoreText = "none"
	}

	settings := []string{
		fmt.Sprintf("Max Pages:   %s", m.formatLimit(m.config.DefaultMaxPages)),
		fmt.Sprintf("Max Depth:   %s", m.formatLimit(m.config.DefaultMaxDepth)),
		fmt.Sprintf("Concurrency: %d workers", m.config.DefaultConcurrency),
		fmt.Sprintf("Ignore:      %s", ignoreText),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("Audit Settings:"),
		"",
		ContentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, settings...)),
		"",
		lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("(Configure defaults in Settings)"),
	)
}

// renderButtons shows action buttons
func (m *NewAuditModel) renderButtons() string {
	// Only show the start button, pre-selected
	startBtn := RenderButton("[ 🚀 Start Audit ]", true)

	return startBtn
}

// handleSelection processes enter key on focused elements
func (m *NewAuditModel) handleSelection() (tea.Model, tea.Cmd) {
	// Always start the audit since there's only one button
	m.starting = true

	// Initialize the audit adapter with the current config
	err := InitializeAuditAdapter(m.config)
	if err != nil {
		LogError("Failed to initialize audit adapter: %v", err)
		// Continue with local-only mode
	}

	// Create AuditConfig from the stored config
	auditConfig := AuditConfig{
		Port:           m.config.DefaultPort,
		Concurrency:    m.config.DefaultConcurrency,
		MaxPages:       m.config.DefaultMaxPages,
		MaxDepth:       m.config.DefaultMaxDepth,
		IgnorePatterns: m.config.DefaultIgnorePatterns, // Use configured ignore patterns
		APIKey:         m.config.APIKey,
	}

	// Create and switch to the existing SimpleAuditModel
	auditModel := NewSimpleAuditModel(auditConfig)
	return auditModel, auditModel.Init()
}

// formatLimit formats limit values for display
func (m *NewAuditModel) formatLimit(limit int) string {
	if limit == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", limit)
}
