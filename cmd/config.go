package cmd

import (
	"fmt"

	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View configuration",
	Long:  `View CLI configuration settings.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Display all configuration settings and account information.`,
	RunE:  runConfigList,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Retrieve a specific configuration value by key.

Valid keys: active-account, site, email`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Active Account:")
	if cfg.ActiveAccount == "" {
		fmt.Println("  (none)")
	} else {
		fmt.Printf("  %s\n", cfg.ActiveAccount)
	}

	fmt.Println("\nAccounts:")
	if len(cfg.Accounts) == 0 {
		fmt.Println("  (none)")
	} else {
		for name, account := range cfg.Accounts {
			active := ""
			if name == cfg.ActiveAccount {
				active = " (active)"
			}
			fmt.Printf("  %s%s:\n", name, active)
			fmt.Printf("    site:  %s\n", account.Site)
			fmt.Printf("    email: %s\n", account.Email)
		}
	}


	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Handle special keys
	switch key {
	case "active-account":
		if cfg.ActiveAccount == "" {
			return fmt.Errorf("no active account")
		}
		fmt.Println(cfg.ActiveAccount)
		return nil
	case "site":
		account, err := cfg.GetActiveAccount()
		if err != nil {
			return err
		}
		fmt.Println(account.Site)
		return nil
	case "email":
		account, err := cfg.GetActiveAccount()
		if err != nil {
			return err
		}
		fmt.Println(account.Email)
		return nil
	}

	// Unknown key
	return fmt.Errorf("unknown configuration key '%s'. Valid keys: active-account, site, email", key)
}
