package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show and manage CLI configuration",
	Long: `Display and manage your SEO CLI configuration settings.

This shows your local configuration for audit settings, ports, API key, and other preferences.

Examples:
  seo config                          # Show current configuration
  seo config set-api-key <your-key>   # Set premium API key`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⚙️  SEO CLI Configuration\n")
		fmt.Printf("═══════════════════════════════════════════════════\n\n")

		cfg, err := config.LoadOrCreateConfig()
		if err != nil {
			fmt.Printf("❌ Configuration error: %v\n", err)
			return
		}

		// API Key Status
		fmt.Printf("🔑 API Key:\n")
		if cfg.APIKey != "" {
			maskedKey := cfg.APIKey[:8] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
			fmt.Printf("  Status: ✅ Configured (%s)\n", maskedKey)
		} else {
			fmt.Printf("  Status: ❌ Not configured\n")
		}
		fmt.Printf("\n")

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

		fmt.Printf("\n💡 Use 'seo config set-api-key <key>' to set your premium API key.\n")
	},
}

var configSetAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key <api-key>",
	Short: "Set premium API key in configuration",
	Long:  `Set your premium API key in the configuration file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := args[0]
		
		if apiKey == "" {
			fmt.Printf("❌ API key cannot be empty\n")
			return
		}

		// Load current config
		cfg, err := config.LoadOrCreateConfig()
		if err != nil {
			fmt.Printf("❌ Configuration error: %v\n", err)
			return
		}

		// Set the API key
		cfg.APIKey = apiKey

		// Save config
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("❌ Failed to save configuration: %v\n", err)
			return
		}

		// Show success message
		maskedKey := apiKey[:8] + "..." + apiKey[len(apiKey)-4:]
		fmt.Printf("✅ API key set successfully: %s\n", maskedKey)
		fmt.Printf("💡 Use 'seo pro status' to verify your premium account status.\n")
	},
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configSetAPIKeyCmd)
	
	// Add to root command
	rootCmd.AddCommand(configCmd)
}