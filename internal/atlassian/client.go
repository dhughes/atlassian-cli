package atlassian

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Fields []string // List of fields to return
	Expand []string // List of parameters to expand
}

// GetJiraIssue retrieves a Jira issue by its key or ID
func (c *Client) GetJiraIssue(issueKey string, opts *GetIssueOptions) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.BaseURL, issueKey)

	// Add query parameters
	if opts != nil {
		query := ""
		if len(opts.Fields) > 0 {
			query += "fields=" + strings.Join(opts.Fields, ",")
		}
		if len(opts.Expand) > 0 {
			if query != "" {
				query += "&"
			}
			query += "expand=" + strings.Join(opts.Expand, ",")
		}
		if query != "" {
			url += "?" + query
		}
	}

	resp, err := c.doRequest("GET", url, nil)
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
	url := fmt.Sprintf("%s/rest/api/3/search", c.BaseURL)

	// Build query parameters
	query := "jql=" + jql

	if opts != nil {
		if len(opts.Fields) > 0 {
			query += "&fields=" + strings.Join(opts.Fields, ",")
		}
		if opts.MaxResults > 0 {
			query += fmt.Sprintf("&maxResults=%d", opts.MaxResults)
		} else {
			query += "&maxResults=50" // Default
		}
		if opts.StartAt > 0 {
			query += fmt.Sprintf("&startAt=%d", opts.StartAt)
		}
	} else {
		query += "&maxResults=50"
	}

	url += "?" + query

	resp, err := c.doRequest("GET", url, nil)
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
