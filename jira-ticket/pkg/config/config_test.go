// Package config tests for environment variable loading and validation.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to clear environment variables and restore them after test
func clearEnvVars(t *testing.T) func() {
	t.Helper()
	origProjectKey := os.Getenv("JIRA_PROJECT_KEY")
	origServerURL := os.Getenv("JIRA_SERVER_URL")

	os.Unsetenv("JIRA_PROJECT_KEY")
	os.Unsetenv("JIRA_SERVER_URL")

	return func() {
		if origProjectKey != "" {
			os.Setenv("JIRA_PROJECT_KEY", origProjectKey)
		} else {
			os.Unsetenv("JIRA_PROJECT_KEY")
		}
		if origServerURL != "" {
			os.Setenv("JIRA_SERVER_URL", origServerURL)
		} else {
			os.Unsetenv("JIRA_SERVER_URL")
		}
	}
}

// Helper to create a temporary .env file
func createEnvFile(t *testing.T, content string) (cleanup func()) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		os.Chdir(origDir)
		t.Fatalf("failed to write .env file: %v", err)
	}

	return func() {
		os.Chdir(origDir)
	}
}

// TestValidate_MissingProjectKey tests that missing JIRA_PROJECT_KEY returns descriptive error.
// Validates: Requirements 4.3
func TestValidate_MissingProjectKey(t *testing.T) {
	cfg := &Config{
		ProjectKey: "",
		ServerURL:  "https://example.atlassian.net",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing JIRA_PROJECT_KEY, got nil")
	}

	errMsg := err.Error()
	if !contains(errMsg, "JIRA_PROJECT_KEY") {
		t.Errorf("error should mention JIRA_PROJECT_KEY, got: %s", errMsg)
	}
	if !contains(errMsg, "not set") {
		t.Errorf("error should indicate variable is not set, got: %s", errMsg)
	}
}

// TestValidate_MissingServerURL tests that missing JIRA_SERVER_URL returns descriptive error.
// Validates: Requirements 4.3
func TestValidate_MissingServerURL(t *testing.T) {
	cfg := &Config{
		ProjectKey: "PROJ",
		ServerURL:  "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing JIRA_SERVER_URL, got nil")
	}

	errMsg := err.Error()
	if !contains(errMsg, "JIRA_SERVER_URL") {
		t.Errorf("error should mention JIRA_SERVER_URL, got: %s", errMsg)
	}
	if !contains(errMsg, "not set") {
		t.Errorf("error should indicate variable is not set, got: %s", errMsg)
	}
}

// TestValidate_MissingBothVariables tests that missing both variables returns descriptive errors.
// Validates: Requirements 4.3
func TestValidate_MissingBothVariables(t *testing.T) {
	cfg := &Config{
		ProjectKey: "",
		ServerURL:  "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing variables, got nil")
	}

	errMsg := err.Error()
	if !contains(errMsg, "JIRA_PROJECT_KEY") {
		t.Errorf("error should mention JIRA_PROJECT_KEY, got: %s", errMsg)
	}
	if !contains(errMsg, "JIRA_SERVER_URL") {
		t.Errorf("error should mention JIRA_SERVER_URL, got: %s", errMsg)
	}
}

// TestValidate_AllVariablesPresent tests that valid config passes validation.
func TestValidate_AllVariablesPresent(t *testing.T) {
	cfg := &Config{
		ProjectKey: "PROJ",
		ServerURL:  "https://example.atlassian.net",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error for valid config, got: %v", err)
	}
}

// TestLoad_FromEnvFile tests that .env file is loaded correctly.
// Validates: Requirements 4.4
func TestLoad_FromEnvFile(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	envContent := `JIRA_PROJECT_KEY=ENVPROJ
JIRA_SERVER_URL=https://env.atlassian.net`

	cleanup := createEnvFile(t, envContent)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.ProjectKey != "ENVPROJ" {
		t.Errorf("expected ProjectKey 'ENVPROJ', got '%s'", cfg.ProjectKey)
	}
	if cfg.ServerURL != "https://env.atlassian.net" {
		t.Errorf("expected ServerURL 'https://env.atlassian.net', got '%s'", cfg.ServerURL)
	}
}

// TestLoad_SystemEnvPrecedence tests that system env vars take precedence over .env file.
// Validates: Requirements 4.5
func TestLoad_SystemEnvPrecedence(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	// Set system environment variables
	os.Setenv("JIRA_PROJECT_KEY", "SYSPROJ")
	os.Setenv("JIRA_SERVER_URL", "https://system.atlassian.net")

	// Create .env file with different values
	envContent := `JIRA_PROJECT_KEY=ENVPROJ
JIRA_SERVER_URL=https://env.atlassian.net`

	cleanup := createEnvFile(t, envContent)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// System env vars should take precedence
	if cfg.ProjectKey != "SYSPROJ" {
		t.Errorf("expected ProjectKey 'SYSPROJ' (system env), got '%s'", cfg.ProjectKey)
	}
	if cfg.ServerURL != "https://system.atlassian.net" {
		t.Errorf("expected ServerURL 'https://system.atlassian.net' (system env), got '%s'", cfg.ServerURL)
	}
}

// TestLoad_PartialSystemEnvPrecedence tests that system env vars take precedence only for set variables.
// Validates: Requirements 4.5
func TestLoad_PartialSystemEnvPrecedence(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	// Set only one system environment variable
	os.Setenv("JIRA_PROJECT_KEY", "SYSPROJ")

	// Create .env file with both values
	envContent := `JIRA_PROJECT_KEY=ENVPROJ
JIRA_SERVER_URL=https://env.atlassian.net`

	cleanup := createEnvFile(t, envContent)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// System env var should take precedence for ProjectKey
	if cfg.ProjectKey != "SYSPROJ" {
		t.Errorf("expected ProjectKey 'SYSPROJ' (system env), got '%s'", cfg.ProjectKey)
	}
	// .env value should be used for ServerURL
	if cfg.ServerURL != "https://env.atlassian.net" {
		t.Errorf("expected ServerURL 'https://env.atlassian.net' (from .env), got '%s'", cfg.ServerURL)
	}
}

// TestLoad_NoEnvFile tests that Load works when no .env file exists.
func TestLoad_NoEnvFile(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	// Set system environment variables
	os.Setenv("JIRA_PROJECT_KEY", "SYSPROJ")
	os.Setenv("JIRA_SERVER_URL", "https://system.atlassian.net")

	// Change to temp dir without .env file
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.ProjectKey != "SYSPROJ" {
		t.Errorf("expected ProjectKey 'SYSPROJ', got '%s'", cfg.ProjectKey)
	}
	if cfg.ServerURL != "https://system.atlassian.net" {
		t.Errorf("expected ServerURL 'https://system.atlassian.net', got '%s'", cfg.ServerURL)
	}
}

// TestLoad_EmptyEnvFile tests that Load handles empty .env file gracefully.
func TestLoad_EmptyEnvFile(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	// Set system environment variables
	os.Setenv("JIRA_PROJECT_KEY", "SYSPROJ")
	os.Setenv("JIRA_SERVER_URL", "https://system.atlassian.net")

	cleanup := createEnvFile(t, "")
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.ProjectKey != "SYSPROJ" {
		t.Errorf("expected ProjectKey 'SYSPROJ', got '%s'", cfg.ProjectKey)
	}
	if cfg.ServerURL != "https://system.atlassian.net" {
		t.Errorf("expected ServerURL 'https://system.atlassian.net', got '%s'", cfg.ServerURL)
	}
}

// TestLoad_NoVariablesSet tests that Load returns empty config when no variables are set.
// Validates: Requirements 4.3 (validation catches this)
func TestLoad_NoVariablesSet(t *testing.T) {
	restore := clearEnvVars(t)
	defer restore()

	// Change to temp dir without .env file
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Load should succeed but return empty values
	if cfg.ProjectKey != "" {
		t.Errorf("expected empty ProjectKey, got '%s'", cfg.ProjectKey)
	}
	if cfg.ServerURL != "" {
		t.Errorf("expected empty ServerURL, got '%s'", cfg.ServerURL)
	}

	// Validate should fail
	err = cfg.Validate()
	if err == nil {
		t.Error("expected Validate() to fail for empty config")
	}
}

// contains is a helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
