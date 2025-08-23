package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/api"
	"github.com/ugolbck/seofordev/internal/crawler"
)

// realStartAudit performs the actual audit start with API integration
func (m *SimpleAuditModel) realStartAudit() tea.Cmd {
	return func() tea.Msg {
		// Create API client
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		// Start audit session
		startReq := api.StartAuditRequest{
			BaseURL:        m.baseURL,
			MaxPages:       m.config.MaxPages,
			MaxDepth:       m.config.MaxDepth,
			IgnorePatterns: m.config.IgnorePatterns,
		}

		startResp, err := client.StartAudit(startReq)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to start audit: %w", err)}
		}

		return AuditStartedMsg{
			AuditID: startResp.AuditID,
		}
	}
}

// realPollProgress performs actual site crawling and API submission
func (m *SimpleAuditModel) realPollProgress() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		// If we haven't submitted pages yet, do discovery and submission
		if !m.pagesSubmitted {
			return m.performSiteDiscovery()
		}

		// Poll for analysis progress
		statusResp, err := client.GetAuditStatus(m.auditID)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to get audit status: %w", err)}
		}

		// Convert API response to our message format
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

// performSiteDiscovery runs the crawler and submits pages to the backend
func (m *SimpleAuditModel) performSiteDiscovery() tea.Msg {
	// Create and run crawler
	c := crawler.NewCrawler(
		m.baseURL,
		m.config.Concurrency,
		m.config.MaxPages,
		m.config.MaxDepth,
		m.config.IgnorePatterns,
	)

	err := c.Start()
	if err != nil {
		return ErrorMsg{Error: fmt.Errorf("crawling failed: %w", err)}
	}

	crawlResults := c.GetResults()
	if len(crawlResults) == 0 {
		return ErrorMsg{Error: fmt.Errorf("no pages found on %s - check if the site is running", m.baseURL)}
	}

	// Convert crawler results to API format
	pages := make([]api.PageData, len(crawlResults))
	for i, result := range crawlResults {
		// Extract links from content (simplified)
		links := extractLinksFromContent(result.Content, result.URL)

		pages[i] = api.PageData{
			URL:        result.URL,
			Content:    result.Content,
			Links:      links,
			Depth:      result.Depth,
			StatusCode: result.StatusCode,
		}
	}

	// Submit pages to backend
	client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
	submitReq := api.SubmitPagesRequest{Pages: pages}

	submitResp, err := client.SubmitPages(m.auditID, submitReq)
	if err != nil {
		return ErrorMsg{Error: fmt.Errorf("failed to submit pages: %w", err)}
	}

	// Create initial page summaries for all discovered pages
	initialPages := make([]PageSummary, len(pages))
	for i, page := range pages {
		status := "pending"
		if page.StatusCode >= 400 {
			status = "failed"
		}

		initialPages[i] = PageSummary{
			URL:    page.URL,
			Score:  0,
			Status: status,
			Issues: 0,
		}
	}

	// Success case - return initial progress with all discovered pages
	return ProgressUpdateMsg{
		Status:        "analyzing",
		PagesFound:    len(pages),
		PagesAnalyzed: 0,
		TotalPages:    submitResp.PagesQueued,
		CurrentPage:   "",
		Pages:         initialPages, // Show all discovered pages immediately
	}
}

// extractLinksFromContent extracts internal links from HTML content
func extractLinksFromContent(content, baseURL string) []string {
	// This is a simplified implementation
	// In production, you'd want to use a proper HTML parser like goquery

	var links []string
	// Basic regex or string parsing to find href attributes
	// For now, return empty slice - implement proper parsing later

	return links
}

// Additional helper for real audit completion
func (m *SimpleAuditModel) realCompleteAudit() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		completionResp, err := client.CompleteAudit(m.auditID)
		if err != nil {
			// Log error but don't fail the audit - results are still available
			return AuditCompletedMsg{
				Summary: "Audit completed with errors",
			}
		}

		// Could show credits used/remaining here
		_ = completionResp

		return AuditCompletedMsg{
			Summary: "Audit completed successfully",
		}
	}
}
