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

var jiraSearchJQLCmd = &cobra.Command{
	Use:   "search-jql <jql-query>",
	Short: "Search Jira issues using JQL",
	Long: `Search for Jira issues using JQL (Jira Query Language).

JQL is a powerful query language for finding issues. Examples:
  project = PROJ AND status = "In Progress"
  assignee = currentUser() AND created >= -7d
  summary ~ "bug" AND priority = High

Examples:
  atl jira search-jql "project = PROJ"
  atl jira search-jql "assignee = currentUser()"
  atl jira search-jql "status = 'In Progress'" --max-results 10
  atl jira search-jql "project = PROJ" --fields summary,status,assignee`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraSearchJQL,
}

var jiraCreateIssueCmd = &cobra.Command{
	Use:   "create-issue",
	Short: "Create a new Jira issue",
	Long: `Create a new issue in a Jira project.

Examples:
  atl jira create-issue --project PROJ --type Task --summary "Do something"
  atl jira create-issue --project PROJ --type Bug --summary "Fix bug" --description "Bug details here"`,
	RunE: runJiraCreateIssue,
}

var jiraAddCommentCmd = &cobra.Command{
	Use:   "add-comment <issueKey> <comment>",
	Short: "Add a comment to a Jira issue",
	Long: `Add a comment to an existing Jira issue.

Examples:
  atl jira add-comment PROJ-123 "This is a comment"
  atl jira add-comment PROJ-123 "Multi-line comment works too"`,
	Args: cobra.ExactArgs(2),
	RunE: runJiraAddComment,
}

var jiraEditIssueCmd = &cobra.Command{
	Use:   "edit-issue <issueKey>",
	Short: "Edit a Jira issue",
	Long: `Update fields on an existing Jira issue.

Examples:
  atl jira edit-issue PROJ-123 --summary "New summary"
  atl jira edit-issue PROJ-123 --description "New description"
  atl jira edit-issue PROJ-123 --summary "Update" --description "Details"`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraEditIssue,
}

var jiraGetTransitionsCmd = &cobra.Command{
	Use:   "get-transitions <issueKey>",
	Short: "Get available transitions for an issue",
	Long: `List all available status transitions for a Jira issue.

Examples:
  atl jira get-transitions PROJ-123`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraGetTransitions,
}

var jiraTransitionIssueCmd = &cobra.Command{
	Use:   "transition-issue <issueKey> <transitionID>",
	Short: "Transition an issue to a new status",
	Long: `Change the status of a Jira issue using a transition ID.

Use 'get-transitions' to see available transition IDs.

Examples:
  atl jira transition-issue PROJ-123 21
  atl jira transition-issue PROJ-123 31`,
	Args: cobra.ExactArgs(2),
	RunE: runJiraTransitionIssue,
}

var jiraLookupAccountIDCmd = &cobra.Command{
	Use:   "lookup-account-id <searchString>",
	Short: "Find user account ID by name or email",
	Long: `Search for Jira users by display name or email address to get their account ID.

Account IDs are needed for assigning issues or setting other user fields.

Examples:
  atl jira lookup-account-id "Doug Hughes"
  atl jira lookup-account-id "doug@example.com"`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraLookupAccountID,
}

var (
	// Flags for get-issue
	jiraGetIssueFields         []string
	jiraGetIssueExpand         []string
	jiraGetIssueProperties     []string
	jiraGetIssueFieldsByKeys   bool
	jiraGetIssueUpdateHistory  bool
	outputJSON                 bool

	// Flags for search-jql
	jiraSearchFields     []string
	jiraSearchMaxResults int
	jiraSearchStartAt    int

	// Flags for create-issue
	jiraCreateProject     string
	jiraCreateType        string
	jiraCreateSummary     string
	jiraCreateDescription string
	jiraCreateAssignee    string
	jiraCreateParent      string
	jiraCreateFields      string

	// Flags for edit-issue
	jiraEditSummary     string
	jiraEditDescription string
	jiraEditAssignee    string
	jiraEditFields      string

	// Flags for add-comment
	jiraCommentVisibilityType  string
	jiraCommentVisibilityValue string

	// Flags for get-transitions
	jiraGetTransitionsExpand                      string
	jiraGetTransitionsTransitionID                string
	jiraGetTransitionsIncludeUnavailable          bool
	jiraGetTransitionsSkipRemoteOnly              bool
	jiraGetTransitionsSortByOpsBarAndStatus       bool

	// Flags for transition-issue
	jiraTransitionFields          string
	jiraTransitionUpdate          string
	jiraTransitionHistoryMetadata string
)

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraGetIssueCmd)
	jiraCmd.AddCommand(jiraSearchJQLCmd)
	jiraCmd.AddCommand(jiraCreateIssueCmd)
	jiraCmd.AddCommand(jiraAddCommentCmd)
	jiraCmd.AddCommand(jiraEditIssueCmd)
	jiraCmd.AddCommand(jiraGetTransitionsCmd)
	jiraCmd.AddCommand(jiraTransitionIssueCmd)
	jiraCmd.AddCommand(jiraLookupAccountIDCmd)

	// Flags for get-issue
	jiraGetIssueCmd.Flags().StringSliceVar(&jiraGetIssueFields, "fields", []string{}, "Comma-separated list of fields to return")
	jiraGetIssueCmd.Flags().StringSliceVar(&jiraGetIssueExpand, "expand", []string{}, "Comma-separated list of parameters to expand")
	jiraGetIssueCmd.Flags().StringSliceVar(&jiraGetIssueProperties, "properties", []string{}, "Comma-separated list of properties to return")
	jiraGetIssueCmd.Flags().BoolVar(&jiraGetIssueFieldsByKeys, "fields-by-keys", false, "Return fields by keys instead of IDs")
	jiraGetIssueCmd.Flags().BoolVar(&jiraGetIssueUpdateHistory, "update-history", false, "Include update history")
	jiraGetIssueCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for search-jql
	jiraSearchJQLCmd.Flags().StringSliceVar(&jiraSearchFields, "fields", []string{}, "Comma-separated list of fields to return")
	jiraSearchJQLCmd.Flags().IntVar(&jiraSearchMaxResults, "max-results", 50, "Maximum number of results to return (max 100)")
	jiraSearchJQLCmd.Flags().IntVar(&jiraSearchStartAt, "start-at", 0, "Starting index for pagination")
	jiraSearchJQLCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for create-issue
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateProject, "project", "", "Project key (required)")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateType, "type", "", "Issue type (required, e.g., Task, Bug, Story)")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateSummary, "summary", "", "Issue summary (required)")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateDescription, "description", "", "Issue description")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateAssignee, "assignee", "", "Assignee account ID")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateParent, "parent", "", "Parent issue key (for creating subtasks)")
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateFields, "fields", "", "Additional fields as JSON object")
	jiraCreateIssueCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	jiraCreateIssueCmd.MarkFlagRequired("project")
	jiraCreateIssueCmd.MarkFlagRequired("type")
	jiraCreateIssueCmd.MarkFlagRequired("summary")

	// Flags for add-comment
	jiraAddCommentCmd.Flags().StringVar(&jiraCommentVisibilityType, "visibility-type", "", "Restrict visibility (group or role)")
	jiraAddCommentCmd.Flags().StringVar(&jiraCommentVisibilityValue, "visibility-value", "", "Group or role name for visibility restriction")
	jiraAddCommentCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for edit-issue
	jiraEditIssueCmd.Flags().StringVar(&jiraEditSummary, "summary", "", "New summary")
	jiraEditIssueCmd.Flags().StringVar(&jiraEditDescription, "description", "", "New description")
	jiraEditIssueCmd.Flags().StringVar(&jiraEditAssignee, "assignee", "", "Assignee account ID")
	jiraEditIssueCmd.Flags().StringVar(&jiraEditFields, "fields", "", "Additional fields as JSON object")
	jiraEditIssueCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-transitions
	jiraGetTransitionsCmd.Flags().StringVar(&jiraGetTransitionsExpand, "expand", "", "Expand details for transitions")
	jiraGetTransitionsCmd.Flags().StringVar(&jiraGetTransitionsTransitionID, "transition-id", "", "Get specific transition by ID")
	jiraGetTransitionsCmd.Flags().BoolVar(&jiraGetTransitionsIncludeUnavailable, "include-unavailable", false, "Include unavailable transitions")
	jiraGetTransitionsCmd.Flags().BoolVar(&jiraGetTransitionsSkipRemoteOnly, "skip-remote-only", false, "Skip remote only condition")
	jiraGetTransitionsCmd.Flags().BoolVar(&jiraGetTransitionsSortByOpsBarAndStatus, "sort-by-ops-bar", false, "Sort by ops bar and status")
	jiraGetTransitionsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for transition-issue
	jiraTransitionIssueCmd.Flags().StringVar(&jiraTransitionFields, "fields", "", "Fields to set during transition as JSON object")
	jiraTransitionIssueCmd.Flags().StringVar(&jiraTransitionUpdate, "update", "", "Update operations as JSON object")
	jiraTransitionIssueCmd.Flags().StringVar(&jiraTransitionHistoryMetadata, "history-metadata", "", "History metadata as JSON object")
	jiraTransitionIssueCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for lookup-account-id
	jiraLookupAccountIDCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
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
		Fields:        jiraGetIssueFields,
		Expand:        jiraGetIssueExpand,
		Properties:    jiraGetIssueProperties,
		FieldsByKeys:  jiraGetIssueFieldsByKeys,
		UpdateHistory: jiraGetIssueUpdateHistory,
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

func runJiraSearchJQL(cmd *cobra.Command, args []string) error {
	jql := args[0]

	// Validate max-results
	if jiraSearchMaxResults > 100 {
		return fmt.Errorf("max-results cannot exceed 100")
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
	opts := &atlassian.SearchJQLOptions{
		Fields:     jiraSearchFields,
		MaxResults: jiraSearchMaxResults,
		StartAt:    jiraSearchStartAt,
	}

	// Search issues
	result, err := client.SearchJiraIssuesJQL(jql, opts)
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
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
		printSearchResults(result)
	}

	return nil
}

func printSearchResults(result map[string]interface{}) {
	issues, _ := result["issues"].([]interface{})
	isLast, _ := result["isLast"].(bool)
	nextPageToken, _ := result["nextPageToken"].(string)

	issueCount := len(issues)
	if issueCount == 0 {
		fmt.Println("No issues found.")
		return
	}

	fmt.Printf("Showing %d issue(s)\n\n", issueCount)

	for i, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			key, _ := issueMap["key"].(string)
			fields, _ := issueMap["fields"].(map[string]interface{})

			fmt.Printf("%d. %s", i+1, key)

			if fields != nil {
				if summary, ok := fields["summary"].(string); ok {
					fmt.Printf(": %s", summary)
				}
			}
			fmt.Println()

			if fields != nil {
				// Show type, status, assignee on same line
				parts := []string{}

				if issueType, ok := fields["issuetype"].(map[string]interface{}); ok {
					if name, ok := issueType["name"].(string); ok {
						parts = append(parts, fmt.Sprintf("Type: %s", name))
					}
				}

				if status, ok := fields["status"].(map[string]interface{}); ok {
					if name, ok := status["name"].(string); ok {
						parts = append(parts, fmt.Sprintf("Status: %s", name))
					}
				}

				if assignee, ok := fields["assignee"].(map[string]interface{}); ok {
					if displayName, ok := assignee["displayName"].(string); ok {
						parts = append(parts, fmt.Sprintf("Assignee: %s", displayName))
					}
				} else {
					parts = append(parts, "Assignee: Unassigned")
				}

				if len(parts) > 0 {
					fmt.Printf("   %s\n", strings.Join(parts, " | "))
				}
			}
			fmt.Println()
		}
	}

	// Show pagination info
	if !isLast && nextPageToken != "" {
		fmt.Printf("---\n")
		fmt.Printf("More issues available.\n")
		fmt.Printf("Note: Pagination with next page tokens is not yet implemented.\n")
		fmt.Printf("Use --max-results to increase the number of results (max 100).\n")
	}

	fmt.Printf("\nFor JSON output with all fields: atl jira search-jql \"<query>\" --json\n")
}

func runJiraCreateIssue(cmd *cobra.Command, args []string) error {
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

	// Parse additional fields if provided
	var additionalFields map[string]interface{}
	if jiraCreateFields != "" {
		if err := json.Unmarshal([]byte(jiraCreateFields), &additionalFields); err != nil {
			return fmt.Errorf("invalid --fields JSON: %w", err)
		}
	}

	// Create issue
	opts := &atlassian.CreateIssueOptions{
		ProjectKey:  jiraCreateProject,
		IssueType:   jiraCreateType,
		Summary:     jiraCreateSummary,
		Description: jiraCreateDescription,
		AssigneeID:  jiraCreateAssignee,
		ParentKey:   jiraCreateParent,
		Fields:      additionalFields,
	}

	result, err := client.CreateJiraIssue(opts)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		key, _ := result["key"].(string)
		id, _ := result["id"].(string)

		// Construct web URL
		webURL := fmt.Sprintf("%s/browse/%s", account.Site, key)
		if !strings.HasPrefix(account.Site, "http") {
			webURL = "https://" + webURL
		}

		fmt.Printf("✓ Created issue: %s\n", key)
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  Link: %s\n", webURL)
		fmt.Printf("\nView details: atl jira get-issue %s\n", key)
	}

	return nil
}

func runJiraAddComment(cmd *cobra.Command, args []string) error {
	issueKey := args[0]
	comment := args[1]

	// Validate visibility flags
	if (jiraCommentVisibilityType != "" && jiraCommentVisibilityValue == "") ||
		(jiraCommentVisibilityType == "" && jiraCommentVisibilityValue != "") {
		return fmt.Errorf("both --visibility-type and --visibility-value must be provided together")
	}

	if jiraCommentVisibilityType != "" && jiraCommentVisibilityType != "group" && jiraCommentVisibilityType != "role" {
		return fmt.Errorf("--visibility-type must be 'group' or 'role'")
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

	// Add comment
	opts := &atlassian.AddCommentOptions{
		Comment:         comment,
		VisibilityType:  jiraCommentVisibilityType,
		VisibilityValue: jiraCommentVisibilityValue,
	}

	result, err := client.AddCommentToIssue(issueKey, opts)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		id, _ := result["id"].(string)
		fmt.Printf("✓ Added comment to %s\n", issueKey)
		fmt.Printf("  Comment ID: %s\n", id)
	}

	return nil
}

func runJiraEditIssue(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	// Check if at least one field is provided
	if jiraEditSummary == "" && jiraEditDescription == "" && jiraEditAssignee == "" && jiraEditFields == "" {
		return fmt.Errorf("at least one field must be provided (--summary, --description, --assignee, or --fields)")
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

	// Build fields to update
	fields := make(map[string]interface{})

	// Parse additional fields first
	if jiraEditFields != "" {
		if err := json.Unmarshal([]byte(jiraEditFields), &fields); err != nil {
			return fmt.Errorf("invalid --fields JSON: %w", err)
		}
	}

	// Specific flags override --fields
	if jiraEditSummary != "" {
		fields["summary"] = jiraEditSummary
	}

	if jiraEditDescription != "" {
		// Convert description to ADF format
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": jiraEditDescription,
						},
					},
				},
			},
		}
	}

	if jiraEditAssignee != "" {
		fields["assignee"] = map[string]interface{}{
			"id": jiraEditAssignee,
		}
	}

	// Edit issue
	err = client.EditJiraIssue(issueKey, fields)
	if err != nil {
		return fmt.Errorf("failed to edit issue: %w", err)
	}

	if outputJSON {
		// JSON output - API returns 204 No Content
		response := map[string]interface{}{
			"status":  204,
			"success": true,
			"message": "Issue updated successfully",
		}
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		fmt.Printf("✓ Updated issue %s\n", issueKey)
		if jiraEditSummary != "" {
			fmt.Printf("  Summary: %s\n", jiraEditSummary)
		}
		if jiraEditDescription != "" {
			fmt.Printf("  Description: updated\n")
		}
		if jiraEditAssignee != "" {
			fmt.Printf("  Assignee: %s\n", jiraEditAssignee)
		}
		if jiraEditFields != "" {
			fmt.Printf("  Additional fields: updated\n")
		}
	}

	return nil
}

func runJiraGetTransitions(cmd *cobra.Command, args []string) error {
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

	// Build options
	opts := &atlassian.GetTransitionsOptions{
		Expand:                        jiraGetTransitionsExpand,
		TransitionID:                  jiraGetTransitionsTransitionID,
		IncludeUnavailableTransitions: jiraGetTransitionsIncludeUnavailable,
		SkipRemoteOnlyCondition:       jiraGetTransitionsSkipRemoteOnly,
		SortByOpsBarAndStatus:         jiraGetTransitionsSortByOpsBarAndStatus,
	}

	// Get transitions
	result, err := client.GetIssueTransitions(issueKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get transitions: %w", err)
	}

	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		transitions, _ := result["transitions"].([]interface{})

		if len(transitions) == 0 {
			fmt.Printf("No transitions available for %s\n", issueKey)
			return nil
		}

		fmt.Printf("Available transitions for %s:\n\n", issueKey)

		for _, t := range transitions {
			if trans, ok := t.(map[string]interface{}); ok {
				id, _ := trans["id"].(string)
				name, _ := trans["name"].(string)

				// Get destination status
				to, _ := trans["to"].(map[string]interface{})
				toName, _ := to["name"].(string)

				fmt.Printf("  ID: %-4s  → %s", id, name)
				if toName != "" {
					fmt.Printf(" (to: %s)", toName)
				}
				fmt.Println()
			}
		}

		fmt.Printf("\nTo transition: atl jira transition-issue %s <transition-id>\n", issueKey)
	}

	return nil
}

func runJiraTransitionIssue(cmd *cobra.Command, args []string) error {
	issueKey := args[0]
	transitionID := args[1]

	// Parse JSON parameters if provided
	var fields, update, historyMetadata map[string]interface{}

	if jiraTransitionFields != "" {
		if err := json.Unmarshal([]byte(jiraTransitionFields), &fields); err != nil {
			return fmt.Errorf("invalid --fields JSON: %w", err)
		}
	}
	if jiraTransitionUpdate != "" {
		if err := json.Unmarshal([]byte(jiraTransitionUpdate), &update); err != nil {
			return fmt.Errorf("invalid --update JSON: %w", err)
		}
	}
	if jiraTransitionHistoryMetadata != "" {
		if err := json.Unmarshal([]byte(jiraTransitionHistoryMetadata), &historyMetadata); err != nil {
			return fmt.Errorf("invalid --history-metadata JSON: %w", err)
		}
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

	// Build transition options
	opts := &atlassian.TransitionIssueOptions{
		TransitionID:    transitionID,
		Fields:          fields,
		Update:          update,
		HistoryMetadata: historyMetadata,
	}

	// Transition issue
	err = client.TransitionIssue(issueKey, opts)
	if err != nil {
		return fmt.Errorf("failed to transition issue: %w", err)
	}

	if outputJSON {
		// JSON output - API returns 204 No Content
		response := map[string]interface{}{
			"status":  204,
			"success": true,
			"message": "Issue transitioned successfully",
		}
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		fmt.Printf("✓ Transitioned issue %s\n", issueKey)
		fmt.Printf("\nView updated issue: atl jira get-issue %s\n", issueKey)
	}

	return nil
}

func runJiraLookupAccountID(cmd *cobra.Command, args []string) error {
	searchString := args[0]

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

	// Lookup users
	users, err := client.LookupAccountID(searchString)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %w", err)
	}

	if outputJSON {
		// JSON output
		output, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output (default)
		if len(users) == 0 {
			fmt.Printf("No users found for '%s'\n", searchString)
			return nil
		}

		fmt.Printf("Found %d user(s):\n\n", len(users))

		for i, user := range users {
			accountID, _ := user["accountId"].(string)
			displayName, _ := user["displayName"].(string)
			emailAddress, _ := user["emailAddress"].(string)
			active, _ := user["active"].(bool)

			statusStr := "active"
			if !active {
				statusStr = "inactive"
			}

			fmt.Printf("%d. %s (%s)\n", i+1, displayName, statusStr)
			fmt.Printf("   Email: %s\n", emailAddress)
			fmt.Printf("   Account ID: %s\n", accountID)
			fmt.Println()
		}

		fmt.Printf("To assign an issue: atl jira edit-issue <key> --assignee <account-id>\n")
	}

	return nil
}
