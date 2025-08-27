package export

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/ugolbck/seofordev/internal/api"
)

// ExportToClipboard copies the given text to the system clipboard
func ExportToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// extractPathFromURL extracts just the path from a URL, removing the domain
func extractPathFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If parsing fails, try to extract path manually
		if strings.Contains(rawURL, "://") {
			parts := strings.SplitN(rawURL, "://", 2)
			if len(parts) == 2 {
				if slashIndex := strings.Index(parts[1], "/"); slashIndex != -1 {
					return parts[1][slashIndex:]
				}
			}
		}
		// Fallback: return the original URL if we can't parse it
		return rawURL
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// Include query parameters if they exist
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

	// Include fragment if it exists
	if parsed.Fragment != "" {
		path += "#" + parsed.Fragment
	}

	return path
}

// FormatAIPromptFromMultiplePages creates an AI-ready prompt from multiple pages with their details
func FormatAIPromptFromMultiplePages(pages []api.PageDetailsResponse) string {
	var prompt strings.Builder

	// Introduction
	prompt.WriteString("Please edit the following pages in order to improve SEO on my website by solving the list of issues associated with each page. ")
	prompt.WriteString("Make sure to stay on topic if you need to edit visible text and to not modify the site's current branding, ")
	prompt.WriteString("make sure to refactor the addition of meta tags as much as possible to not duplicate code everywhere, ")
	prompt.WriteString("and make sure that all your changes do not break the site's code in any way.\n\n")

	// Summary - count actual failed checks
	totalIssues := 0
	for _, page := range pages {
		for _, check := range page.Page.Checks {
			if !check.Passed {
				totalIssues++
			}
		}
	}

	prompt.WriteString(fmt.Sprintf("## Summary\n"))
	prompt.WriteString(fmt.Sprintf("- Total Pages: %d\n", len(pages)))
	prompt.WriteString(fmt.Sprintf("- Total Issues: %d\n\n", totalIssues))

	// Individual pages
	for i, pageDetails := range pages {
		page := pageDetails.Page

		// Count actual failed checks for accurate issue count
		actualIssues := 0
		for _, check := range page.Checks {
			if !check.Passed {
				actualIssues++
			}
		}

		prompt.WriteString(fmt.Sprintf("## %d. Path: %s\n\n", i+1, extractPathFromURL(page.URL)))
		prompt.WriteString(fmt.Sprintf("- SEO Score: %.1f/100\n", page.SEOScore))
		prompt.WriteString(fmt.Sprintf("- Issues: %d\n\n", actualIssues))

		// Current metadata
		prompt.WriteString("### Current Metadata\n")
		prompt.WriteString(fmt.Sprintf("- Title: %s\n", page.Title))
		prompt.WriteString(fmt.Sprintf("- Meta Description: %s\n", page.MetaDescription))
		prompt.WriteString(fmt.Sprintf("- H1: %s\n", page.H1))
		prompt.WriteString("\n")

		// Issues - count actual failed checks rather than relying on IssuesCount
		failedChecks := 0
		for _, check := range page.Checks {
			if !check.Passed {
				failedChecks++
			}
		}

		if failedChecks > 0 {
			prompt.WriteString("### Issues to Fix (from most important to less important to fix)\n\n")

			checkIndex := 0
			for _, check := range page.Checks {
				if !check.Passed {
					checkIndex++
					prompt.WriteString(fmt.Sprintf("**%d.%d.** Issue: %s (Weight: %d)\n\n", i+1, checkIndex, check.Message, check.Weight))
				}
			}
		} else if page.IssuesCount > 0 {
			// Fallback: if API says there are issues but we can't find failed checks
			prompt.WriteString("### Issues to Fix (from most important to less important to fix)\n\n")
			prompt.WriteString(fmt.Sprintf("API reports %d issues, but detailed check data is not available.\n", page.IssuesCount))
			prompt.WriteString("This may indicate an issue with the API response format.\n\n")
		}

		prompt.WriteString("---\n\n")
	}

	return prompt.String()
}

// FormatContentBrief formats a content brief for display or export
func FormatContentBrief(brief string) string {
	return brief
}