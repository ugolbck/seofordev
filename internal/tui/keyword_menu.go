package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ugolbck/seofordev/internal/api"
)

// KeywordMenuModel provides keyword-related options
type KeywordMenuModel struct {
	width       int
	height      int
	selectedIdx int

	// Configuration
	config *Config

	// Plan status
	hasPaidPlan bool
	credits     int // -1 means not loaded yet
	planLoaded  bool
}

type keywordMenuItem struct {
	title       string
	description string
	icon        string
}

var keywordMenuItems = []keywordMenuItem{
	{"Generate Keywords", "Generate autocomplete keywords", "✨"},
	{"Keyword History", "View previous generations", "📊"},
}

// NewKeywordMenuModel creates a new keyword menu
func NewKeywordMenuModel(config *Config) *KeywordMenuModel {
	return &KeywordMenuModel{
		selectedIdx: 0,
		config:      config,
		credits:     -1, // Not loaded yet
		planLoaded:  false,
	}
}

// Init implements tea.Model
func (m *KeywordMenuModel) Init() tea.Cmd {
	// Fetch plan status if we have an API key
	if m.config.APIKey != "" {
		return m.fetchPlanStatus()
	}
	return nil
}

// fetchPlanStatus fetches the user's plan status and credit balance from the API validation endpoint
func (m *KeywordMenuModel) fetchPlanStatus() tea.Cmd {
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
func (m *KeywordMenuModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// Update implements tea.Model
func (m *KeywordMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selectedIdx < len(keywordMenuItems)-1 {
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
func (m *KeywordMenuModel) View() string {
	// Calculate responsive width based on terminal size
	contentWidth := m.width - 4 // Account for padding
	if contentWidth < 60 {
		contentWidth = 60 // Minimum width
	}
	if contentWidth > 160 {
		contentWidth = 160 // Maximum width for readability
	}

	// Title with consistent left alignment
	title := TitleStyle.Render("🔑 Keyword Generator")

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
	for i, item := range keywordMenuItems {
		var line string

		if i == m.selectedIdx {
			// Selected item: show title and description on same line, but truncate if needed
			fullLine := fmt.Sprintf("▶ %s %s - %s", item.icon, item.title, item.description)

			// Truncate the line if it's too long to fit in the available width
			maxLineLength := contentWidth - 2 // Leave some margin
			if len(fullLine) > maxLineLength {
				// Calculate how much we can show
				titlePart := fmt.Sprintf("▶ %s %s", item.icon, item.title)
				availableForDesc := maxLineLength - len(titlePart) - 3 // Account for " - "

				if availableForDesc > 10 { // Only show description if we have reasonable space
					truncatedDesc := item.description
					if len(truncatedDesc) > availableForDesc {
						truncatedDesc = truncatedDesc[:availableForDesc-3] + "..."
					}
					line = fmt.Sprintf("%s - %s", titlePart, truncatedDesc)
				} else {
					line = titlePart
				}
			} else {
				line = fullLine
			}

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
func (m *KeywordMenuModel) handleSelection() (tea.Model, tea.Cmd) {
	switch m.selectedIdx {
	case 0: // Generate Keywords
		if m.config.APIKey == "" {
			return m, tea.Printf("❌ Please configure your API key in Settings first")
		}

		// Check if user has paid plan
		if m.planLoaded && !m.hasPaidPlan {
			return m, tea.Printf("🔒 Keyword generation requires a paid plan. Please upgrade at seofor.dev")
		}

		// Create keyword generation model
		generateModel := NewKeywordGenerationModel(m.config)
		return generateModel, generateModel.Init()

	case 1: // Keyword History
		if m.config.APIKey == "" {
			return m, tea.Printf("❌ Please configure your API key in Settings first")
		}

		// Check if user has paid plan
		if m.planLoaded && !m.hasPaidPlan {
			return m, tea.Printf("🔒 Keyword history requires a paid plan. Please upgrade at seofor.dev")
		}

		// Create keyword history model
		historyModel := NewKeywordHistoryModel(m.config)
		return historyModel, historyModel.Init()
	}

	return m, nil
}
