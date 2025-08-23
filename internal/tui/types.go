package tui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/crawler"
)

// AuditPhase represents the current phase of the audit process
type AuditPhase int

const (
	PhaseSessionCreation AuditPhase = iota
	PhaseSiteDiscovery
	PhaseCreditCheck
	PhasePageAnalysis
	PhaseSessionCompletion
	PhaseResultsNavigation
)

// PageStatus represents the status of a page during audit
type PageStatus int

const (
	StatusPending PageStatus = iota
	StatusCrawling
	StatusAnalyzing
	StatusCompleted
	StatusError
	StatusWarning
)

// PageResult represents an analyzed page with its results
type PageResult struct {
	URL        string                 `json:"url"`
	Status     PageStatus             `json:"status"`
	Score      int                    `json:"score"`
	Checks     map[string]CheckResult `json:"checks"`
	Analysis   map[string]interface{} `json:"analysis"`
	Error      string                 `json:"error,omitempty"`
	CrawledAt  time.Time              `json:"crawled_at"`
	AnalyzedAt time.Time              `json:"analyzed_at"`
}

// CheckResult represents the result of an individual SEO check
type CheckResult struct {
	Message string      `json:"message"`
	Passed  bool        `json:"passed"`
	Value   interface{} `json:"value"`
}

// AuditSession represents a complete audit session
type AuditSession struct {
	RunID           string               `json:"run_id"`
	BaseURL         string               `json:"base_url"`
	StartedAt       time.Time            `json:"started_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
	Phase           AuditPhase           `json:"phase"`
	Pages           []PageResult         `json:"pages"`
	SiteMap         map[string][]string  `json:"site_map"` // URL -> linked URLs
	CrawlResults    []crawler.PageResult `json:"crawl_results"`
	CreditsUsed     int                  `json:"credits_used"`
	CreditsRequired int                  `json:"credits_required"`
	Summary         *AuditSummary        `json:"summary,omitempty"`
	Config          AuditConfig          `json:"config"`
}

// AuditConfig holds the configuration for an audit
type AuditConfig struct {
	Port           int      `json:"port"`
	Concurrency    int      `json:"concurrency"`
	MaxPages       int      `json:"max_pages"`
	MaxDepth       int      `json:"max_depth"`
	IgnorePatterns []string `json:"ignore_patterns"`
	APIKey         string   `json:"-"` // Don't serialize the API key
}

// GetEffectiveBaseURL returns the base URL to use, checking environment variable override first
func (c *AuditConfig) GetEffectiveBaseURL() string {
	// Check for environment variable override first (for development only)
	if envURL := os.Getenv("SEO_BASE_URL"); envURL != "" {
		return envURL
	}

	// Always use production URL for users
	return "https://seofor.dev"
}

// AuditSummary provides high-level audit results
type AuditSummary struct {
	TotalPages      int            `json:"total_pages"`
	AverageScore    float64        `json:"average_score"`
	IssuesFound     int            `json:"issues_found"`
	CriticalIssues  int            `json:"critical_issues"`
	WarningIssues   int            `json:"warning_issues"`
	PassedChecks    int            `json:"passed_checks"`
	FailedChecks    int            `json:"failed_checks"`
	TopIssues       []string       `json:"top_issues"`
	Recommendations []string       `json:"recommendations"`
	ScoreByPage     map[string]int `json:"score_by_page"`
}

// Bubble Tea Messages
type (
	// Phase transition messages
	PhaseCompleteMsg struct {
		Phase AuditPhase
		Data  interface{}
	}

	// Session creation messages
	SessionCreatedMsg struct {
		RunID string
	}

	// Site discovery messages
	CrawlStartedMsg  struct{}
	CrawlProgressMsg struct {
		Discovered int
		Crawled    int
		Queue      int
	}
	CrawlCompletedMsg struct {
		Results []crawler.PageResult
		SiteMap map[string][]string
	}

	// Credit check messages
	CreditCheckMsg struct {
		Required  int
		Available int
	}
	CreditConfirmedMsg struct{}

	// Page analysis messages
	AnalysisStartedMsg  struct{}
	AnalysisProgressMsg struct {
		Page    string
		Current int
		Total   int
		Result  *PageResult
	}
	AnalysisCompletedMsg struct {
		Summary AuditSummary
	}

	// Error messages
	ErrorMsg struct {
		Error error
	}

	// Tick messages for animations
	TickMsg time.Time

	// Notification messages
	NotificationMsg struct {
		Message string
		Type    NotificationType
	}

	// Hide notification message
	HideNotificationMsg struct{}

	// Brief generation messages
	BriefGenerationStartedMsg struct {
		BriefID string
		Keyword string
	}

	BriefGenerationProgressMsg struct {
		BriefID string
		Status  string
	}

	BriefGenerationCompletedMsg struct {
		BriefID     string
		Brief       string
		Status      string
		CreditsUsed int
	}

	BriefGenerationFailedMsg struct {
		BriefID string
		Error   string
	}

	// Message to trigger brief generation from keyword selection
	GenerateBriefForKeywordMsg struct {
		Keyword string
	}
)

// NotificationType represents the type of notification
type NotificationType int

const (
	NotificationSuccess NotificationType = iota
	NotificationError
	NotificationInfo
)

// Common interface for all models
type Model interface {
	tea.Model
	SetSize(width, height int)
	IsComplete() bool
}
