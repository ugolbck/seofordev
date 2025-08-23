package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client handles API communication with the SEO backend
type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// APIError represents an API error with status code and body
type APIError struct {
	StatusCode int
	Body       []byte
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// InsufficientCreditsError represents an API error when user has insufficient credits
// This struct matches the Python backend InsufficientCreditsErrorResponse
type InsufficientCreditsError struct {
	Message                 string `json:"error"`
	CreditsRequired         int    `json:"credits_required"`
	CurrentBalance          int    `json:"current_balance"`
	PagesThatCanBeProcessed int    `json:"pages_that_can_be_processed"`
}

func (e *InsufficientCreditsError) Error() string {
	return fmt.Sprintf("insufficient credits: need %d, have %d", e.CreditsRequired, e.CurrentBalance)
}

// NewClient creates a new API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// API Request/Response structures

type StartAuditRequest struct {
	BaseURL        string   `json:"base_url"`
	MaxPages       int      `json:"max_pages"`
	MaxDepth       int      `json:"max_depth"`
	IgnorePatterns []string `json:"ignore_patterns"`
}

type StartAuditResponse struct {
	AuditID string `json:"audit_id"`
}

type SubmitPagesRequest struct {
	Pages []PageData `json:"pages"`
}

type SubmitPagesResponse struct {
	Status      string `json:"status"`
	PagesQueued int    `json:"pages_queued"`
	AuditStatus string `json:"audit_status"`
}

type PageData struct {
	URL        string   `json:"url"`
	Content    string   `json:"content"`
	Links      []string `json:"links"`
	Depth      int      `json:"depth"`
	StatusCode int      `json:"status_code"`
}

type AuditStatusResponse struct {
	Status   string        `json:"status"` // "discovering", "analyzing", "complete"
	Progress ProgressInfo  `json:"progress"`
	Pages    []PageSummary `json:"pages"`
	Summary  *AuditSummary `json:"summary,omitempty"`
}

type ProgressInfo struct {
	PagesFound    int    `json:"pages_found"`
	PagesAnalyzed int    `json:"pages_analyzed"`
	TotalPages    int    `json:"total_pages"`
	CurrentPage   string `json:"current_page,omitempty"`
}

type PageSummary struct {
	URL    string `json:"url"`
	Score  int    `json:"score"`
	Status string `json:"status"` // "pending", "analyzing", "complete"
	Issues int    `json:"issues"`
}

type AuditSummary struct {
	TotalPages      int      `json:"total_pages"`
	AverageScore    float64  `json:"average_score"`
	IssuesFound     int      `json:"issues_found"`
	CriticalIssues  int      `json:"critical_issues"`
	TopIssues       []string `json:"top_issues"`
	Recommendations []string `json:"recommendations"`
}

type CompleteAuditResponse struct {
	Summary AuditSummary `json:"summary"`
}

// SEOCheckResponse represents an individual SEO check result
type SEOCheckResponse struct {
	CheckName string      `json:"check_name"`
	Passed    bool        `json:"passed"`
	Value     interface{} `json:"value,omitempty"` // Can be number, string, object, or null
	Message   string      `json:"message"`
	Weight    int         `json:"weight"`
}

// PageDetailsResponse represents detailed information about a single page
type PageDetailsResponse struct {
	Status string `json:"status"`
	Page   struct {
		URL                string             `json:"url"`
		StatusCode         int                `json:"status_code"`
		Title              string             `json:"title"`
		MetaDescription    string             `json:"meta_description"`
		H1                 string             `json:"h1"`
		CanonicalURL       string             `json:"canonical_url"`
		WordCount          int                `json:"word_count"`
		SEOScore           float64            `json:"seo_score"`
		AnalysisStatus     string             `json:"analysis_status"`
		Indexable          bool               `json:"indexable"`
		IndexabilityReason string             `json:"indexability_reason"`
		Checks             []SEOCheckResponse `json:"checks"`
		AnalyzedAt         string             `json:"analyzed_at,omitempty"`
		IssuesCount        int                `json:"issues_count"`
	} `json:"page"`
}

// AuditViewResponse represents a single audit in the history list
type AuditDetailPageResponse struct {
	ID             string  `json:"id"`
	URL            string  `json:"url"`
	AnalysisStatus string  `json:"analysis_status"`
	SEOScore       float64 `json:"seo_score"`
	AnalyzedAt     *string `json:"analyzed_at,omitempty"`
	IssuesCount    int     `json:"issues_count"`
}

type AuditViewResponse struct {
	ID           string                    `json:"id"`
	CreatedAt    string                    `json:"created_at"`
	Status       string                    `json:"status"`
	OverallScore float64                   `json:"overall_score"`
	Pages        []AuditDetailPageResponse `json:"pages"`
}

// AuditListResponse represents the list of audits
type AuditListResponse struct {
	Audits []AuditViewResponse `json:"audits"`
}

// API Methods

// StartAudit initiates a new audit session
func (c *Client) StartAudit(req StartAuditRequest) (*StartAuditResponse, error) {
	url := fmt.Sprintf("%s/api/audit/start/", c.BaseURL)

	var resp StartAuditResponse
	if err := c.makeRequest("POST", url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to start audit: %w", err)
	}

	return &resp, nil
}

// SubmitPages sends discovered pages for analysis
func (c *Client) SubmitPages(auditID string, req SubmitPagesRequest) (*SubmitPagesResponse, error) {
	url := fmt.Sprintf("%s/api/audit/%s/pages/", c.BaseURL, auditID)

	var resp SubmitPagesResponse
	if err := c.makeRequest("POST", url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to submit pages: %w", err)
	}

	return &resp, nil
}

// GetAuditStatus retrieves current audit progress and results
func (c *Client) GetAuditStatus(auditID string) (*AuditStatusResponse, error) {
	url := fmt.Sprintf("%s/api/audit/%s/status/", c.BaseURL, auditID)

	var resp AuditStatusResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get audit status: %w", err)
	}

	return &resp, nil
}

// CompleteAudit finalizes the audit session
func (c *Client) CompleteAudit(auditID string) (*CompleteAuditResponse, error) {
	url := fmt.Sprintf("%s/api/audit/%s/complete/", c.BaseURL, auditID)

	var resp CompleteAuditResponse
	if err := c.makeRequest("POST", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to complete audit: %w", err)
	}

	return &resp, nil
}

// GetPageDetails retrieves detailed analysis for a specific page
func (c *Client) GetPageDetails(auditID, pageURL string) (*PageDetailsResponse, error) {
	// Use query parameter instead of path parameter to avoid URL encoding issues
	baseURL := fmt.Sprintf("%s/api/audit/%s/page/", c.BaseURL, auditID)

	// Properly encode the page URL as a query parameter
	params := url.Values{}
	params.Add("url", pageURL)

	fullURL := baseURL + "?" + params.Encode()

	var resp PageDetailsResponse
	if err := c.makeRequest("GET", fullURL, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get page details: %w", err)
	}

	return &resp, nil
}

// GetAuditHistory fetches the list of user's audits
func (c *Client) GetAuditHistory() (*AuditListResponse, error) {
	url := fmt.Sprintf("%s/api/audit/history/", c.BaseURL)

	var resp AuditListResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get audit history: %w", err)
	}

	return &resp, nil
}

// Keyword Generation API Structures and Methods

// KeywordData represents a single keyword with its metrics
type KeywordData struct {
	Keyword    string   `json:"keyword"`
	Volume     *int     `json:"volume"`        // Nullable - can be nil
	Difficulty *float64 `json:"difficulty"`    // Nullable - can be nil
	CPC        *float64 `json:"cpc,omitempty"` // Nullable - can be nil
}

// GenerateKeywordsRequest represents the request for keyword generation
type GenerateKeywordsRequest struct {
	SeedKeyword string `json:"seed_keyword"`
}

// GenerateKeywordsResponse represents the response from keyword generation
type GenerateKeywordsResponse struct {
	ID           string        `json:"id"`
	SeedKeyword  string        `json:"seed_keyword"`
	Keywords     []KeywordData `json:"keywords"`
	Status       string        `json:"status"`
	CreditsUsed  int           `json:"credits_used"`
	GeneratedAt  string        `json:"generated_at"`
	TotalResults int           `json:"total_results"`
}

// KeywordGenerationHistoryItem represents a single generation in history
type KeywordGenerationHistoryItem struct {
	ID           string        `json:"id"`
	SeedKeyword  string        `json:"seed_keyword"`
	Keywords     []KeywordData `json:"keywords"`
	Status       string        `json:"status"`
	CreditsUsed  int           `json:"credits_used"`
	GeneratedAt  string        `json:"generated_at"`
	TotalResults int           `json:"total_results"`
}

// KeywordHistoryResponse represents the response from keyword history
type KeywordHistoryResponse struct {
	Generations []KeywordGenerationHistoryItem `json:"generations"`
}

// GenerateKeywords generates keywords for a given seed keyword
func (c *Client) GenerateKeywords(seedKeyword string) (*GenerateKeywordsResponse, error) {
	url := fmt.Sprintf("%s/api/keyword/generate/", c.BaseURL)

	req := GenerateKeywordsRequest{
		SeedKeyword: seedKeyword,
	}

	var resp GenerateKeywordsResponse
	if err := c.makeRequest("POST", url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to generate keywords: %w", err)
	}

	return &resp, nil
}

// GetKeywordHistory fetches the list of user's keyword generations
func (c *Client) GetKeywordHistory() (*KeywordHistoryResponse, error) {
	url := fmt.Sprintf("%s/api/keyword/history/", c.BaseURL)

	var resp KeywordHistoryResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get keyword history: %w", err)
	}

	return &resp, nil
}

// CreditBalanceResponse represents the user's credit balance
type CreditBalanceResponse struct {
	Credits int `json:"credits"`
}

// GetCreditBalance fetches the user's current credit balance
func (c *Client) GetCreditBalance() (*CreditBalanceResponse, error) {
	url := fmt.Sprintf("%s/api/credits/balance/", c.BaseURL)

	var resp CreditBalanceResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get credit balance: %w", err)
	}

	return &resp, nil
}

// Brief Generation API Structures and Methods

// GenerateBriefRequest represents the request for brief generation
type GenerateBriefRequest struct {
	Keyword string `json:"keyword"`
}

// GenerateBriefResponse represents the response from brief generation
type GenerateBriefResponse struct {
	ID          string `json:"id"`           // Brief ID for polling
	Brief       string `json:"brief"`
	Status      string `json:"status"`
	CreditsUsed int    `json:"credits_used"`
	GeneratedAt string `json:"generated_at"`
	TotalResults int   `json:"total_results"`
}

// BriefResponse represents the response from brief status check
type BriefResponse struct {
	ID          string  `json:"id"`
	Keyword     string  `json:"keyword"`
	Brief       *string `json:"brief"` // Nullable
	Status      string  `json:"status"`
	CreditsUsed int     `json:"credits_used"`
	GeneratedAt string  `json:"generated_at"`
}

// BriefHistoryItem represents a single brief generation in history
type BriefHistoryItem struct {
	ID          string  `json:"id"`
	Keyword     string  `json:"keyword"`
	Brief       *string `json:"brief"` // Nullable
	Status      string  `json:"status"`
	CreditsUsed int     `json:"credits_used"`
	GeneratedAt string  `json:"generated_at"`
}

// BriefHistoryResponse represents the response from brief history
type BriefHistoryResponse struct {
	Briefs []BriefHistoryItem `json:"briefs"`
}

// GenerateBrief generates a brief for a given keyword
func (c *Client) GenerateBrief(keyword string) (*GenerateBriefResponse, error) {
	url := fmt.Sprintf("%s/api/brief/generate/", c.BaseURL)

	req := GenerateBriefRequest{
		Keyword: keyword,
	}

	var resp GenerateBriefResponse
	if err := c.makeRequest("POST", url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to generate brief: %w", err)
	}

	return &resp, nil
}

// GetBriefStatus retrieves the status and content of a specific brief
func (c *Client) GetBriefStatus(briefID string) (*BriefResponse, error) {
	url := fmt.Sprintf("%s/api/brief/%s/", c.BaseURL, briefID)

	var resp BriefResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get brief status: %w", err)
	}

	return &resp, nil
}

// GetBriefHistory fetches the list of user's brief generations
func (c *Client) GetBriefHistory() (*BriefHistoryResponse, error) {
	url := fmt.Sprintf("%s/api/brief/history/", c.BaseURL)

	var resp BriefHistoryResponse
	if err := c.makeRequest("GET", url, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get brief history: %w", err)
	}

	return &resp, nil
}

// makeRequest is a helper method for making HTTP requests
func (c *Client) makeRequest(method, url string, reqBody interface{}, respBody interface{}) error {
	var body io.Reader

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("User-Agent", "SEO-CLI/2.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       bodyBytes,
			Message:    string(bodyBytes),
		}
	}

	if respBody != nil {
		if err := json.Unmarshal(bodyBytes, respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
