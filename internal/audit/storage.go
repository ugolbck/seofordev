package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalAudit represents a complete audit stored locally
type LocalAudit struct {
	ID               string                  `json:"id"`
	BaseURL          string                  `json:"base_url"`
	CreatedAt        time.Time               `json:"created_at"`
	CompletedAt      *time.Time              `json:"completed_at,omitempty"`
	Status           string                  `json:"status"`
	PagesDiscovered  int                     `json:"pages_discovered"`
	PagesAnalyzed    int                     `json:"pages_analyzed"`
	OverallScore     *float64                `json:"overall_score,omitempty"`
	AvgPageScore     *float64                `json:"avg_page_score,omitempty"`
	Config           AuditConfig             `json:"config"`
	Pages            []LocalPageAnalysis     `json:"pages"`
	Summary          *LocalAuditSummary      `json:"summary,omitempty"`
}

// LocalPageAnalysis represents a single page's SEO analysis
type LocalPageAnalysis struct {
	ID             string                  `json:"id"`
	URL            string                  `json:"url"`
	StatusCode     int                     `json:"status_code"`
	Depth          int                     `json:"depth"`
	AnalysisStatus string                  `json:"analysis_status"`
	SEOScore       *float64                `json:"seo_score,omitempty"`
	AnalyzedAt     *time.Time              `json:"analyzed_at,omitempty"`
	
	// SEO Elements
	Title              string   `json:"title"`
	MetaDescription    string   `json:"meta_description"`
	H1                 string   `json:"h1"`
	H1Count            int      `json:"h1_count"`
	CanonicalURL       string   `json:"canonical_url"`
	WordCount          int      `json:"word_count"`
	TitleLength        int      `json:"title_length"`
	DescriptionLength  int      `json:"description_length"`
	InternalLinksCount int      `json:"internal_links_count"`
	ExternalLinksCount int      `json:"external_links_count"`
	TotalLinksCount    int      `json:"total_links_count"`
	ImagesTotal        int      `json:"images_total"`
	ImagesWithoutAlt   int      `json:"images_without_alt"`
	IsIndexable        bool     `json:"is_indexable"`
	IndexabilityReason string   `json:"indexability_reason"`
	HasNoindex         bool     `json:"has_noindex"`
	HasNofollow        bool     `json:"has_nofollow"`
	DetectedLanguage   string   `json:"detected_language"`
	HasViewportMeta    bool     `json:"has_viewport_meta"`
	HasCharset         bool     `json:"has_charset"`
	HasStructuredData  bool     `json:"has_structured_data"`
	IssuesCount        int      `json:"issues_count"`
	
	// Check results
	Checks map[string]CheckResult `json:"checks"`
	
	// Links (for internal link analysis)
	InternalLinks []LinkInfo `json:"internal_links,omitempty"`
}

// LocalAuditSummary represents audit summary statistics
type LocalAuditSummary struct {
	TotalPages              int      `json:"total_pages"`
	AverageScore           float64   `json:"average_score"`
	IssuesFound            int       `json:"issues_found"`
	CriticalIssues         int       `json:"critical_issues"`
	WarningIssues          int       `json:"warning_issues"`
	PassedChecks           int       `json:"passed_checks"`
	FailedChecks           int       `json:"failed_checks"`
	TopIssues              []string  `json:"top_issues"`
	Recommendations        []string  `json:"recommendations"`
	ScoreByPage            map[string]int `json:"score_by_page"`
	PagesMissingTitle      int       `json:"pages_missing_title"`
	PagesMissingDescription int      `json:"pages_missing_description"`
	PagesMissingH1         int       `json:"pages_missing_h1"`
	PagesScore90Plus       int       `json:"pages_score_90_plus"`
	PagesScore7089         int       `json:"pages_score_70_89"`
	PagesScore5069         int       `json:"pages_score_50_69"`
	PagesScoreBelow50      int       `json:"pages_score_below_50"`
	DuplicateTitlesCount   int       `json:"duplicate_titles_count"`
	DuplicateDescriptionsCount int   `json:"duplicate_descriptions_count"`
	OrphanedPagesCount     int       `json:"orphaned_pages_count"`
}

// AuditConfig represents audit configuration
type AuditConfig struct {
	Port           int      `json:"port"`
	Concurrency    int      `json:"concurrency"`
	MaxPages       int      `json:"max_pages"`
	MaxDepth       int      `json:"max_depth"`
	IgnorePatterns []string `json:"ignore_patterns"`
}

// LocalStorage handles local audit storage
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage() (*LocalStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".seo", "audits")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audits directory: %w", err)
	}

	return &LocalStorage{
		baseDir: baseDir,
	}, nil
}

// CreateAudit creates a new audit record
func (s *LocalStorage) CreateAudit(baseURL string, config AuditConfig) (*LocalAudit, error) {
	audit := &LocalAudit{
		ID:              uuid.New().String(),
		BaseURL:         baseURL,
		CreatedAt:       time.Now(),
		Status:          "discovering",
		PagesDiscovered: 0,
		PagesAnalyzed:   0,
		Config:          config,
		Pages:           make([]LocalPageAnalysis, 0),
	}

	if err := s.SaveAudit(audit); err != nil {
		return nil, fmt.Errorf("failed to save new audit: %w", err)
	}

	return audit, nil
}

// SaveAudit saves an audit to local storage
func (s *LocalStorage) SaveAudit(audit *LocalAudit) error {
	filename := fmt.Sprintf("%s.json", audit.ID)
	filepath := filepath.Join(s.baseDir, filename)

	data, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal audit: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write audit file: %w", err)
	}

	return nil
}

// LoadAudit loads an audit from local storage
func (s *LocalStorage) LoadAudit(auditID string) (*LocalAudit, error) {
	filename := fmt.Sprintf("%s.json", auditID)
	filepath := filepath.Join(s.baseDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit file: %w", err)
	}

	var audit LocalAudit
	if err := json.Unmarshal(data, &audit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit: %w", err)
	}

	return &audit, nil
}

// ListAudits returns a list of all stored audits
func (s *LocalStorage) ListAudits() ([]*LocalAudit, error) {
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read audits directory: %w", err)
	}

	var audits []*LocalAudit
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			auditID := strings.TrimSuffix(file.Name(), ".json")
			audit, err := s.LoadAudit(auditID)
			if err != nil {
				// Skip corrupted files
				continue
			}
			audits = append(audits, audit)
		}
	}

	// Sort by creation date (newest first)
	sort.Slice(audits, func(i, j int) bool {
		return audits[i].CreatedAt.After(audits[j].CreatedAt)
	})

	return audits, nil
}

// DeleteAudit deletes an audit from local storage
func (s *LocalStorage) DeleteAudit(auditID string) error {
	filename := fmt.Sprintf("%s.json", auditID)
	filepath := filepath.Join(s.baseDir, filename)

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete audit file: %w", err)
	}

	return nil
}

// AddPageAnalysis adds a page analysis to an audit
func (s *LocalStorage) AddPageAnalysis(auditID string, page LocalPageAnalysis) error {
	audit, err := s.LoadAudit(auditID)
	if err != nil {
		return fmt.Errorf("failed to load audit: %w", err)
	}

	// Add or update page
	found := false
	for i, existingPage := range audit.Pages {
		if existingPage.URL == page.URL {
			audit.Pages[i] = page
			found = true
			break
		}
	}

	if !found {
		audit.Pages = append(audit.Pages, page)
	}

	// Update counters
	audit.PagesAnalyzed = len(audit.Pages)
	
	// Calculate average score
	totalScore := 0.0
	validPages := 0
	for _, p := range audit.Pages {
		if p.SEOScore != nil {
			totalScore += *p.SEOScore
			validPages++
		}
	}
	
	if validPages > 0 {
		avgScore := totalScore / float64(validPages)
		audit.AvgPageScore = &avgScore
	}

	return s.SaveAudit(audit)
}

// CompleteAudit marks an audit as completed and generates summary
func (s *LocalStorage) CompleteAudit(auditID string) error {
	audit, err := s.LoadAudit(auditID)
	if err != nil {
		return fmt.Errorf("failed to load audit: %w", err)
	}

	now := time.Now()
	audit.CompletedAt = &now
	audit.Status = "completed"

	// Generate summary
	audit.Summary = s.generateSummary(audit)
	audit.OverallScore = &audit.Summary.AverageScore

	return s.SaveAudit(audit)
}

// generateSummary generates audit summary statistics
func (s *LocalStorage) generateSummary(audit *LocalAudit) *LocalAuditSummary {
	summary := &LocalAuditSummary{
		TotalPages:     len(audit.Pages),
		TopIssues:      make([]string, 0),
		Recommendations: make([]string, 0),
		ScoreByPage:    make(map[string]int),
	}

	if len(audit.Pages) == 0 {
		return summary
	}

	// Track common issues
	issueTracker := make(map[string]int)
	totalScore := 0.0
	validPages := 0
	totalChecks := 0
	passedChecks := 0
	failedChecks := 0

	// Track duplicates
	titles := make(map[string]int)
	descriptions := make(map[string]int)

	for _, page := range audit.Pages {
		if page.SEOScore != nil {
			score := int(*page.SEOScore)
			totalScore += *page.SEOScore
			validPages++
			summary.ScoreByPage[page.URL] = score

			// Score distribution
			switch {
			case score >= 90:
				summary.PagesScore90Plus++
			case score >= 70:
				summary.PagesScore7089++
			case score >= 50:
				summary.PagesScore5069++
			default:
				summary.PagesScoreBelow50++
			}
		}

		// Count missing elements
		if strings.TrimSpace(page.Title) == "" {
			summary.PagesMissingTitle++
			issueTracker["Missing title tag"]++
		} else {
			titles[page.Title]++
		}

		if strings.TrimSpace(page.MetaDescription) == "" {
			summary.PagesMissingDescription++
			issueTracker["Missing meta description"]++
		} else {
			descriptions[page.MetaDescription]++
		}

		if page.H1Count == 0 {
			summary.PagesMissingH1++
			issueTracker["Missing H1 heading"]++
		}

		// Count check results
		for _, check := range page.Checks {
			totalChecks++
			if check.Passed {
				passedChecks++
			} else {
				failedChecks++
				issueTracker[check.Message]++
			}
		}

		// Count issues
		summary.IssuesFound += page.IssuesCount
		if page.IssuesCount > 5 {
			summary.CriticalIssues++
		} else if page.IssuesCount > 0 {
			summary.WarningIssues++
		}
	}

	// Calculate averages
	if validPages > 0 {
		summary.AverageScore = totalScore / float64(validPages)
	}
	summary.PassedChecks = passedChecks
	summary.FailedChecks = failedChecks

	// Count duplicates
	for _, count := range titles {
		if count > 1 {
			summary.DuplicateTitlesCount++
		}
	}
	for _, count := range descriptions {
		if count > 1 {
			summary.DuplicateDescriptionsCount++
		}
	}

	// Generate top issues (most common problems)
	type issueCount struct {
		issue string
		count int
	}
	var issues []issueCount
	for issue, count := range issueTracker {
		if count > 0 {
			issues = append(issues, issueCount{issue, count})
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].count > issues[j].count
	})

	// Take top 5 issues
	for i, issue := range issues {
		if i >= 5 {
			break
		}
		summary.TopIssues = append(summary.TopIssues, fmt.Sprintf("%s (%d pages)", issue.issue, issue.count))
	}

	// Generate recommendations based on common issues
	summary.Recommendations = s.generateRecommendations(summary, issueTracker)

	return summary
}

// generateRecommendations generates actionable recommendations
func (s *LocalStorage) generateRecommendations(summary *LocalAuditSummary, issues map[string]int) []string {
	var recommendations []string

	if summary.PagesMissingTitle > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Add title tags to %d pages", summary.PagesMissingTitle))
	}

	if summary.PagesMissingDescription > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Add meta descriptions to %d pages", summary.PagesMissingDescription))
	}

	if summary.PagesMissingH1 > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Add H1 headings to %d pages", summary.PagesMissingH1))
	}

	if summary.DuplicateTitlesCount > 1 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d duplicate title tags", summary.DuplicateTitlesCount))
	}

	if summary.DuplicateDescriptionsCount > 1 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d duplicate meta descriptions", summary.DuplicateDescriptionsCount))
	}

	if issues["Page is missing viewport meta tag (important for mobile)"] > 0 {
		recommendations = append(recommendations, "Add viewport meta tag for mobile optimization")
	}

	if summary.AverageScore < 70 {
		recommendations = append(recommendations, "Focus on improving overall SEO scores")
	}

	// Limit to top 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations
}