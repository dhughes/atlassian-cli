package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doughughes/atlassian-cli/internal/atlassian"
	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var confluenceCmd = &cobra.Command{
	Use:   "confluence",
	Short: "Work with Confluence pages, spaces, and search",
	Long:  `Interact with Confluence Cloud using the REST API.`,
}

var confluenceSearchCQLCmd = &cobra.Command{
	Use:   "search-cql <cql-query>",
	Short: "Search Confluence content using CQL",
	Long: `Search for Confluence pages using CQL (Confluence Query Language).

CQL is a powerful query language for finding content. Examples:
  type = page AND space = POL
  title ~ "Onboarding" AND type = page
  text ~ "documentation" AND space = ENG

Examples:
  atl confluence search-cql "space = POL"
  atl confluence search-cql "title ~ 'FX Onboarding'"
  atl confluence search-cql "type = page AND space = POL" --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceSearchCQL,
}

var confluenceGetPageCmd = &cobra.Command{
	Use:   "get-page <pageID>",
	Short: "Get a Confluence page by ID",
	Long: `Retrieve detailed information about a Confluence page by its ID.

The page ID can be extracted from the page URL.
Example URL: https://site.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Title
Page ID: 123456789

Examples:
  atl confluence get-page 3984293906
  atl confluence get-page 3984293906 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceGetPage,
}

var (
	// Flags for search-cql
	confluenceSearchLimit  int
	confluenceSearchCursor string
)

func init() {
	rootCmd.AddCommand(confluenceCmd)
	confluenceCmd.AddCommand(confluenceSearchCQLCmd)
	confluenceCmd.AddCommand(confluenceGetPageCmd)

	// Flags for search-cql
	confluenceSearchCQLCmd.Flags().IntVar(&confluenceSearchLimit, "limit", 25, "Maximum number of results (max 250)")
	confluenceSearchCQLCmd.Flags().StringVar(&confluenceSearchCursor, "cursor", "", "Pagination cursor")
	confluenceSearchCQLCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page
	confluenceGetPageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
}

func runConfluenceSearchCQL(cmd *cobra.Command, args []string) error {
	cql := args[0]

	// Validate limit
	if confluenceSearchLimit > 250 {
		return fmt.Errorf("limit cannot exceed 250")
	}

	// Load config and get active account
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	// Create client
	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	// Build request options
	opts := &atlassian.SearchCQLOptions{
		Limit:  confluenceSearchLimit,
		Cursor: confluenceSearchCursor,
	}

	// Search content
	result, err := client.SearchConfluenceCQL(cql, opts)
	if err != nil {
		return fmt.Errorf("failed to search content: %w", err)
	}

	// Output
	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		printConfluenceSearchResults(result, account.Site)
	}

	return nil
}

func runConfluenceGetPage(cmd *cobra.Command, args []string) error {
	pageID := args[0]

	// Load config and get active account
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	// Create client
	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	// Get page
	page, err := client.GetConfluencePage(pageID)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}

	// Output
	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(page, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		printConfluencePagePretty(page, account.Site)
	}

	return nil
}

func printConfluenceSearchResults(result map[string]interface{}, site string) {
	results, _ := result["results"].([]interface{})
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No content found.")
		return
	}

	fmt.Printf("Found %d result(s)\n\n", int(size))

	for i, item := range results {
		if content, ok := item.(map[string]interface{}); ok {
			id, _ := content["id"].(string)
			title, _ := content["title"].(string)
			contentType, _ := content["type"].(string)

			space, _ := content["space"].(map[string]interface{})
			spaceName, _ := space["name"].(string)
			spaceKey, _ := space["key"].(string)

			// Construct web URL
			webURL := ""
			if links, ok := content["_links"].(map[string]interface{}); ok {
				if webui, ok := links["webui"].(string); ok {
					webURL = fmt.Sprintf("%s/wiki%s", site, webui)
					if !strings.HasPrefix(site, "http") {
						webURL = "https://" + webURL
					}
				}
			}

			fmt.Printf("%d. %s (ID: %s)\n", i+1, title, id)
			fmt.Printf("   Type: %s | Space: %s (%s)\n", contentType, spaceName, spaceKey)
			if webURL != "" {
				fmt.Printf("   Link: %s\n", webURL)
			}
			fmt.Println()
		}
	}

	fmt.Printf("\nTo view a page: atl confluence get-page <page-id>\n")
	fmt.Printf("For JSON output: atl confluence search-cql \"<query>\" --json\n")
}

func printConfluencePagePretty(page map[string]interface{}, site string) {
	id, _ := page["id"].(string)
	title, _ := page["title"].(string)
	pageType, _ := page["type"].(string)

	space, _ := page["space"].(map[string]interface{})
	spaceName, _ := space["name"].(string)
	spaceKey, _ := space["key"].(string)

	version, _ := page["version"].(map[string]interface{})
	versionNum, _ := version["number"].(float64)

	history, _ := page["history"].(map[string]interface{})
	createdBy, _ := history["createdBy"].(map[string]interface{})
	createdByName, _ := createdBy["displayName"].(string)

	// Construct web URL
	webURL := ""
	if links, ok := page["_links"].(map[string]interface{}); ok {
		if webui, ok := links["webui"].(string); ok {
			webURL = fmt.Sprintf("%s/wiki%s", site, webui)
			if !strings.HasPrefix(site, "http") {
				webURL = "https://" + webURL
			}
		}
	}

	fmt.Printf("Page: %s (ID: %s)\n", title, id)
	fmt.Printf("Type: %s\n", pageType)
	fmt.Printf("Space: %s (%s)\n", spaceName, spaceKey)
	fmt.Printf("Version: %d\n", int(versionNum))
	if createdByName != "" {
		fmt.Printf("Created by: %s\n", createdByName)
	}
	if webURL != "" {
		fmt.Printf("Link: %s\n", webURL)
	}

	// Get page body content
	body, _ := page["body"].(map[string]interface{})
	if body != nil {
		storage, _ := body["storage"].(map[string]interface{})
		if storage != nil {
			value, _ := storage["value"].(string)
			if value != "" {
				fmt.Printf("\nContent:\n")
				// Convert HTML to readable text
				contentText := atlassian.HTMLToText(value)
				if contentText != "" {
					// Indent content
					lines := strings.Split(contentText, "\n")
					for _, line := range lines {
						fmt.Printf("  %s\n", line)
					}
				} else {
					fmt.Printf("  (empty)\n")
				}
			}
		}
	}

	fmt.Printf("\n---\n")
	fmt.Printf("For JSON output: atl confluence get-page %s --json\n", id)
}
