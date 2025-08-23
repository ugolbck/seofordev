package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// AuditState represents the current state of the audit
type AuditState int

const (
	StateSetup AuditState = iota
	StateActive
	StateResults
	StateDetails // New state for showing page details
)

// SimpleAuditModel is a clean, user-focused audit interface
type SimpleAuditModel struct {
	width  int
	height int

	// Configuration
	config AuditConfig

	// Current state
	state AuditState

	// Audit session
	auditID string
	baseURL string

	// Progress tracking
	pagesFound     int
	pagesAnalyzed  int
	totalPages     int
	currentPage    string
	startTime      time.Time
	pagesSubmitted bool // Track if pages have been submitted to avoid duplicate submissions

	// Results
	pages        []PageSummary
	selectedPage int

	// Scrolling
	listScrollOffset int // Track scroll position in page list

	// Page details state
	currentPageDetails  *api.PageDetailsResponse
	loadingPageDetails  bool
	detailsScrollOffset int // Track scroll position in page details view

	// Timeout handling
	stuckPages map[string]time.Time

	// UI state
	spinner       int
	spinnerFrames []string

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Error handling
	error    error
	quitting bool
}

// PageSummary represents a page in the results list
type PageSummary struct {
	URL    string `json:"url"`
	Score  int    `json:"score"`
	Status string `json:"status"` // "pending", "analyzing", "complete"
	Issues int    `json:"issues"`
}

// NewSimpleAuditModel creates a new simplified audit model
func NewSimpleAuditModel(config AuditConfig) *SimpleAuditModel {
	baseURL := fmt.Sprintf("http://localhost:%d", config.Port)

	return &SimpleAuditModel{
		config:        config,
		state:         StateSetup,
		baseURL:       baseURL,
		pages:         []PageSummary{},
		selectedPage:  0,
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stuckPages:    make(map[string]time.Time),
	}
}

// NewSimpleAuditModelWithURL creates a new simplified audit model with a specific URL
func NewSimpleAuditModelWithURL(config AuditConfig, specificURL string) *SimpleAuditModel {
	return &SimpleAuditModel{
		config:        config,
		state:         StateSetup,
		baseURL:       specificURL,
		pages:         []PageSummary{},
		selectedPage:  0,
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stuckPages:    make(map[string]time.Time),
	}
}

// Init implements tea.Model
func (m *SimpleAuditModel) Init() tea.Cmd {
	// Fetch credits if we have an API key
	var cmds []tea.Cmd
	cmds = append(cmds, m.startAudit())
	cmds = append(cmds, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	}))

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *SimpleAuditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		LogUIEvent("SimpleAudit", "KeyPress", fmt.Sprintf("key=%s state=%d", msg.String(), m.state))
		return m.handleKeypress(msg)

	case TickMsg:
		return m.handleTick()

	case AuditStartedMsg:
		m.auditID = msg.AuditID
		m.state = StateActive
		m.startTime = time.Now()
		return m, m.pollProgress()

	case ProgressUpdateMsg:
		m.pagesFound = msg.PagesFound
		m.pagesAnalyzed = msg.PagesAnalyzed
		m.totalPages = msg.TotalPages
		m.currentPage = msg.CurrentPage
		m.pages = msg.Pages

		// Reset scroll position when pages are updated
		m.listScrollOffset = 0
		m.selectedPage = 0

		// Mark pages as submitted if this is the first update after discovery
		if msg.PagesFound > 0 && !m.pagesSubmitted {
			m.pagesSubmitted = true
		}

		// Check if audit is complete (all pages processed - either complete or failed)
		if m.isAuditComplete() {
			m.state = StateResults
			// Complete the audit session with local processing
			return m, m.localCompleteAudit()
		}

		// Check for timeout (5 minutes max)
		if time.Since(m.startTime) > 5*time.Minute {
			return m, func() tea.Msg {
				return ErrorMsg{Error: fmt.Errorf("audit timed out after 5 minutes - this may be a backend issue")}
			}
		}

		// Continue polling for progress updates
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return tea.Batch(m.pollProgress())()
		})

	case AuditCompletedMsg:
		// Audit completion is final - no more polling
		return m, nil

	case PageDetailMsg:
		// Show detailed page view (simplified for now)
		page := msg.Page
		detailText := fmt.Sprintf("📄 %s\n\nStatus: %s", page.URL, page.Status)
		if page.Status == "complete" {
			detailText += fmt.Sprintf("\nScore: %d/100\nIssues: %d", page.Score, page.Issues)
		}
		// For now just display info, later could show full details
		return m, func() tea.Msg {
			return tea.Println(detailText)
		}

	case PageDetailsMsg:
		m.loadingPageDetails = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.currentPageDetails = msg.Details
		}
		return m, nil

	case ErrorMsg:
		m.error = msg.Error
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
func (m *SimpleAuditModel) View() string {
	if m.error != nil {
		return m.renderError()
	}

	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	switch m.state {
	case StateSetup:
		return m.renderSetup()
	case StateActive:
		return m.renderActive()
	case StateResults:
		return m.renderResults()
	case StateDetails:
		return m.renderPageDetails()
	default:
		return "Loading..."
	}
}

// renderSetup shows the initial audit parameters
func (m *SimpleAuditModel) renderSetup() string {
	title := TitleStyle.Render("🔍 Starting SEO Audit")

	// Audit parameters
	params := []string{
		fmt.Sprintf("Site: %s", lipgloss.NewStyle().Foreground(SecondaryColor).Render(m.baseURL)),
		fmt.Sprintf("Max Pages: %s", m.formatLimit(m.config.MaxPages)),
		fmt.Sprintf("Max Depth: %s", m.formatLimit(m.config.MaxDepth)),
	}

	if len(m.config.IgnorePatterns) > 0 {
		params = append(params, fmt.Sprintf("Ignoring: %v", m.config.IgnorePatterns))
	}

	paramsList := ContentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, params...))

	status := InfoStatusStyle.Render("Initializing...")

	// Build content with optional notification
	content := []string{
		title,
		"",
		paramsList,
		"",
		status,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			content...,
		),
	)
}

// renderActive shows crawling and analysis progress
func (m *SimpleAuditModel) renderActive() string {
	title := TitleStyle.Render("🔍 SEO Audit in Progress")

	// Site info
	siteInfo := SubtitleStyle.Render(m.baseURL)

	// Progress section
	progressTitle := lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Render("Discovery & Analysis")

	// Stats
	stats := m.renderProgressStats()

	// Current activity
	activity := m.renderCurrentActivity()

	// Recent pages list
	recentPages := m.renderRecentPages()

	var content []string
	content = append(content, title, siteInfo, "", progressTitle, stats, "", activity, "", recentPages)

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

// renderResults shows the interactive results browser
func (m *SimpleAuditModel) renderResults() string {
	if len(m.pages) == 0 {
		// Build content with optional notification
		content := []string{
			ErrorStatusStyle.Render("❌ No Pages Found"),
			"",
			"No pages were discovered during the audit.",
			"Check if the target site is accessible.",
		}

		// Add notification if present
		if m.notification != nil {
			content = append(content, "", RenderNotification(*m.notification))
		}

		content = append(content, "", RenderKeyHelp(map[string]string{"q": "Quit"}))

		return AppStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left, content...),
		)
	}

	title := SuccessStatusStyle.Render("✅ Audit Complete!")

	// Summary stats
	avgScore := m.calculateAverageScore()
	completed, failed, _ := m.getAuditSummary()

	var summary string
	if failed > 0 {
		summary = fmt.Sprintf("%d completed, %d failed | Average Score: %d/100",
			completed, failed, avgScore)
	} else {
		summary = fmt.Sprintf("Analyzed %d pages | Average Score: %d/100",
			completed, avgScore)
	}
	summaryText := SubtitleStyle.Render(summary)

	// Page list
	pagesList := m.renderPagesList()

	// Help section
	help := RenderKeyHelp(map[string]string{
		"↑↓":    "Navigate pages",
		"Enter": "View details",
		"e":     "Export AI prompt to clipboard",
		"Esc":   "Back to menu",
		"q":     "Quit",
	})

	// Build content with optional notification
	content := []string{
		title,
		"",
		summaryText,
		"",
		pagesList,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", help)

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

// renderProgressStats shows discovery and analysis statistics
func (m *SimpleAuditModel) renderProgressStats() string {
	discoveredText := lipgloss.NewStyle().
		Foreground(InfoColor).
		Render(fmt.Sprintf("Pages Found: %d", m.pagesFound))

	completed, failed, _ := m.getAuditSummary()
	processed := completed + failed

	var analyzedText string
	if failed > 0 {
		analyzedText = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render(fmt.Sprintf("Processed: %d (%d✅ %d❌)", processed, completed, failed))
	} else {
		analyzedText = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render(fmt.Sprintf("Analyzed: %d", processed))
	}

	var progressText string
	if m.totalPages > 0 {
		progressText = lipgloss.NewStyle().
			Foreground(WarningColor).
			Render(fmt.Sprintf("Total: %d", m.totalPages))
	} else {
		progressText = lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("Discovering...")
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		discoveredText, "  ",
		analyzedText, "  ",
		progressText,
	)
}

// renderCurrentActivity shows what's currently happening
func (m *SimpleAuditModel) renderCurrentActivity() string {
	spinnerIcon := m.spinnerFrames[m.spinner]

	completed, failed, analyzing := m.getAuditSummary()

	var activity string
	if analyzing > 0 {
		if m.currentPage != "" {
			activity = fmt.Sprintf("%s Analyzing: %s", spinnerIcon, m.currentPage)
		} else {
			activity = fmt.Sprintf("%s Processing %d remaining pages...", spinnerIcon, analyzing)
		}
	} else if m.totalPages == 0 {
		activity = fmt.Sprintf("%s Discovering pages...", spinnerIcon)
	} else {
		// All pages processed
		activity = fmt.Sprintf("✅ Analysis complete - %d pages processed", completed+failed)
	}

	// Show timeout warning if audit has been running too long
	if analyzing > 0 {
		elapsed := time.Since(m.startTime)
		if elapsed > 2*time.Minute {
			timeoutWarning := lipgloss.NewStyle().
				Foreground(WarningColor).
				Render(fmt.Sprintf("(Running for %s - may be a backend issue)", elapsed.Round(time.Second)))
			activity = fmt.Sprintf("%s %s", activity, timeoutWarning)
		}
	}

	return InfoStatusStyle.Render(activity)
}

// renderRecentPages shows the latest page results
func (m *SimpleAuditModel) renderRecentPages() string {
	if len(m.pages) == 0 {
		return lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("🔍 Waiting for pages...")
	}

	title := lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Render("📄 Page Results:")

	var pageLines []string

	// Show last 8 pages for better visibility
	start := 0
	if len(m.pages) > 8 {
		start = len(m.pages) - 8
	}

	for i := start; i < len(m.pages); i++ {
		page := m.pages[i]
		status := m.getStatusIcon(page.Status)

		var scoreText string
		var statusColor lipgloss.Color

		switch page.Status {
		case "complete":
			scoreColor := ScoreColor(page.Score)
			scoreText = lipgloss.NewStyle().
				Foreground(scoreColor).
				Bold(true).
				Render(fmt.Sprintf("Score: %d", page.Score))
			statusColor = SuccessColor
		case "failed", "error":
			scoreText = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Render("Failed")
			statusColor = ErrorColor
		default:
			scoreText = lipgloss.NewStyle().
				Foreground(WarningColor).
				Render("analyzing...")
			statusColor = WarningColor
		}

		// Format URL to be shorter if needed
		displayURL := m.formatURL(page.URL)

		line := fmt.Sprintf("  %s %s", status,
			lipgloss.NewStyle().Foreground(statusColor).Render(displayURL))

		if scoreText != "" {
			line = fmt.Sprintf("%s\n    %s", line, scoreText)
		}

		pageLines = append(pageLines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, pageLines...)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		content,
	)
}

// renderPagesList shows the interactive pages list for results browsing
func (m *SimpleAuditModel) renderPagesList() string {
	if len(m.pages) == 0 {
		return lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("No pages to display.")
	}

	maxVisible := m.getMaxVisibleItems()
	start := m.listScrollOffset
	end := start + maxVisible
	if end > len(m.pages) {
		end = len(m.pages)
	}

	var lines []string

	for i := start; i < end; i++ {
		page := m.pages[i]
		status := m.getStatusIcon(page.Status)

		var scoreDisplay string
		var lineColor lipgloss.Color

		switch page.Status {
		case "complete":
			scoreColor := ScoreColor(page.Score)
			scoreDisplay = fmt.Sprintf("Score: %d/100", page.Score)
			lineColor = scoreColor
		case "failed", "error":
			scoreDisplay = "Failed"
			lineColor = ErrorColor
		default:
			scoreDisplay = "Processing..."
			lineColor = WarningColor
		}

		// Format URL for display
		displayURL := m.formatURL(page.URL)

		line := fmt.Sprintf("%s %s - %s", status, displayURL, scoreDisplay)

		style := ListItemStyle
		if i == m.selectedPage {
			style = SelectedItemStyle
		}

		styledLine := style.
			Foreground(lineColor).
			Render(line)

		lines = append(lines, styledLine)
	}

	// Add scroll indicators
	var scrollIndicators []string
	if start > 0 {
		scrollIndicators = append(scrollIndicators, InfoStatusStyle.Render("↑ More pages above"))
	}
	if end < len(m.pages) {
		scrollIndicators = append(scrollIndicators, InfoStatusStyle.Render("↓ More pages below"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if len(scrollIndicators) > 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", lipgloss.JoinVertical(lipgloss.Left, scrollIndicators...))
	}

	return content
}

// renderPageDetails shows detailed information for the selected page
func (m *SimpleAuditModel) renderPageDetails() string {
	if m.selectedPage >= len(m.pages) {
		return AppStyle.Render("Invalid page selection")
	}

	selectedPage := m.pages[m.selectedPage]

	// Title with URL
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(InfoColor).
		Render(fmt.Sprintf("📄 Page Analysis: %s", selectedPage.URL))

	// Add page indicator
	pageIndicator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true).
		Render(fmt.Sprintf("Page %d of %d", m.selectedPage+1, len(m.pages)))

	if m.loadingPageDetails {
		return AppStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				title,
				"",
				pageIndicator,
				"",
				InfoStatusStyle.Render("🔍 Loading detailed analysis..."),
				"",
				RenderKeyHelp(map[string]string{"Enter": "Back to List", "q": "Quit"}),
			),
		)
	}

	// Status section
	var statusColor lipgloss.Color
	var statusText string
	switch selectedPage.Status {
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
	if m.currentPageDetails != nil {
		overviewSection = m.renderPageOverview()
	}

	// Score section
	var scoreSection string
	if selectedPage.Status == "complete" {
		var score int
		var scoreColor lipgloss.Color

		// Use detailed score if available, otherwise fall back to summary score
		if m.currentPageDetails != nil {
			score = int(m.currentPageDetails.Page.SEOScore)
		} else {
			score = selectedPage.Score
		}

		scoreColor = ScoreColor(score)
		scoreSection = lipgloss.NewStyle().
			Foreground(scoreColor).
			Bold(true).
			Render(fmt.Sprintf("Overall Score: %d/100", score))
	}

	// Detailed issues section
	issuesSection := m.renderDetailedIssues()

	// Help text - keep this fixed at the bottom
	help := RenderKeyHelp(map[string]string{
		"↑↓":  "Scroll",
		"←→":  "Navigate pages",
		"e":   "Export AI prompt to clipboard",
		"Esc": "Back to List",
		"q":   "Quit",
	})

	LogUIEvent("SimpleAudit", "RenderPageDetails", fmt.Sprintf("page=%d/%d loading=%v", m.selectedPage+1, len(m.pages), m.loadingPageDetails))

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

	// Apply scrolling to main content only, leaving space for help text
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

// Enhanced renderDetailedIssues method
func (m *SimpleAuditModel) renderDetailedIssues() string {
	if m.currentPageDetails == nil {
		return ContentStyle.Render("No detailed analysis available.")
	}

	if len(m.currentPageDetails.Page.Checks) == 0 {
		return ContentStyle.Render("No SEO checks performed.")
	}

	var sections []string
	var passedChecks []string
	var failedChecks []string

	// Group checks by pass/fail status and sort by weight (importance)
	for _, check := range m.currentPageDetails.Page.Checks {
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

// renderPageOverview shows key page information
func (m *SimpleAuditModel) renderPageOverview() string {
	if m.currentPageDetails == nil {
		return ""
	}

	details := m.currentPageDetails.Page

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

// renderScrollableContent renders content with scroll indicators
func (m *SimpleAuditModel) renderScrollableContent(content string, maxHeight int) string {
	// Ensure maxHeight is at least 1 to prevent slice bounds errors
	if maxHeight <= 0 {
		maxHeight = 1
	}

	lines := strings.Split(content, "\n")

	if len(lines) <= maxHeight {
		return content
	}

	// Apply scroll offset
	start := m.detailsScrollOffset
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

	// Add scroll indicators
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

// Helper methods

// formatURL removes redundant localhost prefix for cleaner display
func (m *SimpleAuditModel) formatURL(url string) string {
	// Remove common localhost prefixes
	prefixes := []string{
		"http://localhost:",
		"https://localhost:",
		"http://127.0.0.1:",
		"https://127.0.0.1:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(url, prefix) {
			// Find the port and path
			afterPrefix := url[len(prefix):]
			if idx := strings.Index(afterPrefix, "/"); idx != -1 {
				// Return just the path part
				return afterPrefix[idx:]
			}
			// If no path, return "/"
			return "/"
		}
	}

	// If not a localhost URL, truncate if too long
	if len(url) > 60 {
		return url[:57] + "..."
	}
	return url
}

func (m *SimpleAuditModel) formatLimit(limit int) string {
	if limit == 0 {
		return lipgloss.NewStyle().Foreground(AccentColor).Render("unlimited")
	}
	return lipgloss.NewStyle().Foreground(AccentColor).Render(fmt.Sprintf("%d", limit))
}

func (m *SimpleAuditModel) getStatusIcon(status string) string {
	switch status {
	case "complete":
		return "✅"
	case "analyzing":
		return "⚡"
	case "error", "failed":
		return "❌"
	default:
		return "⏳"
	}
}

func (m *SimpleAuditModel) calculateAverageScore() int {
	if len(m.pages) == 0 {
		return 0
	}

	total := 0
	count := 0
	for _, page := range m.pages {
		if page.Status == "complete" {
			total += page.Score
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / count
}

// getAuditSummary returns completion stats
func (m *SimpleAuditModel) getAuditSummary() (completed, failed, analyzing int) {
	for _, page := range m.pages {
		switch page.Status {
		case "complete":
			completed++
		case "failed", "error":
			failed++
		case "analyzing", "pending":
			analyzing++
		}
	}
	return
}

// getMaxVisibleItems calculates how many items can be displayed
func (m *SimpleAuditModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, and spacing
	// Title: ~1 line, Summary: ~4 lines, Help: ~3 lines, spacing: ~3 lines
	usedHeight := 11
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

// isAuditComplete checks if all pages are either complete or failed
func (m *SimpleAuditModel) isAuditComplete() bool {
	if len(m.pages) == 0 || m.totalPages == 0 {
		return false
	}

	// Check if all pages have finished processing (complete or failed)
	for _, page := range m.pages {
		if page.Status == "analyzing" || page.Status == "pending" {
			return false
		}
	}

	// Also ensure we have the expected number of pages
	return len(m.pages) >= m.totalPages
}

func (m *SimpleAuditModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == StateResults {
			m.quitting = true
			return m, tea.Quit
		}
		// During audit, just quit (could add confirmation)
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.state == StateResults {
			// Return to audit menu
			return m, func() tea.Msg {
				return BackToAuditMenuMsg{}
			}
		} else if m.state == StateDetails {
			// Go back to results list
			m.state = StateResults
			m.currentPageDetails = nil
			m.detailsScrollOffset = 0
			return m, nil
		}
		return m, nil

	case "up":
		if m.state == StateResults && m.selectedPage > 0 {
			m.selectedPage--
			// Adjust scroll if needed
			if m.selectedPage < m.listScrollOffset {
				m.listScrollOffset = m.selectedPage
			}
		} else if m.state == StateDetails {
			// Scroll up in page details view
			if m.detailsScrollOffset > 0 {
				m.detailsScrollOffset--
			}
		}
		return m, nil

	case "down":
		if m.state == StateResults && m.selectedPage < len(m.pages)-1 {
			m.selectedPage++
			// Adjust scroll if needed
			maxVisible := m.getMaxVisibleItems()
			if m.selectedPage >= m.listScrollOffset+maxVisible {
				m.listScrollOffset = m.selectedPage - maxVisible + 1
			}
		} else if m.state == StateDetails {
			// Scroll down in page details view
			m.detailsScrollOffset++
		}
		return m, nil

	case "left":
		if m.state == StateDetails && m.selectedPage > 0 {
			// Navigate to previous page
			m.selectedPage--
			m.loadingPageDetails = true
			m.detailsScrollOffset = 0 // Reset scroll position
			return m, m.localFetchPageDetails(m.pages[m.selectedPage])
		}
		return m, nil

	case "right":
		if m.state == StateDetails && m.selectedPage < len(m.pages)-1 {
			// Navigate to next page
			m.selectedPage++
			m.loadingPageDetails = true
			m.detailsScrollOffset = 0 // Reset scroll position
			return m, m.localFetchPageDetails(m.pages[m.selectedPage])
		}
		return m, nil

	case "enter":
		if m.state == StateResults && len(m.pages) > 0 {
			m.state = StateDetails
			m.loadingPageDetails = true
			m.detailsScrollOffset = 0 // Reset scroll position
			// Fetch detailed page analysis
			return m, m.localFetchPageDetails(m.pages[m.selectedPage])
		}
		return m, nil

	case "e":
		if m.state == StateResults {
			// Export all pages with their details to clipboard as AI prompt
			return m, m.localExportAllPagesToClipboard()
		} else if m.state == StateDetails && m.currentPageDetails != nil {
			// Export current page details to clipboard as AI prompt
			return m, ExportPageDetailsToClipboardWithNotification(m.currentPageDetails)
		}
		return m, nil
	}

	return m, nil
}

func (m *SimpleAuditModel) handleTick() (tea.Model, tea.Cmd) {
	if m.state == StateActive {
		// Update spinner
		m.spinner = (m.spinner + 1) % len(m.spinnerFrames)

		// Only continue ticking if we're still in active state
		return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})
	}

	// Stop ticking when not in active state
	return m, nil
}

func (m *SimpleAuditModel) renderError() string {
	// Generic error rendering - no more credit error handling needed for audits
	title := ErrorStatusStyle.Render("❌ Audit Failed")
	message := ContentStyle.Render(m.error.Error())
	help := HelpStyle.Render("Press ctrl+c to exit")

	// Build content with optional notification
	content := []string{title, message, help}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center, content...),
	)
}

// Local audit methods using local processing

func (m *SimpleAuditModel) startAudit() tea.Cmd {
	return m.localStartAudit()
}

func (m *SimpleAuditModel) pollProgress() tea.Cmd {
	return m.localPollProgress()
}

// Message types for the simplified model
type (
	AuditStartedMsg struct {
		AuditID string
	}

	ProgressUpdateMsg struct {
		Status        string
		PagesFound    int
		PagesAnalyzed int
		TotalPages    int
		CurrentPage   string
		Pages         []PageSummary
	}

	AuditCompletedMsg struct {
		Summary string
	}

	PageDetailMsg struct {
		Page PageSummary
	}

	PageDetailsMsg struct {
		Details *api.PageDetailsResponse
		Error   error
	}
)

// fetchPageDetails is now replaced by localFetchPageDetails in local_integration.go

// exportAllPagesToClipboard is now replaced by localExportAllPagesToClipboard in local_integration.go
