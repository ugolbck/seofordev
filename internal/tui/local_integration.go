package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/api"
)


// Global local audit adapter instance
var auditAdapter *LocalAuditAdapter

// Global config reference for audit functionality
var globalConfig *Config


// InitializeAuditAdapter initializes the global audit adapter with the current config
func InitializeAuditAdapter(config *Config) error {
	globalConfig = config
	adapter, err := NewLocalAuditAdapter()
	if err != nil {
		return fmt.Errorf("failed to initialize audit adapter: %w", err)
	}
	auditAdapter = adapter
	return nil
}

// GetCurrentConfig returns the current global config
func GetCurrentConfig() *Config {
	return globalConfig
}

// localStartAudit performs the actual audit start with local processing
func (m *SimpleAuditModel) localStartAudit() tea.Cmd {
	return func() tea.Msg {
		
		// Ensure audit adapter is initialized
		if auditAdapter == nil {
			return ErrorMsg{Error: fmt.Errorf("audit adapter is nil - restart the application")}
		}
		
		if GetCurrentConfig() == nil {
			return ErrorMsg{Error: fmt.Errorf("global config is nil - restart the application")}
		}
		

		// Start local audit session
		auditID, err := auditAdapter.StartAudit(m.config, m.baseURL)
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
		// Ensure audit adapter is initialized
		if auditAdapter == nil || GetCurrentConfig() == nil {
			// We need the main config, but we only have AuditConfig here
			// This should be initialized from the main menu before audits start
			return ErrorMsg{Error: fmt.Errorf("audit system not properly initialized - restart the application")}
		}

		// If we haven't submitted pages yet, do discovery and submission
		if !m.pagesSubmitted {
			return auditAdapter.PerformSiteDiscoveryAndSubmit(m.auditID, m.baseURL, m.config)
		}

		// Poll for analysis progress
		statusResp, err := auditAdapter.GetAuditStatus(m.auditID)
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
		// Ensure audit adapter is initialized
		if auditAdapter == nil || GetCurrentConfig() == nil {
			return PageDetailsMsg{Error: fmt.Errorf("audit system not properly initialized - restart the application")}
		}

		// Get page details from local storage
		details, err := auditAdapter.GetPageDetails(m.auditID, page.URL)
		if err != nil {
			return PageDetailsMsg{Error: fmt.Errorf("failed to fetch page details: %w", err)}
		}

		return PageDetailsMsg{Details: details}
	}
}

// localCompleteAudit finalizes the audit using local processing
func (m *SimpleAuditModel) localCompleteAudit() tea.Cmd {
	return func() tea.Msg {
		// Ensure audit adapter is initialized
		if auditAdapter == nil || GetCurrentConfig() == nil {
			return AuditCompletedMsg{Summary: "Audit completed with initialization error"}
		}

		completionResp, err := auditAdapter.CompleteAudit(m.auditID)
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

		// Ensure audit adapter is initialized
		if auditAdapter == nil || GetCurrentConfig() == nil {
			return NotificationMsg{
				Message: "Audit system not properly initialized - restart the application",
				Type:    NotificationError,
			}
		}

		var pageDetails []api.PageDetailsResponse
		var errors []string

		// Fetch details for each page from local storage
		for _, page := range m.pages {
			details, err := auditAdapter.GetPageDetails(m.auditID, page.URL)

			if err != nil {
				// Log the error but continue with other pages
				errorMsg := fmt.Sprintf("Failed to fetch details for %s: %v", page.URL, err)
				errors = append(errors, errorMsg)
				continue
			}

			// Count actual failed checks
			failedChecks := 0
			for _, check := range details.Page.Checks {
				if !check.Passed {
					failedChecks++
				}
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
			fallbackPrompt := fmt.Sprintf("Failed to fetch detailed page analysis from local storage. Errors:\n%s\n\nBasic page summary:\n", 
				fmt.Sprintf("%v", errors))
			for i, page := range m.pages {
				fallbackPrompt += fmt.Sprintf("%d. %s - Score: %d, Issues: %d\n", i+1, page.URL, page.Score, page.Issues)
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