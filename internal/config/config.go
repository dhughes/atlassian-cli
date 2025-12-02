package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	ActiveAccount string              `json:"active_account,omitempty"`
	Accounts      map[string]*Account `json:"accounts,omitempty"`
}

// Account represents an Atlassian account configuration
type Account struct {
	Site  string `json:"site"`
	Email string `json:"email"`
	Token string `json:"token"`
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "atlassian")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Accounts: make(map[string]*Account),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Accounts == nil {
		cfg.Accounts = make(map[string]*Account)
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetActiveAccount returns the currently active account
func (c *Config) GetActiveAccount() (*Account, error) {
	if c.ActiveAccount == "" {
		return nil, fmt.Errorf("no active account set")
	}

	account, ok := c.Accounts[c.ActiveAccount]
	if !ok {
		return nil, fmt.Errorf("active account '%s' not found", c.ActiveAccount)
	}

	return account, nil
}

// SetAccount adds or updates an account
func (c *Config) SetAccount(name string, account *Account) {
	if c.Accounts == nil {
		c.Accounts = make(map[string]*Account)
	}
	c.Accounts[name] = account
}
