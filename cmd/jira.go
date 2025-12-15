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

The --description flag supports MARKDOWN formatting including:
  - Headings (# H1, ## H2, etc.)
  - Bold (**text**), italic (*text*), code (` + "`code`" + `)
  - Lists (bullets and numbered)
  - Code blocks (` + "```language" + `)
  - Links, blockquotes, and more

Examples:
  atl jira create-issue --project PROJ --type Task --summary "Do something"
  atl jira create-issue --project PROJ --type Bug --summary "Fix bug" --description "**Important:** Bug details here"`,
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

The --description flag supports MARKDOWN formatting (headings, bold, lists, code blocks, etc).

Examples:
  atl jira edit-issue PROJ-123 --summary "New summary"
  atl jira edit-issue PROJ-123 --description "## Updated\n\n- Point 1\n- Point 2"
  atl jira edit-issue PROJ-123 --summary "Update" --description "Details with **bold**"`,
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

var jiraGetProjectsCmd = &cobra.Command{
	Use:   "get-projects",
	Short: "List visible Jira projects",
	Long: `List Jira projects the user has permission to access.

Examples:
  atl jira get-projects
  atl jira get-projects --action create
  atl jira get-projects --search "Product"`,
	RunE: runJiraGetProjects,
}

var jiraGetProjectIssueTypesCmd = &cobra.Command{
	Use:   "get-project-issue-types <projectKey>",
	Short: "Get issue types for a project",
	Long: `Get available issue type metadata for a Jira project.

Examples:
  atl jira get-project-issue-types ABC
  atl jira get-project-issue-types ABC --json`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraGetProjectIssueTypes,
}

var jiraGetRemoteLinksCmd = &cobra.Command{
	Use:   "get-remote-links <issueKey>",
	Short: "Get remote links for an issue",
	Long: `Get remote issue links (e.g., Confluence pages) for a Jira issue.

Examples:
  atl jira get-remote-links ABC-123
  atl jira get-remote-links ABC-123 --global-id "appId=456&pageId=123"`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraGetRemoteLinks,
}

var jiraGetCreateMetaCmd = &cobra.Command{
	Use:   "get-create-meta <projectKey> <issueTypeId>",
	Short: "Get create metadata for an issue type",
	Long: `Get field metadata including allowed values for creating issues.

This shows all fields (required and optional) and their constraints,
including allowed values for custom fields like select lists.

Examples:
  atl jira get-create-meta ABC 10001
  atl jira get-create-meta ABC 10001 --json`,
	Args: cobra.ExactArgs(2),
	RunE: runJiraGetCreateMeta,
}

var jiraGetFieldOptionsCmd = &cobra.Command{
	Use:   "get-field-options <fieldKey>",
	Short: "Get allowed values for a custom field",
	Long: `Get allowed values for a custom select field.

Field keys are typically in the format "customfield_XXXXX".
Requires --project and --issue-type-id flags.

Examples:
  atl jira get-field-options customfield_10369 --project ABC --issue-type-id 10002
  atl jira get-field-options customfield_10369 --project ABC --issue-type-id 10002 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runJiraGetFieldOptions,
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

	// Flags for get-projects
	jiraProjectsAction         string
	jiraProjectsSearch         string
	jiraProjectsExpandIssueTypes bool
	jiraProjectsMaxResults     int
	jiraProjectsStartAt        int

	// Flags for get-project-issue-types
	jiraIssueTypesMaxResults int
	jiraIssueTypesStartAt    int

	// Flags for get-remote-links
	jiraRemoteLinksGlobalID string

	// Flags for get-field-options
	jiraFieldOptionsProject     string
	jiraFieldOptionsIssueTypeID string
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
	jiraCmd.AddCommand(jiraGetProjectsCmd)
	jiraCmd.AddCommand(jiraGetProjectIssueTypesCmd)
	jiraCmd.AddCommand(jiraGetRemoteLinksCmd)
	jiraCmd.AddCommand(jiraGetCreateMetaCmd)
	jiraCmd.AddCommand(jiraGetFieldOptionsCmd)

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
	jiraCreateIssueCmd.Flags().StringVar(&jiraCreateDescription, "description", "", "Issue description (supports markdown formatting)")
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
	jiraEditIssueCmd.Flags().StringVar(&jiraEditDescription, "description", "", "New description (supports markdown formatting)")
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

	// Flags for get-projects
	jiraGetProjectsCmd.Flags().StringVar(&jiraProjectsAction, "action", "view", "Filter by permission (view, browse, edit, create)")
	jiraGetProjectsCmd.Flags().StringVar(&jiraProjectsSearch, "search", "", "Search projects by name or key")
	jiraGetProjectsCmd.Flags().BoolVar(&jiraProjectsExpandIssueTypes, "expand-issue-types", false, "Include issue types in response")
	jiraGetProjectsCmd.Flags().IntVar(&jiraProjectsMaxResults, "max-results", 50, "Maximum results to return")
	jiraGetProjectsCmd.Flags().IntVar(&jiraProjectsStartAt, "start-at", 0, "Starting index for pagination")
	jiraGetProjectsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-project-issue-types
	jiraGetProjectIssueTypesCmd.Flags().IntVar(&jiraIssueTypesMaxResults, "max-results", 50, "Maximum results to return")
	jiraGetProjectIssueTypesCmd.Flags().IntVar(&jiraIssueTypesStartAt, "start-at", 0, "Starting index for pagination")
	jiraGetProjectIssueTypesCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-remote-links
	jiraGetRemoteLinksCmd.Flags().StringVar(&jiraRemoteLinksGlobalID, "global-id", "", "Filter by global ID")
	jiraGetRemoteLinksCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-create-meta
	jiraGetCreateMetaCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Flags for get-field-options
	jiraGetFieldOptionsCmd.Flags().StringVar(&jiraFieldOptionsProject, "project", "", "Project key for context (required)")
	jiraGetFieldOptionsCmd.Flags().StringVar(&jiraFieldOptionsIssueTypeID, "issue-type-id", "", "Issue type ID for context (required)")
	jiraGetFieldOptionsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
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

func printIssuePretty(issue map[string]any) {
	// Extract common fields
	key, _ := issue["key"].(string)
	fields, _ := issue["fields"].(map[string]any)

	fmt.Printf("Issue: %s\n", key)

	if fields != nil {
		if summary, ok := fields["summary"].(string); ok {
			fmt.Printf("Summary: %s\n", summary)
		}

		if issueType, ok := fields["issuetype"].(map[string]any); ok {
			if name, ok := issueType["name"].(string); ok {
				fmt.Printf("Type: %s\n", name)
			}
		}

		if status, ok := fields["status"].(map[string]any); ok {
			if name, ok := status["name"].(string); ok {
				fmt.Printf("Status: %s\n", name)
			}
		}

		if priority, ok := fields["priority"].(map[string]any); ok {
			if name, ok := priority["name"].(string); ok {
				fmt.Printf("Priority: %s\n", name)
			}
		}

		if assignee, ok := fields["assignee"].(map[string]any); ok {
			if displayName, ok := assignee["displayName"].(string); ok {
				fmt.Printf("Assignee: %s\n", displayName)
			}
		} else {
			fmt.Printf("Assignee: Unassigned\n")
		}

		if reporter, ok := fields["reporter"].(map[string]any); ok {
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

func printSearchResults(result map[string]any) {
	issues, _ := result["issues"].([]any)
	isLast, _ := result["isLast"].(bool)
	nextPageToken, _ := result["nextPageToken"].(string)

	issueCount := len(issues)
	if issueCount == 0 {
		fmt.Println("No issues found.")
		return
	}

	fmt.Printf("Showing %d issue(s)\n\n", issueCount)

	for i, issue := range issues {
		if issueMap, ok := issue.(map[string]any); ok {
			key, _ := issueMap["key"].(string)
			fields, _ := issueMap["fields"].(map[string]any)

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

				if issueType, ok := fields["issuetype"].(map[string]any); ok {
					if name, ok := issueType["name"].(string); ok {
						parts = append(parts, fmt.Sprintf("Type: %s", name))
					}
				}

				if status, ok := fields["status"].(map[string]any); ok {
					if name, ok := status["name"].(string); ok {
						parts = append(parts, fmt.Sprintf("Status: %s", name))
					}
				}

				if assignee, ok := fields["assignee"].(map[string]any); ok {
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
	var additionalFields map[string]any
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
	fields := make(map[string]any)

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
		// Convert markdown description to ADF format
		adf, err := atlassian.MarkdownToADF(jiraEditDescription)
		if err != nil {
			return fmt.Errorf("failed to convert description to ADF: %w", err)
		}
		fields["description"] = adf
	}

	if jiraEditAssignee != "" {
		fields["assignee"] = map[string]any{
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
		response := map[string]any{
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
		transitions, _ := result["transitions"].([]any)

		if len(transitions) == 0 {
			fmt.Printf("No transitions available for %s\n", issueKey)
			return nil
		}

		fmt.Printf("Available transitions for %s:\n\n", issueKey)

		for _, t := range transitions {
			if trans, ok := t.(map[string]any); ok {
				id, _ := trans["id"].(string)
				name, _ := trans["name"].(string)

				// Get destination status
				to, _ := trans["to"].(map[string]any)
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
	var fields, update, historyMetadata map[string]any

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
		response := map[string]any{
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

func runJiraGetProjects(cmd *cobra.Command, args []string) error {
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

	// Get projects
	opts := &atlassian.GetVisibleProjectsOptions{
		Action:           jiraProjectsAction,
		SearchString:     jiraProjectsSearch,
		ExpandIssueTypes: jiraProjectsExpandIssueTypes,
		MaxResults:       jiraProjectsMaxResults,
		StartAt:          jiraProjectsStartAt,
	}

	projects, err := client.GetVisibleProjects(opts)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		fmt.Printf("Found %d project(s):\n\n", len(projects))

		for i, proj := range projects {
			key, _ := proj["key"].(string)
			name, _ := proj["name"].(string)
			projectType, _ := proj["projectTypeKey"].(string)

			fmt.Printf("%d. %s (%s)\n", i+1, name, key)
			fmt.Printf("   Type: %s\n", projectType)

			// Show issue types if expanded
			if jiraProjectsExpandIssueTypes {
				if issueTypes, ok := proj["issueTypes"].([]any); ok && len(issueTypes) > 0 {
					fmt.Printf("   Issue types: ")
					types := []string{}
					for _, it := range issueTypes {
						if itMap, ok := it.(map[string]any); ok {
							if name, ok := itMap["name"].(string); ok {
								types = append(types, name)
							}
						}
					}
					fmt.Printf("%s\n", strings.Join(types, ", "))
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func runJiraGetProjectIssueTypes(cmd *cobra.Command, args []string) error {
	projectKey := args[0]

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

	// Get issue types
	opts := &atlassian.GetProjectIssueTypesOptions{
		MaxResults: jiraIssueTypesMaxResults,
		StartAt:    jiraIssueTypesStartAt,
	}

	issueTypes, err := client.GetProjectIssueTypes(projectKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get issue types: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(issueTypes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issueTypes) == 0 {
			fmt.Printf("No issue types found for project %s\n", projectKey)
			return nil
		}

		fmt.Printf("Issue types for project %s:\n\n", projectKey)

		for i, it := range issueTypes {
			name, _ := it["name"].(string)
			id, _ := it["id"].(string)
			description, _ := it["description"].(string)
			subtask, _ := it["subtask"].(bool)

			typeStr := "Issue"
			if subtask {
				typeStr = "Subtask"
			}

			fmt.Printf("%d. %s (ID: %s) [%s]\n", i+1, name, id, typeStr)
			if description != "" {
				fmt.Printf("   %s\n", description)
			}
			fmt.Println()
		}
	}

	return nil
}

func runJiraGetRemoteLinks(cmd *cobra.Command, args []string) error {
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

	// Get remote links
	opts := &atlassian.GetRemoteLinksOptions{
		GlobalID: jiraRemoteLinksGlobalID,
	}

	links, err := client.GetIssueRemoteLinks(issueKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get remote links: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(links, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(links) == 0 {
			fmt.Printf("No remote links found for %s\n", issueKey)
			return nil
		}

		fmt.Printf("Remote links for %s:\n\n", issueKey)

		for i, link := range links {
			id, _ := link["id"].(string)
			obj, _ := link["object"].(map[string]any)
			url, _ := obj["url"].(string)
			title, _ := obj["title"].(string)

			fmt.Printf("%d. %s\n", i+1, title)
			fmt.Printf("   URL: %s\n", url)
			fmt.Printf("   ID: %s\n", id)
			fmt.Println()
		}
	}

	return nil
}

func runJiraGetCreateMeta(cmd *cobra.Command, args []string) error {
	projectKey := args[0]
	issueTypeID := args[1]

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

	// Get create metadata
	metadata, err := client.GetCreateMeta(projectKey, issueTypeID)
	if err != nil {
		return fmt.Errorf("failed to get create metadata: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output - fields are returned as an array, not a map
		fieldsArray, _ := metadata["fields"].([]any)
		if len(fieldsArray) == 0 {
			fmt.Printf("No fields found for project %s issue type %s\n", projectKey, issueTypeID)
			return nil
		}

		fmt.Printf("Create metadata for project %s, issue type %s:\n\n", projectKey, issueTypeID)

		// Separate required and optional fields
		var requiredFields []any
		var optionalFields []any

		for _, fieldVal := range fieldsArray {
			fieldMap, _ := fieldVal.(map[string]any)
			required, _ := fieldMap["required"].(bool)
			if required {
				requiredFields = append(requiredFields, fieldVal)
			} else {
				optionalFields = append(optionalFields, fieldVal)
			}
		}

		// Display required fields first
		if len(requiredFields) > 0 {
			fmt.Println("Required Fields:")
			for _, fieldVal := range requiredFields {
				fieldMap, _ := fieldVal.(map[string]any)
				key, _ := fieldMap["key"].(string)
				printFieldInfo(fieldVal, key)
			}
		}

		// Display optional fields
		if len(optionalFields) > 0 {
			fmt.Println("\nOptional Fields:")
			for _, fieldVal := range optionalFields {
				fieldMap, _ := fieldVal.(map[string]any)
				key, _ := fieldMap["key"].(string)
				printFieldInfo(fieldVal, key)
			}
		}
	}

	return nil
}

func printFieldInfo(val any, key string) {
	fieldMap, _ := val.(map[string]any)
	name, _ := fieldMap["name"].(string)
	schema, _ := fieldMap["schema"].(map[string]any)
	fieldType, _ := schema["type"].(string)

	fmt.Printf("\n  %s (%s)\n", name, key)
	fmt.Printf("    Type: %s\n", fieldType)

	// Show allowed values if present
	allowedValues, _ := fieldMap["allowedValues"].([]any)
	if len(allowedValues) > 0 {
		fmt.Printf("    Allowed values:\n")
		for _, av := range allowedValues {
			avMap, _ := av.(map[string]any)
			value, _ := avMap["value"].(string)
			id, _ := avMap["id"].(string)
			if value != "" {
				fmt.Printf("      - %s (id: %s)\n", value, id)
			} else {
				// For some fields like issue types
				valueName, _ := avMap["name"].(string)
				if valueName != "" {
					fmt.Printf("      - %s (id: %s)\n", valueName, id)
				}
			}
		}
	}
}

func runJiraGetFieldOptions(cmd *cobra.Command, args []string) error {
	fieldKey := args[0]

	// Validate required flags
	if jiraFieldOptionsProject == "" {
		return fmt.Errorf("--project flag is required")
	}
	if jiraFieldOptionsIssueTypeID == "" {
		return fmt.Errorf("--issue-type-id flag is required")
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

	// Get field options
	options, err := client.GetFieldOptions(fieldKey, jiraFieldOptionsProject, jiraFieldOptionsIssueTypeID)
	if err != nil {
		return fmt.Errorf("failed to get field options: %w", err)
	}

	if outputJSON {
		output, err := json.MarshalIndent(options, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Pretty output
		fieldName, _ := options["name"].(string)
		allowedValues, _ := options["allowedValues"].([]any)

		if fieldName == "" {
			fieldName = fieldKey
		}

		fmt.Printf("Field: %s (%s)\n\n", fieldName, fieldKey)

		if len(allowedValues) == 0 {
			fmt.Println("No allowed values found for this field")
			return nil
		}

		fmt.Println("Allowed values:")
		for i, av := range allowedValues {
			avMap, _ := av.(map[string]any)
			value, _ := avMap["value"].(string)
			id, _ := avMap["id"].(string)

			fmt.Printf("%d. %s\n", i+1, value)
			fmt.Printf("   ID: %s\n", id)

			// Show how to use in --fields
			if i == 0 {
				fmt.Printf("   Usage: --fields '{\"customfield_XXXXX\": {\"id\": \"%s\"}}'\n", id)
			}
			fmt.Println()
		}
	}

	return nil
}
