package audit

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ugolbck/seofordev/internal/crawler"
)

// ProcessorStatus represents the status of audit processing
type ProcessorStatus string

const (
	StatusDiscovering ProcessorStatus = "discovering"
	StatusAnalyzing   ProcessorStatus = "analyzing" 
	StatusCompleted   ProcessorStatus = "completed"
	StatusFailed      ProcessorStatus = "failed"
)

// PageStatus represents the status of individual page processing
type PageStatus string

const (
	PageStatusPending   PageStatus = "pending"
	PageStatusAnalyzing PageStatus = "analyzing"
	PageStatusCompleted PageStatus = "completed"
	PageStatusFailed    PageStatus = "failed"
)

// Processor handles the complete audit workflow locally
type Processor struct {
	storage    *LocalStorage
	analyzer   *Analyzer
	audit      *LocalAudit
	mu         sync.RWMutex
	processing map[string]bool // Track which pages are being processed
}

// NewProcessor creates a new audit processor
func NewProcessor() (*Processor, error) {
	storage, err := NewLocalStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &Processor{
		storage:    storage,
		analyzer:   NewAnalyzer(),
		processing: make(map[string]bool),
	}, nil
}

// StartAudit initializes a new audit session
func (p *Processor) StartAudit(baseURL string, config AuditConfig) (*LocalAudit, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	audit, err := p.storage.CreateAudit(baseURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit: %w", err)
	}

	p.audit = audit
	log.Printf("🚀 Started new audit: %s for %s", audit.ID, baseURL)
	
	return audit, nil
}

// SubmitPages processes discovered pages and starts analysis
func (p *Processor) SubmitPages(auditID string, pages []crawler.PageResult) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.audit == nil || p.audit.ID != auditID {
		// Load audit if not in memory
		audit, err := p.storage.LoadAudit(auditID)
		if err != nil {
			return fmt.Errorf("audit not found: %w", err)
		}
		p.audit = audit
	}

	// Update audit status
	p.audit.Status = string(StatusAnalyzing)
	p.audit.PagesDiscovered = len(pages)

	// Start processing pages concurrently
	concurrency := p.audit.Config.Concurrency
	if concurrency <= 0 {
		concurrency = 3 // Default concurrency
	}

	// Create a channel to control concurrency
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	log.Printf("📥 Processing %d pages with concurrency %d", len(pages), concurrency)

	for _, page := range pages {
		// Skip failed pages (status code != 200)
		if page.StatusCode != 200 {
			p.addFailedPage(page)
			continue
		}

		wg.Add(1)
		go func(pageData crawler.PageResult) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if err := p.processPage(pageData); err != nil {
				log.Printf("❌ Failed to process page %s: %v", pageData.URL, err)
			}
		}(page)
	}

	// Wait for all pages to be processed
	go func() {
		wg.Wait()
		p.completeAudit()
	}()

	return nil
}

// processPage analyzes a single page
func (p *Processor) processPage(pageData crawler.PageResult) error {
	pageID := uuid.New().String()
	
	p.mu.Lock()
	p.processing[pageData.URL] = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.processing, pageData.URL)
		p.mu.Unlock()
	}()

	log.Printf("🔍 Analyzing page: %s", pageData.URL)

	// Create initial page record
	page := LocalPageAnalysis{
		ID:             pageID,
		URL:            pageData.URL,
		StatusCode:     pageData.StatusCode,
		Depth:          pageData.Depth,
		AnalysisStatus: string(PageStatusAnalyzing),
	}

	// Perform SEO analysis
	analysis, err := p.analyzer.AnalyzeContent(pageData.Content, pageData.URL)
	if err != nil {
		log.Printf("❌ Analysis failed for %s: %v", pageData.URL, err)
		page.AnalysisStatus = string(PageStatusFailed)
		page.IndexabilityReason = fmt.Sprintf("Analysis failed: %v", err)
		return p.storage.AddPageAnalysis(p.audit.ID, page)
	}

	// Run SEO checks
	checker := NewChecker(analysis, pageData.StatusCode)
	checkResults := checker.RunAllChecks()

	// Populate page data from analysis
	p.populatePageFromAnalysis(&page, analysis, checkResults)

	now := time.Now()
	page.AnalyzedAt = &now
	page.AnalysisStatus = string(PageStatusCompleted)

	log.Printf("✅ Completed analysis for %s (score: %.1f)", pageData.URL, *page.SEOScore)

	// Save page to audit
	return p.storage.AddPageAnalysis(p.audit.ID, page)
}

// populatePageFromAnalysis fills page data from analysis results
func (p *Processor) populatePageFromAnalysis(page *LocalPageAnalysis, analysis *AnalysisResult, checks *CheckResults) {
	// Basic info
	page.IsIndexable = checks.Indexable
	page.IndexabilityReason = checks.IndexabilityReason
	score := checks.Score
	page.SEOScore = &score

	// SEO elements
	page.Title = analysis.Title
	page.MetaDescription = analysis.Description
	page.TitleLength = analysis.Content.TitleLength
	page.DescriptionLength = analysis.Content.DescriptionLength
	page.WordCount = analysis.Content.WordCount

	// Headings
	if len(analysis.H1) > 0 {
		page.H1 = analysis.H1[0]
	}
	page.H1Count = analysis.Headings.H1Count

	// Technical
	page.CanonicalURL = analysis.Technical.Canonical
	page.HasViewportMeta = analysis.Technical.ViewportMeta
	page.HasCharset = analysis.Technical.CharsetDeclared
	page.HasStructuredData = analysis.Schema.HasStructuredData

	// Links
	page.InternalLinksCount = analysis.Links.InternalCount
	page.ExternalLinksCount = analysis.Links.ExternalCount
	page.TotalLinksCount = analysis.Links.TotalCount
	page.InternalLinks = analysis.Links.Internal

	// Images
	page.ImagesTotal = analysis.Images.TotalCount
	page.ImagesWithoutAlt = analysis.Images.WithoutAltCount

	// Robots
	if analysis.Robots != nil {
		page.HasNoindex = analysis.Robots.NoIndex
		page.HasNofollow = analysis.Robots.NoFollow
	}

	// Language
	page.DetectedLanguage = analysis.Language

	// Count issues (failed checks)
	issuesCount := 0
	for _, check := range checks.Checks {
		if !check.Passed {
			issuesCount++
		}
	}
	page.IssuesCount = issuesCount

	// Store check results
	page.Checks = checks.Checks
}

// addFailedPage adds a page that couldn't be crawled
func (p *Processor) addFailedPage(pageData crawler.PageResult) {
	page := LocalPageAnalysis{
		ID:                 uuid.New().String(),
		URL:                pageData.URL,
		StatusCode:         pageData.StatusCode,
		Depth:              pageData.Depth,
		AnalysisStatus:     string(PageStatusFailed),
		IndexabilityReason: fmt.Sprintf("HTTP %d - Page not accessible", pageData.StatusCode),
		IsIndexable:        false,
		IssuesCount:        1,
	}

	score := 0.0
	page.SEOScore = &score

	p.storage.AddPageAnalysis(p.audit.ID, page)
}

// completeAudit finalizes the audit
func (p *Processor) completeAudit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.audit == nil {
		return
	}

	log.Printf("🏁 Completing audit %s", p.audit.ID)

	if err := p.storage.CompleteAudit(p.audit.ID); err != nil {
		log.Printf("❌ Failed to complete audit: %v", err)
		return
	}

	// Reload audit to get updated data
	if updatedAudit, err := p.storage.LoadAudit(p.audit.ID); err == nil {
		p.audit = updatedAudit
	}

	log.Printf("✅ Audit completed: %d pages analyzed", p.audit.PagesAnalyzed)
}

// GetAuditStatus returns current audit status and progress
func (p *Processor) GetAuditStatus(auditID string) (*LocalAudit, error) {
	// Always load from storage to get latest data
	audit, err := p.storage.LoadAudit(auditID)
	if err != nil {
		return nil, fmt.Errorf("audit not found: %w", err)
	}

	return audit, nil
}

// GetPageDetails returns detailed analysis for a specific page
func (p *Processor) GetPageDetails(auditID, pageURL string) (*LocalPageAnalysis, error) {
	audit, err := p.storage.LoadAudit(auditID)
	if err != nil {
		return nil, fmt.Errorf("audit not found: %w", err)
	}

	for _, page := range audit.Pages {
		if page.URL == pageURL {
			return &page, nil
		}
	}

	return nil, fmt.Errorf("page not found in audit")
}

// ListAudits returns all stored audits
func (p *Processor) ListAudits() ([]*LocalAudit, error) {
	return p.storage.ListAudits()
}

// DeleteAudit removes an audit from storage
func (p *Processor) DeleteAudit(auditID string) error {
	return p.storage.DeleteAudit(auditID)
}