package cmd

import (
	"fmt"

	"github.com/doughughes/atlassian-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify CLI configuration settings.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Display all configuration settings including defaults and account information.`,
	RunE:  runConfigList,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  `Retrieve a specific configuration value by key.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. This affects default values used by commands.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Unset a configuration value",
	Long:  `Remove a configuration value, reverting to built-in defaults.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigUnset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
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
			fmt.Printf("    site:     %s\n", account.Site)
			fmt.Printf("    email:    %s\n", account.Email)
			fmt.Printf("    cloud-id: %s\n", account.CloudID)
		}
	}

	fmt.Println("\nDefaults:")
	if len(cfg.Defaults) == 0 {
		fmt.Println("  (none)")
	} else {
		for key, value := range cfg.Defaults {
			fmt.Printf("  %s: %s\n", key, value)
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
	case "cloud-id":
		account, err := cfg.GetActiveAccount()
		if err != nil {
			return err
		}
		fmt.Println(account.CloudID)
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

	// Check defaults
	value := cfg.GetDefault(key)
	if value == "" {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	fmt.Println(value)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Handle special keys
	switch key {
	case "cloud-id":
		account, err := cfg.GetActiveAccount()
		if err != nil {
			return err
		}
		account.CloudID = value
	case "active-account", "site", "email", "token":
		return fmt.Errorf("'%s' cannot be set directly, use 'atl auth' commands", key)
	default:
		// Set as default
		cfg.SetDefault(key, value)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if key exists
	if cfg.GetDefault(key) == "" {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	delete(cfg.Defaults, key)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Unset %s\n", key)
	return nil
}
