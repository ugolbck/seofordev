package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/services"
)

var keywordCmd = &cobra.Command{
	Use:   "keyword",
	Short: "Generate keywords and manage keyword history",
	Long: `Generate SEO keywords from seed keywords and manage your keyword generation history.
	
Examples:
  seo keyword generate "coffee shop"    # Generate keywords for seed
  seo keyword history                   # Show keyword generation history
  seo keyword show <generation-id>      # Show detailed keyword results`,
}

var keywordGenerateCmd = &cobra.Command{
	Use:   "generate <seed-keyword>",
	Short: "Generate keywords from a seed keyword",
	Long:  `Generate SEO keywords and their metrics from a seed keyword.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		seedKeyword := args[0]
		
		keywordService, err := services.NewKeywordService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		result, err := keywordService.GenerateKeywords(seedKeyword)
		if err != nil {
			fmt.Printf("❌ Failed to generate keywords: %v\n", err)
			return
		}

		fmt.Printf("\n🔍 Keywords for '%s'\n", result.SeedKeyword)
		fmt.Printf("════════════════════════════════════════════════════════════\n\n")
		fmt.Printf("📊 Generated %d keywords using %d credits\n", result.TotalResults, result.CreditsUsed)
		fmt.Printf("📅 Generated: %s\n\n", result.GeneratedAt)

		if len(result.Keywords) == 0 {
			fmt.Printf("No keywords generated.\n")
			return
		}

		fmt.Printf("📝 Keywords:\n")
		fmt.Printf("────────────────────────────────────────────────────────────\n")
		
		for i, keyword := range result.Keywords {
			if i >= 50 { // Limit display to first 50
				fmt.Printf("    ... and %d more keywords\n", len(result.Keywords)-50)
				break
			}

			volume := "N/A"
			if keyword.Volume != nil {
				volume = fmt.Sprintf("%d", *keyword.Volume)
			}

			difficulty := "N/A"
			if keyword.Difficulty != nil {
				difficulty = fmt.Sprintf("%.1f", *keyword.Difficulty)
			}

			cpc := "N/A"
			if keyword.CPC != nil {
				cpc = fmt.Sprintf("$%.2f", *keyword.CPC)
			}

			fmt.Printf("  %-30s Vol: %-8s Diff: %-6s CPC: %s\n", 
				truncate(keyword.Keyword, 30), volume, difficulty, cpc)
		}
		
		fmt.Printf("\n💡 Use these keywords in your content to improve SEO!\n")
	},
}

var keywordHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show keyword generation history",
	Long:  `Show your keyword generation history with summaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		keywordService, err := services.NewKeywordService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		history, err := keywordService.GetHistory()
		if err != nil {
			fmt.Printf("❌ Failed to load keyword history: %v\n", err)
			return
		}

		if len(history.Generations) == 0 {
			fmt.Println("No keyword generations found. Run 'seo keyword generate <seed>' to get started.")
			return
		}

		fmt.Printf("\n🔍 Keyword Generation History (%d generations)\n\n", len(history.Generations))

		for _, gen := range history.Generations {
			status := "✅"
			if gen.Status != "completed" {
				status = "⏳"
			}

			fmt.Printf("%s %s - '%s'\n", status, gen.ID, gen.SeedKeyword)
			fmt.Printf("    Generated: %s\n", gen.GeneratedAt)
			fmt.Printf("    Keywords: %d, Credits: %d\n\n", gen.TotalResults, gen.CreditsUsed)
		}

		fmt.Printf("Use 'seo keyword show <id>' to view detailed results\n")
	},
}

var keywordShowCmd = &cobra.Command{
	Use:   "show <generation-id>",
	Short: "Show keyword generation details",
	Long:  `Show the detailed results of a specific keyword generation.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generationID := args[0]
		
		keywordService, err := services.NewKeywordService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		generation, err := keywordService.GetKeywordGeneration(generationID)
		if err != nil {
			fmt.Printf("❌ Failed to load keyword generation: %v\n", err)
			return
		}

		fmt.Printf("\n🔍 Keyword Generation Details\n")
		fmt.Printf("═══════════════════════════════════════════════════════════\n\n")
		fmt.Printf("🆔 ID: %s\n", generation.ID)
		fmt.Printf("🌱 Seed Keyword: %s\n", generation.SeedKeyword)
		fmt.Printf("📅 Generated: %s\n", generation.GeneratedAt)
		fmt.Printf("📊 Credits Used: %d\n", generation.CreditsUsed)
		fmt.Printf("✅ Status: %s\n", generation.Status)
		fmt.Printf("📈 Total Results: %d\n\n", generation.TotalResults)

		if len(generation.Keywords) == 0 {
			fmt.Printf("No keywords generated.\n")
			return
		}

		fmt.Printf("📝 Keywords (%d):\n", len(generation.Keywords))
		fmt.Printf("────────────────────────────────────────────────────────────\n")
		
		for i, keyword := range generation.Keywords {
			volume := "N/A"
			if keyword.Volume != nil {
				volume = fmt.Sprintf("%d", *keyword.Volume)
			}

			difficulty := "N/A"
			if keyword.Difficulty != nil {
				difficulty = fmt.Sprintf("%.1f", *keyword.Difficulty)
			}

			cpc := "N/A"
			if keyword.CPC != nil {
				cpc = fmt.Sprintf("$%.2f", *keyword.CPC)
			}

			fmt.Printf("%3d. %-35s Vol: %-8s Diff: %-6s CPC: %s\n", 
				i+1, truncate(keyword.Keyword, 35), volume, difficulty, cpc)
		}
		
		fmt.Printf("\n💡 Use these keywords in your content to improve SEO!\n")
	},
}


// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	// Add subcommands
	keywordCmd.AddCommand(keywordGenerateCmd)
	keywordCmd.AddCommand(keywordHistoryCmd)
	keywordCmd.AddCommand(keywordShowCmd)
	
	// Add to root command
	rootCmd.AddCommand(keywordCmd)
}