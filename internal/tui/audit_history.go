package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// AuditHistoryModel displays the list of user's audits
type AuditHistoryModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Audit data
	audits        []api.AuditViewResponse
	selectedAudit int
	loading       bool
	error         error

	// Scrolling
	listScrollOffset int // Track scroll position in audit list

	// Navigation
	quitting bool
}

// NewAuditHistoryModel creates a new audit history model
func NewAuditHistoryModel(config *Config) *AuditHistoryModel {
	return &AuditHistoryModel{
		config:        config,
		selectedAudit: 0,
		loading:       true,
		audits:        []api.AuditViewResponse{},
	}
}

// Init implements tea.Model
func (m *AuditHistoryModel) Init() tea.Cmd {
	return m.fetchAuditHistory()
}

// Update implements tea.Model
func (m *AuditHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case AuditHistoryLoadedMsg:
		m.loading = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.audits = msg.Audits
			// Start at the top (most recent audit) and reset scroll
			m.selectedAudit = 0
			m.listScrollOffset = 0
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *AuditHistoryModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.error != nil {
		return m.renderError()
	}

	if m.loading {
		return m.renderLoading()
	}

	if len(m.audits) == 0 {
		return m.renderEmpty()
	}

	return m.renderAuditList()
}

// renderLoading shows a loading state
func (m *AuditHistoryModel) renderLoading() string {
	title := TitleStyle.Render("📊 Audit History")
	loading := InfoStatusStyle.Render("🔍 Loading your audit history...")
	help := RenderKeyHelp(map[string]string{"Esc": "Back to menu"})

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			loading,
			"",
			help,
		),
	)
}

// renderEmpty shows when no audits are found
func (m *AuditHistoryModel) renderEmpty() string {
	title := TitleStyle.Render("📊 Audit History")
	empty := ContentStyle.Render("No audits found. Run your first audit to see it here!")
	help := RenderKeyHelp(map[string]string{"Esc": "Back to menu"})

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			empty,
			"",
			help,
		),
	)
}

// renderAuditList shows the list of audits
func (m *AuditHistoryModel) renderAuditList() string {
	title := TitleStyle.Render("📊 Audit History")

	// Summary
	summary := m.renderSummary()

	// Audit list
	auditList := m.renderAuditItems()

	// Help
	help := RenderKeyHelp(map[string]string{
		"↑↓":    "Navigate/Scroll",
		"Enter": "View Details",
		"Esc":   "Back to menu",
	})

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			summary,
			"",
			auditList,
			"",
			help,
		),
	)
}

// renderSummary shows audit statistics
func (m *AuditHistoryModel) renderSummary() string {
	if len(m.audits) == 0 {
		return ""
	}

	// Calculate statistics
	var totalScore float64
	var completedCount int

	for _, audit := range m.audits {
		if audit.Status == "completed" {
			totalScore += audit.OverallScore
			completedCount++
		}
	}

	var avgScore float64
	if completedCount > 0 {
		avgScore = totalScore / float64(completedCount)
	}

	stats := []string{
		fmt.Sprintf("Total Audits: %d", len(m.audits)),
		fmt.Sprintf("Completed: %d", completedCount),
		fmt.Sprintf("Average Score: %.1f/100", avgScore),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("📈 Summary:"),
		"",
		ContentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, stats...)),
	)
}

// renderAuditItems renders the list of audit items
func (m *AuditHistoryModel) renderAuditItems() string {
	if len(m.audits) == 0 {
		return ""
	}

	maxVisible := m.getMaxVisibleItems()

	// Calculate visible range
	start := m.listScrollOffset
	end := start + maxVisible
	if end > len(m.audits) {
		end = len(m.audits)
	}

	// Ensure start is valid
	if start >= len(m.audits) {
		start = 0
		m.listScrollOffset = 0
	}

	var lines []string

	// Add scroll indicator if there are items above
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↑ More audits above")
		lines = append(lines, scrollUp)
	}

	// Add visible audit items
	for i := start; i < end; i++ {
		line := m.renderAuditItem(i, m.audits[i])
		lines = append(lines, line)
	}

	// Add scroll indicator if there are items below
	if end < len(m.audits) {
		scrollDown := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↓ More audits below")
		lines = append(lines, scrollDown)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("📋 Your Audits:"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// renderAuditItem renders a single audit item
func (m *AuditHistoryModel) renderAuditItem(index int, audit api.AuditViewResponse) string {
	// Parse the creation date
	createdAt, err := time.Parse(time.RFC3339, audit.CreatedAt)
	var dateStr string
	if err != nil {
		dateStr = audit.CreatedAt // Fallback to raw string
	} else {
		dateStr = createdAt.Format("Jan 02, 2006 15:04")
	}

	// Format the score
	var scoreDisplay string
	if audit.Status == "completed" {
		scoreDisplay = fmt.Sprintf("%.1f/100", audit.OverallScore)
	} else {
		scoreDisplay = "N/A"
	}

	// Status icon
	statusIcon := m.getStatusIcon(audit.Status)

	// Build the line as plain text first
	line := fmt.Sprintf("%s %s - %s",
		statusIcon,
		scoreDisplay,
		dateStr,
	)

	// Apply selection styling to the complete line
	style := ListItemStyle
	if index == m.selectedAudit {
		style = SelectedItemStyle
	}

	return style.Render(line)
}

// getStatusIcon returns the appropriate icon for the audit status
func (m *AuditHistoryModel) getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "running":
		return "⚡"
	case "pending":
		return "⏳"
	default:
		return "❓"
	}
}

// renderError shows error state
func (m *AuditHistoryModel) renderError() string {
	title := ErrorStatusStyle.Render("❌ Failed to Load Audit History")
	message := ContentStyle.Render(m.error.Error())
	help := RenderKeyHelp(map[string]string{"Esc": "Back to menu"})

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			message,
			"",
			help,
		),
	)
}

// handleKeypress handles keyboard input
func (m *AuditHistoryModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, func() tea.Msg { return BackToAuditMenuMsg{} }

	case "up", "k":
		if m.selectedAudit > 0 {
			m.selectedAudit--
			// Adjust scroll offset to keep selected item visible
			if m.selectedAudit < m.listScrollOffset {
				m.listScrollOffset = m.selectedAudit
			}
		}
		return m, nil

	case "down", "j":
		if m.selectedAudit < len(m.audits)-1 {
			m.selectedAudit++
			// Adjust scroll offset to keep selected item visible
			maxVisible := m.listScrollOffset + m.getMaxVisibleItems() - 1
			if m.selectedAudit > maxVisible {
				m.listScrollOffset = m.selectedAudit - m.getMaxVisibleItems() + 1
			}
		}
		return m, nil

	case "enter":
		if len(m.audits) > 0 && m.selectedAudit < len(m.audits) {
			selectedAudit := m.audits[m.selectedAudit]
			// Navigate to audit details view
			return m, func() tea.Msg {
				return NavigateToAuditDetailsMsg{
					Audit: selectedAudit,
				}
			}
		}
		return m, nil

	}

	return m, nil
}

// fetchAuditHistory fetches the audit history from local storage
func (m *AuditHistoryModel) fetchAuditHistory() tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return AuditHistoryLoadedMsg{Error: fmt.Errorf("failed to initialize local audit: %w", err)}
			}
		}

		// Get audits from local storage
		localAudits, err := localAuditAdapter.ListAudits()
		if err != nil {
			return AuditHistoryLoadedMsg{Error: fmt.Errorf("failed to fetch local audit history: %w", err)}
		}

		// Convert local audits to API format for compatibility
		apiAudits := make([]api.AuditViewResponse, len(localAudits))
		for i, localAudit := range localAudits {
			// Convert pages to API format
			pages := make([]api.AuditDetailPageResponse, len(localAudit.Pages))
			for j, page := range localAudit.Pages {
				score := 0.0
				if page.SEOScore != nil {
					score = *page.SEOScore
				}
				
				var analyzedAt *string
				if page.AnalyzedAt != nil {
					timeStr := page.AnalyzedAt.Format(time.RFC3339)
					analyzedAt = &timeStr
				}

				pages[j] = api.AuditDetailPageResponse{
					ID:             page.ID,
					URL:            page.URL,
					AnalysisStatus: page.AnalysisStatus,
					SEOScore:       score,
					AnalyzedAt:     analyzedAt,
					IssuesCount:    page.IssuesCount,
				}
			}

			// Calculate overall score
			overallScore := 0.0
			if localAudit.OverallScore != nil {
				overallScore = *localAudit.OverallScore
			} else if localAudit.AvgPageScore != nil {
				overallScore = *localAudit.AvgPageScore
			}

			apiAudits[i] = api.AuditViewResponse{
				ID:           localAudit.ID,
				CreatedAt:    localAudit.CreatedAt.Format(time.RFC3339),
				Status:       localAudit.Status,
				OverallScore: overallScore,
				Pages:        pages,
			}
		}

		return AuditHistoryLoadedMsg{Audits: apiAudits}
	}
}

// Message types
type AuditHistoryLoadedMsg struct {
	Audits []api.AuditViewResponse
	Error  error
}

type NavigateToAuditDetailsMsg struct {
	Audit api.AuditViewResponse
}

// getMaxVisibleItems calculates how many audit items can fit on screen
func (m *AuditHistoryModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, spacing, and potential scroll indicators
	// Title: ~1 line, Summary: ~5 lines, Help: ~3 lines, spacing: ~4 lines
	// Scroll indicators: ~2 lines (if needed)
	// Leave some buffer for safety
	availableHeight := m.height - 18
	if availableHeight <= 0 {
		availableHeight = 8 // Minimum reasonable height
	}
	return availableHeight
}
