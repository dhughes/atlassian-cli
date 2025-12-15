package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := &Config{
		ActiveAccount: "test",
		Accounts: map[string]*Account{
			"test": {
				Site:  "test.atlassian.net",
				Email: "test@example.com",
				Token: "test-token",
			},
		},
	}

	if cfg.ActiveAccount != "test" {
		t.Errorf("Expected active account 'test', got %q", cfg.ActiveAccount)
	}

	if len(cfg.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(cfg.Accounts))
	}
}

func TestGetActiveAccount_Success(t *testing.T) {
	cfg := &Config{
		ActiveAccount: "main",
		Accounts: map[string]*Account{
			"main": {
				Site:  "company.atlassian.net",
				Email: "user@company.com",
				Token: "token123",
			},
		},
	}

	account, err := cfg.GetActiveAccount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if account.Site != "company.atlassian.net" {
		t.Errorf("Expected site 'company.atlassian.net', got %q", account.Site)
	}
	if account.Email != "user@company.com" {
		t.Errorf("Expected email 'user@company.com', got %q", account.Email)
	}
	if account.Token != "token123" {
		t.Errorf("Expected token 'token123', got %q", account.Token)
	}
}

func TestGetActiveAccount_NoActiveAccount(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]*Account{
			"test": {
				Site:  "test.atlassian.net",
				Email: "test@example.com",
				Token: "token",
			},
		},
	}

	_, err := cfg.GetActiveAccount()
	if err == nil {
		t.Error("Expected error for no active account, got nil")
	}

	if err.Error() != "no active account set" {
		t.Errorf("Expected 'no active account set' error, got %q", err.Error())
	}
}

func TestGetActiveAccount_AccountNotFound(t *testing.T) {
	cfg := &Config{
		ActiveAccount: "missing",
		Accounts: map[string]*Account{
			"other": {
				Site:  "other.atlassian.net",
				Email: "other@example.com",
				Token: "token",
			},
		},
	}

	_, err := cfg.GetActiveAccount()
	if err == nil {
		t.Error("Expected error for missing account, got nil")
	}

	expectedError := "active account 'missing' not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSetAccount_NewAccount(t *testing.T) {
	cfg := &Config{
		Accounts: make(map[string]*Account),
	}

	account := &Account{
		Site:  "new.atlassian.net",
		Email: "new@example.com",
		Token: "new-token",
	}

	cfg.SetAccount("new", account)

	if len(cfg.Accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(cfg.Accounts))
	}

	savedAccount := cfg.Accounts["new"]
	if savedAccount.Site != "new.atlassian.net" {
		t.Errorf("Expected site 'new.atlassian.net', got %q", savedAccount.Site)
	}
}

func TestSetAccount_UpdateExistingAccount(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]*Account{
			"existing": {
				Site:  "old.atlassian.net",
				Email: "old@example.com",
				Token: "old-token",
			},
		},
	}

	updatedAccount := &Account{
		Site:  "updated.atlassian.net",
		Email: "updated@example.com",
		Token: "updated-token",
	}

	cfg.SetAccount("existing", updatedAccount)

	if len(cfg.Accounts) != 1 {
		t.Errorf("Expected 1 account (updated), got %d", len(cfg.Accounts))
	}

	savedAccount := cfg.Accounts["existing"]
	if savedAccount.Site != "updated.atlassian.net" {
		t.Errorf("Expected updated site, got %q", savedAccount.Site)
	}
}

func TestSetAccount_InitializesMapIfNil(t *testing.T) {
	cfg := &Config{}

	account := &Account{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "token",
	}

	cfg.SetAccount("test", account)

	if cfg.Accounts == nil {
		t.Fatal("Expected Accounts map to be initialized")
	}

	if len(cfg.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(cfg.Accounts))
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atlassian-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create a config
	cfg := &Config{
		ActiveAccount: "test",
		Accounts: map[string]*Account{
			"test": {
				Site:  "test.atlassian.net",
				Email: "test@example.com",
				Token: "test-token",
			},
			"other": {
				Site:  "other.atlassian.net",
				Email: "other@example.com",
				Token: "other-token",
			},
		},
	}

	// Save it
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", mode)
	}

	// Load it back
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify contents match
	if loadedCfg.ActiveAccount != cfg.ActiveAccount {
		t.Errorf("Active account mismatch: expected %q, got %q", cfg.ActiveAccount, loadedCfg.ActiveAccount)
	}

	if len(loadedCfg.Accounts) != len(cfg.Accounts) {
		t.Errorf("Account count mismatch: expected %d, got %d", len(cfg.Accounts), len(loadedCfg.Accounts))
	}

	for name, account := range cfg.Accounts {
		loadedAccount, exists := loadedCfg.Accounts[name]
		if !exists {
			t.Errorf("Account %q missing after load", name)
			continue
		}

		if loadedAccount.Site != account.Site {
			t.Errorf("Site mismatch for %q: expected %q, got %q", name, account.Site, loadedAccount.Site)
		}
		if loadedAccount.Email != account.Email {
			t.Errorf("Email mismatch for %q: expected %q, got %q", name, account.Email, loadedAccount.Email)
		}
		if loadedAccount.Token != account.Token {
			t.Errorf("Token mismatch for %q: expected %q, got %q", name, account.Token, loadedAccount.Token)
		}
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atlassian-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load non-existent config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error for nonexistent config, got %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.ActiveAccount != "" {
		t.Errorf("Expected empty active account, got %q", cfg.ActiveAccount)
	}

	if cfg.Accounts == nil {
		t.Error("Expected initialized Accounts map")
	}

	if len(cfg.Accounts) != 0 {
		t.Errorf("Expected empty accounts map, got %d accounts", len(cfg.Accounts))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atlassian-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Write invalid JSON
	err = os.WriteFile(configPath, []byte("invalid json{{{"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Try to load it
	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestConfigPath_CreatesDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atlassian-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Get config path (should create directory)
	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}

	// Verify directory was created
	configDir := filepath.Dir(configPath)
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	// Verify directory permissions
	mode := info.Mode().Perm()
	if mode != 0700 {
		t.Errorf("Expected directory permissions 0700, got %o", mode)
	}
}

func TestJSONSerialization(t *testing.T) {
	cfg := &Config{
		ActiveAccount: "main",
		Accounts: map[string]*Account{
			"main": {
				Site:  "company.atlassian.net",
				Email: "user@company.com",
				Token: "secret-token",
			},
			"secondary": {
				Site:  "other.atlassian.net",
				Email: "user@other.com",
				Token: "other-token",
			},
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal back
	var loadedCfg Config
	err = json.Unmarshal(data, &loadedCfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify match
	if loadedCfg.ActiveAccount != cfg.ActiveAccount {
		t.Errorf("Active account mismatch after JSON round-trip")
	}

	if len(loadedCfg.Accounts) != len(cfg.Accounts) {
		t.Errorf("Account count mismatch after JSON round-trip")
	}
}

func TestAccount_AllFields(t *testing.T) {
	account := &Account{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token-123",
	}

	if account.Site != "test.atlassian.net" {
		t.Errorf("Expected site 'test.atlassian.net', got %q", account.Site)
	}
	if account.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %q", account.Email)
	}
	if account.Token != "test-token-123" {
		t.Errorf("Expected token 'test-token-123', got %q", account.Token)
	}
}

func TestMultipleAccounts(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atlassian-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config with multiple accounts
	cfg := &Config{
		ActiveAccount: "work",
		Accounts: map[string]*Account{
			"work": {
				Site:  "work.atlassian.net",
				Email: "user@work.com",
				Token: "work-token",
			},
			"personal": {
				Site:  "personal.atlassian.net",
				Email: "user@personal.com",
				Token: "personal-token",
			},
			"client": {
				Site:  "client.atlassian.net",
				Email: "contractor@client.com",
				Token: "client-token",
			},
		},
	}

	// Save and reload
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all accounts are present
	if len(loadedCfg.Accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(loadedCfg.Accounts))
	}

	// Verify active account
	activeAccount, err := loadedCfg.GetActiveAccount()
	if err != nil {
		t.Fatalf("Failed to get active account: %v", err)
	}

	if activeAccount.Site != "work.atlassian.net" {
		t.Errorf("Expected active account to be work account, got %q", activeAccount.Site)
	}
}
