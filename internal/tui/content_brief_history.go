package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// ContentBriefHistoryModel displays the list of past brief generations
type ContentBriefHistoryModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// History data
	briefs      []api.BriefHistoryItem
	selectedIdx int
	loading     bool
	error       error

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Scrolling
	listScrollOffset int // Track scroll position in brief list

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Navigation
	quitting bool
}

// BriefHistoryLoadedMsg handles loading brief history results
type BriefHistoryLoadedMsg struct {
	Briefs []api.BriefHistoryItem
	Error  error
}

// NewContentBriefHistoryModel creates a new content brief history model
func NewContentBriefHistoryModel(config *Config) *ContentBriefHistoryModel {
	return &ContentBriefHistoryModel{
		config:      config,
		selectedIdx: 0,
		loading:     true,
		briefs:      []api.BriefHistoryItem{},
		credits:     -1, // Not loaded yet
	}
}

// Init implements tea.Model
func (m *ContentBriefHistoryModel) Init() tea.Cmd {
	// Fetch both history and credits
	var cmds []tea.Cmd
	cmds = append(cmds, m.fetchBriefHistory())
	if m.config != nil && m.config.APIKey != "" {
		cmds = append(cmds, m.fetchCredits())
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *ContentBriefHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case BriefHistoryLoadedMsg:
		m.loading = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.briefs = msg.Briefs
			// Start at the top (most recent brief) and reset scroll
			m.selectedIdx = 0
			m.listScrollOffset = 0
		}
		return m, nil

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
func (m *ContentBriefHistoryModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.loading {
		return m.renderLoading()
	}

	if m.error != nil {
		return m.renderError()
	}

	if len(m.briefs) == 0 {
		return m.renderEmpty()
	}

	return m.renderBriefsList()
}

// renderLoading shows loading state
func (m *ContentBriefHistoryModel) renderLoading() string {
	title := TitleStyle.Render("📊 Content Brief History")
	loading := ContentStyle.Render("⏳ Loading your content briefs...")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			loading,
		),
	)
}

// renderError shows error state
func (m *ContentBriefHistoryModel) renderError() string {
	title := TitleStyle.Render("📊 Content Brief History")
	errorText := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Render(fmt.Sprintf("❌ Error: %v", m.error))

	statusBar := RenderStatusBar(map[string]string{
		"r":   "Retry",
		"Esc": "Back",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			errorText,
			"",
			statusBar,
		),
	)
}

// renderEmpty shows when there are no briefs
func (m *ContentBriefHistoryModel) renderEmpty() string {
	title := TitleStyle.Render("📊 Content Brief History")
	empty := ContentStyle.Render("No content briefs found. Generate your first brief!")

	statusBar := RenderStatusBar(map[string]string{
		"n":   "New Brief",
		"Esc": "Back",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			empty,
			"",
			statusBar,
		),
	)
}

// renderBriefsList renders the main briefs list view
func (m *ContentBriefHistoryModel) renderBriefsList() string {
	title := TitleStyle.Render("📊 Content Brief History")

	// Calculate items to show
	maxVisible := m.getMaxVisibleItems()
	start := m.listScrollOffset
	end := start + maxVisible
	if end > len(m.briefs) {
		end = len(m.briefs)
	}

	var items []string

	// Add scroll indicator if there are items above
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↑ More briefs above")
		items = append(items, scrollUp)
	}

	// Add brief items
	for i := start; i < end; i++ {
		brief := m.briefs[i]
		isSelected := i == m.selectedIdx
		items = append(items, m.renderBriefItem(i, brief, isSelected))
	}

	// Add scroll indicator if there are items below
	if end < len(m.briefs) {
		scrollDown := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↓ More briefs below")
		items = append(items, scrollDown)
	}

	briefsList := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Status bar
	statusBar := RenderStatusBar(map[string]string{
		"↑↓":    "Navigate",
		"Enter": "View Brief",
		"e":     "Export to clipboard",
		"n":     "New Brief",
		"Esc":   "Back",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	// Build content with optional notification
	content := []string{
		title,
		"",
		briefsList,
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

// renderBriefItem renders a single brief item
func (m *ContentBriefHistoryModel) renderBriefItem(index int, brief api.BriefHistoryItem, isSelected bool) string {
	// Parse the generated date
	generatedAt, err := time.Parse(time.RFC3339, brief.GeneratedAt)
	var dateStr string
	if err != nil {
		dateStr = brief.GeneratedAt // Fallback to raw string
	} else {
		dateStr = generatedAt.Format("Jan 2, 2006 at 3:04 PM")
	}

	// Status indicator
	var statusIcon string
	switch brief.Status {
	case "completed":
		statusIcon = "✅"
	case "failed":
		statusIcon = "❌"
	case "processing":
		statusIcon = "⏳"
	case "pending":
		statusIcon = "⏸"
	default:
		statusIcon = "❓"
	}

	// Brief preview (first 80 chars)
	var briefPreview string
	if brief.Brief != nil && *brief.Brief != "" {
		preview := strings.ReplaceAll(*brief.Brief, "\n", " ")
		if len(preview) > 80 {
			briefPreview = preview[:77] + "..."
		} else {
			briefPreview = preview
		}
	} else {
		briefPreview = "No content available"
	}

	// Build item content
	line1 := fmt.Sprintf("%s %s", statusIcon, brief.Keyword)
	line2 := fmt.Sprintf("Generated: %s | Credits: %d", dateStr, brief.CreditsUsed)
	line3 := fmt.Sprintf("Preview: %s", briefPreview)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(line1),
		lipgloss.NewStyle().Foreground(MutedColor).Render(line2),
		lipgloss.NewStyle().Foreground(TextColor).Render(line3),
	)

	// Apply selection styling
	if isSelected {
		return SelectedItemStyle.Render(content)
	}

	return ListItemStyle.Render(content)
}

// handleKeypress handles keyboard input
func (m *ContentBriefHistoryModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		return m, func() tea.Msg { return BackToContentBriefMenuMsg{} }

	case "r":
		if m.error != nil {
			// Retry loading
			m.loading = true
			m.error = nil
			return m, m.fetchBriefHistory()
		}
		return m, nil

	case "n":
		// Navigate to new brief generation
		generationModel := NewContentBriefGenerationModel()
		generationModel.config = m.config
		return generationModel, generationModel.Init()

	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
			// Adjust scroll if needed
			if m.selectedIdx < m.listScrollOffset {
				m.listScrollOffset = m.selectedIdx
			}
		}
		return m, nil

	case "down", "j":
		if m.selectedIdx < len(m.briefs)-1 {
			m.selectedIdx++
			// Adjust scroll if needed
			maxVisible := m.getMaxVisibleItems()
			if m.selectedIdx >= m.listScrollOffset+maxVisible {
				m.listScrollOffset = m.selectedIdx - maxVisible + 1
			}
		}
		return m, nil

	case "enter", " ":
		if len(m.briefs) > 0 && m.selectedIdx < len(m.briefs) {
			// View the selected brief
			return m.viewBrief(m.briefs[m.selectedIdx])
		}
		return m, nil

	case "e":
		// Export selected brief to clipboard
		if len(m.briefs) > 0 && m.selectedIdx < len(m.briefs) {
			selectedBrief := m.briefs[m.selectedIdx]
			if selectedBrief.Brief != nil && *selectedBrief.Brief != "" {
				return m, ExportContentBriefToClipboardWithNotification(*selectedBrief.Brief)
			}
		}
		return m, nil
	}

	return m, nil
}

// viewBrief opens the brief in a detailed view
func (m *ContentBriefHistoryModel) viewBrief(brief api.BriefHistoryItem) (tea.Model, tea.Cmd) {
	// For now, we'll create a simple brief viewer
	// In the future, this could be a more sophisticated viewer
	detailModel := NewBriefDetailModel(m.config, brief)
	return detailModel, detailModel.Init()
}

// fetchBriefHistory fetches the brief history from the API
func (m *ContentBriefHistoryModel) fetchBriefHistory() tea.Cmd {
	return func() tea.Msg {
		if m.config == nil || m.config.APIKey == "" {
			return BriefHistoryLoadedMsg{Error: fmt.Errorf("API key not configured")}
		}

		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		history, err := client.GetBriefHistory()
		if err != nil {
			return BriefHistoryLoadedMsg{Error: fmt.Errorf("failed to fetch brief history: %w", err)}
		}
		return BriefHistoryLoadedMsg{Briefs: history.Briefs}
	}
}

// fetchCredits fetches the current credit balance from the API
func (m *ContentBriefHistoryModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// getMaxVisibleItems calculates how many items can be displayed
func (m *ContentBriefHistoryModel) getMaxVisibleItems() int {
	// Account for title, status bar, and spacing
	// Title: ~1 line, Status bar: ~3 lines, spacing: ~4 lines
	usedHeight := 8
	availableHeight := m.height - usedHeight

	// Each brief item takes about 3-4 lines (keyword, date/credits, preview)
	itemHeight := 4
	maxItems := availableHeight / itemHeight

	// Ensure we show at least 5 items if possible
	if maxItems < 5 {
		maxItems = 5
	}

	return maxItems
}
