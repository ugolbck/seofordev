package tui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ugolbck/seofordev/internal/api"
)

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

// FormatAIPromptFromAudit creates an AI-ready prompt from an entire audit
func FormatAIPromptFromAudit(audit api.AuditViewResponse) string {
	var prompt strings.Builder

	// Introduction
	prompt.WriteString("Please edit the following pages in order to improve SEO on my website by solving the list of issues associated with each page. ")
	prompt.WriteString("Make sure to stay on topic if you need to edit visible text and to not modify the site's current branding, ")
	prompt.WriteString("make sure to refactor the addition of meta tags as much as possible to not duplicate code everywhere, ")
	prompt.WriteString("and make sure that all your changes do not break the site's code in any way.\n\n")

	// Audit summary
	prompt.WriteString(fmt.Sprintf("## Audit Summary\n"))
	prompt.WriteString(fmt.Sprintf("- Audit ID: %s\n", audit.ID))
	prompt.WriteString(fmt.Sprintf("- Overall Score: %.1f/100\n", audit.OverallScore))
	prompt.WriteString(fmt.Sprintf("- Total Pages: %d\n", len(audit.Pages)))
	prompt.WriteString(fmt.Sprintf("- Created: %s\n\n", audit.CreatedAt))

	// Pages and issues
	prompt.WriteString("## Pages to Fix\n\n")

	for i, page := range audit.Pages {
		if page.IssuesCount > 0 {
			prompt.WriteString(fmt.Sprintf("### %d. Path: %s\n", i+1, extractPathFromURL(page.URL)))
			prompt.WriteString(fmt.Sprintf("- SEO Score: %.1f/100\n", page.SEOScore))
			prompt.WriteString(fmt.Sprintf("- Issues Found: %d\n", page.IssuesCount))
			prompt.WriteString(fmt.Sprintf("- Status: %s\n\n", page.AnalysisStatus))
		}
	}

	prompt.WriteString("**Note:** You'll need to fetch detailed issues for each page to see specific problems to fix. ")
	prompt.WriteString("This prompt provides the overview of pages that need attention.\n")

	return prompt.String()
}

// FormatAIPromptFromPageDetails creates an AI-ready prompt from a specific page's details
func FormatAIPromptFromPageDetails(pageDetails *api.PageDetailsResponse) string {
	var prompt strings.Builder

	// Introduction
	prompt.WriteString("Please edit the following page to improve SEO by solving the list of issues identified below. ")
	prompt.WriteString("Make sure to stay on topic if you need to edit visible text and to not modify the site's current branding, ")
	prompt.WriteString("make sure to refactor the addition of meta tags as much as possible to not duplicate code everywhere, ")
	prompt.WriteString("and make sure that all your changes do not break the site's code in any way.\n\n")

	// Page info
	page := pageDetails.Page
	prompt.WriteString(fmt.Sprintf("## Path: %s\n\n", extractPathFromURL(page.URL)))
	prompt.WriteString(fmt.Sprintf("- SEO Score: %.1f/100\n", page.SEOScore))
	prompt.WriteString(fmt.Sprintf("- Total Issues: %d\n", page.IssuesCount))
	prompt.WriteString(fmt.Sprintf("- Word Count: %d\n", page.WordCount))
	prompt.WriteString(fmt.Sprintf("- Status Code: %d\n\n", page.StatusCode))

	// Current page metadata
	prompt.WriteString("### Current Page Metadata\n")
	prompt.WriteString(fmt.Sprintf("- Title: %s\n", page.Title))
	prompt.WriteString(fmt.Sprintf("- Meta Description: %s\n", page.MetaDescription))
	prompt.WriteString(fmt.Sprintf("- H1: %s\n", page.H1))
	if page.CanonicalURL != "" {
		prompt.WriteString(fmt.Sprintf("- Canonical URL: %s\n", page.CanonicalURL))
	}
	prompt.WriteString(fmt.Sprintf("- Indexable: %t\n", page.Indexable))
	if !page.Indexable {
		prompt.WriteString(fmt.Sprintf("- Indexability Issue: %s\n", page.IndexabilityReason))
	}
	prompt.WriteString("\n")

	// Issues to fix - count actual failed checks
	failedChecks := 0
	for _, check := range page.Checks {
		if !check.Passed {
			failedChecks++
		}
	}

	prompt.WriteString("### Issues to Fix (from most important to less important to fix)\n\n")

	if failedChecks > 0 {
		checkIndex := 0
		for _, check := range page.Checks {
			if !check.Passed {
				checkIndex++
				prompt.WriteString(fmt.Sprintf("**%d.** Issue: %s (Weight: %d)\n\n", checkIndex, check.Message, check.Weight))
			}
		}
	} else if page.IssuesCount > 0 {
		// Fallback: if API says there are issues but we can't find failed checks
		prompt.WriteString(fmt.Sprintf("API reports %d issues, but detailed check data is not available.\n", page.IssuesCount))
		prompt.WriteString("This may indicate an issue with the API response format.\n\n")
	} else {
		prompt.WriteString("No specific issues found - page appears to be well optimized!\n")
	}

	return prompt.String()
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

// ExportToClipboard copies the given text to the system clipboard
func ExportToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ExportAuditToClipboard exports an entire audit to clipboard as AI prompt
func ExportAuditToClipboard(audit api.AuditViewResponse) error {
	prompt := FormatAIPromptFromAudit(audit)
	return ExportToClipboard(prompt)
}

// ExportPageDetailsToClipboard exports a single page's details to clipboard as AI prompt
func ExportPageDetailsToClipboard(pageDetails *api.PageDetailsResponse) error {
	prompt := FormatAIPromptFromPageDetails(pageDetails)
	return ExportToClipboard(prompt)
}

// ExportMultiplePagesToClipboard exports multiple pages to clipboard as AI prompt
func ExportMultiplePagesToClipboard(pages []api.PageDetailsResponse) error {
	prompt := FormatAIPromptFromMultiplePages(pages)
	return ExportToClipboard(prompt)
}

// ExportPageDetailsToClipboardWithNotification exports a single page's details to clipboard and returns a notification
func ExportPageDetailsToClipboardWithNotification(pageDetails *api.PageDetailsResponse) tea.Cmd {
	return func() tea.Msg {
		err := ExportPageDetailsToClipboard(pageDetails)
		if err != nil {
			return NotificationMsg{
				Message: "Failed to copy AI prompt to clipboard",
				Type:    NotificationError,
			}
		}
		return NotificationMsg{
			Message: "AI prompt copied to clipboard successfully!",
			Type:    NotificationSuccess,
		}
	}
}

// ExportMultiplePagesToClipboardWithNotification exports multiple pages to clipboard and returns a notification
func ExportMultiplePagesToClipboardWithNotification(pages []api.PageDetailsResponse) tea.Cmd {
	return func() tea.Msg {
		err := ExportMultiplePagesToClipboard(pages)
		if err != nil {
			return NotificationMsg{
				Message: "Failed to copy AI prompt to clipboard",
				Type:    NotificationError,
			}
		}

		pageCount := len(pages)
		message := fmt.Sprintf("AI prompt for %d page(s) copied to clipboard successfully!", pageCount)
		return NotificationMsg{
			Message: message,
			Type:    NotificationSuccess,
		}
	}
}

// ExportContentBriefToClipboard exports a content brief to clipboard
func ExportContentBriefToClipboard(brief string) error {
	return ExportToClipboard(brief)
}

// ExportContentBriefToClipboardWithNotification exports a content brief to clipboard and returns a notification
func ExportContentBriefToClipboardWithNotification(brief string) tea.Cmd {
	return func() tea.Msg {
		err := ExportContentBriefToClipboard(brief)
		if err != nil {
			return NotificationMsg{
				Message: "Failed to copy content brief to clipboard",
				Type:    NotificationError,
			}
		}
		return NotificationMsg{
			Message: "Content brief copied to clipboard successfully!",
			Type:    NotificationSuccess,
		}
	}
}
