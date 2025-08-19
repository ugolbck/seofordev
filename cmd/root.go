/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/playwright"
	"github.com/ugolbck/seofordev/internal/tui"
	"github.com/ugolbck/seofordev/internal/tui/logger"
	"github.com/ugolbck/seofordev/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "oss_seo",
	Version: version.GetVersion(),
	Short:   "Open-source SEO for indie hackers",
	Long:    `Open-source SEO tools for indie hackers and developers. Run 'seo'.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging (only in debug mode)
		if err := logger.InitLogger(); err != nil {
			// Only show error if we're actually trying to log (debug mode)
			if os.Getenv("SEO_DEBUG") == "1" || os.Getenv("DEBUG") == "1" {
				fmt.Printf("⚠️  Failed to initialize debug logging: %v\n", err)
			}
			// Continue without logging rather than failing
		}
		defer logger.CloseLogger()

		logger.LogInfo("Application starting")

		// Ensure Playwright is installed before starting TUI
		if err := playwright.EnsurePlaywrightInstalled(); err != nil {
			fmt.Printf("❌ Failed to set up Playwright: %v\n", err)
			fmt.Printf("   Please check your internet connection and try again.\n")
			os.Exit(1)
		}

		// Check minimum version requirement
		if err := version.CheckMinimumVersion(); err != nil {
			fmt.Printf("❌ %v\n\n", err)
			fmt.Printf("To update to the latest version, run:\n")
			fmt.Printf("  curl -sSfL https://seofor.dev/install.sh | bash\n\n")
			fmt.Printf("Or download manually from:\n")
			fmt.Printf("  https://github.com/ugolbck/seofordev/releases/latest\n")
			os.Exit(1)
		}

		// Check for updates (don't block startup if this fails)
		versionResult := tui.CheckForUpdates()

		// Define the start model
		var startModel tea.Model

		// Create the start model with the version check result
		startModel = tui.NewMainMenuModelWithVersionCheck(versionResult)

		// Configure tea program
		p := tea.NewProgram(
			startModel,
			tea.WithAltScreen(), // Use alternative screen buffer
		)

		// Start the program
		if _, err := p.Run(); err != nil {
			fmt.Printf("❌ Application failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// No CLI flags needed - everything is handled through the TUI (for now)
}
