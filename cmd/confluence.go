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
  type = page AND space = TEAM
  title ~ "Onboarding" AND type = page
  text ~ "documentation" AND space = ENG

Examples:
  atl confluence search-cql "space = TEAM"
  atl confluence search-cql "title ~ 'Team Onboarding'"
  atl confluence search-cql "type = page AND space = TEAM" --limit 10`,
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
  atl confluence get-page 3984293906 --status draft
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

Note: Version number handling differs by status:
  - Published pages: Use current version + 1
  - Draft pages: Always use version 1 (drafts don't increment)

Examples:
  atl confluence update-page 3984293906 --title "Updated Title" --body "<p>New content</p>" --version 16
  atl confluence update-page 123456 --title "Draft" --body "<p>Content</p>" --version 1 --status draft`,
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
	// Flags for get-page
	confluenceGetPageStatus string

	// Flags for search-cql
	confluenceSearchLimit      int
	confluenceSearchCursor     string
	confluenceSearchCqlContext string
	confluenceSearchExpand     string
	confluenceSearchNext       bool
	confluenceSearchPrev       bool

	// Flags for get-spaces
	confluenceSpaceKeys           []string
	confluenceSpaceIDs            []string
	confluenceSpaceType           string
	confluenceSpaceStatus         string
	confluenceSpaceLabels         []string
	confluenceSpaceFavoritedBy    string
	confluenceSpaceNotFavoritedBy string
	confluenceSpaceSort           string
	confluenceSpaceDescFormat     string
	confluenceSpaceIncludeIcon    bool
	confluenceSpaceLimit          int
	confluenceSpaceCursor         string

	// Flags for get-pages-in-space
	confluencePagesTitle    string
	confluencePagesStatus   string
	confluencePagesLimit    int
	confluencePagesCursor   string
	confluencePagesDepth    string
	confluencePagesSort     string
	confluencePagesSubtype  string

	// Flags for create-page
	confluenceCreateSpace   string
	confluenceCreateTitle   string
	confluenceCreateBody    string
	confluenceCreateParent  string
	confluenceCreatePrivate bool

	// Flags for update-page
	confluenceUpdateTitle         string
	confluenceUpdateBody          string
	confluenceUpdateVersion       int
	confluenceUpdateParent        string
	confluenceUpdateSpace         string
	confluenceUpdateStatus        string
	confluenceUpdateVersionMsg    string

	// Flags for add-comment
	confluenceCommentParentID     string
	confluenceCommentAttachmentID string
	confluenceCommentCustomID     string

	// Flags for get-page-descendants
	confluenceDescendantsDepth int
	confluenceDescendantsLimit int

	// Flags for get-page-comments
	confluenceCommentsLimit  int
	confluenceCommentsStart  int
	confluenceCommentsStatus string

	// Flags for create-inline-comment
	confluenceInlineTextSelection      string
	confluenceInlineMatchIndex         int
	confluenceInlineMatchCount         int
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
	confluenceCmd.AddCommand(confluenceGetPageAncestorsCmd)
	confluenceCmd.AddCommand(confluenceGetPageDescendantsCmd)
	confluenceCmd.AddCommand(confluenceGetPageCommentsCmd)
	confluenceCmd.AddCommand(confluenceCreateInlineCommentCmd)

	// Flags for search-cql
	confluenceSearchCQLCmd.Flags().IntVar(&confluenceSearchLimit, "limit", 25, "Maximum number of results (max 250)")
	confluenceSearchCQLCmd.Flags().StringVar(&confluenceSearchCursor, "cursor", "", "Pagination cursor")
	confluenceSearchCQLCmd.Flags().StringVar(&confluenceSearchCqlContext, "cql-context", "", "CQL context for query execution")
	confluenceSearchCQLCmd.Flags().StringVar(&confluenceSearchExpand, "expand", "", "Properties to expand")
	confluenceSearchCQLCmd.Flags().BoolVar(&confluenceSearchNext, "next", false, "Include next page link")
	confluenceSearchCQLCmd.Flags().BoolVar(&confluenceSearchPrev, "prev", false, "Include previous page link")
	confluenceSearchCQLCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page
	confluenceGetPageCmd.Flags().StringVar(&confluenceGetPageStatus, "status", "", "Page status (current, draft, archived, trashed)")
	confluenceGetPageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-spaces
	confluenceGetSpacesCmd.Flags().StringSliceVar(&confluenceSpaceKeys, "keys", []string{}, "Filter by space keys")
	confluenceGetSpacesCmd.Flags().StringSliceVar(&confluenceSpaceIDs, "ids", []string{}, "Filter by space IDs")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceType, "type", "", "Filter by type (global, personal, etc)")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceStatus, "status", "", "Filter by status (current, archived)")
	confluenceGetSpacesCmd.Flags().StringSliceVar(&confluenceSpaceLabels, "labels", []string{}, "Filter by labels")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceFavoritedBy, "favorited-by", "", "Filter spaces favorited by user account ID")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceNotFavoritedBy, "not-favorited-by", "", "Filter spaces not favorited by user account ID")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceSort, "sort", "", "Sort order (id, -id, key, -key, name, -name)")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceDescFormat, "description-format", "", "Format for space descriptions (plain, view)")
	confluenceGetSpacesCmd.Flags().BoolVar(&confluenceSpaceIncludeIcon, "include-icon", false, "Include space icon information")
	confluenceGetSpacesCmd.Flags().IntVar(&confluenceSpaceLimit, "limit", 25, "Maximum number of spaces to return")
	confluenceGetSpacesCmd.Flags().StringVar(&confluenceSpaceCursor, "cursor", "", "Pagination cursor")
	confluenceGetSpacesCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-pages-in-space
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesTitle, "title", "", "Filter by page title")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesStatus, "status", "", "Filter by status (current, archived, deleted, trashed)")
	confluenceGetPagesInSpaceCmd.Flags().IntVar(&confluencePagesLimit, "limit", 25, "Maximum number of pages to return")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesCursor, "cursor", "", "Pagination cursor")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesDepth, "depth", "", "Filter by depth (all, root)")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesSort, "sort", "", "Sort order (id, -id, title, -title, etc)")
	confluenceGetPagesInSpaceCmd.Flags().StringVar(&confluencePagesSubtype, "subtype", "", "Filter by subtype (live for live docs, page for regular pages)")
	confluenceGetPagesInSpaceCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for create-page
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateSpace, "space", "", "Space key (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateTitle, "title", "", "Page title (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateBody, "body", "", "Page body in HTML storage format (required)")
	confluenceCreatePageCmd.Flags().StringVar(&confluenceCreateParent, "parent", "", "Parent page ID")
	confluenceCreatePageCmd.Flags().BoolVar(&confluenceCreatePrivate, "private", false, "Create as private page")
	confluenceCreatePageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	confluenceCreatePageCmd.MarkFlagRequired("space")
	confluenceCreatePageCmd.MarkFlagRequired("title")
	confluenceCreatePageCmd.MarkFlagRequired("body")

	// Flags for update-page
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateTitle, "title", "", "New page title (required)")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateBody, "body", "", "New page body in HTML storage format (required)")
	confluenceUpdatePageCmd.Flags().IntVar(&confluenceUpdateVersion, "version", 0, "New version number (required, must be current version + 1)")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateParent, "parent", "", "New parent page ID")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateSpace, "space", "", "New space key")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateStatus, "status", "", "Page status (current, draft)")
	confluenceUpdatePageCmd.Flags().StringVar(&confluenceUpdateVersionMsg, "version-message", "", "Version message describing changes")
	confluenceUpdatePageCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	confluenceUpdatePageCmd.MarkFlagRequired("title")
	confluenceUpdatePageCmd.MarkFlagRequired("body")
	confluenceUpdatePageCmd.MarkFlagRequired("version")

	// Flags for add-comment
	confluenceAddCommentCmd.Flags().StringVar(&confluenceCommentParentID, "parent-comment-id", "", "Parent comment ID for replies")
	confluenceAddCommentCmd.Flags().StringVar(&confluenceCommentAttachmentID, "attachment-id", "", "Attachment ID to add to comment")
	confluenceAddCommentCmd.Flags().StringVar(&confluenceCommentCustomID, "custom-content-id", "", "Custom content ID to add to comment")
	confluenceAddCommentCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page-ancestors
	confluenceGetPageAncestorsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page-descendants
	confluenceGetPageDescendantsCmd.Flags().IntVar(&confluenceDescendantsDepth, "depth", 0, "Maximum depth to traverse")
	confluenceGetPageDescendantsCmd.Flags().IntVar(&confluenceDescendantsLimit, "limit", 25, "Maximum number of descendants")
	confluenceGetPageDescendantsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-page-comments
	confluenceGetPageCommentsCmd.Flags().IntVar(&confluenceCommentsLimit, "limit", 25, "Maximum number of comments")
	confluenceGetPageCommentsCmd.Flags().IntVar(&confluenceCommentsStart, "start", 0, "Starting index for pagination")
	confluenceGetPageCommentsCmd.Flags().StringVar(&confluenceCommentsStatus, "status", "", "Filter by status")
	confluenceGetPageCommentsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for create-inline-comment
	confluenceCreateInlineCommentCmd.Flags().StringVar(&confluenceInlineTextSelection, "text-selection", "", "Text to highlight (required for inline)")
	confluenceCreateInlineCommentCmd.Flags().IntVar(&confluenceInlineMatchIndex, "match-index", 0, "Match index (0-based)")
	confluenceCreateInlineCommentCmd.Flags().IntVar(&confluenceInlineMatchCount, "match-count", 1, "Total number of matches")
	confluenceCreateInlineCommentCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	confluenceCreateInlineCommentCmd.MarkFlagRequired("text-selection")
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
		Limit:      confluenceSearchLimit,
		Cursor:     confluenceSearchCursor,
		CqlContext: confluenceSearchCqlContext,
		Expand:     confluenceSearchExpand,
		Next:       confluenceSearchNext,
		Prev:       confluenceSearchPrev,
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
	opts := &atlassian.GetPageOptions{
		Status: confluenceGetPageStatus,
	}

	page, err := client.GetConfluencePage(pageID, opts)
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

func printConfluenceSearchResults(result map[string]any, site string) {
	results, _ := result["results"].([]any)
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No content found.")
		return
	}

	fmt.Printf("Found %d result(s)\n\n", int(size))

	for i, item := range results {
		if content, ok := item.(map[string]any); ok {
			id, _ := content["id"].(string)
			title, _ := content["title"].(string)
			contentType, _ := content["type"].(string)

			space, _ := content["space"].(map[string]any)
			spaceName, _ := space["name"].(string)
			spaceKey, _ := space["key"].(string)

			// Construct web URL
			webURL := ""
			if links, ok := content["_links"].(map[string]any); ok {
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

func printConfluencePagePretty(page map[string]any, site string) {
	id, _ := page["id"].(string)
	title, _ := page["title"].(string)
	pageType, _ := page["type"].(string)

	space, _ := page["space"].(map[string]any)
	spaceName, _ := space["name"].(string)
	spaceKey, _ := space["key"].(string)

	version, _ := page["version"].(map[string]any)
	versionNum, _ := version["number"].(float64)

	history, _ := page["history"].(map[string]any)
	createdBy, _ := history["createdBy"].(map[string]any)
	createdByName, _ := createdBy["displayName"].(string)

	// Construct web URL
	webURL := ""
	if links, ok := page["_links"].(map[string]any); ok {
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
	body, _ := page["body"].(map[string]any)
	if body != nil {
		storage, _ := body["storage"].(map[string]any)
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
	opts := &atlassian.GetSpacesOptions{
		Keys:              confluenceSpaceKeys,
		IDs:               confluenceSpaceIDs,
		Type:              confluenceSpaceType,
		Status:            confluenceSpaceStatus,
		Labels:            confluenceSpaceLabels,
		FavoritedBy:       confluenceSpaceFavoritedBy,
		NotFavoritedBy:    confluenceSpaceNotFavoritedBy,
		Sort:              confluenceSpaceSort,
		DescriptionFormat: confluenceSpaceDescFormat,
		IncludeIcon:       confluenceSpaceIncludeIcon,
		Limit:             confluenceSpaceLimit,
		Cursor:            confluenceSpaceCursor,
	}

	result, err := client.GetConfluenceSpaces(opts)
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
		Cursor:   confluencePagesCursor,
		Depth:    confluencePagesDepth,
		Sort:     confluencePagesSort,
		Subtype:  confluencePagesSubtype,
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
		SpaceKey:  confluenceCreateSpace,
		Title:     confluenceCreateTitle,
		Body:      confluenceCreateBody,
		ParentID:  confluenceCreateParent,
		IsPrivate: confluenceCreatePrivate,
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
		if links, ok := result["_links"].(map[string]any); ok {
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
		PageID:         pageID,
		Title:          confluenceUpdateTitle,
		Body:           confluenceUpdateBody,
		Version:        confluenceUpdateVersion,
		ParentID:       confluenceUpdateParent,
		SpaceKey:       confluenceUpdateSpace,
		Status:         confluenceUpdateStatus,
		VersionMessage: confluenceUpdateVersionMsg,
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
		version, _ := result["version"].(map[string]any)
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
		PageID:          pageID,
		Comment:         comment,
		ParentCommentID: confluenceCommentParentID,
		AttachmentID:    confluenceCommentAttachmentID,
		CustomContentID: confluenceCommentCustomID,
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

func printSpacesList(result map[string]any, site string) {
	results, _ := result["results"].([]any)
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No spaces found.")
		return
	}

	fmt.Printf("Found %d space(s)\n\n", int(size))

	for i, item := range results {
		if space, ok := item.(map[string]any); ok {
			key, _ := space["key"].(string)
			name, _ := space["name"].(string)
			spaceType, _ := space["type"].(string)

			// Construct web URL
			webURL := ""
			if links, ok := space["_links"].(map[string]any); ok {
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

func printPagesList(result map[string]any, site string) {
	results, _ := result["results"].([]any)
	size, _ := result["size"].(float64)

	if len(results) == 0 {
		fmt.Println("No pages found.")
		return
	}

	fmt.Printf("Found %d page(s)\n\n", int(size))

	for i, item := range results {
		if page, ok := item.(map[string]any); ok {
			id, _ := page["id"].(string)
			title, _ := page["title"].(string)

			// Construct web URL
			webURL := ""
			if links, ok := page["_links"].(map[string]any); ok {
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

var confluenceGetPageAncestorsCmd = &cobra.Command{
	Use:   "get-page-ancestors <pageID>",
	Short: "Get parent pages of a Confluence page",
	Long: `Retrieve the ancestor (parent) pages of a Confluence page, showing the page hierarchy.

Examples:
  atl confluence get-page-ancestors 3984293906`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceGetPageAncestors,
}

var confluenceGetPageDescendantsCmd = &cobra.Command{
	Use:   "get-page-descendants <pageID>",
	Short: "Get child pages of a Confluence page",
	Long: `Retrieve the descendant (child) pages of a Confluence page.

Examples:
  atl confluence get-page-descendants 3984293906
  atl confluence get-page-descendants 3984293906 --depth 2`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceGetPageDescendants,
}

var confluenceGetPageCommentsCmd = &cobra.Command{
	Use:   "get-page-comments <pageID>",
	Short: "Get comments on a Confluence page",
	Long: `Retrieve comments on a Confluence page.

Examples:
  atl confluence get-page-comments 3984293906
  atl confluence get-page-comments 3984293906 --limit 50`,
	Args: cobra.ExactArgs(1),
	RunE: runConfluenceGetPageComments,
}

var confluenceCreateInlineCommentCmd = &cobra.Command{
	Use:   "create-inline-comment <pageID> <comment>",
	Short: "Create an inline comment on a Confluence page",
	Long: `Create an inline comment attached to specific text on a Confluence page.

Examples:
  atl confluence create-inline-comment 3984293906 "<p>Great point!</p>" \
    --text-selection "specific text to highlight" \
    --match-index 0 \
    --match-count 1`,
	Args: cobra.ExactArgs(2),
	RunE: runConfluenceCreateInlineComment,
}

func runConfluenceGetPageAncestors(cmd *cobra.Command, args []string) error {
	pageID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	ancestors, err := client.GetPageAncestors(pageID)
	if err != nil {
		return fmt.Errorf("failed to get ancestors: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(ancestors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(ancestors) == 0 {
			fmt.Println("No ancestors (this is a root page)")
			return nil
		}

		fmt.Printf("Page hierarchy for page %s:\n\n", pageID)

		for i, ancestor := range ancestors {
			title, _ := ancestor["title"].(string)
			id, _ := ancestor["id"].(string)
			fmt.Printf("%d. %s (ID: %s)\n", i+1, title, id)
		}
	}

	return nil
}

func runConfluenceGetPageDescendants(cmd *cobra.Command, args []string) error {
	pageID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	opts := &atlassian.GetPageDescendantsOptions{
		Depth: confluenceDescendantsDepth,
		Limit: confluenceDescendantsLimit,
	}

	result, err := client.GetPageDescendants(pageID, opts)
	if err != nil {
		return fmt.Errorf("failed to get descendants: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		results, _ := result["results"].([]any)

		if len(results) == 0 {
			fmt.Printf("No child pages found for page %s\n", pageID)
			return nil
		}

		fmt.Printf("Child pages of page %s:\n\n", pageID)

		for i, item := range results {
			if page, ok := item.(map[string]any); ok {
				title, _ := page["title"].(string)
				id, _ := page["id"].(string)
				fmt.Printf("%d. %s (ID: %s)\n", i+1, title, id)
			}
		}
	}

	return nil
}

func runConfluenceGetPageComments(cmd *cobra.Command, args []string) error {
	pageID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	opts := &atlassian.GetPageCommentsOptions{
		Limit:  confluenceCommentsLimit,
		Start:  confluenceCommentsStart,
		Status: confluenceCommentsStatus,
	}

	result, err := client.GetPageComments(pageID, opts)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		results, _ := result["results"].([]any)

		if len(results) == 0 {
			fmt.Printf("No comments found for page %s\n", pageID)
			return nil
		}

		fmt.Printf("Comments on page %s:\n\n", pageID)

		for i, item := range results {
			if comment, ok := item.(map[string]any); ok {
				id, _ := comment["id"].(string)
				title, _ := comment["title"].(string)
				body, _ := comment["body"].(map[string]any)

				fmt.Printf("%d. Comment ID: %s\n", i+1, id)
				if title != "" {
					fmt.Printf("   Title: %s\n", title)
				}

				if body != nil {
					storage, _ := body["storage"].(map[string]any)
					if storage != nil {
						value, _ := storage["value"].(string)
						if value != "" {
							// Convert HTML to text
							commentText := atlassian.HTMLToText(value)
							if len(commentText) > 100 {
								commentText = commentText[:100] + "..."
							}
							fmt.Printf("   %s\n", commentText)
						}
					}
				}
				fmt.Println()
			}
		}
	}

	return nil
}

func runConfluenceCreateInlineComment(cmd *cobra.Command, args []string) error {
	pageID := args[0]
	comment := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	opts := &atlassian.CreateInlineCommentOptions{
		PageID:                  pageID,
		Comment:                 comment,
		TextSelection:           confluenceInlineTextSelection,
		TextSelectionMatchIndex: confluenceInlineMatchIndex,
		TextSelectionMatchCount: confluenceInlineMatchCount,
	}

	result, err := client.CreateInlineComment(opts)
	if err != nil {
		return fmt.Errorf("failed to create inline comment: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		id, _ := result["id"].(string)
		fmt.Printf("✓ Created inline comment on page %s\n", pageID)
		fmt.Printf("  Comment ID: %s\n", id)
		fmt.Printf("  Highlighted text: %s\n", confluenceInlineTextSelection)
	}

	return nil
}
