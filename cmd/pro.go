package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/config"
	"github.com/ugolbck/seofordev/internal/services"
)

var proCmd = &cobra.Command{
	Use:   "pro",
	Short: "Premium features and account management",
	Long: `Manage your seofor.dev premium subscription, API keys, and credit balance.

These commands are for users with paid seofor.dev plans who want to use premium features
like keyword generation and content brief creation.
	
Examples:
  seo pro status           # Show premium account status and credit balance  
  seo pro setup            # Setup premium API key`,
}

var proStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show account status and credit balance",
	Long:  `Display your premium account status, API key status, and current credit balance.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔑 seofor.dev Premium Account Status\n")
		fmt.Printf("═══════════════════════════════════════════════════\n\n")

		// Check API key configuration
		cfg, err := config.LoadOrCreateConfig()
		if err != nil {
			fmt.Printf("❌ Configuration error: %v\n", err)
			return
		}

		if cfg.APIKey == "" {
			fmt.Printf("❌ API Key: Not configured\n")
			fmt.Printf("📊 Credits: N/A\n\n")
			fmt.Printf("💡 Setup your premium API key:\n")
			fmt.Printf("  1. Subscribe to a plan: https://seofor.dev/payments/pricing\n")
			fmt.Printf("  2. Get your API key: https://seofor.dev/dashboard\n")
			fmt.Printf("  3. Run: seo pro setup\n")
			return
		}

		// Mask API key for display
		maskedKey := cfg.APIKey[:8] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
		fmt.Printf("✅ API Key: %s\n", maskedKey)

		// Get credit balance
		keywordService, err := services.NewKeywordService()
		if err != nil {
			fmt.Printf("⚠️  Credits: Unable to check (API key may be invalid)\n")
			fmt.Printf("   Error: %v\n\n", err)
			return
		}

		balance, err := keywordService.GetCreditBalance()
		if err != nil {
			fmt.Printf("⚠️  Credits: Unable to check\n")
			fmt.Printf("   Error: %v\n\n", err)
			return
		}

		fmt.Printf("💰 Credits: %d\n", balance.Credits)

		if balance.Credits < 10 {
			fmt.Printf("⚠️  Status: Low credit balance\n")
		} else if balance.Credits < 50 {
			fmt.Printf("✅ Status: Good credit balance\n")
		} else {
			fmt.Printf("🎉 Status: Excellent credit balance\n")
		}

		fmt.Printf("\n📝 Credit Usage:\n")
		fmt.Printf("  • Keyword generation: 10 credits per generation\n")
		fmt.Printf("  • Content brief creation: 20 credits per brief\n")
		fmt.Printf("\n💳 Manage subscription: https://seofor.dev/dashboard\n")
	},
}

var proSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup premium API key",
	Long:  `Setup wizard for configuring your seofor.dev premium API key.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔑 seofor.dev Premium API Key Setup\n")
		fmt.Printf("═══════════════════════════════════════════════════\n\n")

		// Check if already configured
		cfg, err := config.LoadOrCreateConfig()
		if err != nil {
			fmt.Printf("❌ Configuration error: %v\n", err)
			return
		}

		if cfg.APIKey != "" {
			maskedKey := cfg.APIKey[:8] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
			fmt.Printf("✅ Premium API key already configured: %s\n\n", maskedKey)
			fmt.Printf("To reconfigure, you can:\n")
			fmt.Printf("  1. Run: seo config set-api-key <new-key>\n")
			fmt.Printf("  2. Edit your config file: ~/.seo/config.yml\n\n")
			fmt.Printf("💳 Manage your subscription: https://seofor.dev/dashboard\n")
			return
		}

		fmt.Printf("📋 Premium Setup Steps:\n\n")
		fmt.Printf("1. Subscribe to a premium plan:\n")
		fmt.Printf("   → Visit: https://seofor.dev/payments/pricing\n")
		fmt.Printf("   → Choose a plan that fits your needs\n")
		fmt.Printf("   → Complete the subscription process\n\n")

		fmt.Printf("2. Get your API key:\n")
		fmt.Printf("   → Visit: https://seofor.dev/dashboard\n")
		fmt.Printf("   → Copy your API key from the dashboard\n\n")

		fmt.Printf("3. Configure your API key:\n\n")

		fmt.Printf("   seo config set-api-key your_api_key_here\n\n")

		fmt.Printf("   Alternative: Edit ~/.seo/config.yml and add:\n")
		fmt.Printf("   api_key: your_api_key_here\n\n")

		fmt.Printf("4. Verify your setup:\n")
		fmt.Printf("   seo pro status\n\n")

		fmt.Printf("💡 Note: Premium features require an active subscription.\n")
	},
}

func init() {
	// Add subcommands
	proCmd.AddCommand(proStatusCmd)
	proCmd.AddCommand(proSetupCmd)

	// Add to root command
	rootCmd.AddCommand(proCmd)
}
