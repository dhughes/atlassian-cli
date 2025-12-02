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

var confluenceGetSpacesCmd = &cobra.Command{
	Use:   "get-spaces",
	Short: "List Confluence spaces",
	Long: `Retrieve a list of Confluence spaces.

Examples:
  atl confluence get-spaces
  atl confluence get-spaces --keys POL,ENG
  atl confluence get-spaces --limit 50`,
	RunE: runConfluenceGetSpaces,
}

var confluenceGetPagesInSpaceCmd = &cobra.Command{
	Use:   "get-pages-in-space <spaceKey>",
	Short: "List pages in a Confluence space",
	Long: `Retrieve pages within a specific Confluence space.

Examples:
  atl confluence get-pages-in-space POL
  atl confluence get-pages-in-space POL --title "Onboarding"
  atl confluence get-pages-in-space POL --limit 50`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceGetPagesInSpace,
}

var confluenceCreatePageCmd = &cobra.Command{
	Use:   "create-page",
	Short: "Create a new Confluence page",
	Long: `Create a new page in a Confluence space.

Examples:
  atl confluence create-page --space POL --title "New Page" --body "<p>Content here</p>"
  atl confluence create-page --space POL --title "Child Page" --body "<p>Content</p>" --parent 123456`,
	RunE: runConfluenceCreatePage,
}

var confluenceUpdatePageCmd = &cobra.Command{
	Use:   "update-page <pageID>",
	Short: "Update a Confluence page",
	Long: `Update an existing Confluence page.

Note: You must provide the current version number + 1 for the update to succeed.

Examples:
  atl confluence update-page 3984293906 --title "Updated Title" --body "<p>New content</p>" --version 16`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceUpdatePage,
}

var confluenceAddCommentCmd = &cobra.Command{
	Use:   "add-comment <pageID> <comment>",
	Short: "Add a comment to a Confluence page",
	Long: `Add a comment to an existing Confluence page.

Examples:
  atl confluence add-comment 3984293906 "<p>This is a comment</p>"`,
	Args: cobra.ExactArgs(2),
	RunE: runConfluenceAddComment,
}

var (
	// Flags for search-cql
	confluenceSearchLimit  int
	confluenceSearchCursor string

	// Flags for get-spaces
	confluenceSpaceKeys  []string
	confluenceSpaceLimit int

	// Flags for get-pages-in-space
	confluencePagesTitle  string
	confluencePagesStatus string
	confluencePagesLimit  int

	// Flags for create-page
	confluenceCreateSpace  string
	confluenceCreateTitle  string
	confluenceCreateBody   string
	confluenceCreateParent string

	// Flags for update-page
	confluenceUpdateTitle   string
	confluenceUpdateBody    string
	confluenceUpdateVersion int
)

func init() {
	rootCmd.AddCommand(confluenceCmd)
	confluenceCmd.AddCommand(confluenceSearchCQLCmd)
	confluenceCmd.AddCommand(confluenceGetPageCmd)
	confluenceCmd.AddCommand(confluenceGetSpacesCmd)
	confluenceCmd.AddCommand(confluenceGetPagesInSpaceCmd)
	confluenceCmd.AddCommand(confluenceCreatePageCmd)
	confluenceCmd.AddCommand(confluenceUpdatePageCmd)
	confluenceCmd.AddCommand(confluenceAddCommentCmd)

	// Flags for search-cql
	confluenceSearchCQLCmd.Flags().IntVar(&confluenceSearchLimit, "limit", 25, "Maximum number of results (max 250)")
	confluenceSearchCQLCmd.Flags().StringVar(&confluenceSearchCursor, "cursor", "", "Pagination cursor")
	confluenceSearchCQLCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page
	confluenceGetPageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-spaces
	confluenceGetSpacesCmd.Flags().StringSliceVar(&confluenceSpaceKeys, "keys", []string{}, "Filter by space keys")
	confluenceGetSpacesCmd.Flags().IntVar(&confluenceSpaceLimit, "limit", 25, "Maximum number of spaces to return")
	confluenceGetSpacesCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-pages-in-space
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesTitle, "title", "", "Filter by page title")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesStatus, "status", "", "Filter by status (current, archived, trashed)")
	confluenceGetPagesInSpaceCmd.Flags().IntVar(&confluencePagesLimit, "limit", 25, "Maximum number of pages to return")
	confluenceGetPagesInSpaceCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for create-page
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateSpace, "space", "", "Space key (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateTitle, "title", "", "Page title (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateBody, "body", "", "Page body in HTML storage format (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateParent, "parent", "", "Parent page ID")
	confluenceCreatePageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	confluenceCreatePageCmd.MarkFlagRequired("space")
	confluenceCreatePageCmd.MarkFlagRequired("title")
	confluenceCreatePageCmd.MarkFlagRequired("body")

	// Flags for update-page
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateTitle, "title", "", "New page title (required)")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateBody, "body", "", "New page body in HTML storage format (required)")
	confluenceUpdatePageCmd.Flags().IntVar(&confluenceUpdateVersion, "version", 0, "New version number (required, must be current version + 1)")
	confluenceUpdatePageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	confluenceUpdatePageCmd.MarkFlagRequired("title")
	confluenceUpdatePageCmd.MarkFlagRequired("body")
	confluenceUpdatePageCmd.MarkFlagRequired("version")

	// Flags for add-comment
	confluenceAddCommentCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
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

func runConfluenceGetSpaces(cmd *cobra.Command, args []string) error {
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

	// Get spaces
	result, err := client.GetConfluenceSpaces(confluenceSpaceKeys, confluenceSpaceLimit)
	if err != nil {
		return fmt.Errorf("failed to get spaces: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printSpacesList(result, account.Site)
	}

	return nil
}

func runConfluenceGetPagesInSpace(cmd *cobra.Command, args []string) error {
	spaceKey := args[0]

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

	// Get pages
	opts := &atlassian.GetPagesInSpaceOptions{
		SpaceKey: spaceKey,
		Title:    confluencePagesTitle,
		Status:   confluencePagesStatus,
		Limit:    confluencePagesLimit,
	}

	result, err := client.GetPagesInSpace(opts)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printPagesList(result, account.Site)
	}

	return nil
}

func runConfluenceCreatePage(cmd *cobra.Command, args []string) error {
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

	// Create page
	opts := &atlassian.CreatePageOptions{
		SpaceKey: confluenceCreateSpace,
		Title:    confluenceCreateTitle,
		Body:     confluenceCreateBody,
		ParentID: confluenceCreateParent,
	}

	result, err := client.CreateConfluencePage(opts)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		id, _ := result["id"].(string)
		title, _ := result["title"].(string)

		// Construct web URL
		webURL := ""
		if links, ok := result["_links"].(map[string]interface{}); ok {
			if webui, ok := links["webui"].(string); ok {
				webURL = fmt.Sprintf("%s/wiki%s", account.Site, webui)
				if !strings.HasPrefix(account.Site, "http") {
					webURL = "https://" + webURL
				}
			}
		}

		fmt.Printf("✓ Created page: %s (ID: %s)\n", title, id)
		if webURL != "" {
			fmt.Printf("  Link: %s\n", webURL)
		}
		fmt.Printf("\nView page: atl confluence get-page %s\n", id)
	}

	return nil
}

func runConfluenceUpdatePage(cmd *cobra.Command, args []string) error {
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

	// Update page
	opts := &atlassian.UpdatePageOptions{
		PageID:  pageID,
		Title:   confluenceUpdateTitle,
		Body:    confluenceUpdateBody,
		Version: confluenceUpdateVersion,
	}

	result, err := client.UpdateConfluencePage(opts)
	if err != nil {
		return fmt.Errorf("failed to update page: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		title, _ := result["title"].(string)
		version, _ := result["version"].(map[string]interface{})
		versionNum, _ := version["number"].(float64)

		fmt.Printf("✓ Updated page: %s\n", title)
		fmt.Printf("  Version: %d\n", int(versionNum))
		fmt.Printf("\nView page: atl confluence get-page %s\n", pageID)
	}

	return nil
}

func runConfluenceAddComment(cmd *cobra.Command, args []string) error {
	pageID := args[0]
	comment := args[1]

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

	// Add comment
	opts := &atlassian.AddPageCommentOptions{
		PageID:  pageID,
		Comment: comment,
	}

	result, err := client.AddConfluencePageComment(opts)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		id, _ := result["id"].(string)
		fmt.Printf("✓ Added comment to page %s\n", pageID)
		fmt.Printf("  Comment ID: %s\n", id)
	}

	return nil
}

func printSpacesList(result map[string]interface{}, site string) {
	results, _ := result["results"].([]interface{})
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No spaces found.")
		return
	}

	fmt.Printf("Found %d space(s)\n\n", int(size))

	for i, item := range results {
		if space, ok := item.(map[string]interface{}); ok {
			key, _ := space["key"].(string)
			name, _ := space["name"].(string)
			spaceType, _ := space["type"].(string)

			// Construct web URL
			webURL := ""
			if links, ok := space["_links"].(map[string]interface{}); ok {
				if webui, ok := links["webui"].(string); ok {
					webURL = fmt.Sprintf("%s/wiki%s", site, webui)
					if !strings.HasPrefix(site, "http") {
						webURL = "https://" + webURL
					}
				}
			}

			fmt.Printf("%d. %s (%s)\n", i+1, name, key)
			fmt.Printf("   Type: %s\n", spaceType)
			if webURL != "" {
				fmt.Printf("   Link: %s\n", webURL)
			}
			fmt.Println()
		}
	}

	fmt.Printf("\nFor JSON output: atl confluence get-spaces --json\n")
}

func printPagesList(result map[string]interface{}, site string) {
	results, _ := result["results"].([]interface{})
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No pages found.")
		return
	}

	fmt.Printf("Found %d page(s)\n\n", int(size))

	for i, item := range results {
		if page, ok := item.(map[string]interface{}); ok {
			id, _ := page["id"].(string)
			title, _ := page["title"].(string)

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

			fmt.Printf("%d. %s (ID: %s)\n", i+1, title, id)
			if webURL != "" {
				fmt.Printf("   Link: %s\n", webURL)
			}
			fmt.Println()
		}
	}

	fmt.Printf("\nTo view a page: atl confluence get-page <page-id>\n")
	fmt.Printf("For JSON output: atl confluence get-pages-in-space <space> --json\n")
}
