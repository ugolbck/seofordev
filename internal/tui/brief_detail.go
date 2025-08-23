package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// BriefDetailModel displays a single brief in detail
type BriefDetailModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Brief data
	brief api.BriefHistoryItem

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Scrolling for long briefs
	scrollOffset int

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Navigation
	quitting bool
}

// NewBriefDetailModel creates a new brief detail model
func NewBriefDetailModel(config *Config, brief api.BriefHistoryItem) *BriefDetailModel {
	return &BriefDetailModel{
		config:  config,
		brief:   brief,
		credits: -1, // Not loaded yet
	}
}

// Init implements tea.Model
func (m *BriefDetailModel) Init() tea.Cmd {
	// Fetch credits if we have an API key
	if m.config != nil && m.config.APIKey != "" {
		return m.fetchCredits()
	}
	return nil
}

// Update implements tea.Model
func (m *BriefDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case CreditsMsg:
		// Update credit balance
		if msg.Error == nil {
			m.credits = msg.Credits
		} else {
			m.credits = -1
		}
		return m, nil

	case NotificationMsg:
		// Store notification and timestamp for auto-hide
		m.notification = &msg
		m.notificationTime = time.Now()
		// Auto-hide notification after 3 seconds
		return m, tea.Tick(time.Second*3, func(time.Time) tea.Msg {
			return HideNotificationMsg{}
		})

	case HideNotificationMsg:
		// Clear notification if it's been long enough
		if time.Since(m.notificationTime) >= time.Second*3 {
			m.notification = nil
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *BriefDetailModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	return m.renderBriefDetail()
}

// renderBriefDetail renders the detailed view of the brief
func (m *BriefDetailModel) renderBriefDetail() string {
	title := TitleStyle.Render("📄 Content Brief Details")

	// Brief metadata
	generatedAt, err := time.Parse(time.RFC3339, m.brief.GeneratedAt)
	var dateStr string
	if err != nil {
		dateStr = m.brief.GeneratedAt // Fallback to raw string
	} else {
		dateStr = generatedAt.Format("Jan 2, 2006 at 3:04 PM")
	}

	// Status indicator
	var statusIcon string
	var statusText string
	switch m.brief.Status {
	case "completed":
		statusIcon = "✅"
		statusText = "Completed"
	case "failed":
		statusIcon = "❌"
		statusText = "Failed"
	case "processing":
		statusIcon = "⏳"
		statusText = "Processing"
	case "pending":
		statusIcon = "⏸"
		statusText = "Pending"
	default:
		statusIcon = "❓"
		statusText = strings.Title(m.brief.Status)
	}

	// Metadata section
	metadata := lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("Brief Information"),
		ContentStyle.Render(fmt.Sprintf("Keyword: %s", m.brief.Keyword)),
		ContentStyle.Render(fmt.Sprintf("Status: %s %s", statusIcon, statusText)),
		ContentStyle.Render(fmt.Sprintf("Generated: %s", dateStr)),
		ContentStyle.Render(fmt.Sprintf("Credits Used: %d", m.brief.CreditsUsed)),
		ContentStyle.Render(fmt.Sprintf("Brief ID: %s", m.brief.ID)),
	)

	metadataBox := BoxStyle.Render(metadata)

	// Brief content section
	var contentSection string
	if m.brief.Brief != nil && *m.brief.Brief != "" {
		contentTitle := SubtitleStyle.Render("Generated Brief")

		// Calculate available height for content
		contentHeight := m.height - 20 // Account for title, metadata, status bar, etc.
		if contentHeight < 10 {
			contentHeight = 10
		}

		// Split content into lines and handle scrolling
		contentLines := strings.Split(*m.brief.Brief, "\n")

		// Apply scroll offset
		startLine := m.scrollOffset
		endLine := startLine + contentHeight
		if endLine > len(contentLines) {
			endLine = len(contentLines)
		}

		var visibleLines []string
		if startLine < len(contentLines) {
			visibleLines = contentLines[startLine:endLine]
		}

		// Show scroll indicators
		var scrollIndicators []string
		if m.scrollOffset > 0 {
			scrollIndicators = append(scrollIndicators,
				lipgloss.NewStyle().Foreground(MutedColor).Italic(true).Render("↑ More content above"))
		}

		scrollIndicators = append(scrollIndicators, visibleLines...)

		if endLine < len(contentLines) {
			scrollIndicators = append(scrollIndicators,
				lipgloss.NewStyle().Foreground(MutedColor).Italic(true).Render("↓ More content below"))
		}

		// Fixed width content frame to prevent dynamic resizing
		fixedFrameWidth := 80
		if m.width > 0 && m.width-10 < fixedFrameWidth {
			fixedFrameWidth = m.width - 10 // Adjust for smaller terminals
		}

		briefContent := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1).
			Width(fixedFrameWidth).
			Height(contentHeight).
			Render(strings.Join(scrollIndicators, "\n"))

		contentSection = lipgloss.JoinVertical(lipgloss.Left,
			contentTitle,
			briefContent,
		)
	} else {
		contentSection = lipgloss.JoinVertical(lipgloss.Left,
			SubtitleStyle.Render("Generated Brief"),
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Render("No content available for this brief."),
		)
	}

	// Status bar - using ordered map to ensure consistent navigation order
	var statusBarHelp map[string]string
	if m.brief.Brief != nil && *m.brief.Brief != "" && len(strings.Split(*m.brief.Brief, "\n")) > 10 {
		statusBarHelp = map[string]string{
			"↑↓":  "Scroll content",
			"e":   "Export to clipboard",
			"b":   "Generate new brief for this keyword",
			"h":   "Back to history",
			"Esc": "Back",
		}
	} else {
		statusBarHelp = map[string]string{
			"e":   "Export to clipboard",
			"b":   "Generate new brief for this keyword",
			"h":   "Back to history",
			"Esc": "Back",
		}
	}

	statusBar := RenderStatusBar(statusBarHelp, m.credits, m.config != nil && m.config.APIKey != "")

	// Build content with optional notification
	content := []string{
		title,
		"",
		metadataBox,
		"",
		contentSection,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", statusBar)

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

// handleKeypress handles keyboard input
func (m *BriefDetailModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Go back to brief history
		historyModel := NewContentBriefHistoryModel(m.config)
		return historyModel, historyModel.Init()

	case "h":
		// Go back to brief history
		historyModel := NewContentBriefHistoryModel(m.config)
		return historyModel, historyModel.Init()

	case "b":
		// Generate new brief for this keyword
		generationModel := NewContentBriefGenerationModelWithKeyword(m.brief.Keyword)
		generationModel.config = m.config
		return generationModel, generationModel.Init()

	case "e":
		// Export content brief to clipboard
		if m.brief.Brief != nil && *m.brief.Brief != "" {
			return m, ExportContentBriefToClipboardWithNotification(*m.brief.Brief)
		}
		return m, nil

	case "up", "k":
		// Scroll up
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "down", "j":
		// Scroll down
		if m.brief.Brief != nil {
			contentLines := strings.Split(*m.brief.Brief, "\n")
			maxScroll := len(contentLines) - 10 // Leave some content visible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}
		}
		return m, nil
	}

	return m, nil
}

// fetchCredits fetches the current credit balance from the API
func (m *BriefDetailModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
