package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/api"
)

// Global local audit adapter instance
var localAuditAdapter *LocalAuditAdapter

// InitializeLocalAuditAdapter initializes the global local audit adapter
func InitializeLocalAuditAdapter() error {
	adapter, err := NewLocalAuditAdapter()
	if err != nil {
		return fmt.Errorf("failed to initialize local audit adapter: %w", err)
	}
	localAuditAdapter = adapter
	return nil
}

// localStartAudit performs the actual audit start with local processing
func (m *SimpleAuditModel) localStartAudit() tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return ErrorMsg{Error: fmt.Errorf("failed to initialize local audit: %w", err)}
			}
		}

		// Start local audit session
		auditID, err := localAuditAdapter.StartAudit(m.config, m.baseURL)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to start local audit: %w", err)}
		}

		return AuditStartedMsg{
			AuditID: auditID,
		}
	}
}

// localPollProgress performs local crawling and progress polling
func (m *SimpleAuditModel) localPollProgress() tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return ErrorMsg{Error: fmt.Errorf("failed to initialize local audit: %w", err)}
			}
		}

		// If we haven't submitted pages yet, do discovery and submission
		if !m.pagesSubmitted {
			return localAuditAdapter.PerformSiteDiscoveryAndSubmit(m.auditID, m.baseURL, m.config)
		}

		// Poll for analysis progress
		statusResp, err := localAuditAdapter.GetAuditStatus(m.auditID)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to get audit status: %w", err)}
		}

		// Convert API response to our message format (same as before)
		pages := make([]PageSummary, len(statusResp.Pages))
		for i, p := range statusResp.Pages {
			pages[i] = PageSummary{
				URL:    p.URL,
				Score:  p.Score,
				Status: p.Status,
				Issues: p.Issues,
			}
		}

		return ProgressUpdateMsg{
			Status:        statusResp.Status,
			PagesFound:    statusResp.Progress.PagesFound,
			PagesAnalyzed: statusResp.Progress.PagesAnalyzed,
			TotalPages:    statusResp.Progress.TotalPages,
			CurrentPage:   statusResp.Progress.CurrentPage,
			Pages:         pages,
		}
	}
}

// localFetchPageDetails fetches detailed analysis for a specific page using local storage
func (m *SimpleAuditModel) localFetchPageDetails(page PageSummary) tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return PageDetailsMsg{Error: fmt.Errorf("failed to initialize local audit: %w", err)}
			}
		}

		// Get page details from local storage
		details, err := localAuditAdapter.GetPageDetails(m.auditID, page.URL)
		if err != nil {
			return PageDetailsMsg{Error: fmt.Errorf("failed to fetch page details: %w", err)}
		}

		return PageDetailsMsg{Details: details}
	}
}

// localCompleteAudit finalizes the audit using local processing
func (m *SimpleAuditModel) localCompleteAudit() tea.Cmd {
	return func() tea.Msg {
		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return AuditCompletedMsg{Summary: "Audit completed with initialization error"}
			}
		}

		completionResp, err := localAuditAdapter.CompleteAudit(m.auditID)
		if err != nil {
			// Log error but don't fail the audit - results are still available
			return AuditCompletedMsg{
				Summary: "Audit completed with errors",
			}
		}

		// Could show summary information here
		_ = completionResp

		return AuditCompletedMsg{
			Summary: "Audit completed successfully",
		}
	}
}

// localExportAllPagesToClipboard fetches all page details and exports them as AI prompt
func (m *SimpleAuditModel) localExportAllPagesToClipboard() tea.Cmd {
	return func() tea.Msg {
		LogInfo("Starting export of %d pages from local audit %s", len(m.pages), m.auditID)

		// Ensure local audit adapter is initialized
		if localAuditAdapter == nil {
			if err := InitializeLocalAuditAdapter(); err != nil {
				return NotificationMsg{
					Message: "Failed to initialize local audit system",
					Type:    NotificationError,
				}
			}
		}

		var pageDetails []api.PageDetailsResponse
		var errors []string

		// Fetch details for each page from local storage
		for i, page := range m.pages {
			LogDebug("Fetching local details for page %d/%d: %s", i+1, len(m.pages), page.URL)
			details, err := localAuditAdapter.GetPageDetails(m.auditID, page.URL)

			if err != nil {
				// Log the error but continue with other pages
				errorMsg := fmt.Sprintf("Failed to fetch details for %s: %v", page.URL, err)
				errors = append(errors, errorMsg)
				LogError("Local page details fetch failed: %s", errorMsg)
				continue
			}

			// Count actual failed checks
			failedChecks := 0
			for _, check := range details.Page.Checks {
				if !check.Passed {
					failedChecks++
				}
			}

			LogDebug("Successfully fetched local details for %s: issues_count=%d, actual failed checks=%d, total checks=%d, score %.1f",
				page.URL, details.Page.IssuesCount, failedChecks, len(details.Page.Checks), details.Page.SEOScore)

			pageDetails = append(pageDetails, *details)
		}

		// Export to clipboard
		if len(pageDetails) > 0 {
			LogInfo("Exporting %d pages with detailed local analysis", len(pageDetails))
			err := ExportMultiplePagesToClipboard(pageDetails)
			LogExport("local audit with details", len(pageDetails), err)

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
			LogError("No local page details could be fetched, creating fallback export")
			fallbackPrompt := fmt.Sprintf("Failed to fetch detailed page analysis from local storage. Errors:\n%s\n\nBasic page summary:\n", 
				fmt.Sprintf("%v", errors))
			for i, page := range m.pages {
				fallbackPrompt += fmt.Sprintf("%d. %s - Score: %d, Issues: %d\n", i+1, page.URL, page.Score, page.Issues)
			}
			err := ExportToClipboard(fallbackPrompt)
			LogExport("local audit fallback", len(m.pages), err)

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