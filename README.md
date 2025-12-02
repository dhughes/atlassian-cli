# Atlassian CLI

A command-line interface for interacting with Atlassian Jira and Confluence APIs.

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

**Note:** Cloud ID is not automatically discovered (it requires OAuth, not Basic Auth). Most commands work without it. If you need to set it manually for specific operations, use:

```bash
./atl config set cloud-id <your-cloud-id>
```

### Check Authentication Status

```bash
./atl auth status
```

Shows:
- Active account name
- Site URL
- Email address
- Cloud ID (if set)
- Credential validity
- Your display name

### Log Out

```bash
./atl auth logout
```

Removes stored credentials for the active account.

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
./atl config get cloud-id
./atl config get active-account
./atl config get site
./atl config get email
```

### Set Configuration Value

```bash
./atl config set <key> <value>
```

Useful for setting defaults like:
```bash
./atl config set default-project FX
./atl config set output-format json
```

### Unset Configuration Value

```bash
./atl config unset <key>
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
      "cloud_id": "",
      "token": "your-api-token"
    }
  },
  "defaults": {}
}
```

**Note:** `cloud_id` is optional. Set it manually if needed: `./atl config set cloud-id <value>`

**Security Note**: The config file is created with 0600 permissions (user read/write only).

## Future Development

### Phase 2: Essential Jira Commands (Coming Soon)
- `atl jira get-issue`
- `atl jira create-issue`
- `atl jira edit-issue`
- `atl jira search-jql`
- And more...

### Phase 3: Essential Confluence Commands
- `atl confluence get-page`
- `atl confluence get-pages-in-space`
- `atl confluence create-page`
- `atl confluence update-page`
- And more...

### Phase 4: Complete MCP Coverage
- All remaining Jira operations
- All remaining Confluence operations
- Meta commands (search, fetch by ARI)
- Multiple account support
- Shell completion

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

## License

MIT
