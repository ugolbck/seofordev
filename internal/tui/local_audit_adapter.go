package tui

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/api"
	"github.com/ugolbck/seofordev/internal/audit"
	"github.com/ugolbck/seofordev/internal/crawler"
)

// LocalAuditAdapter provides a local implementation of audit functionality
// It replaces API calls with local processing using the audit.Processor
type LocalAuditAdapter struct {
	processor *audit.Processor
	audits    map[string]*audit.LocalAudit // Cache active audits
	mu        sync.RWMutex
}

// NewLocalAuditAdapter creates a new local audit adapter
func NewLocalAuditAdapter() (*LocalAuditAdapter, error) {
	processor, err := audit.NewProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit processor: %w", err)
	}

	return &LocalAuditAdapter{
		processor: processor,
		audits:    make(map[string]*audit.LocalAudit),
	}, nil
}

// StartAudit starts a local audit session
func (a *LocalAuditAdapter) StartAudit(config AuditConfig, baseURL string) (string, error) {
	auditConfig := audit.AuditConfig{
		Port:           config.Port,
		Concurrency:    config.Concurrency,
		MaxPages:       config.MaxPages,
		MaxDepth:       config.MaxDepth,
		IgnorePatterns: config.IgnorePatterns,
	}

	localAudit, err := a.processor.StartAudit(baseURL, auditConfig)
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	a.audits[localAudit.ID] = localAudit
	a.mu.Unlock()

	return localAudit.ID, nil
}

// PerformSiteDiscoveryAndSubmit handles crawling and submitting pages for analysis
func (a *LocalAuditAdapter) PerformSiteDiscoveryAndSubmit(auditID string, baseURL string, config AuditConfig) tea.Msg {
	// Create and run crawler
	c := crawler.NewCrawler(
		baseURL,
		config.Concurrency,
		config.MaxPages,
		config.MaxDepth,
		config.IgnorePatterns,
	)

	err := c.Start()
	if err != nil {
		return ErrorMsg{Error: fmt.Errorf("crawling failed: %w", err)}
	}

	crawlResults := c.GetResults()
	if len(crawlResults) == 0 {
		return ErrorMsg{Error: fmt.Errorf("no pages found on %s - check if the site is running", baseURL)}
	}

	// Submit pages to local processor for analysis
	err = a.processor.SubmitPages(auditID, crawlResults)
	if err != nil {
		return ErrorMsg{Error: fmt.Errorf("failed to submit pages for analysis: %w", err)}
	}

	// Create initial page summaries for all discovered pages
	initialPages := make([]PageSummary, len(crawlResults))
	for i, result := range crawlResults {
		status := "pending"
		if result.StatusCode >= 400 {
			status = "failed"
		}

		initialPages[i] = PageSummary{
			URL:    result.URL,
			Score:  0,
			Status: status,
			Issues: 0,
		}
	}

	// Success case - return initial progress with all discovered pages
	return ProgressUpdateMsg{
		Status:        "analyzing",
		PagesFound:    len(crawlResults),
		PagesAnalyzed: 0,
		TotalPages:    len(crawlResults),
		CurrentPage:   "",
		Pages:         initialPages, // Show all discovered pages immediately
	}
}

// GetAuditStatus retrieves current audit progress and results from local storage
func (a *LocalAuditAdapter) GetAuditStatus(auditID string) (*api.AuditStatusResponse, error) {
	localAudit, err := a.processor.GetAuditStatus(auditID)
	if err != nil {
		return nil, err
	}

	// Convert local audit to API format
	pages := make([]api.PageSummary, len(localAudit.Pages))
	for i, page := range localAudit.Pages {
		score := 0
		if page.SEOScore != nil {
			score = int(*page.SEOScore)
		}

		status := convertAnalysisStatus(page.AnalysisStatus)
		
		pages[i] = api.PageSummary{
			URL:    page.URL,
			Score:  score,
			Status: status,
			Issues: page.IssuesCount,
		}
	}

	// Count pages by status
	pagesAnalyzed := 0
	currentPage := ""
	for _, page := range localAudit.Pages {
		if page.AnalysisStatus == "completed" || page.AnalysisStatus == "failed" {
			pagesAnalyzed++
		} else if page.AnalysisStatus == "analyzing" && currentPage == "" {
			currentPage = page.URL
		}
	}

	progress := api.ProgressInfo{
		PagesFound:    localAudit.PagesDiscovered,
		PagesAnalyzed: pagesAnalyzed,
		TotalPages:    len(localAudit.Pages),
		CurrentPage:   currentPage,
	}

	var summary *api.AuditSummary
	if localAudit.Summary != nil {
		summary = &api.AuditSummary{
			TotalPages:      localAudit.Summary.TotalPages,
			AverageScore:    localAudit.Summary.AverageScore,
			IssuesFound:     localAudit.Summary.IssuesFound,
			CriticalIssues:  localAudit.Summary.CriticalIssues,
			TopIssues:       localAudit.Summary.TopIssues,
			Recommendations: localAudit.Summary.Recommendations,
		}
	}

	return &api.AuditStatusResponse{
		Status:   localAudit.Status,
		Progress: progress,
		Pages:    pages,
		Summary:  summary,
	}, nil
}

// GetPageDetails retrieves detailed analysis for a specific page
func (a *LocalAuditAdapter) GetPageDetails(auditID, pageURL string) (*api.PageDetailsResponse, error) {
	pageDetails, err := a.processor.GetPageDetails(auditID, pageURL)
	if err != nil {
		return nil, err
	}

	// Convert local page analysis to API format
	checks := make([]api.SEOCheckResponse, 0, len(pageDetails.Checks))
	for checkName, check := range pageDetails.Checks {
		checks = append(checks, api.SEOCheckResponse{
			CheckName: checkName,
			Passed:    check.Passed,
			Value:     check.Value,
			Message:   check.Message,
			Weight:    check.Weight,
		})
	}

	var analyzedAt string
	if pageDetails.AnalyzedAt != nil {
		analyzedAt = pageDetails.AnalyzedAt.Format(time.RFC3339)
	}

	score := 0.0
	if pageDetails.SEOScore != nil {
		score = *pageDetails.SEOScore
	}

	return &api.PageDetailsResponse{
		Status: "success",
		Page: struct {
			URL                string                 `json:"url"`
			StatusCode         int                    `json:"status_code"`
			Title              string                 `json:"title"`
			MetaDescription    string                 `json:"meta_description"`
			H1                 string                 `json:"h1"`
			CanonicalURL       string                 `json:"canonical_url"`
			WordCount          int                    `json:"word_count"`
			SEOScore           float64                `json:"seo_score"`
			AnalysisStatus     string                 `json:"analysis_status"`
			Indexable          bool                   `json:"indexable"`
			IndexabilityReason string                 `json:"indexability_reason"`
			Checks             []api.SEOCheckResponse `json:"checks"`
			AnalyzedAt         string                 `json:"analyzed_at,omitempty"`
			IssuesCount        int                    `json:"issues_count"`
		}{
			URL:                pageDetails.URL,
			StatusCode:         pageDetails.StatusCode,
			Title:              pageDetails.Title,
			MetaDescription:    pageDetails.MetaDescription,
			H1:                 pageDetails.H1,
			CanonicalURL:       pageDetails.CanonicalURL,
			WordCount:          pageDetails.WordCount,
			SEOScore:           score,
			AnalysisStatus:     pageDetails.AnalysisStatus,
			Indexable:          pageDetails.IsIndexable,
			IndexabilityReason: pageDetails.IndexabilityReason,
			Checks:             checks,
			AnalyzedAt:         analyzedAt,
			IssuesCount:        pageDetails.IssuesCount,
		},
	}, nil
}

// CompleteAudit finalizes the audit (no-op for local audits - they complete automatically)
func (a *LocalAuditAdapter) CompleteAudit(auditID string) (*api.CompleteAuditResponse, error) {
	localAudit, err := a.processor.GetAuditStatus(auditID)
	if err != nil {
		return nil, err
	}

	var summary api.AuditSummary
	if localAudit.Summary != nil {
		summary = api.AuditSummary{
			TotalPages:      localAudit.Summary.TotalPages,
			AverageScore:    localAudit.Summary.AverageScore,
			IssuesFound:     localAudit.Summary.IssuesFound,
			CriticalIssues:  localAudit.Summary.CriticalIssues,
			TopIssues:       localAudit.Summary.TopIssues,
			Recommendations: localAudit.Summary.Recommendations,
		}
	}

	return &api.CompleteAuditResponse{
		Summary: summary,
	}, nil
}

// ListAudits returns all local audits
func (a *LocalAuditAdapter) ListAudits() ([]*audit.LocalAudit, error) {
	return a.processor.ListAudits()
}

// DeleteAudit removes an audit from local storage  
func (a *LocalAuditAdapter) DeleteAudit(auditID string) error {
	a.mu.Lock()
	delete(a.audits, auditID)
	a.mu.Unlock()
	
	return a.processor.DeleteAudit(auditID)
}


// Helper function to convert analysis status from local format to API format
func convertAnalysisStatus(localStatus string) string {
	switch localStatus {
	case "pending":
		return "pending"
	case "analyzing":
		return "analyzing"
	case "completed":
		return "complete"
	case "failed":
		return "failed"
	default:
		return "pending"
	}
}