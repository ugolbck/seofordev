/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/playwright"
	"github.com/ugolbck/seofordev/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "seo",
	Version: version.GetVersion(),
	Short:   "SEO tools for indie hackers - command line interface",
	Long: `SEO tools for indie hackers - a command line interface for SEO tasks.

This tool provides commands for:
- Website auditing - run local SEO audits and export AI prompts to fix issues
- Keyword research - generate keyword suggestions  
- Content creation - generate SEO content briefs and export AI prompts to generate articles
- Search Engine Indexation - notify search engines about your latest changes, and get quoted in ChatGPT

Use the available commands below to get started, or run 'seo <command> --help' for detailed usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging (only in debug mode)
		if os.Getenv("SEO_DEBUG") == "1" || os.Getenv("DEBUG") == "1" {
			log.SetLevel(log.DebugLevel)
		}

		log.Debug("SEO CLI starting")

		// Ensure Playwright is installed for audit functionality
		if err := playwright.EnsurePlaywrightInstalled(); err != nil {
			log.Warn("Failed to set up Playwright", "error", err)
			fmt.Printf("⚠️  Playwright setup failed - audit functionality may not work\n")
			fmt.Printf("   Please check your internet connection and try running an audit to trigger setup.\n\n")
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

		fmt.Printf(`                    ___               _             
                   / __)             | |            
  ___ _____  ___ _| |__ ___   ____ __| |_____ _   _ 
 /___) ___ |/ _ (_   __) _ \ / ___) _  | ___ | | | |
|___ | ____| |_| || | | |_| | |_ ( (_| | ____|\ V / 
(___/|_____)\___/ |_|  \___/|_(_) \____|_____) \_/  
                                                    
`)
		fmt.Printf("🚀 SEO Tools CLI %s\n\n", version.GetVersion())

		fmt.Printf("Available commands:\n")
		fmt.Printf("  seo audit run            # Run localhost SEO audit (free)\n")
		fmt.Printf("  seo audit list           # List audit history\n")
		fmt.Printf("  seo config               # Show CLI configuration\n")
		fmt.Printf("  seo keyword generate     # Generate keywords (premium)\n")
		fmt.Printf("  seo brief generate       # Generate content briefs (premium)\n")
		fmt.Printf("  seo index submit         # Submit URLs to search engines via IndexNow\n")
		fmt.Printf("  seo pro status           # Check premium account status\n")
		fmt.Printf("  seo pro setup            # Setup premium features\n")
		fmt.Printf("  seo --help               # Show all commands\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  seo audit run                        # Audit localhost:3000\n")
		fmt.Printf("  seo audit run --port 8080            # Audit localhost:8080\n")
		fmt.Printf("  seo config                           # Show configuration\n")
		fmt.Printf("  seo keyword generate \"coffee shop\"   # Generate keywords (premium)\n")
		fmt.Printf("  seo pro status                       # Check premium account\n\n")
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
