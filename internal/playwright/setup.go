package playwright

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/playwright-community/playwright-go"
)

const (
	playwrightVersion = "0.5200.0"
)

// GetPlaywrightDir returns the directory where Playwright should be installed
func GetPlaywrightDir() string {
	// For testing: use TEST_PLAYWRIGHT_DIR environment variable if set
	if testDir := os.Getenv("TEST_PLAYWRIGHT_DIR"); testDir != "" {
		return testDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".seo", "playwright")
	}
	return filepath.Join(homeDir, ".seo", "playwright")
}

// EnsurePlaywrightInstalled checks if Playwright is installed and installs it seamlessly if needed
// This should be called before the TUI starts to ensure setup is visible to the user
func EnsurePlaywrightInstalled() error {
	driverDir := GetPlaywrightDir()

	// Check if already installed with correct version
	versionMarker := filepath.Join(driverDir, "VERSION_v"+playwrightVersion)
	if _, err := os.Stat(versionMarker); err == nil {
		return nil // Already installed with correct version
	}

	// First-time setup message
	fmt.Println("🎭 Setting up Playwright for web crawling...")
	fmt.Println("📥 Downloading browser components (~150MB, one-time setup)...")
	fmt.Println("   This may take a few minutes depending on your connection...")

	// Create directory structure safely
	if err := os.MkdirAll(driverDir, 0755); err != nil {
		return fmt.Errorf("could not create Playwright directory: %w", err)
	}

	// Clean any old installations
	if err := cleanOldPlaywrightInstalls(driverDir); err != nil {
		fmt.Printf("⚠️  Warning: could not clean old installations: %v\n", err)
	}

	// Set environment variable to control browser installation location
	browsersDir := filepath.Join(driverDir, "browsers")
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", browsersDir)

	// Install Playwright with custom driver directory
	runOptions := &playwright.RunOptions{
		DriverDirectory: driverDir,
		Browsers:        []string{"chromium"}, // Only install Chromium
	}

	fmt.Println("   Installing Playwright runtime and Chromium browser...")
	if err := playwright.Install(runOptions); err != nil {
		return fmt.Errorf("could not install Playwright: %w", err)
	}

	// Create version marker for future checks
	if err := os.WriteFile(versionMarker, []byte(playwrightVersion), 0644); err != nil {
		return fmt.Errorf("could not create version marker: %w", err)
	}

	fmt.Println("✅ Playwright setup complete!")
	fmt.Println("")
	return nil
}

// CheckPlaywrightInstalled verifies that Playwright is properly installed (for crawler use)
func CheckPlaywrightInstalled() error {
	driverDir := GetPlaywrightDir()

	// Check if installation marker exists
	versionMarker := filepath.Join(driverDir, "VERSION_v"+playwrightVersion)
	if _, err := os.Stat(versionMarker); os.IsNotExist(err) {
		return fmt.Errorf("Playwright installation not found. Please restart the application to reinstall")
	}

	// Check if browser directory exists
	browserPath := filepath.Join(driverDir, "browsers")
	if _, err := os.Stat(browserPath); os.IsNotExist(err) {
		return fmt.Errorf("Playwright browsers not found. Please restart the application to reinstall")
	}

	return nil
}

// cleanOldPlaywrightInstalls removes old version markers and cleans up if needed
func cleanOldPlaywrightInstalls(driverDir string) error {
	// Check if we're changing versions
	versionMarker := filepath.Join(driverDir, "VERSION_v"+playwrightVersion)
	if _, err := os.Stat(versionMarker); err == nil {
		return nil // Same version, no cleanup needed
	}

	// Different version detected - clean everything for fresh install
	fmt.Println("   Detected version change, cleaning old installation...")

	// Remove entire directory and recreate for clean slate
	if err := os.RemoveAll(driverDir); err != nil {
		return fmt.Errorf("could not clean old installation: %w", err)
	}

	// Recreate directory
	if err := os.MkdirAll(driverDir, 0755); err != nil {
		return fmt.Errorf("could not recreate directory: %w", err)
	}

	return nil
}
