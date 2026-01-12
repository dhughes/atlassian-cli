package atlassian

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents an Atlassian API client
type Client struct {
	Email   string
	Token   string
	BaseURL string
	client  *http.Client
}

// NewClient creates a new Atlassian API client
func NewClient(email, token, site string) *Client {
	baseURL := site
	if !strings.HasPrefix(site, "http") {
		baseURL = "https://" + site
	}

	return &Client{
		Email:   email,
		Token:   token,
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// basicAuth returns the Basic auth header value
func (c *Client) basicAuth() string {
	auth := c.Email + ":" + c.Token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// AccessibleResource represents an Atlassian cloud resource
type AccessibleResource struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Scopes     []string `json:"scopes"`
	AvatarURL  string   `json:"avatarUrl"`
}

// GetAccessibleResources fetches the list of accessible Atlassian cloud resources
// Note: This endpoint requires OAuth and will fail with Basic Auth (API tokens)
func (c *Client) GetAccessibleResources() ([]AccessibleResource, error) {
	url := "https://api.atlassian.com/oauth/token/accessible-resources"

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get resources (status %d): %s", resp.StatusCode, string(body))
	}

	var resources []AccessibleResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resources, nil
}

// UserInfo represents the current user's information
type UserInfo struct {
	AccountID   string `json:"accountId"`
	AccountType string `json:"accountType"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
	Locale      string `json:"locale"`
}

// GetCurrentUser fetches information about the authenticated user
func (c *Client) GetCurrentUser() (*UserInfo, error) {
	url := fmt.Sprintf("%s/rest/api/3/myself", c.BaseURL)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info (status %d): %s", resp.StatusCode, string(body))
	}

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// TestAuthentication verifies that the credentials are valid by calling the Jira API
func (c *Client) TestAuthentication() error {
	_, err := c.GetCurrentUser()
	return err
}

// GetIssueOptions contains optional parameters for getting an issue
type GetIssueOptions struct {
	Fields        []string // List of fields to return
	Expand        []string // List of parameters to expand
	Properties    []string // List of properties to return
	FieldsByKeys  bool     // Return fields by keys instead of IDs
	UpdateHistory bool     // Include update history
}

// GetJiraIssue retrieves a Jira issue by its key or ID
func (c *Client) GetJiraIssue(issueKey string, opts *GetIssueOptions) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s", c.BaseURL, issueKey)

	// Add query parameters
	if opts != nil {
		params := url.Values{}
		if len(opts.Fields) > 0 {
			params.Add("fields", strings.Join(opts.Fields, ","))
		}
		if len(opts.Expand) > 0 {
			params.Add("expand", strings.Join(opts.Expand, ","))
		}
		if len(opts.Properties) > 0 {
			params.Add("properties", strings.Join(opts.Properties, ","))
		}
		if opts.FieldsByKeys {
			params.Add("fieldsByKeys", "true")
		}
		if opts.UpdateHistory {
			params.Add("updateHistory", "true")
		}
		if len(params) > 0 {
			apiURL += "?" + params.Encode()
		}
	}

	resp, err := c.doRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get issue (status %d): %s", resp.StatusCode, string(body))
	}

	var issue map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return issue, nil
}

// SearchJQLOptions contains optional parameters for JQL search
type SearchJQLOptions struct {
	Fields      []string // List of fields to return
	MaxResults  int      // Maximum number of results (default 50, max 100)
	StartAt     int      // Starting index for pagination
}

// SearchJiraIssuesJQL searches for Jira issues using JQL (Jira Query Language)
func (c *Client) SearchJiraIssuesJQL(jql string, opts *SearchJQLOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.BaseURL)

	// Build query parameters using url.Values for proper encoding
	params := url.Values{}
	params.Add("jql", jql)

	// Default fields to request if none specified
	defaultFields := "summary,status,issuetype,assignee,priority,reporter,created,updated"

	if opts != nil {
		if len(opts.Fields) > 0 {
			params.Add("fields", strings.Join(opts.Fields, ","))
		} else {
			params.Add("fields", defaultFields)
		}
		if opts.MaxResults > 0 {
			params.Add("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		} else {
			params.Add("maxResults", "50") // Default
		}
		if opts.StartAt > 0 {
			params.Add("startAt", fmt.Sprintf("%d", opts.StartAt))
		}
	} else {
		params.Add("fields", defaultFields)
		params.Add("maxResults", "50")
	}

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search issues (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CreateIssueOptions contains parameters for creating an issue
type CreateIssueOptions struct {
	ProjectKey  string
	IssueType   string
	Summary     string
	Description string
	AssigneeID  string
	ParentKey   string
	PriorityID  string
	Fields      map[string]any // Additional custom fields
}

// CreateJiraIssue creates a new Jira issue
func (c *Client) CreateJiraIssue(opts *CreateIssueOptions) (map[string]any, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue", c.BaseURL)

	// Build request body
	fields := map[string]any{
		"project": map[string]any{
			"key": opts.ProjectKey,
		},
		"issuetype": map[string]any{
			"name": opts.IssueType,
		},
		"summary": opts.Summary,
	}

	// Add optional fields
	if opts.Description != "" {
		// Convert markdown description to ADF format
		adf, err := MarkdownToADF(opts.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to convert description to ADF: %w", err)
		}
		fields["description"] = adf
	}

	if opts.AssigneeID != "" {
		fields["assignee"] = map[string]any{
			"id": opts.AssigneeID,
		}
	}

	if opts.ParentKey != "" {
		fields["parent"] = map[string]any{
			"key": opts.ParentKey,
		}
	}

	if opts.PriorityID != "" {
		fields["priority"] = map[string]any{
			"id": opts.PriorityID,
		}
	}

	// Add any additional custom fields
	if opts.Fields != nil {
		for k, v := range opts.Fields {
			fields[k] = v
		}
	}

	body := map[string]any{
		"fields": fields,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create issue (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// AddCommentOptions contains parameters for adding a comment
type AddCommentOptions struct {
	Comment        string
	VisibilityType string // "group" or "role"
	VisibilityValue string // Group or role name
}

// AddCommentToIssue adds a comment to a Jira issue
func (c *Client) AddCommentToIssue(issueKey string, opts *AddCommentOptions) (map[string]any, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", c.BaseURL, issueKey)

	// Build comment body in ADF format
	body := map[string]any{
		"body": map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []any{
				map[string]any{
					"type": "paragraph",
					"content": []any{
						map[string]any{
							"type": "text",
							"text": opts.Comment,
						},
					},
				},
			},
		},
	}

	// Add visibility if specified
	if opts.VisibilityType != "" && opts.VisibilityValue != "" {
		body["visibility"] = map[string]any{
			"type":  opts.VisibilityType,
			"value": opts.VisibilityValue,
		}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to add comment (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// EditJiraIssue updates fields on a Jira issue
func (c *Client) EditJiraIssue(issueKey string, fields map[string]any) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.BaseURL, issueKey)

	body := map[string]any{
		"fields": fields,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("PUT", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to edit issue (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetTransitionsOptions contains optional parameters for getting transitions
type GetTransitionsOptions struct {
	Expand                      string
	TransitionID                string
	IncludeUnavailableTransitions bool
	SkipRemoteOnlyCondition     bool
	SortByOpsBarAndStatus       bool
}

// GetIssueTransitions gets available transitions for an issue
func (c *Client) GetIssueTransitions(issueKey string, opts *GetTransitionsOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.BaseURL, issueKey)

	params := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			params.Add("expand", opts.Expand)
		}
		if opts.TransitionID != "" {
			params.Add("transitionId", opts.TransitionID)
		}
		if opts.IncludeUnavailableTransitions {
			params.Add("includeUnavailableTransitions", "true")
		}
		if opts.SkipRemoteOnlyCondition {
			params.Add("skipRemoteOnlyCondition", "true")
		}
		if opts.SortByOpsBarAndStatus {
			params.Add("sortByOpsBarAndStatus", "true")
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get transitions (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// TransitionIssueOptions contains parameters for transitioning an issue
type TransitionIssueOptions struct {
	TransitionID    string
	Fields          map[string]any
	Update          map[string]any
	HistoryMetadata map[string]any
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(issueKey string, opts *TransitionIssueOptions) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.BaseURL, issueKey)

	body := map[string]any{
		"transition": map[string]any{
			"id": opts.TransitionID,
		},
	}

	// Add optional parameters
	if opts.Fields != nil && len(opts.Fields) > 0 {
		body["fields"] = opts.Fields
	}
	if opts.Update != nil && len(opts.Update) > 0 {
		body["update"] = opts.Update
	}
	if opts.HistoryMetadata != nil && len(opts.HistoryMetadata) > 0 {
		body["historyMetadata"] = opts.HistoryMetadata
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to transition issue (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// LookupAccountID searches for Jira users by display name or email
func (c *Client) LookupAccountID(searchString string) ([]map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/user/search", c.BaseURL)

	params := url.Values{}
	params.Add("query", searchString)

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to lookup account (status %d): %s", resp.StatusCode, string(body))
	}

	var users []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// GetVisibleProjectsOptions contains parameters for getting visible projects
type GetVisibleProjectsOptions struct {
	Action         string // view, browse, edit, create
	SearchString   string
	ExpandIssueTypes bool
	MaxResults     int
	StartAt        int
}

// GetVisibleProjects lists projects the user has access to
func (c *Client) GetVisibleProjects(opts *GetVisibleProjectsOptions) ([]map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/project/search", c.BaseURL)

	params := url.Values{}

	if opts != nil {
		if opts.Action != "" {
			params.Add("action", opts.Action)
		}
		if opts.SearchString != "" {
			params.Add("query", opts.SearchString)
		}
		if opts.ExpandIssueTypes {
			params.Add("expand", "issueTypes")
		}
		if opts.MaxResults > 0 {
			params.Add("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}
		if opts.StartAt > 0 {
			params.Add("startAt", fmt.Sprintf("%d", opts.StartAt))
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get projects (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract values array
	values, _ := result["values"].([]any)
	projects := make([]map[string]any, 0, len(values))
	for _, v := range values {
		if proj, ok := v.(map[string]any); ok {
			projects = append(projects, proj)
		}
	}

	return projects, nil
}

// GetProjectIssueTypesOptions contains parameters for getting project issue types
type GetProjectIssueTypesOptions struct {
	MaxResults int
	StartAt    int
}

// GetProjectIssueTypes gets issue type metadata for a project
func (c *Client) GetProjectIssueTypes(projectKey string, opts *GetProjectIssueTypesOptions) ([]map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/issue/createmeta/%s/issuetypes", c.BaseURL, projectKey)

	params := url.Values{}
	if opts != nil {
		if opts.MaxResults > 0 {
			params.Add("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}
		if opts.StartAt > 0 {
			params.Add("startAt", fmt.Sprintf("%d", opts.StartAt))
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get issue types (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract values array
	values, _ := result["values"].([]any)
	issueTypes := make([]map[string]any, 0, len(values))
	for _, v := range values {
		if it, ok := v.(map[string]any); ok {
			issueTypes = append(issueTypes, it)
		}
	}

	return issueTypes, nil
}

// GetRemoteLinksOptions contains parameters for getting remote issue links
type GetRemoteLinksOptions struct {
	GlobalID string
}

// GetIssueRemoteLinks gets remote links for a Jira issue
func (c *Client) GetIssueRemoteLinks(issueKey string, opts *GetRemoteLinksOptions) ([]map[string]any, error) {
	baseURL := fmt.Sprintf("%s/rest/api/3/issue/%s/remotelink", c.BaseURL, issueKey)

	params := url.Values{}
	if opts != nil && opts.GlobalID != "" {
		params.Add("globalId", opts.GlobalID)
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get remote links (status %d): %s", resp.StatusCode, string(body))
	}

	var links []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return links, nil
}

// SearchCQLOptions contains optional parameters for CQL search
type SearchCQLOptions struct {
	Limit      int
	Cursor     string
	CqlContext string
	Expand     string
	Next       bool
	Prev       bool
}

// SearchConfluenceCQL searches Confluence content using CQL (Confluence Query Language)
func (c *Client) SearchConfluenceCQL(cql string, opts *SearchCQLOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/search", c.BaseURL)

	// Build query parameters using url.Values for proper encoding
	params := url.Values{}
	params.Add("cql", cql)

	if opts != nil {
		if opts.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Add("limit", "25")
		}
		if opts.Cursor != "" {
			params.Add("cursor", opts.Cursor)
		}
		if opts.CqlContext != "" {
			params.Add("cqlcontext", opts.CqlContext)
		}
		if opts.Expand != "" {
			params.Add("expand", opts.Expand)
		}
		if opts.Next {
			params.Add("next", "true")
		}
		if opts.Prev {
			params.Add("prev", "true")
		}
	} else {
		params.Add("limit", "25")
	}

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search content (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPageOptions contains parameters for getting a page
type GetPageOptions struct {
	Status string // Page status: current, draft, archived, trashed
}

// GetConfluencePage retrieves a Confluence page by ID
func (c *Client) GetConfluencePage(pageID string, opts *GetPageOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/%s", c.BaseURL, pageID)

	// Request body content expanded
	params := url.Values{}
	params.Add("expand", "body.storage,version,space,history")

	// Add status if specified (defaults to current if not specified)
	if opts != nil && opts.Status != "" {
		params.Add("status", opts.Status)
	}

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get page (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetSpacesOptions contains parameters for getting spaces
type GetSpacesOptions struct {
	Keys              []string
	IDs               []string
	Type              string
	Status            string
	Labels            []string
	FavoritedBy       string
	NotFavoritedBy    string
	Sort              string
	DescriptionFormat string
	IncludeIcon       bool
	Limit             int
	Cursor            string
}

// GetConfluenceSpaces retrieves Confluence spaces
func (c *Client) GetConfluenceSpaces(opts *GetSpacesOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/space", c.BaseURL)

	params := url.Values{}

	if opts != nil {
		if len(opts.Keys) > 0 {
			for _, key := range opts.Keys {
				params.Add("spaceKey", key)
			}
		}
		if len(opts.IDs) > 0 {
			params.Add("spaceId", strings.Join(opts.IDs, ","))
		}
		if opts.Type != "" {
			params.Add("type", opts.Type)
		}
		if opts.Status != "" {
			params.Add("status", opts.Status)
		}
		if len(opts.Labels) > 0 {
			for _, label := range opts.Labels {
				params.Add("label", label)
			}
		}
		if opts.FavoritedBy != "" {
			params.Add("favourite", opts.FavoritedBy)
		}
		if opts.NotFavoritedBy != "" {
			params.Add("favouriteUserKey", opts.NotFavoritedBy)
		}
		if opts.Sort != "" {
			params.Add("sort", opts.Sort)
		}
		if opts.DescriptionFormat != "" {
			params.Add("expand", "description."+opts.DescriptionFormat)
		}
		if opts.IncludeIcon {
			params.Add("expand", "icon")
		}
		if opts.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Add("limit", "25")
		}
		if opts.Cursor != "" {
			params.Add("cursor", opts.Cursor)
		}
	} else {
		params.Add("limit", "25")
	}

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get spaces (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPagesInSpaceOptions contains parameters for getting pages in a space
type GetPagesInSpaceOptions struct {
	SpaceKey string
	Title    string
	Status   string
	Limit    int
	Cursor   string
	Depth    string
	Sort     string
	Subtype  string
}

// GetPagesInSpace retrieves pages within a Confluence space
func (c *Client) GetPagesInSpace(opts *GetPagesInSpaceOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	params := url.Values{}
	params.Add("type", "page")
	params.Add("spaceKey", opts.SpaceKey)

	if opts.Title != "" {
		params.Add("title", opts.Title)
	}
	if opts.Status != "" {
		params.Add("status", opts.Status)
	}
	if opts.Cursor != "" {
		params.Add("cursor", opts.Cursor)
	}
	if opts.Depth != "" {
		params.Add("depth", opts.Depth)
	}
	if opts.Sort != "" {
		params.Add("orderby", opts.Sort)
	}
	if opts.Subtype != "" {
		params.Add("expand", "metadata.properties.subtype")
		// Note: Actual subtype filtering may require different approach
	}
	if opts.Limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Add("limit", "25")
	}

	fullURL := baseURL + "?" + params.Encode()

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get pages (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CreatePageOptions contains parameters for creating a page
type CreatePageOptions struct {
	SpaceKey  string
	Title     string
	Body      string
	ParentID  string
	IsPrivate bool
}

// CreateConfluencePage creates a new Confluence page
func (c *Client) CreateConfluencePage(opts *CreatePageOptions) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	body := map[string]any{
		"type":  "page",
		"title": opts.Title,
		"space": map[string]any{
			"key": opts.SpaceKey,
		},
		"body": map[string]any{
			"storage": map[string]any{
				"value":          opts.Body,
				"representation": "storage",
			},
		},
	}

	if opts.ParentID != "" {
		body["ancestors"] = []any{
			map[string]any{
				"id": opts.ParentID,
			},
		}
	}

	if opts.IsPrivate {
		body["metadata"] = map[string]any{
			"properties": map[string]any{
				"editor": map[string]any{
					"key":   "editor",
					"value": "v2",
				},
			},
		}
		// Note: Private pages may require additional permissions setup
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create page (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// UpdatePageOptions contains parameters for updating a page
type UpdatePageOptions struct {
	PageID         string
	Title          string
	Body           string
	Version        int
	ParentID       string
	SpaceKey       string
	Status         string
	VersionMessage string
}

// UpdateConfluencePage updates an existing Confluence page
func (c *Client) UpdateConfluencePage(opts *UpdatePageOptions) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content/%s", c.BaseURL, opts.PageID)

	body := map[string]any{
		"type":  "page",
		"title": opts.Title,
		"version": map[string]any{
			"number": opts.Version,
		},
		"body": map[string]any{
			"storage": map[string]any{
				"value":          opts.Body,
				"representation": "storage",
			},
		},
	}

	if opts.VersionMessage != "" {
		body["version"].(map[string]any)["message"] = opts.VersionMessage
	}

	if opts.ParentID != "" {
		body["ancestors"] = []any{
			map[string]any{
				"id": opts.ParentID,
			},
		}
	}

	if opts.SpaceKey != "" {
		body["space"] = map[string]any{
			"key": opts.SpaceKey,
		}
	}

	if opts.Status != "" {
		body["status"] = opts.Status
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("PUT", apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update page (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// AddPageCommentOptions contains parameters for adding a comment to a page
type AddPageCommentOptions struct {
	PageID           string
	Comment          string
	ParentCommentID  string
	AttachmentID     string
	CustomContentID  string
}

// AddConfluencePageComment adds a comment to a Confluence page
func (c *Client) AddConfluencePageComment(opts *AddPageCommentOptions) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	body := map[string]any{
		"type": "comment",
		"container": map[string]any{
			"id":   opts.PageID,
			"type": "page",
		},
		"body": map[string]any{
			"storage": map[string]any{
				"value":          opts.Comment,
				"representation": "storage",
			},
		},
	}

	// Add optional parameters
	if opts.ParentCommentID != "" {
		body["container"] = map[string]any{
			"id":   opts.ParentCommentID,
			"type": "comment",
		}
	}
	// Note: attachmentId and customContentId may require different API structure

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to add comment (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPageAncestors gets the parent pages of a Confluence page
func (c *Client) GetPageAncestors(pageID string) ([]map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/%s/ancestor", c.BaseURL, pageID)

	resp, err := c.doRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get ancestors (status %d): %s", resp.StatusCode, string(body))
	}

	var ancestors []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&ancestors); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return ancestors, nil
}

// GetPageDescendantsOptions contains parameters for getting page descendants
type GetPageDescendantsOptions struct {
	Depth int
	Limit int
}

// GetPageDescendants gets child pages of a Confluence page
func (c *Client) GetPageDescendants(pageID string, opts *GetPageDescendantsOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/%s/descendant/page", c.BaseURL, pageID)

	params := url.Values{}
	if opts != nil {
		if opts.Depth > 0 {
			params.Add("depth", fmt.Sprintf("%d", opts.Depth))
		}
		if opts.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get descendants (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPageCommentsOptions contains parameters for getting page comments
type GetPageCommentsOptions struct {
	Limit  int
	Start  int
	Status string
}

// GetPageComments gets comments for a Confluence page
func (c *Client) GetPageComments(pageID string, opts *GetPageCommentsOptions) (map[string]any, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/%s/child/comment", c.BaseURL, pageID)

	params := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Start > 0 {
			params.Add("start", fmt.Sprintf("%d", opts.Start))
		}
		if opts.Status != "" {
			params.Add("status", opts.Status)
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get comments (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CreateInlineCommentOptions contains parameters for creating an inline comment
type CreateInlineCommentOptions struct {
	PageID                   string
	Comment                  string
	TextSelection            string
	TextSelectionMatchIndex  int
	TextSelectionMatchCount  int
}

// CreateInlineComment creates an inline comment on a Confluence page
func (c *Client) CreateInlineComment(opts *CreateInlineCommentOptions) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	body := map[string]any{
		"type": "comment",
		"container": map[string]any{
			"id":   opts.PageID,
			"type": "page",
		},
		"body": map[string]any{
			"storage": map[string]any{
				"value":          opts.Comment,
				"representation": "storage",
			},
		},
	}

	// Add inline comment properties
	if opts.TextSelection != "" {
		body["metadata"] = map[string]any{
			"properties": map[string]any{
				"inline-comment-properties": map[string]any{
					"textSelection":           opts.TextSelection,
					"textSelectionMatchIndex": opts.TextSelectionMatchIndex,
					"textSelectionMatchCount": opts.TextSelectionMatchCount,
				},
			},
		}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create inline comment (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetCreateMeta gets field metadata for creating issues of a specific type
func (c *Client) GetCreateMeta(projectKey string, issueTypeID string) (map[string]any, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/createmeta/%s/issuetypes/%s", c.BaseURL, projectKey, issueTypeID)

	resp, err := c.doRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get create metadata (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetFieldOptions gets allowed values for a custom field
// Requires project key and issue type ID to retrieve field configuration
// Uses createmeta API to retrieve field configuration
func (c *Client) GetFieldOptions(fieldKey string, projectKey string, issueTypeID string) (map[string]any, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("project key is required to get field options")
	}
	if issueTypeID == "" {
		return nil, fmt.Errorf("issue type ID is required to get field options")
	}

	// Get createmeta for this issue type
	createMetaURL := fmt.Sprintf("%s/rest/api/3/issue/createmeta/%s/issuetypes/%s", c.BaseURL, projectKey, issueTypeID)

	resp2, err := c.doRequest("GET", createMetaURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return nil, fmt.Errorf("failed to get create metadata (status %d): %s", resp2.StatusCode, string(body))
	}

	var createMetaResult map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&createMetaResult); err != nil {
		return nil, fmt.Errorf("failed to decode create metadata response: %w", err)
	}

	// Search for the field in the fields array
	fieldsArray, _ := createMetaResult["fields"].([]any)
	for _, fieldVal := range fieldsArray {
		fieldMap, _ := fieldVal.(map[string]any)
		key, _ := fieldMap["key"].(string)
		if key == fieldKey {
			return fieldMap, nil
		}
	}

	return nil, fmt.Errorf("field %s not found in project %s", fieldKey, projectKey)
}

// IssueLinkType represents an issue link type
type IssueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
	Self    string `json:"self"`
}

// GetIssueLinkTypes retrieves all issue link types
func (c *Client) GetIssueLinkTypes() ([]IssueLinkType, error) {
	url := fmt.Sprintf("%s/rest/api/3/issueLinkType", c.BaseURL)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get issue link types (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		IssueLinkTypes []IssueLinkType `json:"issueLinkTypes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.IssueLinkTypes, nil
}

// LinkIssueOptions contains options for linking issues
type LinkIssueOptions struct {
	TypeName      string
	InwardIssue   string
	OutwardIssue  string
	CommentBody   string
}

// LinkIssues creates a link between two issues
func (c *Client) LinkIssues(opts *LinkIssueOptions) error {
	if opts.TypeName == "" {
		return fmt.Errorf("type name is required")
	}
	if opts.InwardIssue == "" {
		return fmt.Errorf("inward issue is required")
	}
	if opts.OutwardIssue == "" {
		return fmt.Errorf("outward issue is required")
	}

	url := fmt.Sprintf("%s/rest/api/3/issueLink", c.BaseURL)

	body := map[string]any{
		"type": map[string]any{
			"name": opts.TypeName,
		},
		"inwardIssue": map[string]any{
			"key": opts.InwardIssue,
		},
		"outwardIssue": map[string]any{
			"key": opts.OutwardIssue,
		},
	}

	if opts.CommentBody != "" {
		body["comment"] = map[string]any{
			"body": opts.CommentBody,
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to link issues (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
