package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuditMenuModel provides audit-related options
type AuditMenuModel struct {
	width       int
	height      int
	selectedIdx int

	// Configuration
	config *Config
}

type auditMenuItem struct {
	title       string
	description string
	icon        string
}

var auditMenuItems = []auditMenuItem{
	{"New Audit (Full)", "Start a new audit of your localhost site", "🆕"},
	{"New Audit (Single Page)", "Start a new audit of your localhost site", "🆕"},
	{"Audit History", "View previous audit results", "📊"},
}

// NewAuditMenuModel creates a new audit menu
func NewAuditMenuModel(config *Config) *AuditMenuModel {
	return &AuditMenuModel{
		selectedIdx: 0,
		config:      config,
	}
}

// Init implements tea.Model
func (m *AuditMenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *AuditMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil

		case "down", "j":
			if m.selectedIdx < len(auditMenuItems)-1 {
				m.selectedIdx++
			}
			return m, nil

		case "enter", " ":
			return m.handleSelection()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *AuditMenuModel) View() string {
	// Calculate responsive width based on terminal size
	contentWidth := m.width - 4 // Account for padding
	if contentWidth < 60 {
		contentWidth = 60 // Minimum width
	}
	if contentWidth > 160 {
		contentWidth = 160 // Maximum width for readability
	}

	// Title with consistent left alignment
	title := TitleStyle.Render("🔍 Localhost Audit")

	// API Key status with consistent formatting
	var apiStatus string
	if m.config.APIKey != "" {
		apiStatus = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("✅ API Key configured (premium features available)")
	} else {
		apiStatus = lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("ℹ️  API Key not set (audits work without API key)")
	}

	// Menu items with consistent formatting
	var menuLines []string
	for i, item := range auditMenuItems {
		var line string

		if i == m.selectedIdx {
			// Selected item: show title and description on same line
			line = fmt.Sprintf("▶ %s %s - %s", item.icon, item.title, item.description)
			menuLines = append(menuLines, SelectedItemStyle.Render(line))
		} else {
			// Non-selected item: show just the title
			line = fmt.Sprintf("  %s %s", item.icon, item.title)
			menuLines = append(menuLines, ListItemStyle.Render(line))
		}
	}

	// Menu section
	menu := lipgloss.JoinVertical(lipgloss.Left, menuLines...)

	// Current default settings
	defaultsInfo := m.renderDefaultSettings(contentWidth)

	// Status bar with help - no credits needed for audit features
	statusBar := RenderKeyHelp(map[string]string{
		"↑↓":    "Navigate",
		"Enter": "Select",
		"Esc":   "Back",
	})

	// Create consistent layout with standardized spacing
	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"", // Single empty line after title
			apiStatus,
			"", // Single empty line after status
			menu,
			"", // Single empty line after menu
			defaultsInfo,
			"", // Single empty line after settings
			statusBar,
		),
	)
}

// renderDefaultSettings shows the current default audit settings
func (m *AuditMenuModel) renderDefaultSettings(contentWidth int) string {
	// Format ignore patterns for display
	var ignoreText string
	if len(m.config.DefaultIgnorePatterns) > 0 {
		ignoreText = strings.Join(m.config.DefaultIgnorePatterns, ", ")
	} else {
		ignoreText = "none"
	}

	settings := []string{
		fmt.Sprintf("Default Port: %d", m.config.DefaultPort),
		fmt.Sprintf("Max Pages: %s", m.formatLimit(m.config.DefaultMaxPages)),
		fmt.Sprintf("Max Depth: %s", m.formatLimit(m.config.DefaultMaxDepth)),
		fmt.Sprintf("Concurrency: %d", m.config.DefaultConcurrency),
		fmt.Sprintf("Ignore Patterns: %s", ignoreText),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Width(contentWidth).Render("Current Default Settings:"),
		"",
		ContentStyle.Width(contentWidth).Render(lipgloss.JoinVertical(lipgloss.Left, settings...)),
	)
}

// handleSelection processes menu selection
func (m *AuditMenuModel) handleSelection() (tea.Model, tea.Cmd) {
	switch m.selectedIdx {
	case 0: // New Audit
		// Create new audit with default settings
		newAuditModel := NewNewAuditModel(m.config)
		return newAuditModel, newAuditModel.Init()

	case 1: // New Audit (Single Page)
		// Create single page audit model
		singlePageAuditModel := NewSinglePageAuditModel(m.config)
		return singlePageAuditModel, singlePageAuditModel.Init()

	case 2: // Audit History
		// Create audit history model
		auditHistoryModel := NewAuditHistoryModel(m.config)
		return auditHistoryModel, auditHistoryModel.Init()
	}

	return m, nil
}

// formatLimit formats limit values for display
func (m *AuditMenuModel) formatLimit(limit int) string {
	if limit == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", limit)
}

