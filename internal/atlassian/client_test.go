package atlassian

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		token         string
		site          string
		expectedURL   string
	}{
		{
			name:        "Site with https prefix",
			email:       "user@example.com",
			token:       "token123",
			site:        "https://company.atlassian.net",
			expectedURL: "https://company.atlassian.net",
		},
		{
			name:        "Site without https prefix",
			email:       "user@example.com",
			token:       "token123",
			site:        "company.atlassian.net",
			expectedURL: "https://company.atlassian.net",
		},
		{
			name:        "Site with http prefix",
			email:       "user@example.com",
			token:       "token123",
			site:        "http://company.atlassian.net",
			expectedURL: "http://company.atlassian.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.email, tt.token, tt.site)

			if client.Email != tt.email {
				t.Errorf("Expected email %q, got %q", tt.email, client.Email)
			}
			if client.Token != tt.token {
				t.Errorf("Expected token %q, got %q", tt.token, client.Token)
			}
			if client.BaseURL != tt.expectedURL {
				t.Errorf("Expected base URL %q, got %q", tt.expectedURL, client.BaseURL)
			}
			if client.client == nil {
				t.Error("Expected HTTP client to be initialized")
			}
		})
	}
}

func TestBasicAuth(t *testing.T) {
	client := NewClient("user@example.com", "secret-token", "company.atlassian.net")

	authHeader := client.basicAuth()

	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("Expected auth header to start with 'Basic ', got %q", authHeader)
	}

	// Decode and verify
	encodedPart := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedPart)
	if err != nil {
		t.Fatalf("Failed to decode auth header: %v", err)
	}

	expected := "user@example.com:secret-token"
	if string(decoded) != expected {
		t.Errorf("Expected decoded auth %q, got %q", expected, string(decoded))
	}
}

func TestGetVisibleProjects_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/project/search") {
			t.Errorf("Expected /rest/api/3/project/search path, got %s", r.URL.Path)
		}

		// Check auth header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Basic auth header, got %q", authHeader)
		}

		// Return mock response
		response := map[string]interface{}{
			"values": []interface{}{
				map[string]interface{}{
					"id":   "10000",
					"key":  "ABC",
					"name": "Test Project",
				},
				map[string]interface{}{
					"id":   "10001",
					"key":  "XYZ",
					"name": "Another Project",
				},
			},
			"total": 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	projects, err := client.GetVisibleProjects(&GetVisibleProjectsOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	firstProject := projects[0]
	if firstProject["key"] != "ABC" {
		t.Errorf("Expected first project key 'ABC', got %v", firstProject["key"])
	}
}

func TestGetVisibleProjects_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errorMessages":["Authentication failed"]}`))
	}))
	defer server.Close()

	client := NewClient("user@example.com", "bad-token", server.URL)

	_, err := client.GetVisibleProjects(&GetVisibleProjectsOptions{})
	if err == nil {
		t.Error("Expected error for unauthorized request, got nil")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Expected error to mention 401 status, got %v", err)
	}
}

func TestGetJiraIssue_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/issue/ABC-123") {
			t.Errorf("Expected issue path, got %s", r.URL.Path)
		}

		// Return mock response
		response := map[string]interface{}{
			"key": "ABC-123",
			"fields": map[string]interface{}{
				"summary": "Test Issue",
				"status": map[string]interface{}{
					"name": "To Do",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	issue, err := client.GetJiraIssue("ABC-123", &GetIssueOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if issue["key"] != "ABC-123" {
		t.Errorf("Expected issue key 'ABC-123', got %v", issue["key"])
	}

	fields, ok := issue["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields map")
	}

	if fields["summary"] != "Test Issue" {
		t.Errorf("Expected summary 'Test Issue', got %v", fields["summary"])
	}
}

func TestCreateJiraIssue_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/issue") {
			t.Errorf("Expected issue path, got %s", r.URL.Path)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify request structure
		fields, ok := requestBody["fields"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected fields in request body")
		}

		project, ok := fields["project"].(map[string]interface{})
		if !ok || project["key"] != "ABC" {
			t.Error("Expected project key 'ABC' in request")
		}

		// Return mock response
		response := map[string]interface{}{
			"key": "ABC-124",
			"id":  "10100",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	opts := &CreateIssueOptions{
		ProjectKey:  "ABC",
		IssueType:   "Task",
		Summary:     "New Task",
		Description: "Test description",
	}

	result, err := client.CreateJiraIssue(opts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["key"] != "ABC-124" {
		t.Errorf("Expected issue key 'ABC-124', got %v", result["key"])
	}
	if result["id"] != "10100" {
		t.Errorf("Expected issue id '10100', got %v", result["id"])
	}
}

func TestSearchJiraIssuesJQL_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/search") {
			t.Errorf("Expected search path, got %s", r.URL.Path)
		}

		// Verify JQL parameter
		jql := r.URL.Query().Get("jql")
		if jql == "" {
			t.Error("Expected JQL parameter in query")
		}

		// Return mock response
		response := map[string]interface{}{
			"issues": []interface{}{
				map[string]interface{}{
					"key": "ABC-100",
					"fields": map[string]interface{}{
						"summary": "First Issue",
					},
				},
				map[string]interface{}{
					"key": "ABC-101",
					"fields": map[string]interface{}{
						"summary": "Second Issue",
					},
				},
			},
			"total": 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	result, err := client.SearchJiraIssuesJQL("project = ABC", &SearchJQLOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	issues, ok := result["issues"].([]interface{})
	if !ok {
		t.Fatal("Expected issues array in result")
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
}

func TestGetConfluencePage_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/wiki/rest/api/content/123") {
			t.Errorf("Expected page path with /wiki/rest/api/content/123, got %s", r.URL.Path)
		}

		// Return mock response
		response := map[string]interface{}{
			"id":    "123",
			"title": "Test Page",
			"body": map[string]interface{}{
				"storage": map[string]interface{}{
					"value": "<p>Page content</p>",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	page, err := client.GetConfluencePage("123", &GetPageOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if page["id"] != "123" {
		t.Errorf("Expected page id '123', got %v", page["id"])
	}
	if page["title"] != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %v", page["title"])
	}
}

func TestGetConfluenceSpaces_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/wiki/rest/api/space") {
			t.Errorf("Expected spaces path with /wiki/rest/api/space, got %s", r.URL.Path)
		}

		// Return mock response
		response := map[string]interface{}{
			"results": []interface{}{
				map[string]interface{}{
					"id":   "100",
					"key":  "TEST",
					"name": "Test Space",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	result, err := client.GetConfluenceSpaces(&GetSpacesOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	results, ok := result["results"].([]interface{})
	if !ok {
		t.Fatal("Expected results array in response")
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 space, got %d", len(results))
	}

	space, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected space to be a map")
	}

	if space["key"] != "TEST" {
		t.Errorf("Expected space key 'TEST', got %v", space["key"])
	}
}

func TestClientTimeout(t *testing.T) {
	client := NewClient("user@example.com", "token", "company.atlassian.net")

	if client.client.Timeout == 0 {
		t.Error("Expected HTTP client to have a timeout set")
	}
}

func TestGetCreateMeta_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/issue/createmeta/ABC/issuetypes/10002") {
			t.Errorf("Expected createmeta path, got %s", r.URL.Path)
		}

		// Return mock response
		response := map[string]interface{}{
			"fields": []interface{}{
				map[string]interface{}{
					"key":      "summary",
					"name":     "Summary",
					"required": true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
				map[string]interface{}{
					"key":      "customfield_10369",
					"name":     "Portfolio Work Type",
					"required": true,
					"schema": map[string]interface{}{
						"type": "option",
					},
					"allowedValues": []interface{}{
						map[string]interface{}{
							"id":    "10689",
							"value": "Growth",
						},
						map[string]interface{}{
							"id":    "10690",
							"value": "KTLO",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	metadata, err := client.GetCreateMeta("ABC", "10002")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	fields, ok := metadata["fields"].([]interface{})
	if !ok {
		t.Fatal("Expected fields array in metadata")
	}

	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Check for custom field
	found := false
	for _, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}
		if fieldMap["key"] == "customfield_10369" {
			found = true
			if fieldMap["name"] != "Portfolio Work Type" {
				t.Errorf("Expected field name 'Portfolio Work Type', got %v", fieldMap["name"])
			}
		}
	}

	if !found {
		t.Error("Expected to find customfield_10369 in response")
	}
}

func TestGetFieldOptions_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will be called twice - once for issue types, once for create meta
		if strings.Contains(r.URL.Path, "/issuetypes/10002") {
			// Return createmeta response
			response := map[string]interface{}{
				"fields": []interface{}{
					map[string]interface{}{
						"key":  "customfield_10369",
						"name": "Portfolio Work Type",
						"allowedValues": []interface{}{
							map[string]interface{}{
								"id":    "10689",
								"value": "Growth",
							},
							map[string]interface{}{
								"id":    "10690",
								"value": "KTLO",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "/issuetypes") {
			// Return issue types response
			response := map[string]interface{}{
				"values": []interface{}{
					map[string]interface{}{
						"id":   "10002",
						"name": "Task",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "token", server.URL)

	fieldInfo, err := client.GetFieldOptions("customfield_10369", "ABC", "10002")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if fieldInfo["name"] != "Portfolio Work Type" {
		t.Errorf("Expected field name 'Portfolio Work Type', got %v", fieldInfo["name"])
	}

	allowedValues, ok := fieldInfo["allowedValues"].([]interface{})
	if !ok {
		t.Fatal("Expected allowedValues array")
	}

	if len(allowedValues) != 2 {
		t.Errorf("Expected 2 allowed values, got %d", len(allowedValues))
	}
}
