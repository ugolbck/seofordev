package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/ugolbck/seofordev/internal/export"
	"github.com/ugolbck/seofordev/internal/services"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run SEO audits and manage audit history",
	Long: `Run SEO audits on your website and manage audit history.
	
Examples:
  seo audit run                    # Audit localhost:3000
  seo audit run --port 8080        # Audit localhost:8080
  seo audit list                   # Show audit history
  seo audit show <audit-id>        # Show audit details
  seo audit export <audit-id>      # Export audit as AI prompt`,
}

var auditRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a localhost SEO audit",
	Long: `Run an SEO audit on your localhost development server. Specify the port with --port flag.

Examples:
  seo audit run                           # Audit localhost:3000 (default)
  seo audit run --port 8080              # Audit localhost:8080
  seo audit run --port 3000 --max-pages 50  # Audit with custom limits`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		maxPages, _ := cmd.Flags().GetInt("max-pages")
		maxDepth, _ := cmd.Flags().GetInt("max-depth")
		ignorePatterns, _ := cmd.Flags().GetStringSlice("ignore")

		baseURL := fmt.Sprintf("http://localhost:%d", port)

		log.Info("Starting localhost SEO audit", "port", port, "url", baseURL)

		auditService, err := services.NewAuditService()
		if err != nil {
			log.Fatal("Failed to initialize audit service", "error", err)
		}

		config := services.AuditConfig{
			Port:           port,
			Concurrency:    concurrency,
			MaxPages:       maxPages,
			MaxDepth:       maxDepth,
			IgnorePatterns: ignorePatterns,
		}

		result, err := auditService.RunAudit(baseURL, config)
		if err != nil {
			log.Fatal("Audit failed", "error", err)
		}

		overallScore := "N/A"
		if result.OverallScore != nil {
			overallScore = fmt.Sprintf("%.1f/100", *result.OverallScore)
		}

		log.Info("Audit completed successfully", 
			"audit_id", result.ID,
			"pages_analyzed", result.PagesAnalyzed,
			"overall_score", overallScore)

		fmt.Printf("\n✅ Audit complete!\n")
		fmt.Printf("   ID: %s\n", result.ID)
		fmt.Printf("   Pages analyzed: %d\n", result.PagesAnalyzed)
		fmt.Printf("   Overall score: %s\n", overallScore)
		fmt.Printf("\nView details: seo audit show %s\n", result.ID)
	},
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit history",
	Long:  `List all stored SEO audits with their basic information.`,
	Run: func(cmd *cobra.Command, args []string) {
		auditService, err := services.NewAuditService()
		if err != nil {
			log.Fatal("Failed to initialize audit service", "error", err)
		}

		audits, err := auditService.ListAudits()
		if err != nil {
			log.Fatal("Failed to load audit history", "error", err)
		}

		if len(audits) == 0 {
			fmt.Println("No audits found. Run 'seo audit run' to create your first audit.")
			return
		}

		fmt.Printf("\n📊 Audit History (%d audits)\n\n", len(audits))
		
		for _, audit := range audits {
			status := "✅"
			if audit.Status != "completed" {
				status = "⏳"
			}

			score := "N/A"
			if audit.OverallScore != nil {
				score = fmt.Sprintf("%.1f/100", *audit.OverallScore)
			}

			fmt.Printf("%s %s - %s\n", status, audit.ID[:8], audit.BaseURL)
			fmt.Printf("    Created: %s\n", audit.CreatedAt.Format("Jan 2, 2006 15:04"))
			fmt.Printf("    Score: %s, Pages: %d\n\n", score, audit.PagesAnalyzed)
		}

		fmt.Printf("Use 'seo audit show <id>' to view details\n")
	},
}

var auditShowCmd = &cobra.Command{
	Use:   "show <audit-id>",
	Short: "Show audit details",
	Long:  `Show detailed results for a specific audit.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		auditID := args[0]
		
		auditService, err := services.NewAuditService()
		if err != nil {
			log.Fatal("Failed to initialize audit service", "error", err)
		}

		audit, err := auditService.GetAudit(auditID)
		if err != nil {
			log.Fatal("Failed to load audit", "audit_id", auditID, "error", err)
		}

		fmt.Printf("\n📊 Audit Details\n")
		fmt.Printf("═══════════════════════════════════════════════════\n\n")
		fmt.Printf("🆔 ID: %s\n", audit.ID)
		fmt.Printf("🌐 URL: %s\n", audit.BaseURL)
		fmt.Printf("📅 Created: %s\n", audit.CreatedAt.Format("January 2, 2006 at 15:04"))
		
		if audit.CompletedAt != nil {
			fmt.Printf("✅ Completed: %s\n", audit.CompletedAt.Format("January 2, 2006 at 15:04"))
		}

		if audit.OverallScore != nil {
			fmt.Printf("📈 Overall Score: %.1f/100\n", *audit.OverallScore)
		}
		
		fmt.Printf("📄 Pages Analyzed: %d\n\n", len(audit.Pages))

		if len(audit.Pages) > 0 {
			fmt.Printf("📄 Pages:\n")
			fmt.Printf("─────────────────────────────────────────────────────\n")
			
			for i, page := range audit.Pages {
				if i >= 10 { // Show first 10 pages
					fmt.Printf("    ... and %d more pages\n", len(audit.Pages)-10)
					break
				}

				score := "N/A"
				if page.SEOScore != nil {
					score = fmt.Sprintf("%.1f", *page.SEOScore)
				}

				status := "✅"
				if page.AnalysisStatus != "completed" {
					status = "⏳"
				}

				fmt.Printf("  %s %s - %s\n", status, score, page.URL)
			}
		}

		if audit.Summary != nil && len(audit.Summary.Recommendations) > 0 {
			fmt.Printf("\n💡 Recommendations:\n")
			fmt.Printf("─────────────────────────────────────────────────────\n")
			for i, rec := range audit.Summary.Recommendations {
				if i >= 5 { // Show first 5 recommendations
					fmt.Printf("    ... and %d more recommendations\n", len(audit.Summary.Recommendations)-5)
					break
				}
				fmt.Printf("  • %s\n", rec)
			}
		}

		fmt.Printf("\n💾 Export: seo audit export %s\n", audit.ID)
	},
}

var auditExportCmd = &cobra.Command{
	Use:   "export <audit-id>",
	Short: "Export audit as AI prompt",
	Long:  `Export audit results as an AI prompt for fixing SEO issues.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		auditID := args[0]
		
		auditService, err := services.NewAuditService()
		if err != nil {
			log.Fatal("Failed to initialize audit service", "error", err)
		}

		prompt, err := auditService.ExportAuditPrompt(auditID)
		if err != nil {
			log.Fatal("Failed to export audit", "audit_id", auditID, "error", err)
		}

		// Try to copy to clipboard
		if err := export.ExportToClipboard(prompt); err != nil {
			log.Warn("Failed to copy to clipboard", "error", err)
			fmt.Printf("\n📄 AI Prompt (copy manually):\n")
			fmt.Printf("════════════════════════════════════════════════════════════════\n")
			fmt.Print(prompt)
		} else {
			fmt.Printf("✅ AI prompt copied to clipboard!\n")
			fmt.Printf("   Paste it into your favorite AI assistant to get SEO fixes.\n")
		}
	},
}

func init() {
	// Add flags to run command
	auditRunCmd.Flags().IntP("port", "p", 3000, "Port for localhost audit")
	auditRunCmd.Flags().IntP("concurrency", "c", 4, "Number of concurrent requests")
	auditRunCmd.Flags().IntP("max-pages", "m", 0, "Maximum pages to audit (0 = unlimited)")
	auditRunCmd.Flags().IntP("max-depth", "d", 0, "Maximum crawl depth (0 = unlimited)")
	auditRunCmd.Flags().StringSliceP("ignore", "i", []string{"/api", "/admin"}, "URL patterns to ignore")

	// Add subcommands
	auditCmd.AddCommand(auditRunCmd)
	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditShowCmd)
	auditCmd.AddCommand(auditExportCmd)

	// Add to root command
	rootCmd.AddCommand(auditCmd)
}