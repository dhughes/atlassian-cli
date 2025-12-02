package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/doughughes/atlassian-cli/internal/atlassian"
	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Meta commands for cross-product operations",
	Long:  `Commands that work across Jira and Confluence or provide meta information.`,
}

var metaUserInfoCmd = &cobra.Command{
	Use:   "user-info",
	Short: "Get current user information",
	Long: `Retrieve information about the currently authenticated user.

Examples:
  atl meta user-info
  atl meta user-info --json`,
	RunE: runMetaUserInfo,
}

var metaGetResourcesCmd = &cobra.Command{
	Use:   "get-resources",
	Short: "Get accessible Atlassian resources",
	Long: `List all Atlassian cloud resources (sites) accessible with current credentials.

Note: This endpoint requires OAuth, not Basic Auth. It will likely fail with API tokens.

Examples:
  atl meta get-resources`,
	RunE: runMetaGetResources,
}

var metaSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across Jira and Confluence (Rovo search)",
	Long: `Perform a unified search across both Jira and Confluence using Rovo search.

Note: This requires Rovo to be enabled and may not work with all Atlassian setups.

Examples:
  atl meta search "authentication"
  atl meta search "FX project documentation"`,
	Args: cobra.ExactArgs(1),
	RunE: runMetaSearch,
}

var metaFetchCmd = &cobra.Command{
	Use:   "fetch <ari>",
	Short: "Fetch a resource by ARI",
	Long: `Fetch any Atlassian resource by its ARI (Atlassian Resource Identifier).

ARI format: ari:cloud:<product>:<cloudId>:<resource-type>/<resource-id>

Examples:
  atl meta fetch "ari:cloud:jira:abc123:issue/10107"
  atl meta fetch "ari:cloud:confluence:abc123:page/123456789"`,
	Args: cobra.ExactArgs(1),
	RunE: runMetaFetch,
}

func init() {
	rootCmd.AddCommand(metaCmd)
	metaCmd.AddCommand(metaUserInfoCmd)
	metaCmd.AddCommand(metaGetResourcesCmd)
	metaCmd.AddCommand(metaSearchCmd)
	metaCmd.AddCommand(metaFetchCmd)

	// Flags
	metaUserInfoCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	metaGetResourcesCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	metaSearchCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	metaFetchCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
}

func runMetaUserInfo(cmd *cobra.Command, args []string) error {
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

	// Get user info
	user, err := client.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Printf("User: %s\n", user.DisplayName)
		fmt.Printf("Account ID: %s\n", user.AccountID)
		fmt.Printf("Email: %s\n", user.Email)
		fmt.Printf("Account Type: %s\n", user.AccountType)
		fmt.Printf("Active: %t\n", user.Active)
		fmt.Printf("Locale: %s\n", user.Locale)
	}

	return nil
}

func runMetaGetResources(cmd *cobra.Command, args []string) error {
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

	// Get resources (Note: This may fail with Basic Auth)
	resources, err := client.GetAccessibleResources()
	if err != nil {
		return fmt.Errorf("failed to get resources: %w\n\nNote: This endpoint requires OAuth and may not work with API tokens", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(resources, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(resources) == 0 {
			fmt.Println("No accessible resources found.")
			return nil
		}

		fmt.Printf("Found %d accessible resource(s):\n\n", len(resources))

		for i, res := range resources {
			fmt.Printf("%d. %s\n", i+1, res.Name)
			fmt.Printf("   ID: %s\n", res.ID)
			fmt.Printf("   URL: %s\n", res.URL)
			fmt.Println()
		}
	}

	return nil
}

func runMetaSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	result, err := client.RovoSearch(query)
	if err != nil {
		return fmt.Errorf("failed to search: %w\n\nNote: Rovo search may not be available on all Atlassian instances", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		results, _ := result["results"].([]interface{})

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("Found results:\n\n")

		for i, item := range results {
			if res, ok := item.(map[string]interface{}); ok {
				title, _ := res["title"].(string)
				resType, _ := res["type"].(string)
				url, _ := res["url"].(string)

				fmt.Printf("%d. %s\n", i+1, title)
				fmt.Printf("   Type: %s\n", resType)
				if url != "" {
					fmt.Printf("   URL: %s\n", url)
				}
				fmt.Println()
			}
		}
	}

	return nil
}

func runMetaFetch(cmd *cobra.Command, args []string) error {
	ari := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}

	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	result, err := client.FetchByARI(ari)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	fmt.Println(string(output))

	return nil
}
