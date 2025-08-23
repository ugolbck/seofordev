package crawler

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	pwsetup "github.com/ugolbck/seofordev/internal/playwright"
)

type PageResult struct {
	URL        string
	Content    string
	Depth      int
	StatusCode int
}

type Crawler struct {
	BaseURL        string
	Concurrency    int
	MaxPages       int
	MaxDepth       int
	IgnorePatterns []string

	// State tracking
	visited   map[string]bool
	results   []PageResult
	pageCount int

	// Queue management
	queue     chan crawlTask
	queueOpen bool

	// Robots.txt handling
	robotsRules []robotsRule
	robotsMu    sync.RWMutex

	// Synchronization
	mu      sync.RWMutex
	queueMu sync.Mutex
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc

	// Playwright resources
	pw      *playwright.Playwright
	browser playwright.Browser
}

type robotsRule struct {
	userAgent string
	disallow  []string
	allow     []string
}

type crawlTask struct {
	URL   string
	Depth int
}

func NewCrawler(baseURL string, concurrency, maxPages, maxDepth int, ignorePatterns []string) *Crawler {
	// Disable all logging output
	log.SetOutput(ioutil.Discard)

	ctx, cancel := context.WithCancel(context.Background())

	return &Crawler{
		BaseURL:        baseURL,
		Concurrency:    concurrency,
		MaxPages:       maxPages,
		MaxDepth:       maxDepth,
		IgnorePatterns: ignorePatterns,
		visited:        make(map[string]bool),
		results:        []PageResult{},
		queue:          make(chan crawlTask, 100),
		queueOpen:      true,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (c *Crawler) Start() error {
	// Initialize Playwright
	if err := c.initPlaywright(); err != nil {
		return err
	}
	defer c.cleanup()

	// Fetch and parse robots.txt
	c.fetchRobotsTxt()

	// Normalize and add base URL to queue
	normalizedBase := c.normalizeURL(c.BaseURL)
	c.addToQueue(crawlTask{URL: normalizedBase, Depth: 0})

	// Start workers
	for i := 0; i < c.Concurrency; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	// Monitor for completion
	go c.monitor()

	// Wait for all workers to complete
	c.wg.Wait()

	return nil
}

func (c *Crawler) initPlaywright() error {
	// Check if Playwright is installed (should have been installed at startup)
	if err := pwsetup.CheckPlaywrightInstalled(); err != nil {
		return fmt.Errorf("Playwright not properly installed: %w", err)
	}

	// Set up paths for Playwright
	driverDir := pwsetup.GetPlaywrightDir()
	browsersDir := filepath.Join(driverDir, "browsers")
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", browsersDir)

	// Run Playwright with custom driver directory
	runOptions := &playwright.RunOptions{
		DriverDirectory: driverDir,
	}

	var err error
	c.pw, err = playwright.Run(runOptions)
	if err != nil {
		return fmt.Errorf("could not launch Playwright: %w", err)
	}

	c.browser, err = c.pw.Chromium.Launch()
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}

	return nil
}

func (c *Crawler) cleanup() {
	if c.browser != nil {
		c.browser.Close()
	}
	if c.pw != nil {
		c.pw.Stop()
	}
}

func (c *Crawler) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case task, ok := <-c.queue:
			if !ok {
				return // Queue closed
			}
			c.crawlPage(task.URL, task.Depth)
		}
	}
}

func (c *Crawler) monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			pageCount := c.pageCount
			c.mu.RUnlock()

			c.queueMu.Lock()
			queueLen := len(c.queue)
			queueOpen := c.queueOpen
			c.queueMu.Unlock()

			// Check if we've reached max pages
			if c.MaxPages > 0 && pageCount >= c.MaxPages {
				c.stop()
				return
			}

			// Check if queue is empty and we're done
			if queueLen == 0 && queueOpen {
				c.stop()
				return
			}
		}
	}
}

func (c *Crawler) stop() {
	c.queueMu.Lock()
	if c.queueOpen {
		c.queueOpen = false
		close(c.queue)
	}
	c.queueMu.Unlock()

	c.cancel()
}

func (c *Crawler) addToQueue(task crawlTask) bool {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()

	if !c.queueOpen {
		return false
	}

	select {
	case c.queue <- task:
		return true
	case <-c.ctx.Done():
		return false
	default:
		// Queue full - silently skip
		return false
	}
}

func (c *Crawler) crawlPage(pageURL string, depth int) {
	normalizedURL := c.normalizeURL(pageURL)

	c.mu.Lock()
	// Check if already visited
	if c.visited[normalizedURL] {
		c.mu.Unlock()
		return
	}

	// Check if we've reached max pages
	if c.MaxPages > 0 && c.pageCount >= c.MaxPages {
		c.mu.Unlock()
		return
	}

	// Check if URL should be ignored
	if c.shouldIgnore(normalizedURL) {
		c.visited[normalizedURL] = true
		c.mu.Unlock()
		return
	}

	// Check robots.txt rules
	if c.isDisallowedByRobots(normalizedURL) {
		c.visited[normalizedURL] = true
		c.mu.Unlock()
		return
	}

	// Mark as visited and increment page count
	c.visited[normalizedURL] = true
	c.pageCount++
	c.mu.Unlock()

	// Create page with timeout
	page, err := c.browser.NewPage()
	if err != nil {
		return
	}
	defer page.Close()

	// Navigate to page
	ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
	defer cancel()

	response, err := page.Goto(normalizedURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   playwright.Float(15000),
	})

	// Default status code for errors
	statusCode := 0
	if err != nil {
		// Store failed page
		c.mu.Lock()
		c.results = append(c.results, PageResult{
			URL:        normalizedURL,
			Content:    "",
			Depth:      depth,
			StatusCode: statusCode,
		})
		c.mu.Unlock()
		return
	}

	// Get status code from response
	if response != nil {
		statusCode = response.Status()
	}

	// Get page content
	content, err := page.Content()
	if err != nil {
		content = "" // Empty content but still record the page with its status code
	}

	// Store result
	c.mu.Lock()
	c.results = append(c.results, PageResult{
		URL:        normalizedURL,
		Content:    content,
		Depth:      depth,
		StatusCode: statusCode,
	})
	c.mu.Unlock()

	// Discover links if we haven't reached max depth
	if c.MaxDepth == 0 || depth < c.MaxDepth {
		c.discoverLinks(page, normalizedURL, depth, ctx)
	}
}

func (c *Crawler) discoverLinks(page playwright.Page, pageURL string, depth int, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-c.ctx.Done():
		return
	default:
	}

	anchors, err := page.QuerySelectorAll("a")
	if err != nil {
		return
	}

	linksFound := 0
	linksQueued := 0

	for _, a := range anchors {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		default:
		}

		href, _ := a.GetAttribute("href")
		if href == "" {
			continue
		}

		abs := resolveURL(pageURL, href)
		if abs == "" {
			continue
		}

		normalizedAbs := c.normalizeURL(abs)
		linksFound++

		// Check if URL is valid for crawling
		if c.isSameHost(normalizedAbs) && !c.shouldIgnore(normalizedAbs) && !c.isDisallowedByRobots(normalizedAbs) {
			c.mu.RLock()
			alreadyVisited := c.visited[normalizedAbs]
			c.mu.RUnlock()

			if !alreadyVisited {
				if c.addToQueue(crawlTask{URL: normalizedAbs, Depth: depth + 1}) {
					linksQueued++
				}
			}
		}
	}
}

// normalizeURL removes fragments and trailing slashes, and converts to lowercase
func (c *Crawler) normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	parsed.Fragment = ""

	// Handle root path consistently - always use "/" for root
	if parsed.Path == "" {
		parsed.Path = "/"
	} else if parsed.Path != "/" && strings.HasSuffix(parsed.Path, "/") {
		// Remove trailing slash from non-root paths
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	// Convert host to lowercase
	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String()
}

func (c *Crawler) shouldIgnore(url string) bool {
	if len(c.IgnorePatterns) == 0 {
		return false
	}

	for _, pattern := range c.IgnorePatterns {
		// Try as regex first
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			regex := strings.Trim(pattern, "/")
			if matched, _ := regexp.MatchString(regex, url); matched {
				return true
			}
		}

		// Try as simple string contains
		if strings.Contains(url, pattern) {
			return true
		}
	}

	return false
}

func resolveURL(base, href string) string {
	if href == "" {
		return ""
	}

	// Skip javascript: and mailto: links
	if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
		return ""
	}

	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if u.IsAbs() {
		return u.String()
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}

func (c *Crawler) isSameHost(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	baseParsed, err := url.Parse(c.BaseURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, baseParsed.Host)
}

// GetResults returns a copy of the results (thread-safe)
func (c *Crawler) GetResults() []PageResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]PageResult, len(c.results))
	copy(results, c.results)
	return results
}

// GetStats returns current crawling statistics
func (c *Crawler) GetStats() (visited int, queued int, results int) {
	c.mu.RLock()
	visited = len(c.visited)
	results = len(c.results)
	c.mu.RUnlock()

	c.queueMu.Lock()
	queued = len(c.queue)
	c.queueMu.Unlock()

	return
}

// fetchRobotsTxt fetches and parses robots.txt from the base URL
func (c *Crawler) fetchRobotsTxt() {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return // Skip robots.txt if base URL is invalid
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", baseURL.Scheme, baseURL.Host)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(robotsURL)
	if err != nil {
		return // No robots.txt or error - allow all
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return // No robots.txt - allow all
	}

	c.parseRobotsTxt(resp)
}

// parseRobotsTxt parses robots.txt content and stores rules
func (c *Crawler) parseRobotsTxt(resp *http.Response) {
	scanner := bufio.NewScanner(resp.Body)
	var currentRule *robotsRule

	c.robotsMu.Lock()
	defer c.robotsMu.Unlock()

	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first colon
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case "user-agent":
			// Save previous rule if exists
			if currentRule != nil {

				c.robotsRules = append(c.robotsRules, *currentRule)
			}
			// Start new rule
			currentRule = &robotsRule{
				userAgent: strings.ToLower(value),
				disallow:  []string{},
				allow:     []string{},
			}

		case "disallow":
			if currentRule != nil && value != "" {
				currentRule.disallow = append(currentRule.disallow, value)
			}

		case "allow":
			if currentRule != nil && value != "" {
				currentRule.allow = append(currentRule.allow, value)
			}
		}
	}

	// Save last rule
	if currentRule != nil {

		c.robotsRules = append(c.robotsRules, *currentRule)
	}
}

// isDisallowedByRobots checks if a URL is disallowed by robots.txt
func (c *Crawler) isDisallowedByRobots(urlStr string) bool {
	c.robotsMu.RLock()
	defer c.robotsMu.RUnlock()

	if len(c.robotsRules) == 0 {
		return false // No robots.txt rules - allow all
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false // Invalid URL - allow
	}

	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	// Check rules for Googlebot first, then * (all crawlers)
	userAgents := []string{"googlebot", "*"}

	for _, ua := range userAgents {
		for _, rule := range c.robotsRules {
			if rule.userAgent != ua {
				continue
			}
			// Find the longest matching pattern (most specific wins)
			var longestMatch string
			var isDisallowed bool

			// Check all Allow patterns
			for _, allowPattern := range rule.allow {
				if c.matchesRobotsPattern(path, allowPattern) && len(allowPattern) > len(longestMatch) {
					longestMatch = allowPattern
					isDisallowed = false
				}
			}

			// Check all Disallow patterns
			for _, disallowPattern := range rule.disallow {
				if c.matchesRobotsPattern(path, disallowPattern) && len(disallowPattern) > len(longestMatch) {
					longestMatch = disallowPattern
					isDisallowed = true
				}
			}

			// If we found a matching pattern, use it
			if longestMatch != "" {
				if isDisallowed {
					return true
				} else {
					return false
				}
			}
		}
	}

	return false // Default allow if no matching rules
}

// matchesRobotsPattern checks if a path matches a robots.txt pattern
func (c *Crawler) matchesRobotsPattern(path, pattern string) bool {
	// Empty pattern means match nothing
	if pattern == "" {
		return false
	}

	// Simple prefix matching (most common case)
	if strings.HasPrefix(path, pattern) {
		return true
	}

	// Handle wildcards - convert robots.txt pattern to regex
	regexPattern := regexp.QuoteMeta(pattern)
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")
	regexPattern = "^" + regexPattern

	matched, err := regexp.MatchString(regexPattern, path)
	if err != nil {
		// Fall back to simple prefix matching if regex fails
		return strings.HasPrefix(path, pattern)
	}

	return matched
}
