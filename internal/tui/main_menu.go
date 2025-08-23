package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ugolbck/seofordev/internal/api"
	"github.com/ugolbck/seofordev/internal/tui/config"
)

// Message types for credit balance
type CreditsMsg struct {
	Credits int
	Error   error
}

// Message types for plan status
type PlanStatusMsg struct {
	HasPaidPlan bool
	Credits     int
	Error       error
}

// MainMenuModel is the primary interface for all SEO tools
type MainMenuModel struct {
	width       int
	height      int
	selectedIdx int
	quitting    bool

	// Current active model (nil means we're in main menu)
	activeModel tea.Model

	// Configuration loaded from file
	config *config.Config

	// User plan status
	hasPaidPlan bool // Whether user has paid plan
	credits     int  // -1 means not loaded yet (only for paid features)
	planLoaded  bool // Whether we've loaded the plan status

	// Version checking
	versionResult *VersionCheckResult
}

type menuItem struct {
	title       string
	description string
	icon        string
}

// getMenuItems returns menu items based on API key availability
func (m *MainMenuModel) getMenuItems() []menuItem {
	items := []menuItem{
		{"Localhost Audit", "Find and fix your local web server pages", "🔍"},
		{"Keyword Generator", "Generate and research SEO keywords", "🔑"},
		{"Content Brief for AI", "Generate content briefs for AI writing", "📄"},
		{"Settings", "Configure API key and default parameters", "⚙️ "},
		{"Help", "Get help and documentation", "❓"},
		{"Quit", "Exit the application", "👋"},
	}

	return items
}

// NewMainMenuModelWithVersionCheck creates a new main menu with version check result
func NewMainMenuModelWithVersionCheck(versionResult VersionCheckResult) *MainMenuModel {
	// Load configuration from file
	newConfig, err := config.LoadConfig()
	if err != nil {
		// If we can't load config, use defaults but log the error
		fmt.Printf("Warning: Could not load config: %v\n", err)
		newConfig = config.GetDefaultConfig()
	}

	return &MainMenuModel{
		selectedIdx:   0,
		config:        newConfig,
		credits:       -1, // Not loaded yet
		planLoaded:    false,
		versionResult: &versionResult,
	}
}

// Init implements tea.Model
func (m *MainMenuModel) Init() tea.Cmd {
	// Fetch user plan status if we have an API key
	if m.config.APIKey != "" {
		return m.fetchPlanStatus()
	}
	return nil
}

// Update implements tea.Model
func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If we have an active model, delegate to it first
	if m.activeModel != nil {
		var cmd tea.Cmd
		m.activeModel, cmd = m.activeModel.Update(msg)

		// Check if the active model wants to return to main menu
		if backMsg, ok := msg.(BackToMainMenuMsg); ok {
			// Handle any data return from the sub-model
			var extraCmd tea.Cmd
			if backMsg.Data != nil {
				extraCmd = m.handleReturnData(backMsg.Data)
			}
			m.activeModel = nil

			// Refresh credits when returning from paid features to main menu
			if m.planLoaded && m.hasPaidPlan {
				creditCmd := m.fetchCredits()
				if extraCmd != nil {
					return m, tea.Batch(cmd, extraCmd, creditCmd)
				}
				return m, tea.Batch(cmd, creditCmd)
			}

			if extraCmd != nil {
				return m, tea.Batch(cmd, extraCmd)
			}
			return m, cmd
		}

		return m, cmd
	}

	// Handle main menu input
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil

		case "down", "j":
			menuItems := m.getMenuItems()
			if m.selectedIdx < len(menuItems)-1 {
				m.selectedIdx++
			}
			return m, nil

		case "enter", " ":
			return m.handleSelection()
		}

	case PlanStatusMsg:
		// Update plan status
		if msg.Error == nil {
			m.hasPaidPlan = msg.HasPaidPlan
			m.credits = msg.Credits
			m.planLoaded = true
		} else {
			// Keep current values on error
			m.planLoaded = false
		}
		return m, nil

	case CreditsMsg:
		// Update credit balance (for paid users only)
		if msg.Error == nil {
			m.credits = msg.Credits
		} else {
			// Keep current credits value on error
			m.credits = -1
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *MainMenuModel) View() string {
	if m.quitting {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Thanks for using seofor.dev! 👋")
	}

	// If we have an active model, show it
	if m.activeModel != nil {
		return m.activeModel.View()
	}

	// Calculate responsive width based on terminal size
	contentWidth := m.width - 4 // Account for padding
	if contentWidth < 60 {
		contentWidth = 60 // Minimum width
	}
	if contentWidth > 160 {
		contentWidth = 160 // Maximum width for readability
	}

	// Breadcrumbs at the top left
	breadcrumbs := BreadcrumbsStyle.Render("home")

	// ASCII art title
	asciiArt := "                 __              _            \n" +
		"                / _|            | |           \n" +
		" ___  ___  ___ | |_ ___  _ __ __| | _____   __\n" +
		"/ __|/ _ \\/ _ \\|  _/ _ \\| '__/ _` |/ _ \\ \\ / /\n" +
		"\\__ \\  __/ (_) | || (_) | |_| (_| |  __/\\ V / \n" +
		"|___|\\___|\\___/|_| \\___/|_(_)\\__,_|\\___| \\_/  \n" +
		"                                              \n" +
		"        🚀 SEO tools for indie hackers"

	// Get dynamic menu items based on plan status
	menuItems := m.getMenuItems()

	// Menu items with consistent formatting
	var menuLines []string
	for i, item := range menuItems {
		var line string

		if i == m.selectedIdx {
			// Selected item: show title and description on same line
			line = fmt.Sprintf("▶ %s %s - %s", item.icon, item.title, item.description)
			style := SelectedItemStyle.Width(contentWidth)
			menuLines = append(menuLines, style.Render(line))
		} else {
			// Non-selected item: show just the title
			line = fmt.Sprintf("  %s %s", item.icon, item.title)
			style := ListItemStyle.Width(contentWidth)
			menuLines = append(menuLines, style.Render(line))
		}

		// Add spacing between menu items (except for the last one)
		if i < len(menuItems)-1 {
			menuLines = append(menuLines, "")
		}
	}

	// Menu section
	menu := lipgloss.JoinVertical(lipgloss.Left, menuLines...)

	title := TitleStyle.Render(asciiArt)
	// Add spacing between title and menu items
	mainContent := lipgloss.JoinVertical(lipgloss.Center, title, "", "", "", "", "", "", "", menu)

	// Helper bar at bottom - always visible with consistent ordering
	helpers := []struct {
		key  string
		desc string
	}{
		{"↑ / j", "up"},
		{"↓ / k", "down"},
		{"enter", "select"},
		{"q / ctrl+c", "quit"},
	}

	// Apply style on all help pairs
	var helperPairs []string
	grayColor := lipgloss.Color("#9CA3AF")       // Gray color for keys
	darkerGrayColor := lipgloss.Color("#6B7280") // Darker gray for descriptions

	for _, helper := range helpers {
		keyText := lipgloss.NewStyle().Foreground(grayColor).Render(helper.key)
		descText := lipgloss.NewStyle().Foreground(darkerGrayColor).Render(helper.desc)
		helperPairs = append(helperPairs, keyText+" "+descText)
	}

	// Join with bullet point separators
	helpTextWithBullets := ""
	for i, pair := range helperPairs {
		if i > 0 {
			helpTextWithBullets += " • "
		}
		helpTextWithBullets += pair
	}

	// Layout with breadcrumbs at top, centered content, helper bar at bottom
	contentHeight := m.height - 5 // Reserve more space for breadcrumbs padding and helper bar
	centeredContent := lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Top, mainContent)

	// Combine everything with proper padding
	return lipgloss.JoinVertical(lipgloss.Left,
		"", // Top padding above breadcrumbs
		lipgloss.NewStyle().Padding(0, 2).Render(breadcrumbs),
		centeredContent,
		"",
		lipgloss.NewStyle().Padding(0, 2).Render(helpTextWithBullets), // Add left padding to helpers
	)
}

// handleSelection processes menu selection
func (m *MainMenuModel) handleSelection() (tea.Model, tea.Cmd) {
	menuItems := m.getMenuItems()

	if m.selectedIdx >= len(menuItems) {
		return m, nil
	}

	selectedItem := menuItems[m.selectedIdx]

	// Handle based on menu item title/type
	switch {
	// case selectedItem.title == "Localhost Audit":
	// 	auditMenu := NewAuditMenuModel(m.config)
	// 	m.activeModel = auditMenu
	// 	return m, auditMenu.Init()

	// case selectedItem.title == "Keyword Generator" && m.planLoaded && m.hasPaidPlan:
	// 	keywordMenu := NewKeywordMenuModel(m.config)
	// 	m.activeModel = keywordMenu
	// 	return m, keywordMenu.Init()

	// case selectedItem.title == "Keyword Generator 🔒":
	// 	return m, tea.Printf("🔒 Keyword Generator requires a paid plan. Please upgrade at seofor.dev")

	// case selectedItem.title == "Content Brief for AI" && m.planLoaded && m.hasPaidPlan:
	// 	contentBriefMenu := NewContentBriefMenuModel(m.config)
	// 	m.activeModel = contentBriefMenu
	// 	return m, contentBriefMenu.Init()

	// case selectedItem.title == "Content Brief for AI 🔒":
	// 	return m, tea.Printf("🔒 Content Brief Generator requires a paid plan. Please upgrade at seofor.dev")

	// case selectedItem.title == "Settings":
	// 	settingsMenu := NewSettingsMenuModel(m.config)
	// 	m.activeModel = settingsMenu
	// 	return m, settingsMenu.Init()

	// case selectedItem.title == "Help":
	// 	helpModel := NewHelpModel()
	// 	m.activeModel = helpModel
	// 	return m, helpModel.Init()

	case selectedItem.title == "Quit":
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleReturnData processes data returned from sub-models
func (m *MainMenuModel) handleReturnData(data interface{}) tea.Cmd {
	switch d := data.(type) {
	case *config.Config:
		// Update config and save it
		m.config = d
		if err := config.SaveConfig(m.config); err != nil {
			// TODO: Show error to user
			fmt.Printf("Warning: Could not save config: %v\n", err)
		}
		// Fetch plan status with the new API key
		if m.config.APIKey != "" {
			return m.fetchPlanStatus()
		}
	}
	return nil
}

// fetchPlanStatus fetches the user's plan status and credit balance from the API validation endpoint
func (m *MainMenuModel) fetchPlanStatus() tea.Cmd {
	return func() tea.Msg {
		resp, err := config.ValidateAPIKeyWithServer(m.config.APIKey, m.config.GetEffectiveBaseURL())
		if err != nil {
			return PlanStatusMsg{Error: err}
		}
		return PlanStatusMsg{
			HasPaidPlan: resp.User.HasPaidPlan,
			Credits:     resp.User.Credits,
			Error:       nil,
		}
	}
}

// fetchCredits fetches the current credit balance from the API (for paid users)
func (m *MainMenuModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// BackToMenuMsg is sent by sub-models to return to main menu
type BackToMainMenuMsg struct {
	Data interface{} // Optional data to pass back
}
