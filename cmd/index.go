package cmd

import (
	"github.com/ugolbck/seofordev/internal/indexnow"

	"github.com/spf13/cobra"

	"fmt"
)

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.AddCommand(setupCmd)
	indexCmd.AddCommand(submitCmd)
	indexCmd.AddCommand(verifyCmd)
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "IndexNow related commands",
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Generate IndexNow key and setup instructions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return indexnow.Setup()
	},
}

var submitCmd = &cobra.Command{
	Use:   "submit <key> <url1> [url2] [url3]...",
	Short: "Submit one or more page URLs to IndexNow for faster indexing",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		urls := args[1:]

		err := indexnow.SubmitURLs(urls, key)
		if err != nil {
			return fmt.Errorf("❌ submission failed: %v", err)
		}

		fmt.Printf("✅ Successfully submitted %d URLs to IndexNow (host: %s)\n", len(urls), urls[0])
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify <domain> <key>",
	Short: "Verify that your IndexNow key file is reachable on your domain",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		key := args[1]

		err := indexnow.VerifyKey(domain, key)
		if err != nil {
			return fmt.Errorf("❌ verification failed: %v", err)
		}

		fmt.Printf("✅ Key verified! %s/%s.txt is accessible and valid.\n", domain, key)
		return nil
	},
}
