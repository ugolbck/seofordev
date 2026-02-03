package export

// SEOCheckResponse represents an individual SEO check result
type SEOCheckResponse struct {
	CheckName string      `json:"check_name"`
	Passed    bool        `json:"passed"`
	Value     interface{} `json:"value,omitempty"`
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
