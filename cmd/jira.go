package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doughughes/atlassian-cli/internal/atlassian"
	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Work with Jira issues, projects, and search",
	Long:  `Interact with Jira Cloud using the REST API v3.`,
}

var jiraGetIssueCmd = &cobra.Command{
	Use:   "get-issue <issueKey>",
	Short: "Get details of a Jira issue",
	Long: `Retrieve detailed information about a Jira issue by its key or ID.

Examples:
  atl jira get-issue PROJ-123
  atl jira get-issue 10000
  atl jira get-issue PROJ-123 --json
  atl jira get-issue PROJ-123 --fields summary,status,assignee`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraGetIssue,
}

var (
	// Flags for get-issue
	jiraGetIssueFields []string
	jiraGetIssueExpand []string
	outputJSON         bool
	outputPretty       bool
)

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraGetIssueCmd)

	// Flags for get-issue
	jiraGetIssueCmd.Flags().StringSliceVar(&jiraGetIssueFields, "fields", []string{}, "Comma-separated list of fields to return")
	jiraGetIssueCmd.Flags().StringSliceVar(&jiraGetIssueExpand, "expand", []string{}, "Comma-separated list of parameters to expand")
	jiraGetIssueCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	jiraGetIssueCmd.Flags().BoolVar(&outputPretty, "pretty", false, "Human-readable formatted output (default)")
}

func runJiraGetIssue(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

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
	opts := &atlassian.GetIssueOptions{
		Fields: jiraGetIssueFields,
		Expand: jiraGetIssueExpand,
	}

	// Get issue
	issue, err := client.GetJiraIssue(issueKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Output
	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		printIssuePretty(issue)
	}

	return nil
}

func printIssuePretty(issue map[string]interface{}) {
	// Extract common fields
	key, _ := issue["key"].(string)
	fields, _ := issue["fields"].(map[string]interface{})

	fmt.Printf("Issue: %s\n", key)

	if fields != nil {
		if summary, ok := fields["summary"].(string); ok {
			fmt.Printf("Summary: %s\n", summary)
		}

		if issueType, ok := fields["issuetype"].(map[string]interface{}); ok {
			if name, ok := issueType["name"].(string); ok {
				fmt.Printf("Type: %s\n", name)
			}
		}

		if status, ok := fields["status"].(map[string]interface{}); ok {
			if name, ok := status["name"].(string); ok {
				fmt.Printf("Status: %s\n", name)
			}
		}

		if priority, ok := fields["priority"].(map[string]interface{}); ok {
			if name, ok := priority["name"].(string); ok {
				fmt.Printf("Priority: %s\n", name)
			}
		}

		if assignee, ok := fields["assignee"].(map[string]interface{}); ok {
			if displayName, ok := assignee["displayName"].(string); ok {
				fmt.Printf("Assignee: %s\n", displayName)
			}
		} else {
			fmt.Printf("Assignee: Unassigned\n")
		}

		if reporter, ok := fields["reporter"].(map[string]interface{}); ok {
			if displayName, ok := reporter["displayName"].(string); ok {
				fmt.Printf("Reporter: %s\n", displayName)
			}
		}

		if created, ok := fields["created"].(string); ok {
			fmt.Printf("Created: %s\n", created)
		}

		if updated, ok := fields["updated"].(string); ok {
			fmt.Printf("Updated: %s\n", updated)
		}

		// Parse and display description using ADF parser
		if description, ok := fields["description"]; ok && description != nil {
			fmt.Printf("\nDescription:\n")
			descText := atlassian.ADFToText(description)
			if descText != "" {
				// Indent description text
				lines := strings.Split(descText, "\n")
				for _, line := range lines {
					fmt.Printf("  %s\n", line)
				}
			} else {
				fmt.Printf("  (empty)\n")
			}
		}
	}

	fmt.Printf("\n---\n")
	fmt.Printf("For JSON output: atl jira get-issue %s --json\n", key)
}
