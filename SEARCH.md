# Searching in Atlassian CLI

Unlike the Atlassian MCP server, this CLI does not support unified Rovo search across both Jira and Confluence. This is because:

1. **Rovo has no public REST API** - Confirmed by Atlassian Community forums
2. **MCP server uses OAuth 2.1** - Accesses privileged endpoints via `https://mcp.atlassian.com/v1/sse`
3. **CLI uses Basic Auth (API tokens)** - Limited to standard REST APIs only

## How to Search

To search across Jira and Confluence, you need to use **both** search commands separately:

### Search Jira Issues (JQL)

```bash
# Search by project
./atl jira search-jql "project = FX"

# Search by assignee
./atl jira search-jql "assignee = currentUser()"

# Complex queries
./atl jira search-jql "project = FX AND status = 'In Progress' AND created >= -7d"

# Combine searches
./atl jira search-jql "summary ~ 'authentication' OR description ~ 'authentication'"
```

**JQL (Jira Query Language) Resources:**
- [JQL Documentation](https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/)

### Search Confluence Pages (CQL)

```bash
# Search by title
./atl confluence search-cql "title ~ 'Onboarding'"

# Search by space
./atl confluence search-cql "space = POL"

# Search by text content
./atl confluence search-cql "text ~ 'authentication'"

# Complex queries
./atl confluence search-cql "title ~ 'FX' AND space = POL AND type = page"
```

**CQL (Confluence Query Language) Resources:**
- [CQL Documentation](https://developer.atlassian.com/server/confluence/advanced-searching-using-cql/)

### Searching Across Both Products

To search for "authentication" across everything:

```bash
# Search Jira
./atl jira search-jql "summary ~ 'authentication' OR description ~ 'authentication'" --json > jira-results.json

# Search Confluence
./atl confluence search-cql "text ~ 'authentication'" --json > confluence-results.json

# Combine results with jq or similar
cat jira-results.json confluence-results.json | jq -s '.'
```

## Why No Unified Search?

The Atlassian MCP server's unified `search` tool uses Rovo, which:
- Requires OAuth 2.1 authentication (not Basic Auth)
- Uses a specialized gateway at `mcp.atlassian.com` (not standard REST API)
- Is designed for the MCP protocol, not direct HTTP access

Our CLI uses API tokens (Basic Auth) for simplicity and to avoid frequent re-authentication issues. This is the trade-off: easier authentication vs. no Rovo access.

## References

- [Is there API Rest access to Rovo?](https://community.atlassian.com/forums/Rovo-questions/Is-there-API-Rest-access-to-Rovo/qaq-p/3037848) - Confirms no public REST API
- [Atlassian Rovo MCP Server](https://support.atlassian.com/atlassian-rovo-mcp-server/docs/use-atlassian-rovo-mcp-server/) - MCP server documentation
