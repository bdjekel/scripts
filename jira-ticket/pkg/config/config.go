// Package config handles environment variable loading and validation.
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the Jira Ticket Generator.
type Config struct {
	// ProjectKey is the Jira project identifier (e.g., "PROJ")
	ProjectKey string
	// ServerURL is the Jira server URL (e.g., "https://company.atlassian.net")
	ServerURL string
}

// Load reads configuration from environment variables and an optional .env file.
// System environment variables take precedence over values in the .env file.
func Load() (*Config, error) {
	// Capture existing system env vars before loading .env
	existingProjectKey := os.Getenv("JIRA_PROJECT_KEY")
	existingServerURL := os.Getenv("JIRA_SERVER_URL")

	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Use system env vars if they were set, otherwise use values from .env
	projectKey := existingProjectKey
	if projectKey == "" {
		projectKey = os.Getenv("JIRA_PROJECT_KEY")
	}

	serverURL := existingServerURL
	if serverURL == "" {
		serverURL = os.Getenv("JIRA_SERVER_URL")
	}

	cfg := &Config{
		ProjectKey: projectKey,
		ServerURL:  serverURL,
	}

	return cfg, nil
}

// Validate checks that all required configuration values are present.
// Returns an error describing which required variables are missing.
func (c *Config) Validate() error {
	var errs []error

	if c.ProjectKey == "" {
		errs = append(errs, errors.New("JIRA_PROJECT_KEY environment variable is not set. Set it or add to .env file"))
	}

	if c.ServerURL == "" {
		errs = append(errs, errors.New("JIRA_SERVER_URL environment variable is not set. Set it or add to .env file"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration error: %w", errors.Join(errs...))
	}

	return nil
}
