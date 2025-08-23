package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// PageDetailsModel displays detailed analysis for a specific page from audit history
type PageDetailsModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Page data
	auditID     string
	pageURL     string
	pageDetails *api.PageDetailsResponse
	loading     bool
	error       error

	// Audit data for navigation
	audit        api.AuditViewResponse
	allPages     []api.AuditDetailPageResponse // All pages from the audit for navigation
	selectedPage int                           // Current page index

	// Scrolling
	scrollOffset int // Track scroll position in page details view

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Navigation
	quitting bool
}

// NewPageDetailsModel creates a new page details model
func NewPageDetailsModel(config *Config, auditID, pageURL string, audit api.AuditViewResponse) *PageDetailsModel {
	// Find the index of the current page
	selectedPage := 0
	for i, page := range audit.Pages {
		if page.URL == pageURL {
			selectedPage = i
			break
		}
	}

	return &PageDetailsModel{
		config:       config,
		auditID:      auditID,
		pageURL:      pageURL,
		audit:        audit,
		allPages:     audit.Pages,
		selectedPage: selectedPage,
		loading:      true,
	}
}

// Init implements tea.Model
func (m *PageDetailsModel) Init() tea.Cmd {
	return m.fetchPageDetails()
}

// Update implements tea.Model
func (m *PageDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case PageDetailsMsg:
		m.loading = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.pageDetails = msg.Details
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
func (m *PageDetailsModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.error != nil {
		return m.renderError()
	}

	if m.loading {
		return m.renderLoading()
	}

	return m.renderPageDetails()
}

// renderLoading shows a loading state
func (m *PageDetailsModel) renderLoading() string {
	title := TitleStyle.Render("📄 Page Analysis")
	loading := InfoStatusStyle.Render("🔍 Loading detailed analysis...")

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

// renderPageDetails shows detailed information for the selected page (matching live audit format)
func (m *PageDetailsModel) renderPageDetails() string {
	if m.selectedPage >= len(m.allPages) {
		return AppStyle.Render("Invalid page selection")
	}

	selectedPage := m.allPages[m.selectedPage]

	// Title with URL (matching live audit format)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(InfoColor).
		Render(fmt.Sprintf("📄 Page Analysis: %s", selectedPage.URL))

	// Add page indicator (matching live audit format)
	pageIndicator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true).
		Render(fmt.Sprintf("Page %d of %d", m.selectedPage+1, len(m.allPages)))

	if m.loading {
		return AppStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				title,
				"",
				pageIndicator,
				"",
				InfoStatusStyle.Render("🔍 Loading detailed analysis..."),
				"",
				RenderKeyHelp(map[string]string{
					"←→":  "Navigate pages",
					"Esc": "Back to audit details",
					"q":   "Quit",
				}),
			),
		)
	}

	// Status section (matching live audit format)
	var statusColor lipgloss.Color
	var statusText string
	switch selectedPage.AnalysisStatus {
	case "complete":
		statusColor = SuccessColor
		statusText = "✅ Analysis Complete"
	case "failed":
		statusColor = ErrorColor
		statusText = "❌ Analysis Failed"
	default:
		statusColor = WarningColor
		statusText = "⏳ Still Analyzing"
	}

	status := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true).
		Render(statusText)

	// Page overview section (when details are available)
	var overviewSection string
	if m.pageDetails != nil {
		overviewSection = m.renderPageOverview()
	}

	// Score section (matching live audit format)
	var scoreSection string
	if selectedPage.AnalysisStatus == "completed" {
		score := int(selectedPage.SEOScore)
		scoreColor := ScoreColor(score)
		scoreSection = lipgloss.NewStyle().
			Foreground(scoreColor).
			Bold(true).
			Render(fmt.Sprintf("Overall Score: %d/100", score))
	}

	// Detailed issues section
	issuesSection := m.renderDetailedIssues()

	// Help text - keep this fixed at the bottom (matching live audit)
	help := RenderKeyHelp(map[string]string{
		"↑↓":  "Scroll",
		"←→":  "Navigate pages",
		"e":   "Export AI prompt to clipboard",
		"Esc": "Back to audit details",
		"q":   "Quit",
	})

	// Build the main content (without help text)
	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		pageIndicator,
		"",
		status,
	)

	if overviewSection != "" {
		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			mainContent,
			"",
			overviewSection,
		)
	}

	if scoreSection != "" {
		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			mainContent,
			"",
			scoreSection,
		)
	}

	mainContent = lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		"",
		issuesSection,
	)

	// Apply scrolling to main content only, leaving space for help text (matching live audit)
	availableHeight := m.height - 8 // Account for padding, borders, and help text
	if availableHeight <= 0 {
		availableHeight = 30 // Minimum reasonable height
	}
	scrollableContent := m.renderScrollableContent(mainContent, availableHeight)

	// Combine scrollable content with fixed help text and optional notification
	contentParts := []string{scrollableContent}

	// Add notification if present
	if m.notification != nil {
		contentParts = append(contentParts, "", RenderNotification(*m.notification))
	}

	contentParts = append(contentParts, "", help)

	finalContent := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	return AppStyle.Render(finalContent)
}

// renderPageOverview shows key page information (matching live audit format)
func (m *PageDetailsModel) renderPageOverview() string {
	if m.pageDetails == nil {
		return ""
	}

	details := m.pageDetails.Page

	var overview []string

	// Basic page info
	if details.Title != "" {
		overview = append(overview, fmt.Sprintf("Title: %s", details.Title))
	}
	if details.MetaDescription != "" {
		overview = append(overview, fmt.Sprintf("Meta Description: %s", details.MetaDescription))
	}
	if details.H1 != "" {
		overview = append(overview, fmt.Sprintf("H1: %s", details.H1))
	}
	if details.CanonicalURL != "" {
		overview = append(overview, fmt.Sprintf("Canonical URL: %s", details.CanonicalURL))
	}

	// Content info
	if details.WordCount > 0 {
		overview = append(overview, fmt.Sprintf("Word Count: %d", details.WordCount))
	}

	// Indexability info
	var indexabilityText string
	if details.Indexable {
		indexabilityText = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("✅ Indexable")
	} else {
		indexabilityText = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render("❌ Not Indexable")
	}
	overview = append(overview, fmt.Sprintf("Indexability: %s", indexabilityText))

	if details.IndexabilityReason != "" {
		overview = append(overview, fmt.Sprintf("Reason: %s", details.IndexabilityReason))
	}

	if len(overview) == 0 {
		return ""
	}

	// Render overview section
	overviewTitle := lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Render("📋 Page Overview:")

	overviewContent := ContentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, overview...),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		overviewTitle,
		"",
		overviewContent,
	)
}

// renderDetailedIssues shows SEO checks and issues (matching live audit format)
func (m *PageDetailsModel) renderDetailedIssues() string {
	if m.pageDetails == nil {
		return ContentStyle.Render("No detailed analysis available.")
	}

	if len(m.pageDetails.Page.Checks) == 0 {
		return ContentStyle.Render("No SEO checks performed.")
	}

	var sections []string
	var passedChecks []string
	var failedChecks []string

	// Group checks by pass/fail status and sort by weight (importance)
	for _, check := range m.pageDetails.Page.Checks {
		checkDisplay := fmt.Sprintf("• %s", check.Message)

		if check.Passed {
			passedChecks = append(passedChecks, checkDisplay)
		} else {
			failedChecks = append(failedChecks, checkDisplay)
		}
	}

	// Render failed checks first (issues to fix) - these are already ordered by weight from backend
	if len(failedChecks) > 0 {
		sections = append(sections,
			lipgloss.NewStyle().
				Foreground(ErrorColor).
				Bold(true).
				Render(fmt.Sprintf("🚨 Issues to Fix (%d):", len(failedChecks))),
			"",
		)

		for _, check := range failedChecks {
			sections = append(sections,
				lipgloss.NewStyle().
					Foreground(ErrorColor).
					Render(check),
			)
		}
	}

	// Render passed checks
	if len(passedChecks) > 0 {
		if len(sections) > 0 {
			sections = append(sections, "", "")
		}

		sections = append(sections,
			lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true).
				Render(fmt.Sprintf("✅ Passed Checks (%d):", len(passedChecks))),
			"",
		)

		for _, check := range passedChecks {
			sections = append(sections,
				lipgloss.NewStyle().
					Foreground(SuccessColor).
					Render(check),
			)
		}
	}

	if len(sections) == 0 {
		return ContentStyle.Render("No check results available.")
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderScrollableContent renders content with scroll indicators (matching live audit)
func (m *PageDetailsModel) renderScrollableContent(content string, maxHeight int) string {
	// Ensure maxHeight is at least 1 to prevent slice bounds errors
	if maxHeight <= 0 {
		maxHeight = 1
	}

	lines := strings.Split(content, "\n")

	if len(lines) <= maxHeight {
		return content
	}

	// Apply scroll offset
	start := m.scrollOffset
	if start >= len(lines) {
		start = len(lines) - maxHeight
	}
	if start < 0 {
		start = 0
	}

	end := start + maxHeight
	if end > len(lines) {
		end = len(lines)
	}

	// Ensure we don't have invalid slice bounds
	if start >= len(lines) || end <= start {
		return "Error: Invalid scroll position"
	}

	// Get visible lines
	visibleLines := lines[start:end]
	content = strings.Join(visibleLines, "\n")

	// Add scroll indicators (matching live audit format)
	var indicators []string
	if start > 0 {
		indicators = append(indicators, "↑ More above")
	}
	if end < len(lines) {
		indicators = append(indicators, "↓ More below")
	}

	if len(indicators) > 0 {
		scrollInfo := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render(strings.Join(indicators, " | "))
		content = content + "\n\n" + scrollInfo
	}

	return content
}

// formatURL removes localhost prefix and truncates long URLs
func (m *PageDetailsModel) formatURL(url string) string {
	// Remove localhost prefixes
	url = strings.TrimPrefix(url, m.config.GetEffectiveBaseURL())
	url = strings.TrimPrefix(url, "https://localhost:8000")
	url = strings.TrimPrefix(url, "http://localhost")
	url = strings.TrimPrefix(url, "https://localhost")

	// Truncate if too long
	maxLength := m.width - 20
	if maxLength <= 0 {
		maxLength = 50
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

// renderCompactHelp renders the help text
func (m *PageDetailsModel) renderCompactHelp() string {
	help := map[string]string{
		"↑↓":  "Scroll",
		"←→":  "Navigate pages",
		"e":   "Export AI prompt to clipboard",
		"Esc": "Back to audit details",
		"q":   "Quit",
	}

	return RenderKeyHelp(help)
}

// renderError renders error state
func (m *PageDetailsModel) renderError() string {
	title := TitleStyle.Render("📄 Page Analysis")
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
func (m *PageDetailsModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Return to audit details
		return m, func() tea.Msg {
			return BackToAuditDetailsMsg{
				Audit: m.audit,
			}
		}

	case "up":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "down":
		// Allow scrolling down (scroll bounds are handled by renderScrollableContent)
		m.scrollOffset++
		return m, nil

	case "left":
		if m.selectedPage > 0 {
			// Navigate to previous page
			m.selectedPage--
			m.loading = true
			m.scrollOffset = 0 // Reset scroll position
			m.pageURL = m.allPages[m.selectedPage].URL
			return m, m.fetchPageDetails()
		}
		return m, nil

	case "right":
		if m.selectedPage < len(m.allPages)-1 {
			// Navigate to next page
			m.selectedPage++
			m.loading = true
			m.scrollOffset = 0 // Reset scroll position
			m.pageURL = m.allPages[m.selectedPage].URL
			return m, m.fetchPageDetails()
		}
		return m, nil

	case "e":
		// Export page details to clipboard as AI prompt
		if m.pageDetails != nil {
			return m, ExportPageDetailsToClipboardWithNotification(m.pageDetails)
		}
		return m, nil
	}

	return m, nil
}

// fetchPageDetails fetches the detailed page analysis
func (m *PageDetailsModel) fetchPageDetails() tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return PageDetailsMsg{Error: fmt.Errorf("failed to initialize local audit: %w", err)}
			}
		}

		details, err := localAuditAdapter.GetPageDetails(m.auditID, m.pageURL)
		if err != nil {
			return PageDetailsMsg{Error: fmt.Errorf("failed to fetch page details: %w", err)}
		}

		return PageDetailsMsg{Details: details}
	}
}
