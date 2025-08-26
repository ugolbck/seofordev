package audit

import (
	"fmt"
	"math"
	"net/url"
	"strings"
)

// CheckResult represents the result of an individual SEO check
type CheckResult struct {
	Passed  bool        `json:"passed"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Weight  int         `json:"weight"`
}

// CheckResults represents all SEO check results
type CheckResults struct {
	Indexable          bool                   `json:"indexable"`
	IndexabilityReason string                 `json:"indexability_reason,omitempty"`
	Score              float64                `json:"score"`
	Checks             map[string]CheckResult `json:"checks"`
}

// Checker performs comprehensive SEO checks on analyzed page data
type Checker struct {
	analysis   *AnalysisResult
	statusCode int
	results    map[string]CheckResult
	weights    map[string]int
}

// NewChecker creates a new SEO checker
func NewChecker(analysis *AnalysisResult, statusCode int) *Checker {
	return &Checker{
		analysis:   analysis,
		statusCode: statusCode,
		results:    make(map[string]CheckResult),
		weights: map[string]int{
			"response_status_code":      95,
			"title_presence":            85,
			"title_length":              75,
			"unique_title_tag":          70,
			"meta_description_presence": 80,
			"meta_description_length":   65,
			"unique_meta_description":   60,
			"h1_presence":               90,
			"unique_h1_heading":         85,
			"h1_length":                 70,
			"h2_presence":               50,
			"content_length":            75,
			"canonical_url_presence":    55,
			"url_matches_canonical":     50,
			"unique_canonical_link":     45,
			"meta_robots_indexing":      65,
			"outlinks_count":            35,
			"external_links_count":      30,
			"missing_alt_attribute":     55,
			"meta_refresh_redirect":     25,
			"viewport_meta":             40,
			"charset_declared":          35,
			"images_optimization":       45,
			"structured_data":           35,
			"page_loading_speed":        60,
			"social_media_meta":         25,
		},
	}
}

// RunAllChecks performs all SEO checks and returns comprehensive results
func (c *Checker) RunAllChecks() *CheckResults {
	// Check indexability first
	if !c.checkIndexability() {
		return &CheckResults{
			Indexable:          false,
			IndexabilityReason: c.getIndexabilityReason(),
			Score:              0,
			Checks:             make(map[string]CheckResult),
		}
	}

	// Run all individual checks
	c.checkResponseStatusCode()
	c.checkTitlePresence()
	c.checkTitleLength()
	c.checkUniqueTitleTag()
	c.checkMetaDescriptionPresence()
	c.checkMetaDescriptionLength()
	c.checkUniqueMetaDescription()
	c.checkH1Presence()
	c.checkUniqueH1Heading()
	c.checkH1Length()
	c.checkH2Presence()
	c.checkContentLength()
	c.checkCanonicalURLPresence()
	c.checkURLMatchesCanonical()
	c.checkUniqueCanonicalLink()
	c.checkMetaRobotsIndexing()
	c.checkOutlinksCount()
	c.checkExternalLinksCount()
	c.checkMissingAltAttribute()
	c.checkMetaRefreshRedirect()
	c.checkViewportMeta()
	c.checkCharsetDeclared()
	c.checkImagesOptimization()
	c.checkStructuredData()
	c.checkPageLoadingSpeed()
	c.checkSocialMediaMeta()

	// Calculate overall score
	score := c.calculateScore()

	return &CheckResults{
		Indexable:          true,
		IndexabilityReason: "",
		Score:              score,
		Checks:             c.results,
	}
}

// checkIndexability determines if the page is indexable by search engines
func (c *Checker) checkIndexability() bool {
	if c.statusCode != 200 {
		return false
	}

	if c.analysis.Robots != nil && c.analysis.Robots.NoIndex {
		return false
	}

	return true
}

// getIndexabilityReason returns the reason why the page is not indexable
func (c *Checker) getIndexabilityReason() string {
	if c.statusCode != 200 {
		return fmt.Sprintf("HTTP %d - Page not accessible", c.statusCode)
	}

	if c.analysis.Robots != nil && c.analysis.Robots.NoIndex {
		return "Meta robots noindex directive"
	}

	return ""
}

// Individual check methods

func (c *Checker) checkResponseStatusCode() {
	passed := c.statusCode == 200
	message := "Page returns HTTP 200"
	if !passed {
		message = fmt.Sprintf("Page returns HTTP %d", c.statusCode)
	}

	c.results["response_status_code"] = CheckResult{
		Passed:  passed,
		Value:   c.statusCode,
		Message: message,
		Weight:  c.weights["response_status_code"],
	}
}

func (c *Checker) checkTitlePresence() {
	title := strings.TrimSpace(c.analysis.Title)
	passed := title != ""
	message := "Page has a title tag"
	if !passed {
		message = "Page is missing a title tag"
	}

	c.results["title_presence"] = CheckResult{
		Passed:  passed,
		Value:   title,
		Message: message,
		Weight:  c.weights["title_presence"],
	}
}

func (c *Checker) checkTitleLength() {
	titleLength := c.analysis.Content.TitleLength
	passed := titleLength >= 25 && titleLength <= 65
	message := fmt.Sprintf("Title length is %d characters (optimal: 25-65)", titleLength)
	if !passed {
		if titleLength < 25 {
			message = fmt.Sprintf("Title is too short (%d chars). Consider 25-65 characters", titleLength)
		} else {
			message = fmt.Sprintf("Title is too long (%d chars). Consider 25-65 characters", titleLength)
		}
	}

	c.results["title_length"] = CheckResult{
		Passed:  passed,
		Value:   titleLength,
		Message: message,
		Weight:  c.weights["title_length"],
	}
}

func (c *Checker) checkUniqueTitleTag() {
	// For single page analysis, always pass (uniqueness requires multiple pages)
	passed := true
	message := "Title appears to be unique (single page analysis)"

	c.results["unique_title_tag"] = CheckResult{
		Passed:  passed,
		Value:   c.analysis.Title,
		Message: message,
		Weight:  c.weights["unique_title_tag"],
	}
}

func (c *Checker) checkMetaDescriptionPresence() {
	description := strings.TrimSpace(c.analysis.Description)
	passed := description != ""
	message := "Page has a meta description"
	if !passed {
		message = "Page is missing a meta description"
	}

	c.results["meta_description_presence"] = CheckResult{
		Passed:  passed,
		Value:   description,
		Message: message,
		Weight:  c.weights["meta_description_presence"],
	}
}

func (c *Checker) checkMetaDescriptionLength() {
	descLength := c.analysis.Content.DescriptionLength
	passed := descLength >= 110 && descLength <= 155
	message := fmt.Sprintf("Meta description length is %d characters (optimal: 110-155)", descLength)
	if !passed {
		if descLength < 110 {
			message = fmt.Sprintf("Meta description is too short (%d chars). Consider 110-155 characters", descLength)
		} else if descLength > 155 {
			message = fmt.Sprintf("Meta description is too long (%d chars). Consider 110-155 characters", descLength)
		}
	}

	c.results["meta_description_length"] = CheckResult{
		Passed:  passed,
		Value:   descLength,
		Message: message,
		Weight:  c.weights["meta_description_length"],
	}
}

func (c *Checker) checkUniqueMetaDescription() {
	// For single page analysis, always pass (uniqueness requires multiple pages)
	passed := true
	message := "Meta description appears to be unique (single page analysis)"

	c.results["unique_meta_description"] = CheckResult{
		Passed:  passed,
		Value:   c.analysis.Description,
		Message: message,
		Weight:  c.weights["unique_meta_description"],
	}
}

func (c *Checker) checkH1Presence() {
	hasH1 := len(c.analysis.H1) > 0
	message := "Page has an H1 heading"
	if !hasH1 {
		message = "Page is missing an H1 heading"
	}

	value := ""
	if len(c.analysis.H1) > 0 {
		value = c.analysis.H1[0]
	}

	c.results["h1_presence"] = CheckResult{
		Passed:  hasH1,
		Value:   value,
		Message: message,
		Weight:  c.weights["h1_presence"],
	}
}

func (c *Checker) checkUniqueH1Heading() {
	h1Count := c.analysis.Headings.H1Count
	passed := h1Count == 1
	message := "Page has exactly 1 H1 heading"
	if !passed {
		if h1Count == 0 {
			message = "Page has no H1 heading"
		} else {
			message = fmt.Sprintf("Page has %d H1 headings (should have exactly 1)", h1Count)
		}
	}

	c.results["unique_h1_heading"] = CheckResult{
		Passed:  passed,
		Value:   h1Count,
		Message: message,
		Weight:  c.weights["unique_h1_heading"],
	}
}

func (c *Checker) checkH1Length() {
	if len(c.analysis.H1) == 0 {
		c.results["h1_length"] = CheckResult{
			Passed:  false,
			Value:   0,
			Message: "No H1 heading to check length",
			Weight:  c.weights["h1_length"],
		}
		return
	}

	h1Length := len(c.analysis.H1[0])
	passed := h1Length >= 15 && h1Length <= 65
	message := fmt.Sprintf("H1 length is %d characters (optimal: 15-65)", h1Length)
	if !passed {
		if h1Length < 15 {
			message = fmt.Sprintf("H1 is too short (%d chars). Consider 15-65 characters", h1Length)
		} else {
			message = fmt.Sprintf("H1 is too long (%d chars). Consider 15-65 characters", h1Length)
		}
	}

	c.results["h1_length"] = CheckResult{
		Passed:  passed,
		Value:   h1Length,
		Message: message,
		Weight:  c.weights["h1_length"],
	}
}

func (c *Checker) checkH2Presence() {
	hasH2 := c.analysis.Headings.H2Count > 0
	message := fmt.Sprintf("Page has %d H2 headings", c.analysis.Headings.H2Count)
	if !hasH2 {
		message = "Page has no H2 headings (consider adding for structure)"
	}

	c.results["h2_presence"] = CheckResult{
		Passed:  hasH2,
		Value:   c.analysis.Headings.H2Count,
		Message: message,
		Weight:  c.weights["h2_presence"],
	}
}

func (c *Checker) checkContentLength() {
	wordCount := c.analysis.Content.WordCount
	passed := wordCount >= 250
	message := fmt.Sprintf("Page has %d words of content", wordCount)
	if !passed {
		message = fmt.Sprintf("Page has only %d words (consider 250+ for better SEO)", wordCount)
	}

	c.results["content_length"] = CheckResult{
		Passed:  passed,
		Value:   wordCount,
		Message: message,
		Weight:  c.weights["content_length"],
	}
}

func (c *Checker) checkCanonicalURLPresence() {
	hasCanonical := c.analysis.Technical.Canonical != ""
	message := "Page has a canonical URL"
	if !hasCanonical {
		message = "Page is missing a canonical URL"
	}

	c.results["canonical_url_presence"] = CheckResult{
		Passed:  hasCanonical,
		Value:   c.analysis.Technical.Canonical,
		Message: message,
		Weight:  c.weights["canonical_url_presence"],
	}
}

func (c *Checker) checkURLMatchesCanonical() {
	canonical := c.analysis.Technical.Canonical
	if canonical == "" {
		c.results["url_matches_canonical"] = CheckResult{
			Passed:  false,
			Value:   "",
			Message: "No canonical URL to compare",
			Weight:  c.weights["url_matches_canonical"],
		}
		return
	}

	// Parse URLs for comparison
	pageURL, err1 := url.Parse(c.analysis.URL)
	canonicalURL, err2 := url.Parse(canonical)

	passed := false
	message := "Canonical URL does not match page URL"

	if err1 == nil && err2 == nil {
		// Compare normalized URLs (ignore trailing slashes, fragments)
		pageNorm := strings.TrimSuffix(pageURL.String(), "/")
		canonicalNorm := strings.TrimSuffix(canonicalURL.String(), "/")
		passed = pageNorm == canonicalNorm
		if passed {
			message = "Canonical URL matches page URL"
		}
	}

	c.results["url_matches_canonical"] = CheckResult{
		Passed:  passed,
		Value:   canonical,
		Message: message,
		Weight:  c.weights["url_matches_canonical"],
	}
}

func (c *Checker) checkUniqueCanonicalLink() {
	// For single page analysis, always pass (uniqueness requires multiple pages)
	passed := true
	message := "Canonical URL appears to be unique (single page analysis)"

	c.results["unique_canonical_link"] = CheckResult{
		Passed:  passed,
		Value:   c.analysis.Technical.Canonical,
		Message: message,
		Weight:  c.weights["unique_canonical_link"],
	}
}

func (c *Checker) checkMetaRobotsIndexing() {
	passed := c.analysis.Robots == nil || !c.analysis.Robots.NoIndex
	message := "Page allows indexing"
	if !passed {
		message = "Page has noindex directive"
	}

	c.results["meta_robots_indexing"] = CheckResult{
		Passed:  passed,
		Value:   c.analysis.Robots != nil && c.analysis.Robots.NoIndex,
		Message: message,
		Weight:  c.weights["meta_robots_indexing"],
	}
}

func (c *Checker) checkOutlinksCount() {
	internalCount := c.analysis.Links.InternalCount
	passed := internalCount >= 1 && internalCount <= 100
	message := fmt.Sprintf("Page has %d internal links", internalCount)
	if !passed {
		if internalCount == 0 {
			message = "Page has no internal links (consider adding for better navigation)"
		} else if internalCount > 100 {
			message = fmt.Sprintf("Page has %d internal links (consider reducing)", internalCount)
		}
	}

	c.results["outlinks_count"] = CheckResult{
		Passed:  passed,
		Value:   internalCount,
		Message: message,
		Weight:  c.weights["outlinks_count"],
	}
}

func (c *Checker) checkExternalLinksCount() {
	externalCount := c.analysis.Links.ExternalCount
	passed := externalCount <= 15
	message := fmt.Sprintf("Page has %d external links", externalCount)
	if !passed {
		message = fmt.Sprintf("Page has %d external links (consider reducing)", externalCount)
	}

	c.results["external_links_count"] = CheckResult{
		Passed:  passed,
		Value:   externalCount,
		Message: message,
		Weight:  c.weights["external_links_count"],
	}
}

func (c *Checker) checkMissingAltAttribute() {
	totalImages := c.analysis.Images.TotalCount
	missingAlt := c.analysis.Images.WithoutAltCount
	passed := missingAlt == 0
	message := fmt.Sprintf("All %d images have alt attributes", totalImages)
	if !passed {
		if totalImages == 0 {
			message = "Page has no images to check"
			passed = true
		} else {
			message = fmt.Sprintf("%d of %d images are missing alt attributes", missingAlt, totalImages)
		}
	}

	c.results["missing_alt_attribute"] = CheckResult{
		Passed:  passed,
		Value:   missingAlt,
		Message: message,
		Weight:  c.weights["missing_alt_attribute"],
	}
}

func (c *Checker) checkMetaRefreshRedirect() {
	hasMetaRefresh := c.analysis.Technical.MetaRefresh
	passed := !hasMetaRefresh
	message := "Page does not use meta refresh redirects"
	if !passed {
		message = "Page uses meta refresh redirect (not recommended for SEO)"
	}

	c.results["meta_refresh_redirect"] = CheckResult{
		Passed:  passed,
		Value:   hasMetaRefresh,
		Message: message,
		Weight:  c.weights["meta_refresh_redirect"],
	}
}

func (c *Checker) checkViewportMeta() {
	hasViewport := c.analysis.Technical.ViewportMeta
	message := "Page has viewport meta tag"
	if !hasViewport {
		message = "Page is missing viewport meta tag (important for mobile)"
	}

	c.results["viewport_meta"] = CheckResult{
		Passed:  hasViewport,
		Value:   hasViewport,
		Message: message,
		Weight:  c.weights["viewport_meta"],
	}
}

func (c *Checker) checkCharsetDeclared() {
	hasCharset := c.analysis.Technical.CharsetDeclared
	message := "Page declares charset"
	if !hasCharset {
		message = "Page is missing charset declaration"
	}

	c.results["charset_declared"] = CheckResult{
		Passed:  hasCharset,
		Value:   hasCharset,
		Message: message,
		Weight:  c.weights["charset_declared"],
	}
}

func (c *Checker) checkImagesOptimization() {
	totalImages := c.analysis.Images.TotalCount
	withAlt := c.analysis.Images.WithAltCount

	passed := true
	message := "Images appear to be optimized"

	if totalImages > 0 {
		altRatio := float64(withAlt) / float64(totalImages)
		passed = altRatio >= 0.8 // 80% of images should have alt text
		if !passed {
			message = fmt.Sprintf("Only %.0f%% of images have alt text (aim for 80%+)", altRatio*100)
		}
	} else {
		message = "No images to optimize"
	}

	c.results["images_optimization"] = CheckResult{
		Passed:  passed,
		Value:   totalImages,
		Message: message,
		Weight:  c.weights["images_optimization"],
	}
}

func (c *Checker) checkStructuredData() {
	hasStructuredData := c.analysis.Schema.HasStructuredData
	message := "Page has structured data"
	if !hasStructuredData {
		message = "Page is missing structured data (JSON-LD or microdata)"
	}

	c.results["structured_data"] = CheckResult{
		Passed:  hasStructuredData,
		Value:   hasStructuredData,
		Message: message,
		Weight:  c.weights["structured_data"],
	}
}

func (c *Checker) checkPageLoadingSpeed() {
	// Simple heuristic based on content size and images
	wordCount := c.analysis.Content.WordCount
	imageCount := c.analysis.Images.TotalCount
	
	// Estimate load speed based on content complexity
	loadScore := 100.0
	if wordCount > 2000 {
		loadScore -= 20 // Heavy text content
	}
	if imageCount > 10 {
		loadScore -= 25 // Many images slow loading
	}
	if wordCount < 100 {
		loadScore -= 15 // Very light content might lack substance
	}
	
	passed := loadScore >= 70
	message := fmt.Sprintf("Page loading characteristics score: %.0f/100", loadScore)
	if !passed {
		message = fmt.Sprintf("Page may load slowly (score: %.0f/100)", loadScore)
	}

	c.results["page_loading_speed"] = CheckResult{
		Passed:  passed,
		Value:   loadScore,
		Message: message,
		Weight:  c.weights["page_loading_speed"],
	}
}

func (c *Checker) checkSocialMediaMeta() {
	// Check for Open Graph or Twitter Card meta tags in the Meta map
	hasOG := false
	hasTwitter := false
	
	if c.analysis.Meta != nil {
		for key := range c.analysis.Meta {
			keyLower := strings.ToLower(key)
			if strings.Contains(keyLower, "og:") {
				hasOG = true
			}
			if strings.Contains(keyLower, "twitter:") {
				hasTwitter = true
			}
		}
	}
	
	passed := hasOG || hasTwitter
	message := "Page has social media meta tags (Open Graph or Twitter Cards)"
	if !passed {
		message = "Page is missing social media meta tags (consider adding Open Graph or Twitter Cards)"
	}

	c.results["social_media_meta"] = CheckResult{
		Passed:  passed,
		Value:   hasOG || hasTwitter,
		Message: message,
		Weight:  c.weights["social_media_meta"],
	}
}

// calculateScore calculates the overall SEO score (0-100)
func (c *Checker) calculateScore() float64 {
	totalWeightedScore := 0.0
	totalPossibleScore := 0.0

	for _, result := range c.results {
		weight := float64(result.Weight)
		if result.Passed {
			totalWeightedScore += weight
		}
		totalPossibleScore += weight
	}

	if totalPossibleScore == 0 {
		return 0
	}

	score := (totalWeightedScore / totalPossibleScore) * 100
	return math.Round(score*10) / 10 // Round to 1 decimal place
}
