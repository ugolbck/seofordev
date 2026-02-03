package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show CLI configuration",
	Long: `Display your SEO CLI configuration settings.

This shows your local configuration for audit settings, ports, and other preferences.

Examples:
  seo config                          # Show current configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⚙️  SEO CLI Configuration\n")
		fmt.Printf("═══════════════════════════════════════════════════\n\n")

		cfg, err := config.LoadOrCreateConfig()
		if err != nil {
			fmt.Printf("❌ Configuration error: %v\n", err)
			return
		}

		// Default Audit Settings
		fmt.Printf("📊 Audit Settings:\n")
		fmt.Printf("  Default Port: %d\n", cfg.DefaultPort)
		fmt.Printf("  Concurrency: %d\n", cfg.DefaultConcurrency)
		if cfg.DefaultMaxPages > 0 {
			fmt.Printf("  Max Pages: %d\n", cfg.DefaultMaxPages)
		} else {
			fmt.Printf("  Max Pages: unlimited\n")
		}
		if cfg.DefaultMaxDepth > 0 {
			fmt.Printf("  Max Depth: %d\n", cfg.DefaultMaxDepth)
		} else {
			fmt.Printf("  Max Depth: unlimited\n")
		}
		fmt.Printf("  Ignore Patterns: %v\n", cfg.DefaultIgnorePatterns)

		// Configuration file location
		homeDir, _ := os.UserHomeDir()
		configPath := fmt.Sprintf("%s/.seo/config.yml", homeDir)
		fmt.Printf("\n📁 Config File: %s\n", configPath)
	},
}

func init() {
	// Add to root command
	rootCmd.AddCommand(configCmd)
}
