package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/doughughes/atlassian-cli/internal/atlassian"
	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	apiMethod   string
	apiHeaders  []string
	apiData     string
	apiInclude  bool
	apiVerbose  bool
)

var apiCmd = &cobra.Command{
	Use:   "api <path>",
	Short: "Make an authenticated request to any Atlassian REST endpoint",
	Long: `Make an authenticated HTTP request to any Atlassian REST endpoint.

The path argument may be:
  - A path relative to the active account's site, e.g. /rest/api/3/myself
    (prepended with https://<active-site>)
  - An absolute URL, e.g. https://api.atlassian.com/oauth/token/accessible-resources

The default method is GET, or POST if a request body is provided via -d.

Examples:
  # GET a Jira issue
  atl api /rest/api/3/issue/ABC-123

  # GET a Confluence page with query string
  atl api '/wiki/rest/api/content/4012966062?status=historical&version=15&expand=body.storage'

  # POST with an inline JSON body
  atl api -d '{"jql":"project = ABC"}' /rest/api/3/search

  # POST with a JSON body from a file
  atl api -d @body.json /rest/api/3/issue/ABC-123/comment

  # POST with a body from stdin
  echo '{"body":"hi"}' | atl api -d @- /rest/api/3/issue/ABC-123/comment

  # Absolute URL
  atl api https://api.atlassian.com/oauth/token/accessible-resources`,
	Args:          cobra.ExactArgs(1),
	RunE:          runAPI,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", "", "HTTP method (default GET, or POST if -d is provided)")
	apiCmd.Flags().StringArrayVarP(&apiHeaders, "header", "H", nil, "Request header in 'Key: Value' format (repeatable)")
	apiCmd.Flags().StringVarP(&apiData, "data", "d", "", "Request body. Prefix with @ to read from a file (@-) for stdin")
	apiCmd.Flags().BoolVarP(&apiInclude, "include", "i", false, "Print response status and headers to stderr before the body")
	apiCmd.Flags().BoolVarP(&apiVerbose, "verbose", "v", false, "Print the outgoing method, URL, and headers to stderr")
}

func runAPI(cmd *cobra.Command, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("path is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	account, err := cfg.GetActiveAccount()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'atl auth login' first")
	}
	client := atlassian.NewClient(account.Email, account.Token, account.Site)

	body, hasBody, err := readBody(apiData, cmd.InOrStdin())
	if err != nil {
		return err
	}

	method := apiMethod
	if method == "" {
		if hasBody {
			method = http.MethodPost
		} else {
			method = http.MethodGet
		}
	}
	method = strings.ToUpper(method)

	url := client.ResolveURL(path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	for _, h := range apiHeaders {
		key, value, ok := splitHeader(h)
		if !ok {
			return fmt.Errorf("invalid header %q (expected 'Key: Value')", h)
		}
		if strings.EqualFold(key, "Authorization") {
			continue
		}
		req.Header.Add(key, value)
	}

	if hasBody && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if apiVerbose {
		writeVerbose(method, url, req.Header)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if apiInclude {
		writeResponseHeaders(resp)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return nil
	}

	// HTTP error: write the raw response body to stderr and exit 1 without
	// letting main.go decorate it with "Error: ...". The body itself is the
	// useful diagnostic — Atlassian returns structured JSON. os.Exit skips
	// the deferred Close above, which is fine — the OS reclaims the fd.
	_, _ = io.Copy(os.Stderr, resp.Body)
	os.Exit(1)
	return nil
}

// readBody resolves the -d flag into a request body. Returns (body, hasBody,
// error). An empty data string is treated as "no body" — callers wanting an
// empty-body POST should pass an explicit -X POST.
func readBody(data string, stdin io.Reader) (io.Reader, bool, error) {
	if data == "" {
		return nil, false, nil
	}
	if strings.HasPrefix(data, "@") {
		source := data[1:]
		if source == "-" {
			b, err := io.ReadAll(stdin)
			if err != nil {
				return nil, true, fmt.Errorf("failed to read body from stdin: %w", err)
			}
			return strings.NewReader(string(b)), true, nil
		}
		b, err := os.ReadFile(source)
		if err != nil {
			return nil, true, fmt.Errorf("could not read body file: %w", err)
		}
		return strings.NewReader(string(b)), true, nil
	}
	return strings.NewReader(data), true, nil
}

// splitHeader parses a "Key: Value" string. Whitespace around the colon is
// trimmed. Returns ok=false when no colon is present.
func splitHeader(h string) (key, value string, ok bool) {
	i := strings.IndexByte(h, ':')
	if i < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(h[:i])
	value = strings.TrimSpace(h[i+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

// writeVerbose prints the outgoing method, URL, and headers to stderr.
// It accounts for two headers injected by Client.Do after this point:
// Authorization (always set, masked here) and Accept (defaulted to JSON when
// the caller did not set it).
func writeVerbose(method, url string, headers http.Header) {
	fmt.Fprintf(os.Stderr, "> %s %s\n", method, url)
	fmt.Fprintln(os.Stderr, "> Authorization: Basic [REDACTED]")
	if headers.Get("Accept") == "" {
		fmt.Fprintln(os.Stderr, "> Accept: application/json")
	}
	for _, k := range sortedHeaderKeys(headers) {
		if strings.EqualFold(k, "Authorization") {
			continue
		}
		for _, v := range headers.Values(k) {
			fmt.Fprintf(os.Stderr, "> %s: %s\n", k, v)
		}
	}
}

func writeResponseHeaders(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "< %s %s\n", resp.Proto, resp.Status)
	for _, k := range sortedHeaderKeys(resp.Header) {
		for _, v := range resp.Header.Values(k) {
			fmt.Fprintf(os.Stderr, "< %s: %s\n", k, v)
		}
	}
}

func sortedHeaderKeys(h http.Header) []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
