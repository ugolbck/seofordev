package tui

import (
	"fmt"
	"net/url"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SinglePageAuditModel handles URL input for single page audits
type SinglePageAuditModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// URL input
	urlInput string
	cursor   int
	validURL bool
	errorMsg string

	// UI state
	starting bool
}

// NewSinglePageAuditModel creates a new single page audit model
func NewSinglePageAuditModel(config *Config) *SinglePageAuditModel {
	return &SinglePageAuditModel{
		config:   config,
		urlInput: "/",
		cursor:   1,
		validURL: true,
	}
}

// Init implements tea.Model
func (m *SinglePageAuditModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *SinglePageAuditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "enter":
			if m.validURL && !m.starting {
				return m.startSinglePageAudit()
			}
			return m, nil

		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "right":
			if m.cursor < len(m.urlInput) {
				m.cursor++
			}
			return m, nil

		case "home", "ctrl+a":
			m.cursor = 0
			return m, nil

		case "end", "ctrl+e":
			m.cursor = len(m.urlInput)
			return m, nil

		case "backspace":
			if m.cursor > 0 {
				m.urlInput = m.urlInput[:m.cursor-1] + m.urlInput[m.cursor:]
				m.cursor--
				m.validateURL()
			}
			return m, nil

		case "delete":
			if m.cursor < len(m.urlInput) {
				m.urlInput = m.urlInput[:m.cursor] + m.urlInput[m.cursor+1:]
				m.validateURL()
			}
			return m, nil

		case "ctrl+u":
			// Clear input
			m.urlInput = ""
			m.cursor = 0
			m.validateURL()
			return m, nil

		default:
			// Handle character input
			char := msg.String()
			if len(char) == 1 && char >= " " && char <= "~" {
				m.urlInput = m.urlInput[:m.cursor] + char + m.urlInput[m.cursor:]
				m.cursor++
				m.validateURL()
			}
			return m, nil
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *SinglePageAuditModel) View() string {
	if m.starting {
		return AppStyle.Render(
			TitleStyle.Render("🚀 Starting Single Page Audit...") + "\n\n" +
				ContentStyle.Render("Initializing audit for your page..."),
		)
	}

	// Title
	title := TitleStyle.Render("📄 Single Page Audit")

	// Description
	description := ContentStyle.Render(
		"Enter a localhost URL to audit a single page.\n" +
			"Examples: /, /blog, http://localhost:8000/about, 127.0.0.1:3000/contact",
	)

	// URL input field
	urlSection := m.renderURLInput()

	// Status section (validation feedback)
	statusSection := m.renderStatus()

	// Help section
	help := RenderKeyHelp(map[string]string{
		"Enter":  "Start audit",
		"Ctrl+U": "Clear input",
		"Esc":    "Back",
	})

	// Create layout
	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			description,
			"",
			urlSection,
			"",
			statusSection,
			"",
			help,
		),
	)
}

// renderURLInput renders the URL input field with cursor
func (m *SinglePageAuditModel) renderURLInput() string {
	// Create input field with cursor
	left := m.urlInput[:m.cursor]
	cursor := "│"
	right := ""
	if m.cursor < len(m.urlInput) {
		right = m.urlInput[m.cursor:]
	}

	inputText := left + cursor + right

	// Style the input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(0, 1).
		Width(80)

	if !m.validURL {
		inputStyle = inputStyle.BorderForeground(ErrorColor)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("URL:"),
		"",
		inputStyle.Render(inputText),
	)
}

// renderStatus shows validation status and preview
func (m *SinglePageAuditModel) renderStatus() string {
	if !m.validURL {
		return lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render("❌ " + m.errorMsg)
	}

	// Show what URL will be audited
	normalizedURL := m.normalizeURL(m.urlInput)
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("✅ Valid localhost URL"),
		"",
		ContentStyle.Render(fmt.Sprintf("Will audit: %s", normalizedURL)),
	)
}

// validateURL checks if the entered URL is valid for localhost auditing
func (m *SinglePageAuditModel) validateURL() {
	m.validURL = false
	m.errorMsg = ""

	if strings.TrimSpace(m.urlInput) == "" {
		m.errorMsg = "URL cannot be empty"
		return
	}

	normalizedURL := m.normalizeURL(m.urlInput)
	if normalizedURL == "" {
		m.errorMsg = "Invalid URL format"
		return
	}

	// Parse the normalized URL
	parsed, err := url.Parse(normalizedURL)
	if err != nil {
		m.errorMsg = "Invalid URL format"
		return
	}

	// Check if it's localhost or 127.0.0.1
	host := strings.ToLower(parsed.Host)
	if !m.isLocalhost(host) {
		m.errorMsg = "URL must be localhost or 127.0.0.1"
		return
	}

	m.validURL = true
}

// normalizeURL converts various input formats to a full URL
func (m *SinglePageAuditModel) normalizeURL(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// If it's just a path (starts with /), add localhost with default port
	if strings.HasPrefix(input, "/") {
		return fmt.Sprintf("http://localhost:%d%s", m.config.DefaultPort, input)
	}

	// If it doesn't have a scheme, try to parse and add http://
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		// Check if it looks like host:port/path or just host/path
		if strings.Contains(input, "/") || strings.Contains(input, ":") {
			input = "http://" + input
		} else {
			// Just a path without leading /, add it
			return fmt.Sprintf("http://localhost:%d/%s", m.config.DefaultPort, input)
		}
	}

	// Parse to validate and normalize
	parsed, err := url.Parse(input)
	if err != nil {
		return ""
	}

	// If no port specified and it's localhost, add default port
	if (parsed.Host == "localhost" || parsed.Host == "127.0.0.1") && !strings.Contains(parsed.Host, ":") {
		parsed.Host = fmt.Sprintf("%s:%d", parsed.Host, m.config.DefaultPort)
	}

	return parsed.String()
}

// isLocalhost checks if a host is localhost or 127.0.0.1 (with optional port)
func (m *SinglePageAuditModel) isLocalhost(host string) bool {
	// Remove port if present
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	return host == "localhost" || host == "127.0.0.1"
}

// startSinglePageAudit creates and starts the audit for the single page
func (m *SinglePageAuditModel) startSinglePageAudit() (tea.Model, tea.Cmd) {
	m.starting = true

	// Initialize the audit adapter with the current config
	err := InitializeAuditAdapter(m.config)
	if err != nil {
		LogError("Failed to initialize audit adapter: %v", err)
		// Continue with local-only mode
	}

	// Get the normalized URL
	normalizedURL := m.normalizeURL(m.urlInput)

	// Create AuditConfig for single page audit
	auditConfig := AuditConfig{
		Port:           m.config.DefaultPort,
		Concurrency:    1,          // Single page, no need for high concurrency
		MaxPages:       1,          // Only one page
		MaxDepth:       0,          // Don't follow links
		IgnorePatterns: []string{}, // No ignore patterns needed for single page
		APIKey:         m.config.APIKey,
	}

	// Create and switch to SimpleAuditModel with the specific URL
	auditModel := NewSimpleAuditModelWithURL(auditConfig, normalizedURL)
	return auditModel, auditModel.Init()
}
