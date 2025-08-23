package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ugolbck/seofordev/internal/api"
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
	config *Config

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

// getMenuItems returns menu items with lock icons for unpaid features
func (m *MainMenuModel) getMenuItems() []menuItem {
	items := []menuItem{
		{"Localhost Audit", "Audit your local website for SEO issues", "🔍"},
		{"Settings", "Configure API key and default parameters", "⚙️ "},
		{"Help", "Get help and documentation", "❓"},
		{"Quit", "Exit the application", "👋"},
	}

	// Add keyword and content brief items with lock icons if no paid plan
	if m.planLoaded && !m.hasPaidPlan {
		// Show locked items for non-paid users
		items = append([]menuItem{items[0]}, append([]menuItem{
			{"Keyword Generator 🔒", "Generate and research SEO keywords (requires paid plan)", "🔑"},
			{"Content Brief for AI 🔒", "Generate content briefs for AI writing (requires paid plan)", "📄"},
		}, items[1:]...)...)
	} else if m.planLoaded && m.hasPaidPlan {
		// Show unlocked items for paid users
		items = append([]menuItem{items[0]}, append([]menuItem{
			{"Keyword Generator", "Generate and research SEO keywords", "🔑"},
			{"Content Brief for AI", "Generate content briefs for AI writing", "📄"},
		}, items[1:]...)...)
	} else {
		// Loading state - show items without lock status
		items = append([]menuItem{items[0]}, append([]menuItem{
			{"Keyword Generator", "Generate and research SEO keywords", "🔑"},
			{"Content Brief for AI", "Generate content briefs for AI writing", "📄"},
		}, items[1:]...)...)
	}

	return items
}

// NewMainMenuModel creates a new main menu
func NewMainMenuModel() *MainMenuModel {
	// Load configuration from file
	config, err := LoadConfig()
	if err != nil {
		// If we can't load config, use defaults but log the error
		fmt.Printf("Warning: Could not load config: %v\n", err)
		config = getDefaultConfig()
	}

	return &MainMenuModel{
		selectedIdx: 0,
		config:      config,
		credits:     -1, // Not loaded yet
		planLoaded:  false,
	}
}

// NewMainMenuModelWithConfig creates a new main menu with provided config
func NewMainMenuModelWithConfig(config *Config) *MainMenuModel {
	m := &MainMenuModel{
		selectedIdx: 0,
		config:      config,
		credits:     -1, // Not loaded yet
		planLoaded:  false,
	}
	return m
}

// NewMainMenuModelWithVersionCheck creates a new main menu with version check result
func NewMainMenuModelWithVersionCheck(versionResult VersionCheckResult) *MainMenuModel {
	// Load configuration from file
	config, err := LoadConfig()
	if err != nil {
		// If we can't load config, use defaults but log the error
		fmt.Printf("Warning: Could not load config: %v\n", err)
		config = getDefaultConfig()
	}

	return &MainMenuModel{
		selectedIdx:   0,
		config:        config,
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
		if backMsg, ok := msg.(BackToMenuMsg); ok {
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

		// Check for navigation to audit details
		if navMsg, ok := msg.(NavigateToAuditDetailsMsg); ok {
			auditDetails := NewAuditDetailsModel(m.config, navMsg.Audit)
			m.activeModel = auditDetails
			return m, auditDetails.Init()
		}

		// Check for navigation to page details
		if pageMsg, ok := msg.(NavigateToPageDetailsMsg); ok {
			// Create a dedicated page details model
			pageDetails := NewPageDetailsModel(m.config, pageMsg.AuditID, pageMsg.PageURL, pageMsg.Audit)
			m.activeModel = pageDetails
			return m, pageDetails.Init()
		}

		// Check for navigation back to audit menu
		if _, ok := msg.(BackToAuditMenuMsg); ok {
			auditMenu := NewAuditMenuModel(m.config)
			m.activeModel = auditMenu
			return m, auditMenu.Init()
		}

		// Check for navigation back to audit details
		if backToDetailsMsg, ok := msg.(BackToAuditDetailsMsg); ok {
			auditDetails := NewAuditDetailsModel(m.config, backToDetailsMsg.Audit)
			m.activeModel = auditDetails
			return m, auditDetails.Init()
		}

		// Check for navigation back to keyword menu
		if _, ok := msg.(BackToKeywordMenuMsg); ok {
			keywordMenu := NewKeywordMenuModel(m.config)
			m.activeModel = keywordMenu
			// Refresh credits when returning from paid features
			var cmds []tea.Cmd
			cmds = append(cmds, keywordMenu.Init())
			if m.planLoaded && m.hasPaidPlan {
				cmds = append(cmds, m.fetchCredits())
			}
			return m, tea.Batch(cmds...)
		}

		// Check for navigation to generation details
		if navMsg, ok := msg.(NavigateToGenerationDetailsMsg); ok {
			generationDetails := NewGenerationDetailsModel(m.config, navMsg.Generation)
			m.activeModel = generationDetails
			return m, generationDetails.Init()
		}

		// Check for navigation back to content brief menu
		if _, ok := msg.(BackToContentBriefMenuMsg); ok {
			contentBriefMenu := NewContentBriefMenuModel(m.config)
			m.activeModel = contentBriefMenu
			// Refresh credits when returning from paid features
			var cmds []tea.Cmd
			cmds = append(cmds, contentBriefMenu.Init())
			if m.planLoaded && m.hasPaidPlan {
				cmds = append(cmds, m.fetchCredits())
			}
			return m, tea.Batch(cmds...)
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
			menuItems := m.getMenuItems()
			if m.selectedIdx == len(menuItems)-1 { // Quit option
				m.quitting = true
				return m, tea.Quit
			}
			m.selectedIdx = len(menuItems) - 1 // Move to quit option
			return m, nil

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
		return AppStyle.Render("Thanks for using SEO Developer Tools! 👋")
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

	// ASCII art title
	asciiArt := "                  __             _            \n" +
		"                / _|            | |           \n" +
		" ___  ___  ___ | |_ ___  _ __ __| | _____   __\n" +
		"/ __|/ _ \\/ _ \\|  _/ _ \\| '__/ _` |/ _ \\ \\ / /\n" +
		"\\__ \\  __/ (_) | || (_) | |_| (_| |  __/\\ V / \n" +
		"|___|\\___|\\___/|_| \\___/|_(_)\\__,_|\\___| \\_/  \n" +
		"                                              \n" +
		"        🚀 SEO tools for indie hackers"

	title := TitleStyle.Width(contentWidth).Render(asciiArt)

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
	}

	// Menu section
	menu := lipgloss.JoinVertical(lipgloss.Left, menuLines...)

	// Status section with consistent alignment
	var statusInfo string
	if m.config.APIKey != "" {
		statusInfo = fmt.Sprintf("API Key: %s", MaskAPIKey(m.config.APIKey))
	} else {
		statusInfo = "⚠️  API Key not configured"
	}

	status := lipgloss.NewStyle().
		Foreground(MutedColor).
		Width(contentWidth).
		Align(lipgloss.Left). // Changed from Center to Left
		Render(statusInfo)

	// Status bar with help and credits (only show credits if user has paid plan)
	var statusBar string
	if m.planLoaded && m.hasPaidPlan {
		statusBar = RenderStatusBar(map[string]string{
			"↑↓":    "Navigate",
			"Enter": "Select",
			"q":     "Quit",
		}, m.credits, m.config.APIKey != "")
	} else {
		statusBar = RenderKeyHelp(map[string]string{
			"↑↓":    "Navigate",
			"Enter": "Select",
			"q":     "Quit",
		})
	}

	// Update notification (if available)
	var updateNotification string
	if m.versionResult != nil && m.versionResult.HasUpdate {
		updateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")). // Orange color
			Bold(true).
			Width(contentWidth).
			Align(lipgloss.Left)

		updateNotification = updateStyle.Render(GetUpdateMessage(*m.versionResult))
	}

	// Create consistent layout with standardized spacing
	var layoutComponents []string
	layoutComponents = append(layoutComponents, title)
	layoutComponents = append(layoutComponents, "") // Single empty line after title
	layoutComponents = append(layoutComponents, menu)
	layoutComponents = append(layoutComponents, "") // Single empty line after menu
	layoutComponents = append(layoutComponents, status)
	layoutComponents = append(layoutComponents, "") // Single empty line after status
	layoutComponents = append(layoutComponents, statusBar)

	// Add update notification if available
	if updateNotification != "" {
		layoutComponents = append(layoutComponents, "") // Single empty line
		layoutComponents = append(layoutComponents, updateNotification)
	}

	layout := lipgloss.JoinVertical(lipgloss.Left, layoutComponents...)

	// Apply the main app style with responsive width
	return AppStyle.Width(contentWidth).Render(layout)
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
	case selectedItem.title == "Localhost Audit":
		auditMenu := NewAuditMenuModel(m.config)
		m.activeModel = auditMenu
		return m, auditMenu.Init()

	case selectedItem.title == "Keyword Generator" && m.planLoaded && m.hasPaidPlan:
		keywordMenu := NewKeywordMenuModel(m.config)
		m.activeModel = keywordMenu
		return m, keywordMenu.Init()

	case selectedItem.title == "Keyword Generator 🔒":
		return m, tea.Printf("🔒 Keyword Generator requires a paid plan. Please upgrade at seofor.dev")

	case selectedItem.title == "Content Brief for AI" && m.planLoaded && m.hasPaidPlan:
		contentBriefMenu := NewContentBriefMenuModel(m.config)
		m.activeModel = contentBriefMenu
		return m, contentBriefMenu.Init()

	case selectedItem.title == "Content Brief for AI 🔒":
		return m, tea.Printf("🔒 Content Brief Generator requires a paid plan. Please upgrade at seofor.dev")

	case selectedItem.title == "Settings":
		settingsMenu := NewSettingsMenuModel(m.config)
		m.activeModel = settingsMenu
		return m, settingsMenu.Init()

	case selectedItem.title == "Help":
		helpModel := NewHelpModel()
		m.activeModel = helpModel
		return m, helpModel.Init()

	case selectedItem.title == "Quit":
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleReturnData processes data returned from sub-models
func (m *MainMenuModel) handleReturnData(data interface{}) tea.Cmd {
	switch d := data.(type) {
	case *Config:
		// Update config and save it
		m.config = d
		if err := SaveConfig(m.config); err != nil {
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
		resp, err := ValidateAPIKeyWithServer(m.config.APIKey, m.config.GetEffectiveBaseURL())
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
type BackToMenuMsg struct {
	Data interface{} // Optional data to pass back
}

// BackToAuditMenuMsg is sent by sub-models to return to audit menu
type BackToAuditMenuMsg struct {
	Data interface{} // Optional data to pass back
}

// BackToAuditDetailsMsg is sent by sub-models to return to audit details
type BackToAuditDetailsMsg struct {
	Audit api.AuditViewResponse // The audit to return to
}
