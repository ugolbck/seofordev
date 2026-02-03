package services

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ugolbck/seofordev/internal/audit"
	"github.com/ugolbck/seofordev/internal/crawler"
	"github.com/ugolbck/seofordev/internal/export"
)

// AuditConfig represents configuration for an audit
type AuditConfig struct {
	Port           int
	Concurrency    int
	MaxPages       int
	MaxDepth       int
	IgnorePatterns []string
}

// AuditResult represents the result of a completed audit
type AuditResult struct {
	ID             string
	BaseURL        string
	CreatedAt      time.Time
	CompletedAt    *time.Time
	Status         string
	PagesAnalyzed  int
	OverallScore   *float64
	Pages          []PageResult
	Summary        *AuditSummary
}

// PageResult represents a single page's audit result
type PageResult struct {
	URL            string
	SEOScore       *float64
	AnalysisStatus string
	IssuesCount    int
}

// AuditSummary represents audit summary information
type AuditSummary struct {
	Recommendations []string
	TopIssues       []string
}

// AuditService provides audit functionality
type AuditService struct {
	processor *audit.Processor
}

// NewAuditService creates a new audit service
func NewAuditService() (*AuditService, error) {
	processor, err := audit.NewProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit processor: %w", err)
	}

	return &AuditService{
		processor: processor,
	}, nil
}

// RunAudit performs a complete audit
func (s *AuditService) RunAudit(baseURL string, config AuditConfig) (*AuditResult, error) {

	// Convert config
	auditConfig := audit.AuditConfig{
		Port:           config.Port,
		Concurrency:    config.Concurrency,
		MaxPages:       config.MaxPages,
		MaxDepth:       config.MaxDepth,
		IgnorePatterns: config.IgnorePatterns,
	}

	// Start audit
	localAudit, err := s.processor.StartAudit(baseURL, auditConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start audit: %w", err)
	}


	// Create and run crawler
	c := crawler.NewCrawler(
		baseURL,
		config.Concurrency,
		config.MaxPages,
		config.MaxDepth,
		config.IgnorePatterns,
	)

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("crawling failed: %w", err)
	}

	crawlResults := c.GetResults()
	if len(crawlResults) == 0 {
		return nil, fmt.Errorf("no pages found on %s - check if the site is running", baseURL)
	}


	// Submit pages for analysis
	if err := s.processor.SubmitPages(localAudit.ID, crawlResults); err != nil {
		return nil, fmt.Errorf("failed to submit pages: %w", err)
	}

	// Wait for completion (poll status)
	return s.waitForCompletion(localAudit.ID)
}

// waitForCompletion polls the audit status until complete
func (s *AuditService) waitForCompletion(auditID string) (*AuditResult, error) {
	for {
		status, err := s.processor.GetAuditStatus(auditID)
		if err != nil {
			return nil, fmt.Errorf("failed to get audit status: %w", err)
		}

		log.Debug("Audit progress",
			"status", status.Status,
			"pages_analyzed", status.PagesAnalyzed,
			"total_pages", len(status.Pages))

		if status.Status == "completed" {
			break
		} else if status.Status == "failed" {
			return nil, fmt.Errorf("audit failed")
		}

		// Poll every 2 seconds
		time.Sleep(2 * time.Second)
	}

	// Complete the audit (note: completeAudit is called internally by the processor)
	// Load final results
	return s.GetAudit(auditID)
}

// ListAudits returns all stored audits
func (s *AuditService) ListAudits() ([]*AuditResult, error) {
	localAudits, err := s.processor.ListAudits()
	if err != nil {
		return nil, fmt.Errorf("failed to list audits: %w", err)
	}

	results := make([]*AuditResult, len(localAudits))
	for i, audit := range localAudits {
		results[i] = s.convertLocalAudit(audit)
	}

	return results, nil
}

// GetAudit returns a specific audit
func (s *AuditService) GetAudit(auditID string) (*AuditResult, error) {
	// Get list of audits and find the matching one
	audits, err := s.processor.ListAudits()
	if err != nil {
		return nil, fmt.Errorf("failed to list audits: %w", err)
	}

	// Try exact match first
	for _, audit := range audits {
		if audit.ID == auditID {
			return s.convertLocalAudit(audit), nil
		}
	}

	// Try partial match (first 8 chars)
	if len(auditID) == 8 {
		for _, audit := range audits {
			if len(audit.ID) >= 8 && audit.ID[:8] == auditID {
				return s.convertLocalAudit(audit), nil
			}
		}
	}

	return nil, fmt.Errorf("audit not found: %s", auditID)
}

// ExportAuditPrompt exports an audit as an AI prompt
func (s *AuditService) ExportAuditPrompt(auditID string) (string, error) {
	audit, err := s.GetAudit(auditID)
	if err != nil {
		return "", err
	}

	// Get the detailed audit data (audit is already a result object)
	// We need to get the LocalAudit to access detailed page data
	localAudits, err := s.processor.ListAudits()
	if err != nil {
		return "", fmt.Errorf("failed to list audits: %w", err)
	}

	var foundAudit interface{}
	for _, la := range localAudits {
		if la.ID == audit.ID {
			foundAudit = la
			break
		}
	}

	if foundAudit == nil {
		return "", fmt.Errorf("audit not found: %s", audit.ID)
	}

	// Get page details for each page using the audit ID directly
	var pageDetails []export.PageDetailsResponse
	for _, page := range audit.Pages {
		details, err := s.processor.GetPageDetails(audit.ID, page.URL)
		if err != nil {
			log.Warn("Failed to get page details", "url", page.URL, "error", err)
			continue
		}

		// Convert audit.PageDetails to export.PageDetailsResponse format
		apiDetails := &export.PageDetailsResponse{
			Status: "success",
			Page: struct {
				URL                string                   `json:"url"`
				StatusCode         int                      `json:"status_code"`
				Title              string                   `json:"title"`
				MetaDescription    string                   `json:"meta_description"`
				H1                 string                   `json:"h1"`
				CanonicalURL       string                   `json:"canonical_url"`
				WordCount          int                      `json:"word_count"`
				SEOScore           float64                  `json:"seo_score"`
				AnalysisStatus     string                   `json:"analysis_status"`
				Indexable          bool                     `json:"indexable"`
				IndexabilityReason string                   `json:"indexability_reason"`
				Checks             []export.SEOCheckResponse `json:"checks"`
				AnalyzedAt         string                   `json:"analyzed_at,omitempty"`
				IssuesCount        int                      `json:"issues_count"`
			}{
				URL:                details.URL,
				StatusCode:         details.StatusCode,
				Title:              details.Title,
				MetaDescription:    details.MetaDescription,
				H1:                 details.H1,
				CanonicalURL:       "",
				WordCount:          details.WordCount,
				SEOScore:           func() float64 { if details.SEOScore != nil && *details.SEOScore > 0 { return *details.SEOScore }; return 0 }(),
				AnalysisStatus:     details.AnalysisStatus,
				Indexable:          details.IsIndexable,
				IndexabilityReason: details.IndexabilityReason,
				Checks:             s.convertChecks(details.Checks),
				IssuesCount:        details.IssuesCount,
			},
		}

		if details.AnalyzedAt != nil {
			apiDetails.Page.AnalyzedAt = details.AnalyzedAt.Format(time.RFC3339)
		}

		pageDetails = append(pageDetails, *apiDetails)
	}

	if len(pageDetails) == 0 {
		return "", fmt.Errorf("no page details found for export")
	}

	prompt := export.FormatAIPromptFromMultiplePages(pageDetails)
	return prompt, nil
}

// convertLocalAudit converts audit.LocalAudit to AuditResult
func (s *AuditService) convertLocalAudit(audit *audit.LocalAudit) *AuditResult {
	pages := make([]PageResult, len(audit.Pages))
	for i, page := range audit.Pages {
		pages[i] = PageResult{
			URL:            page.URL,
			SEOScore:       page.SEOScore,
			AnalysisStatus: page.AnalysisStatus,
			IssuesCount:    page.IssuesCount,
		}
	}

	var summary *AuditSummary
	if audit.Summary != nil {
		summary = &AuditSummary{
			Recommendations: audit.Summary.Recommendations,
			TopIssues:       audit.Summary.TopIssues,
		}
	}

	return &AuditResult{
		ID:             audit.ID,
		BaseURL:        audit.BaseURL,
		CreatedAt:      audit.CreatedAt,
		CompletedAt:    audit.CompletedAt,
		Status:         audit.Status,
		PagesAnalyzed:  len(audit.Pages),
		OverallScore:   audit.OverallScore,
		Pages:          pages,
		Summary:        summary,
	}
}

// convertChecks converts map[string]audit.CheckResult to []export.SEOCheckResponse
func (s *AuditService) convertChecks(checks map[string]audit.CheckResult) []export.SEOCheckResponse {
	result := make([]export.SEOCheckResponse, 0, len(checks))

	for name, check := range checks {
		result = append(result, export.SEOCheckResponse{
			CheckName: name,
			Passed:    check.Passed,
			Value:     check.Value,
			Message:   check.Message,
			Weight:    check.Weight,
		})
	}

	return result
}
