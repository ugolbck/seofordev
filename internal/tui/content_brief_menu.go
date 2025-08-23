package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ugolbck/seofordev/internal/api"
)

// ContentBriefMenuModel provides content brief-related options
type ContentBriefMenuModel struct {
	width       int
	height      int
	selectedIdx int

	// Current active model (nil means we're in content brief menu)
	activeModel tea.Model

	// Configuration
	config *Config

	// Plan status
	hasPaidPlan bool
	credits     int // -1 means not loaded yet
	planLoaded  bool
}

type contentBriefMenuItem struct {
	title       string
	description string
	icon        string
}

var contentBriefMenuItems = []contentBriefMenuItem{
	{"New Content Brief Generation", "Generate a new content brief for AI writing", "✨"},
	{"Content Brief History", "View previous content brief generations", "📊"},
}

// NewContentBriefMenuModel creates a new content brief menu
func NewContentBriefMenuModel(config *Config) *ContentBriefMenuModel {
	return &ContentBriefMenuModel{
		selectedIdx: 0,
		config:      config,
		credits:     -1, // Not loaded yet
		planLoaded:  false,
	}
}

// Init implements tea.Model
func (m *ContentBriefMenuModel) Init() tea.Cmd {
	// Fetch plan status if we have an API key
	if m.config.APIKey != "" {
		return m.fetchPlanStatus()
	}
	return nil
}

// Update implements tea.Model
func (m *ContentBriefMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If we have an active model, delegate to it first
	if m.activeModel != nil {
		var cmd tea.Cmd
		m.activeModel, cmd = m.activeModel.Update(msg)

		// Check if the active model wants to return to content brief menu
		if _, ok := msg.(BackToContentBriefMenuMsg); ok {
			m.activeModel = nil
			return m, cmd
		}

		return m, cmd
	}

	// Handle content brief menu input
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
			if m.selectedIdx < len(contentBriefMenuItems)-1 {
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
			m.planLoaded = false
		}
		return m, nil

	case CreditsMsg:
		// Update credit balance
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
func (m *ContentBriefMenuModel) View() string {
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

	// Title with consistent left alignment
	title := TitleStyle.Render("📄 Content Brief for AI")

	// API Key status with consistent formatting
	var apiStatus string
	if m.config.APIKey != "" {
		apiStatus = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("✅ API Key configured")
	} else {
		apiStatus = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render("❌ API Key not set (go to Settings)")
	}

	// Menu items with consistent formatting
	var menuLines []string
	for i, item := range contentBriefMenuItems {
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

	// Status bar with help and credits (only show credits if user has paid plan)
	var statusBar string
	if m.planLoaded && m.hasPaidPlan {
		statusBar = RenderStatusBar(map[string]string{
			"↑↓":    "Navigate",
			"Enter": "Select",
			"Esc":   "Back",
		}, m.credits, m.config.APIKey != "")
	} else {
		statusBar = RenderKeyHelp(map[string]string{
			"↑↓":    "Navigate",
			"Enter": "Select",
			"Esc":   "Back",
		})
	}

	// Create consistent layout with standardized spacing
	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"", // Single empty line after title
			apiStatus,
			"", // Single empty line after status
			menu,
			"", // Single empty line after menu
			statusBar,
		),
	)
}

// handleSelection processes menu selection
func (m *ContentBriefMenuModel) handleSelection() (tea.Model, tea.Cmd) {
	switch m.selectedIdx {
	case 0: // New Content Brief Generation
		// Check if user has paid plan
		if m.planLoaded && !m.hasPaidPlan {
			return m, tea.Printf("🔒 Content brief generation requires a paid plan. Please upgrade at seofor.dev")
		}

		generationModel := NewContentBriefGenerationModel()
		generationModel.config = m.config
		m.activeModel = generationModel
		return m, generationModel.Init()

	case 1: // Content Brief History
		// Check if user has paid plan
		if m.planLoaded && !m.hasPaidPlan {
			return m, tea.Printf("🔒 Content brief history requires a paid plan. Please upgrade at seofor.dev")
		}

		historyModel := NewContentBriefHistoryModel(m.config)
		m.activeModel = historyModel
		return m, historyModel.Init()
	}

	return m, nil
}

// fetchPlanStatus fetches the user's plan status and credit balance from the API validation endpoint
func (m *ContentBriefMenuModel) fetchPlanStatus() tea.Cmd {
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

// fetchCredits fetches the current credit balance from the API
func (m *ContentBriefMenuModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// BackToContentBriefMenuMsg is sent by sub-models to return to content brief menu
type BackToContentBriefMenuMsg struct {
	Data interface{} // Optional data to pass back
}
