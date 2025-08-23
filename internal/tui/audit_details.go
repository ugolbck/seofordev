package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// AuditDetailsModel displays the pages for a selected audit
type AuditDetailsModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Audit data
	audit        api.AuditViewResponse
	selectedPage int
	loading      bool
	error        error

	// Scrolling
	listScrollOffset int // Track scroll position in page list

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Navigation
	quitting bool
}

// NewAuditDetailsModel creates a new audit details model
func NewAuditDetailsModel(config *Config, audit api.AuditViewResponse) *AuditDetailsModel {
	return &AuditDetailsModel{
		config:           config,
		audit:            audit,
		selectedPage:     0,
		loading:          false,
		listScrollOffset: 0,
	}
}

// Init implements tea.Model
func (m *AuditDetailsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *AuditDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

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
func (m *AuditDetailsModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.error != nil {
		return m.renderError()
	}

	if m.loading {
		return m.renderLoading()
	}

	if len(m.audit.Pages) == 0 {
		return m.renderEmpty()
	}

	return m.renderPageList()
}

// renderLoading shows a loading state
func (m *AuditDetailsModel) renderLoading() string {
	title := TitleStyle.Render("📄 Audit Details")
	loading := InfoStatusStyle.Render("🔍 Loading pages...")

	// Build content with optional notification
	content := []string{
		title,
		"",
		loading,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", m.renderCompactHelp())

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderEmpty shows when there are no pages
func (m *AuditDetailsModel) renderEmpty() string {
	title := TitleStyle.Render("📄 Audit Details")
	empty := ContentStyle.Render("No pages found for this audit.")

	// Build content with optional notification
	content := []string{
		title,
		"",
		m.renderAuditSummary(),
		"",
		empty,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", m.renderCompactHelp())

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderPageList renders the main page list view
func (m *AuditDetailsModel) renderPageList() string {
	title := TitleStyle.Render("📄 Audit Details")

	// Build content with optional notification
	content := []string{
		title,
		"",
		m.renderAuditSummary(),
		"",
		m.renderPagesList(),
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", m.renderCompactHelp())

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderAuditSummary shows audit information
func (m *AuditDetailsModel) renderAuditSummary() string {
	createdAt, _ := time.Parse(time.RFC3339, m.audit.CreatedAt)
	dateStr := createdAt.Format("Jan 2, 2006 at 3:04 PM")

	summary := lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("Audit Information"),
		ContentStyle.Render(fmt.Sprintf("Created: %s", dateStr)),
		ContentStyle.Render(fmt.Sprintf("Status: %s", m.audit.Status)),
		ContentStyle.Render(fmt.Sprintf("Overall Score: %.1f/100", m.audit.OverallScore)),
		ContentStyle.Render(fmt.Sprintf("Pages Analyzed: %d", len(m.audit.Pages))),
	)

	return BoxStyle.Render(summary)
}

// renderPagesList renders the list of pages
func (m *AuditDetailsModel) renderPagesList() string {
	title := SubtitleStyle.Render("Pages")

	maxVisible := m.getMaxVisibleItems()
	start := m.listScrollOffset
	end := start + maxVisible
	if end > len(m.audit.Pages) {
		end = len(m.audit.Pages)
	}

	var items []string
	for i := start; i < end; i++ {
		page := m.audit.Pages[i]
		isSelected := i == m.selectedPage
		items = append(items, m.renderPageItem(i, page, isSelected))
	}

	// Add scroll indicators
	var scrollIndicators []string
	if start > 0 {
		scrollIndicators = append(scrollIndicators, InfoStatusStyle.Render("↑ More pages above"))
	}
	if end < len(m.audit.Pages) {
		scrollIndicators = append(scrollIndicators, InfoStatusStyle.Render("↓ More pages below"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	if len(scrollIndicators) > 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", lipgloss.JoinVertical(lipgloss.Left, scrollIndicators...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

// renderPageItem renders a single page item
func (m *AuditDetailsModel) renderPageItem(index int, page api.AuditDetailPageResponse, isSelected bool) string {
	// Format URL to remove localhost prefix and truncate if needed
	displayURL := m.formatURL(page.URL)

	// Format score with color coding
	var scoreStr string
	if page.SEOScore >= 80 {
		scoreStr = fmt.Sprintf("%.0f", page.SEOScore)
	} else if page.SEOScore >= 60 {
		scoreStr = fmt.Sprintf("%.0f", page.SEOScore)
	} else {
		scoreStr = fmt.Sprintf("%.0f", page.SEOScore)
	}

	// Format status icon
	statusIcon := m.getStatusIcon(page.AnalysisStatus)

	// Create the line in the same format as live audit: "✅ /about - Score: 95/100"
	line := fmt.Sprintf("%s %s - Score: %s/100", statusIcon, displayURL, scoreStr)

	if isSelected {
		return SelectedItemStyle.Render(line)
	}
	return ListItemStyle.Render(line)
}

// formatURL removes localhost prefix and truncates long URLs
func (m *AuditDetailsModel) formatURL(url string) string {
	// Remove localhost prefixes
	url = strings.TrimPrefix(url, m.config.GetEffectiveBaseURL())
	url = strings.TrimPrefix(url, "https://localhost:8000")
	url = strings.TrimPrefix(url, "http://localhost")
	url = strings.TrimPrefix(url, "https://localhost")

	// Truncate if too long (leave space for indicators)
	maxLength := m.width - 20
	if maxLength <= 0 {
		maxLength = 50 // Default minimum width
	}
	if len(url) > maxLength {
		if maxLength <= 3 {
			url = "..."
		} else {
			url = url[:maxLength-3] + "..."
		}
	}

	return url
}

// getStatusIcon returns an icon for the page status
func (m *AuditDetailsModel) getStatusIcon(status string) string {
	switch status {
	case "complete":
		return "✅"
	case "analyzing":
		return "🔄"
	case "pending":
		return "⏳"
	case "error":
		return "❌"
	default:
		return "❓"
	}
}

// renderCompactHelp renders the help text
func (m *AuditDetailsModel) renderCompactHelp() string {
	help := map[string]string{
		"↑↓":    "Navigate pages",
		"Enter": "View page details",
		"e":     "Export AI prompt to clipboard",
		"Esc":   "Back to audit menu",
		"q":     "Quit",
	}

	return RenderKeyHelp(help)
}

// renderError renders error state
func (m *AuditDetailsModel) renderError() string {
	title := TitleStyle.Render("📄 Audit Details")
	errorMsg := ErrorStatusStyle.Render(fmt.Sprintf("Error: %v", m.error))

	// Build content with optional notification
	content := []string{
		title,
		"",
		errorMsg,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", m.renderCompactHelp())

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// handleKeypress handles keyboard input
func (m *AuditDetailsModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Return to audit menu
		return m, func() tea.Msg {
			return BackToAuditMenuMsg{}
		}

	case "up":
		if m.selectedPage > 0 {
			m.selectedPage--
			// Adjust scroll if needed
			if m.selectedPage < m.listScrollOffset {
				m.listScrollOffset = m.selectedPage
			}
		}
		return m, nil

	case "down":
		if m.selectedPage < len(m.audit.Pages)-1 {
			m.selectedPage++
			// Adjust scroll if needed
			maxVisible := m.getMaxVisibleItems()
			if m.selectedPage >= m.listScrollOffset+maxVisible {
				m.listScrollOffset = m.selectedPage - maxVisible + 1
			}
		}
		return m, nil

	case "enter":
		if len(m.audit.Pages) > 0 && m.selectedPage < len(m.audit.Pages) {
			selectedPage := m.audit.Pages[m.selectedPage]
			// Navigate to page details view
			return m, func() tea.Msg {
				return NavigateToPageDetailsMsg{
					AuditID: m.audit.ID,
					PageURL: selectedPage.URL,
					Audit:   m.audit,
				}
			}
		}
		return m, nil

	case "e":
		// Export audit with all page details to clipboard as AI prompt
		return m, m.exportAuditWithDetails()
	}

	return m, nil
}

// getMaxVisibleItems calculates how many items can be displayed
func (m *AuditDetailsModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, and spacing
	// Title: ~1 line, Summary: ~6 lines, Help: ~3 lines, spacing: ~4 lines
	usedHeight := 14
	availableHeight := m.height - usedHeight

	// Each page item takes 1 line
	itemHeight := 1
	maxItems := availableHeight / itemHeight

	// Ensure we show at least 12-15 items if possible
	if maxItems < 12 {
		maxItems = 12
	}

	return maxItems
}

// Message types for navigation
type NavigateToPageDetailsMsg struct {
	AuditID string
	PageURL string
	Audit   api.AuditViewResponse
}

// exportAuditWithDetails fetches all page details and exports them as AI prompt
func (m *AuditDetailsModel) exportAuditWithDetails() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		var pageDetails []api.PageDetailsResponse
		var errors []string

		// Fetch details for each page
		for _, page := range m.audit.Pages {
			details, err := client.GetPageDetails(m.audit.ID, page.URL)
			if err != nil {
				// Log the error but continue with other pages
				errors = append(errors, fmt.Sprintf("Failed to fetch details for %s: %v", page.URL, err))
				continue
			}
			pageDetails = append(pageDetails, *details)
		}

		// Export to clipboard
		if len(pageDetails) > 0 {
			err := ExportMultiplePagesToClipboard(pageDetails)
			if err != nil {
				return NotificationMsg{
					Message: "Failed to copy AI prompt to clipboard",
					Type:    NotificationError,
				}
			}

			message := fmt.Sprintf("AI prompt for %d page(s) copied to clipboard successfully!", len(pageDetails))
			return NotificationMsg{
				Message: message,
				Type:    NotificationSuccess,
			}
		} else {
			// If no page details were fetched, create a fallback export with basic info
			fallbackPrompt := fmt.Sprintf("Failed to fetch detailed page analysis. API errors:\n%s\n\nBasic audit summary:\n", strings.Join(errors, "\n"))
			fallbackPrompt += fmt.Sprintf("Audit ID: %s\nOverall Score: %.1f/100\nTotal Pages: %d\n\n", m.audit.ID, m.audit.OverallScore, len(m.audit.Pages))
			for i, page := range m.audit.Pages {
				fallbackPrompt += fmt.Sprintf("%d. %s - Score: %.1f, Issues: %d\n", i+1, page.URL, page.SEOScore, page.IssuesCount)
			}
			err := ExportToClipboard(fallbackPrompt)
			if err != nil {
				return NotificationMsg{
					Message: "Failed to copy AI prompt to clipboard",
					Type:    NotificationError,
				}
			}

			return NotificationMsg{
				Message: "AI prompt (basic summary) copied to clipboard successfully!",
				Type:    NotificationSuccess,
			}
		}
	}
}
