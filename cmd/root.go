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
	"github.com/ugolbck/seofordev/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "seo",
	Version: version.GetVersion(),
	Short:   "SEO tools for indie hackers - interactive TUI interface",
	Long: `SEO tools for indie hackers - an interactive command line interface for SEO tasks.

This tool provides a unified interface for:
- Website auditing - export AI prompts to fix your site in one click
- Keyword suggestions 
- Content generation - export AI prompt to generate content for your site (coming soon)

Simply run 'seo' to launch the interactive interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging (only in debug mode)
		if err := tui.InitLogger(); err != nil {
			// Only show error if we're actually trying to log (debug mode)
			if os.Getenv("SEO_DEBUG") == "1" || os.Getenv("DEBUG") == "1" {
				fmt.Printf("⚠️  Failed to initialize debug logging: %v\n", err)
			}
			// Continue without logging rather than failing
		}
		defer tui.CloseLogger()

		tui.LogInfo("Application starting")

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

		// Always start with main menu - API key is now optional
		// Configuration loading is now handled inside the main menu model
		startModel := tui.NewMainMenuModelWithVersionCheck(versionResult)

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
	// No CLI flags needed - everything is handled through the TUI
}
