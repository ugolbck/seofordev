package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ugolbck/seofordev/internal/version"
)

const (
	// Repository information
	REPO_OWNER = "ugolbck"
	REPO_NAME  = "seofordev"

	// GitHub API endpoint
	GITHUB_RELEASES_API = "https://api.github.com/repos/ugolbck/seofordev/releases/latest"
)

// GitHubRelease represents the GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// VersionCheckResult contains the result of version checking
type VersionCheckResult struct {
	HasUpdate      bool
	CurrentVersion string
	LatestVersion  string
	Error          error
}

// CheckForUpdates queries GitHub API to check if there's a newer version available
func CheckForUpdates() VersionCheckResult {
	currentVersion := version.GetVersion()
	result := VersionCheckResult{
		CurrentVersion: currentVersion,
		HasUpdate:      false,
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request to GitHub API
	resp, err := client.Get(GITHUB_RELEASES_API)
	if err != nil {
		result.Error = fmt.Errorf("failed to check for updates: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		return result
	}

	// Parse JSON response
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		result.Error = fmt.Errorf("failed to parse GitHub API response: %v", err)
		return result
	}

	result.LatestVersion = release.TagName

	// Compare versions (simple string comparison for now)
	// This assumes semantic versioning like v1.0.0, v1.0.1, etc.
	if compareVersions(currentVersion, release.TagName) < 0 {
		result.HasUpdate = true
	}

	return result
}

// compareVersions compares two version strings
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split versions into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Ensure both have the same number of parts
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	// Pad shorter version with zeros
	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}

	return 0
}

// GetUpdateMessage returns a formatted message about available updates
func GetUpdateMessage(result VersionCheckResult) string {
	if !result.HasUpdate {
		return ""
	}

	return fmt.Sprintf(
		"🚀 Update available! %s → %s\n"+
			"Run: curl -sSfL https://seofor.dev/install.sh | bash",
		result.CurrentVersion,
		result.LatestVersion,
	)
}
