# Atlassian CLI

A command-line interface for interacting with Atlassian Jira and Confluence APIs.

> **Experimental Project**: This tool was developed as an experiment in using AI coding assistants (Claude Code) to generate functional CLI tools that interact with public REST APIs. The entire codebase, including comprehensive unit tests, was created through AI-assisted development.

## Why This Tool?

The `atl` CLI was created to address practical challenges encountered with the official Atlassian MCP server, particularly frequent reauthentication requirements that disrupted workflows. Inspired by GitHub's `gh` CLI—which proved more reliable and useful than the GitHub MCP server—this project explores a similar approach for Atlassian products.

**Key Goals:**
- **Reliability**: Persistent authentication using API tokens (no OAuth token expiration)
- **Dual-purpose**: Designed for both human use and AI/automation integration
- **Simplicity**: Straightforward commands following familiar CLI patterns
- **Extensibility**: Easy to integrate into scripts and automated workflows

The result is a tool that provides feature parity with the Atlassian MCP server while being more reliable, supporting long-lived API tokens (not OAuth), and equally effective for both manual and automated use cases.

## Installation

### Prerequisites

- Go 1.25.4 or later (managed via asdf)

### Building

```bash
# Install dependencies
go mod download

# Build the binary
go build -o atl .

# Optionally, install to your PATH
# go install
```

## Authentication

### Initial Setup

Generate an API token at: https://id.atlassian.com/manage-profile/security/api-tokens

Then log in:

```bash
./atl auth login
```

You'll be prompted for:
- **Atlassian site URL**: Your organization's URL (e.g., `yourcompany.atlassian.net`)
- **Email**: Your Atlassian account email
- **API token**: The token you generated (input is hidden)

The CLI will automatically:
- Verify your credentials using the Jira API
- Retrieve your display name
- Save the configuration to `~/.config/atlassian/config.json`

### Check Authentication Status

```bash
./atl auth status
```

Shows:
- Active account name
- Site URL
- Email address
- Credential validity
- Your display name

### Log Out

```bash
./atl auth logout
```

Removes stored credentials for the active account.

## Quick Start

### Jira Examples

```bash
# Get issue details
./atl jira get-issue ABC-123

# Search issues
./atl jira search-jql "project = ABC AND status = 'In Progress'"

# Create an issue
./atl jira create-issue \
  --project ABC \
  --type Task \
  --summary "Implement new feature" \
  --description "## Details\n- First step\n- Second step"

# Discover required fields for creating issues
./atl jira get-create-meta ABC 10002
./atl jira get-field-options customfield_10369 --project ABC --issue-type-id 10002

# Create issue with custom fields
./atl jira create-issue \
  --project ABC \
  --type Task \
  --summary "New ticket" \
  --fields '{"customfield_10369": {"id": "10690"}}'

# Transition issue to new status
./atl jira get-transitions ABC-123
./atl jira transition-issue ABC-123 31
```

### Confluence Examples

```bash
# Get page content
./atl confluence get-page 123456789

# List pages in a space
./atl confluence get-pages-in-space TEAM

# Create a page
./atl confluence create-page \
  --space TEAM \
  --title "New Documentation" \
  --body "<p>Page content here</p>"

# Update existing page
./atl confluence update-page 123456789 \
  --title "Updated Title" \
  --body "<p>Updated content</p>" \
  --version 2

# Search Confluence
./atl confluence search-cql "title ~ 'API' AND space = TEAM"

# Add a comment
./atl confluence add-comment 123456789 "Great documentation!"
```

## Configuration

### View All Configuration

```bash
./atl config list
```

Shows:
- Active account
- All configured accounts
- Default settings

### Get Specific Configuration Value

```bash
./atl config get active-account
./atl config get site
./atl config get email
```


## Project Structure

```
.
├── main.go                    # Entry point
├── cmd/                       # Command implementations
│   ├── root.go               # Root command
│   ├── auth.go               # Auth commands (login, status, logout)
│   └── config.go             # Config commands (get, set, list, unset)
├── internal/
│   ├── atlassian/            # Atlassian API client
│   │   └── client.go
│   └── config/               # Configuration management
│       └── config.go
└── README.md
```

## Configuration File

Configuration is stored at `~/.config/atlassian/config.json`:

```json
{
  "active_account": "mycompany",
  "accounts": {
    "mycompany": {
      "site": "mycompany.atlassian.net",
      "email": "your-email@example.com",
      "token": "your-api-token"
    }
  }
}
```

**Security Note**: The config file is created with 0600 permissions (user read/write only).

## Searching

This CLI does not support unified Rovo search like the Atlassian MCP server. Rovo requires OAuth 2.1 and has no public REST API.

**To search, use product-specific commands:**

```bash
# Search Jira issues with JQL
./atl jira search-jql "project = PROJ AND status = 'In Progress'"

# Search Confluence pages with CQL
./atl confluence search-cql "title ~ 'Documentation' AND space = TEAM"
```

See [SEARCH.md](SEARCH.md) for detailed examples and guidance.

## Feature Status

### ✅ Completed Features

**Core Infrastructure:**
- Authentication with API tokens (persistent, no expiration)
- Configuration management (`~/.config/atlassian/config.json`)
- Multiple account support with account switching
- JSON output for all commands (via `--json` flag)
- Secure credential storage (0600 file permissions)

**Jira Commands:**
- Issue operations: `get-issue`, `create-issue`, `edit-issue`
- Search: `search-jql`
- Comments: `add-comment`
- Workflow: `get-transitions`, `transition-issue`
- Project info: `get-projects`, `get-project-issue-types`
- Field discovery: `get-create-meta`, `get-field-options`
- User lookup: `lookup-account-id`
- Remote links: `get-remote-links`

**Confluence Commands:**
- Page operations: `get-page`, `create-page`, `update-page`
- Space navigation: `get-pages-in-space`, `get-spaces`
- Page hierarchy: `get-page-ancestors`, `get-page-descendants`
- Comments: `get-page-comments`, `add-comment`, `create-inline-comment`
- Search: `search-cql`

**Content Formatting:**
- Markdown-to-ADF conversion for Jira descriptions
- HTML rendering for Confluence pages (readable terminal output)
- Pretty-printed output by default

## Development

### Run Without Building

```bash
go run . auth status
```

### Watch Mode

```bash
# In one terminal
go build -o atl . && ./atl auth status
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run tests for a specific package
go test ./internal/atlassian
go test ./internal/config

# Run a specific test
go test ./internal/atlassian -run TestADFToText_SimpleParagraph
```

**Test Coverage:**
- `internal/config`: 79.5% coverage
- `internal/atlassian`: 41.5% coverage

Tests include:
- ADF (Atlassian Document Format) parsing and conversion
- HTML to text conversion
- Markdown to ADF conversion
- Configuration management (load, save, account management)
- API client functionality (with mocked HTTP responses)

## License

MIT
