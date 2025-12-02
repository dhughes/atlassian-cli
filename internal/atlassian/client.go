package atlassian

import (
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
func (c *Client) GetJiraIssue(issueKey string, opts *GetIssueOptions) (map[string]interface{}, error) {
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

	var issue map[string]interface{}
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
func (c *Client) SearchJiraIssuesJQL(jql string, opts *SearchJQLOptions) (map[string]interface{}, error) {
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

	var result map[string]interface{}
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
	Fields      map[string]interface{} // Additional custom fields
}

// CreateJiraIssue creates a new Jira issue
func (c *Client) CreateJiraIssue(opts *CreateIssueOptions) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue", c.BaseURL)

	// Build request body
	fields := map[string]interface{}{
		"project": map[string]interface{}{
			"key": opts.ProjectKey,
		},
		"issuetype": map[string]interface{}{
			"name": opts.IssueType,
		},
		"summary": opts.Summary,
	}

	// Add optional fields
	if opts.Description != "" {
		// Convert description to ADF format (simple paragraph)
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": opts.Description,
						},
					},
				},
			},
		}
	}

	if opts.AssigneeID != "" {
		fields["assignee"] = map[string]interface{}{
			"id": opts.AssigneeID,
		}
	}

	if opts.ParentKey != "" {
		fields["parent"] = map[string]interface{}{
			"key": opts.ParentKey,
		}
	}

	if opts.PriorityID != "" {
		fields["priority"] = map[string]interface{}{
			"id": opts.PriorityID,
		}
	}

	// Add any additional custom fields
	if opts.Fields != nil {
		for k, v := range opts.Fields {
			fields[k] = v
		}
	}

	body := map[string]interface{}{
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

	var result map[string]interface{}
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
func (c *Client) AddCommentToIssue(issueKey string, opts *AddCommentOptions) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", c.BaseURL, issueKey)

	// Build comment body in ADF format
	body := map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"content": []interface{}{
						map[string]interface{}{
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
		body["visibility"] = map[string]interface{}{
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// EditJiraIssue updates fields on a Jira issue
func (c *Client) EditJiraIssue(issueKey string, fields map[string]interface{}) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.BaseURL, issueKey)

	body := map[string]interface{}{
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
func (c *Client) GetIssueTransitions(issueKey string, opts *GetTransitionsOptions) (map[string]interface{}, error) {
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// TransitionIssueOptions contains parameters for transitioning an issue
type TransitionIssueOptions struct {
	TransitionID    string
	Fields          map[string]interface{}
	Update          map[string]interface{}
	HistoryMetadata map[string]interface{}
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(issueKey string, opts *TransitionIssueOptions) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.BaseURL, issueKey)

	body := map[string]interface{}{
		"transition": map[string]interface{}{
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
func (c *Client) LookupAccountID(searchString string) ([]map[string]interface{}, error) {
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

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
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
func (c *Client) SearchConfluenceCQL(cql string, opts *SearchCQLOptions) (map[string]interface{}, error) {
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetConfluencePage retrieves a Confluence page by ID
func (c *Client) GetConfluencePage(pageID string) (map[string]interface{}, error) {
	baseURL := fmt.Sprintf("%s/wiki/rest/api/content/%s", c.BaseURL, pageID)

	// Request body content expanded
	params := url.Values{}
	params.Add("expand", "body.storage,version,space,history")

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

	var result map[string]interface{}
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
func (c *Client) GetConfluenceSpaces(opts *GetSpacesOptions) (map[string]interface{}, error) {
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

	var result map[string]interface{}
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
func (c *Client) GetPagesInSpace(opts *GetPagesInSpaceOptions) (map[string]interface{}, error) {
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

	var result map[string]interface{}
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
func (c *Client) CreateConfluencePage(opts *CreatePageOptions) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	body := map[string]interface{}{
		"type":  "page",
		"title": opts.Title,
		"space": map[string]interface{}{
			"key": opts.SpaceKey,
		},
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          opts.Body,
				"representation": "storage",
			},
		},
	}

	if opts.ParentID != "" {
		body["ancestors"] = []interface{}{
			map[string]interface{}{
				"id": opts.ParentID,
			},
		}
	}

	if opts.IsPrivate {
		body["metadata"] = map[string]interface{}{
			"properties": map[string]interface{}{
				"editor": map[string]interface{}{
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

	var result map[string]interface{}
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
func (c *Client) UpdateConfluencePage(opts *UpdatePageOptions) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content/%s", c.BaseURL, opts.PageID)

	body := map[string]interface{}{
		"type":  "page",
		"title": opts.Title,
		"version": map[string]interface{}{
			"number": opts.Version,
		},
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          opts.Body,
				"representation": "storage",
			},
		},
	}

	if opts.VersionMessage != "" {
		body["version"].(map[string]interface{})["message"] = opts.VersionMessage
	}

	if opts.ParentID != "" {
		body["ancestors"] = []interface{}{
			map[string]interface{}{
				"id": opts.ParentID,
			},
		}
	}

	if opts.SpaceKey != "" {
		body["space"] = map[string]interface{}{
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

	var result map[string]interface{}
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
func (c *Client) AddConfluencePageComment(opts *AddPageCommentOptions) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/wiki/rest/api/content", c.BaseURL)

	body := map[string]interface{}{
		"type": "comment",
		"container": map[string]interface{}{
			"id":   opts.PageID,
			"type": "page",
		},
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          opts.Comment,
				"representation": "storage",
			},
		},
	}

	// Add optional parameters
	if opts.ParentCommentID != "" {
		body["container"] = map[string]interface{}{
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
