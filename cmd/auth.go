package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/doughughes/atlassian-cli/internal/atlassian"
	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Atlassian",
	Long:  `Manage authentication credentials for Atlassian Cloud.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to an Atlassian account",
	Long: `Authenticate with Atlassian Cloud by providing your site URL, email, and API token.

Your API token can be generated at: https://id.atlassian.com/manage-profile/security/api-tokens`,
	RunE: runLogin,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Display the current authentication status and active account information.`,
	RunE:  runStatus,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of an Atlassian account",
	Long:  `Remove stored credentials for the active account.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for site URL
	fmt.Print("Atlassian site URL (e.g., yourcompany.atlassian.net): ")
	site, _ := reader.ReadString('\n')
	site = strings.TrimSpace(site)
	if site == "" {
		return fmt.Errorf("site URL is required")
	}

	// Prompt for email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Prompt for API token (hidden input)
	fmt.Print("API token (hidden): ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after hidden input
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("API token is required")
	}

	// Create client and test authentication
	fmt.Println("\nVerifying credentials...")
	client := atlassian.NewClient(email, token, site)

	user, err := client.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("authentication failed: %w\n\nPlease verify:\n  - Your site URL is correct (e.g., yourcompany.atlassian.net)\n  - Your email is correct\n  - Your API token is valid (generate at https://id.atlassian.com)\n  - You have access to this Atlassian site", err)
	}

	// Save configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use site domain as account name (e.g., "mycompany" from "mycompany.atlassian.net")
	configAccountName := strings.Split(site, ".")[0]

	cfg.SetAccount(configAccountName, &config.Account{
		Site:    site,
		Email:   email,
		CloudID: "", // Cloud ID will be set manually if needed via config command
		Token:   token,
	})
	cfg.ActiveAccount = configAccountName

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✓ Successfully authenticated as %s\n", user.DisplayName)
	fmt.Printf("✓ Email: %s\n", email)
	fmt.Printf("✓ Site: %s\n", site)
	fmt.Printf("\nConfiguration saved. You can now use 'atl' commands.\n")
	fmt.Printf("\nNote: Cloud ID not set. If needed for specific commands, set it with:\n")
	fmt.Printf("  atl config set cloud-id <your-cloud-id>\n")

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ActiveAccount == "" {
		fmt.Println("Not logged in. Run 'atl auth login' to authenticate.")
		return nil
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		return err
	}

	fmt.Printf("Logged in to: %s\n", cfg.ActiveAccount)
	fmt.Printf("  Site:     %s\n", account.Site)
	fmt.Printf("  Email:    %s\n", account.Email)
	fmt.Printf("  Cloud ID: %s\n", account.CloudID)

	// Test if credentials are still valid
	fmt.Print("  Status:   ")
	client := atlassian.NewClient(account.Email, account.Token, account.Site)
	if err := client.TestAuthentication(); err != nil {
		fmt.Println("✗ Invalid (credentials may have expired)")
		return nil
	}

	// Get user info
	user, err := client.GetCurrentUser()
	if err != nil {
		fmt.Println("✓ Valid")
	} else {
		fmt.Printf("✓ Valid (%s)\n", user.DisplayName)
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ActiveAccount == "" {
		fmt.Println("Not logged in.")
		return nil
	}

	accountName := cfg.ActiveAccount
	delete(cfg.Accounts, accountName)
	cfg.ActiveAccount = ""

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Logged out of %s\n", accountName)
	return nil
}
