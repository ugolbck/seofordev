package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/export"
	"github.com/ugolbck/seofordev/internal/services"
)

var briefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Generate content briefs and manage brief history",
	Long: `Generate SEO content briefs from keywords and manage your content brief history.
	
Examples:
  seo brief generate "coffee recipes"   # Generate brief for keyword
  seo brief history                     # Show brief generation history
  seo brief show <brief-id>             # Show specific brief content`,
}

var briefGenerateCmd = &cobra.Command{
	Use:   "generate <keyword>",
	Short: "Generate a content brief for a keyword",
	Long: `Generate a detailed SEO content brief for a given keyword. The brief will include content structure, target keywords, and optimization recommendations.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyword := args[0]
		
		briefService, err := services.NewBriefService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Start brief generation
		result, err := briefService.GenerateBrief(keyword)
		if err != nil {
			fmt.Printf("❌ Failed to generate brief: %v\n", err)
			return
		}

		fmt.Printf("📝 Generating content brief for '%s'...\n", keyword)
		fmt.Printf("⏳ Brief ID: %s\n", result.ID)
		fmt.Printf("📊 Credits used: %d\n\n", result.CreditsUsed)

		// Wait for completion
		briefResponse, err := briefService.WaitForBrief(result.ID)
		if err != nil {
			fmt.Printf("❌ Brief generation failed: %v\n", err)
			return
		}

		// Display the brief
		if briefResponse.Brief != nil && *briefResponse.Brief != "" {
			fmt.Printf("✅ Content brief generated successfully!\n\n")
			
			// Check if user wants to copy to clipboard
			copyToClipboard, _ := cmd.Flags().GetBool("copy")
			if copyToClipboard {
				if err := export.ExportToClipboard(*briefResponse.Brief); err != nil {
					log.Warn("Failed to copy to clipboard", "error", err)
					fmt.Printf("⚠️ Could not copy to clipboard, showing content instead:\n\n")
					fmt.Print(*briefResponse.Brief)
				} else {
					fmt.Printf("📋 Brief copied to clipboard!\n")
					fmt.Printf("   Paste it into your content management system or editor.\n")
				}
			} else {
				fmt.Printf("📄 Content Brief:\n")
				fmt.Printf("════════════════════════════════════════════════════════════════\n\n")
				fmt.Print(*briefResponse.Brief)
				fmt.Printf("\n\n💡 Tip: Use --copy flag to copy the brief to clipboard\n")
			}
		} else {
			fmt.Printf("⚠️ Brief generated but content is empty\n")
		}
	},
}

var briefHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show content brief history",
	Long:  `Show your content brief generation history.`,
	Run: func(cmd *cobra.Command, args []string) {
		briefService, err := services.NewBriefService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		history, err := briefService.GetHistory()
		if err != nil {
			fmt.Printf("❌ Failed to load brief history: %v\n", err)
			return
		}

		if len(history.Briefs) == 0 {
			fmt.Println("No content briefs found. Run 'seo brief generate <keyword>' to get started.")
			return
		}

		fmt.Printf("\n📝 Content Brief History (%d briefs)\n\n", len(history.Briefs))

		for _, brief := range history.Briefs {
			status := "✅"
			if brief.Status != "completed" {
				status = "⏳"
			}

			fmt.Printf("%s %s - '%s'\n", status, brief.ID, brief.Keyword)
			fmt.Printf("    Generated: %s\n", brief.GeneratedAt)
			fmt.Printf("    Credits: %d\n\n", brief.CreditsUsed)
		}

		fmt.Printf("Use 'seo brief show <id>' to view brief content\n")
	},
}

var briefShowCmd = &cobra.Command{
	Use:   "show <brief-id>",
	Short: "Show content brief details",
	Long:  `Show the content of a specific content brief.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		briefID := args[0]
		
		briefService, err := services.NewBriefService()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		brief, err := briefService.GetBriefStatus(briefID)
		if err != nil {
			fmt.Printf("❌ Failed to load brief: %v\n", err)
			return
		}

		fmt.Printf("\n📝 Content Brief Details\n")
		fmt.Printf("═══════════════════════════════════════════════════════════\n\n")
		fmt.Printf("🆔 ID: %s\n", brief.ID)
		fmt.Printf("🔍 Keyword: %s\n", brief.Keyword)
		fmt.Printf("📅 Generated: %s\n", brief.GeneratedAt)
		fmt.Printf("📊 Credits Used: %d\n", brief.CreditsUsed)
		fmt.Printf("✅ Status: %s\n\n", brief.Status)

		if brief.Brief != nil && *brief.Brief != "" {
			// Check if user wants to copy to clipboard
			copyToClipboard, _ := cmd.Flags().GetBool("copy")
			if copyToClipboard {
				if err := export.ExportToClipboard(*brief.Brief); err != nil {
					log.Warn("Failed to copy to clipboard", "error", err)
					fmt.Printf("⚠️ Could not copy to clipboard, showing content instead:\n\n")
					fmt.Print(*brief.Brief)
				} else {
					fmt.Printf("📋 Brief copied to clipboard!\n")
					return
				}
			} else {
				fmt.Printf("📄 Content:\n")
				fmt.Printf("─────────────────────────────────────────────────────────\n\n")
				fmt.Print(*brief.Brief)
				fmt.Printf("\n\n💡 Tip: Use --copy flag to copy the brief to clipboard\n")
			}
		} else {
			fmt.Printf("⚠️ No content available for this brief\n")
		}
	},
}

func init() {
	// Add flags
	briefGenerateCmd.Flags().Bool("copy", false, "Copy the generated brief to clipboard")
	briefShowCmd.Flags().Bool("copy", false, "Copy the brief content to clipboard")
	
	// Add subcommands
	briefCmd.AddCommand(briefGenerateCmd)
	briefCmd.AddCommand(briefHistoryCmd)
	briefCmd.AddCommand(briefShowCmd)
	
	// Add to root command
	rootCmd.AddCommand(briefCmd)
}