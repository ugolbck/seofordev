package audit

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// AnalysisResult represents the result of analyzing a single page
type AnalysisResult struct {
	URL         string                 `json:"url"`
	ParsedURL   *ParsedURL             `json:"parsed_url"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	H1          []string               `json:"h1"`
	H2          []string               `json:"h2"`
	Headings    *HeadingData           `json:"headings"`
	Content     *ContentData           `json:"content"`
	Links       *LinkData              `json:"links"`
	Images      *ImageData             `json:"images"`
	Technical   *TechnicalData         `json:"technical"`
	Robots      *RobotsData            `json:"robots"`
	Schema      *SchemaData            `json:"schema"`
	Language    string                 `json:"language"`
	Meta        map[string]interface{} `json:"meta"`
}

// ParsedURL represents URL components
type ParsedURL struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path"`
	Query    string `json:"query,omitempty"`
	Fragment string `json:"fragment,omitempty"`
}

// HeadingData represents heading analysis
type HeadingData struct {
	H1Count int      `json:"h1_count"`
	H2Count int      `json:"h2_count"`
	H3Count int      `json:"h3_count"`
	H4Count int      `json:"h4_count"`
	H5Count int      `json:"h5_count"`
	H6Count int      `json:"h6_count"`
	H1      []string `json:"h1"`
	H2      []string `json:"h2"`
}

// ContentData represents content analysis
type ContentData struct {
	WordCount         int    `json:"word_count"`
	TitleLength       int    `json:"title_length"`
	DescriptionLength int    `json:"description_length"`
	TextContent       string `json:"text_content,omitempty"`
}

// LinkData represents link analysis
type LinkData struct {
	InternalCount int        `json:"internal_count"`
	ExternalCount int        `json:"external_count"`
	TotalCount    int        `json:"total_count"`
	Internal      []LinkInfo `json:"internal,omitempty"`
	External      []LinkInfo `json:"external,omitempty"`
}

// LinkInfo represents individual link information
type LinkInfo struct {
	URL        string `json:"url"`
	AnchorText string `json:"anchor_text"`
	NoFollow   bool   `json:"nofollow"`
}

// ImageData represents image analysis
type ImageData struct {
	TotalCount     int `json:"total_count"`
	WithoutAltCount int `json:"without_alt_count"`
	WithAltCount   int `json:"with_alt_count"`
}

// TechnicalData represents technical SEO elements
type TechnicalData struct {
	Canonical        string `json:"canonical"`
	ViewportMeta     bool   `json:"viewport_meta"`
	CharsetDeclared  bool   `json:"charset_declared"`
	MetaRefresh      bool   `json:"meta_refresh"`
	OpenGraph        bool   `json:"open_graph"`
	TwitterCard      bool   `json:"twitter_card"`
}

// RobotsData represents robots meta directives
type RobotsData struct {
	NoIndex  bool `json:"noindex"`
	NoFollow bool `json:"nofollow"`
	NoCache  bool `json:"nocache"`
}

// SchemaData represents structured data analysis
type SchemaData struct {
	HasStructuredData bool     `json:"has_structured_data"`
	Types             []string `json:"types,omitempty"`
}

// Analyzer performs SEO analysis on HTML content
type Analyzer struct{}

// NewAnalyzer creates a new SEO analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeContent performs comprehensive SEO analysis on HTML content
func (a *Analyzer) AnalyzeContent(htmlContent string, pageURL string) (*AnalysisResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &AnalysisResult{
		URL:       pageURL,
		ParsedURL: a.parseURL(pageURL),
		Meta:      make(map[string]interface{}),
	}

	// Extract all SEO elements
	a.extractMetaData(doc, result)
	a.extractHeadings(doc, result)
	a.extractContent(doc, result)
	a.extractLinks(doc, pageURL, result)
	a.extractImages(doc, result)
	a.extractTechnicalSEO(doc, result)

	return result, nil
}

// parseURL parses URL components
func (a *Analyzer) parseURL(rawURL string) *ParsedURL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return &ParsedURL{}
	}

	return &ParsedURL{
		Scheme:   parsed.Scheme,
		Host:     parsed.Host,
		Port:     parsed.Port(),
		Path:     parsed.Path,
		Query:    parsed.RawQuery,
		Fragment: parsed.Fragment,
	}
}

// extractMetaData extracts title, description, and other meta elements
func (a *Analyzer) extractMetaData(doc *goquery.Document, result *AnalysisResult) {
	// Title
	result.Title = strings.TrimSpace(doc.Find("title").First().Text())

	// Meta description
	desc, _ := doc.Find(`meta[name="description"]`).First().Attr("content")
	result.Description = strings.TrimSpace(desc)

	// Other meta tags
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				result.Meta[name] = content
			}
		}
		if property, exists := s.Attr("property"); exists {
			if content, exists := s.Attr("content"); exists {
				result.Meta[property] = content
			}
		}
	})

	// Robots meta
	result.Robots = a.extractRobotsMeta(doc)
}

// extractHeadings extracts heading structure
func (a *Analyzer) extractHeadings(doc *goquery.Document, result *AnalysisResult) {
	headings := &HeadingData{
		H1: make([]string, 0),
		H2: make([]string, 0),
	}

	// Extract H1s
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			headings.H1 = append(headings.H1, text)
			headings.H1Count++
		}
	})

	// Extract H2s
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			headings.H2 = append(headings.H2, text)
			headings.H2Count++
		}
	})

	// Count other headings
	headings.H3Count = doc.Find("h3").Length()
	headings.H4Count = doc.Find("h4").Length()
	headings.H5Count = doc.Find("h5").Length()
	headings.H6Count = doc.Find("h6").Length()

	result.Headings = headings
	result.H1 = headings.H1
	result.H2 = headings.H2
}

// extractContent extracts and analyzes text content
func (a *Analyzer) extractContent(doc *goquery.Document, result *AnalysisResult) {
	// Remove script and style elements
	doc.Find("script, style, nav, header, footer").Remove()

	// Extract text content
	textContent := strings.TrimSpace(doc.Find("body").Text())
	
	// Clean up whitespace
	re := regexp.MustCompile(`\s+`)
	textContent = re.ReplaceAllString(textContent, " ")

	// Count words
	words := strings.Fields(textContent)
	wordCount := len(words)

	result.Content = &ContentData{
		WordCount:         wordCount,
		TitleLength:       len(result.Title),
		DescriptionLength: len(result.Description),
		TextContent:       textContent,
	}
}

// extractLinks extracts and analyzes links
func (a *Analyzer) extractLinks(doc *goquery.Document, pageURL string, result *AnalysisResult) {
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		result.Links = &LinkData{}
		return
	}

	links := &LinkData{
		Internal: make([]LinkInfo, 0),
		External: make([]LinkInfo, 0),
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		// Skip javascript: and mailto: links
		if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
			return
		}

		// Resolve relative URLs
		linkURL, err := baseURL.Parse(href)
		if err != nil {
			return
		}

		anchorText := strings.TrimSpace(s.Text())
		rel, _ := s.Attr("rel")
		nofollow := strings.Contains(strings.ToLower(rel), "nofollow")

		linkInfo := LinkInfo{
			URL:        linkURL.String(),
			AnchorText: anchorText,
			NoFollow:   nofollow,
		}

		// Determine if internal or external
		if strings.EqualFold(linkURL.Host, baseURL.Host) {
			links.Internal = append(links.Internal, linkInfo)
			links.InternalCount++
		} else {
			links.External = append(links.External, linkInfo)
			links.ExternalCount++
		}
		links.TotalCount++
	})

	result.Links = links
}

// extractImages analyzes images and alt attributes
func (a *Analyzer) extractImages(doc *goquery.Document, result *AnalysisResult) {
	images := &ImageData{}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		images.TotalCount++
		
		if alt, exists := s.Attr("alt"); exists && strings.TrimSpace(alt) != "" {
			images.WithAltCount++
		} else {
			images.WithoutAltCount++
		}
	})

	result.Images = images
}

// extractTechnicalSEO extracts technical SEO elements
func (a *Analyzer) extractTechnicalSEO(doc *goquery.Document, result *AnalysisResult) {
	technical := &TechnicalData{}

	// Canonical URL
	if canonical, exists := doc.Find(`link[rel="canonical"]`).First().Attr("href"); exists {
		technical.Canonical = canonical
	}

	// Viewport meta
	technical.ViewportMeta = doc.Find(`meta[name="viewport"]`).Length() > 0

	// Charset
	technical.CharsetDeclared = doc.Find(`meta[charset]`).Length() > 0 || 
								doc.Find(`meta[http-equiv="Content-Type"]`).Length() > 0

	// Meta refresh
	technical.MetaRefresh = doc.Find(`meta[http-equiv="refresh"]`).Length() > 0

	// Open Graph
	technical.OpenGraph = doc.Find(`meta[property^="og:"]`).Length() > 0

	// Twitter Card
	technical.TwitterCard = doc.Find(`meta[name^="twitter:"]`).Length() > 0

	result.Technical = technical

	// Structured data (basic detection)
	result.Schema = &SchemaData{
		HasStructuredData: doc.Find(`script[type="application/ld+json"]`).Length() > 0 ||
						  doc.Find(`[itemscope]`).Length() > 0,
	}

	// Language detection (basic)
	if lang, exists := doc.Find("html").Attr("lang"); exists {
		result.Language = lang
	}
}

// extractRobotsMeta extracts robots meta directives
func (a *Analyzer) extractRobotsMeta(doc *goquery.Document) *RobotsData {
	robots := &RobotsData{}

	robotsContent, _ := doc.Find(`meta[name="robots"]`).First().Attr("content")
	if robotsContent != "" {
		content := strings.ToLower(robotsContent)
		robots.NoIndex = strings.Contains(content, "noindex")
		robots.NoFollow = strings.Contains(content, "nofollow")
		robots.NoCache = strings.Contains(content, "nocache") || strings.Contains(content, "noarchive")
	}

	return robots
}