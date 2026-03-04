// Package jiracli provides a wrapper around the jira-cli subprocess.
package jiracli

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Adapter defines the interface for Jira CLI operations.
type Adapter interface {
	// CheckInstalled verifies jira-cli is available in PATH.
	CheckInstalled() error

	// CreateEpic creates an Epic and returns the ticket key.
	CreateEpic(summary, priority, assignee string) (*TicketResult, error)

	// CreateStory creates a Story linked to an Epic.
	CreateStory(summary, description, epicKey, priority, assignee string) (*TicketResult, error)

	// SearchEpic finds an existing Epic by summary.
	SearchEpic(summary string) (*SearchResult, error)

	// SearchStory finds an existing Story by summary under an Epic.
	SearchStory(summary, epicKey string) (*SearchResult, error)
}

// CommandExecutor abstracts command execution for testing.
type CommandExecutor interface {
	Run(name string, args ...string) ([]byte, error)
	LookPath(file string) (string, error)
}

// RealExecutor executes actual system commands.
type RealExecutor struct{}

// Run executes a command and returns its combined output.
func (e *RealExecutor) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// Include stderr in error for better diagnostics
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// LookPath searches for an executable in PATH.
func (e *RealExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// CLIAdapter implements the Adapter interface using jira-cli subprocess calls.
type CLIAdapter struct {
	config   Config
	executor CommandExecutor
}

// NewAdapter creates a new CLIAdapter with the given configuration.
func NewAdapter(config Config) *CLIAdapter {
	return &CLIAdapter{
		config:   config,
		executor: &RealExecutor{},
	}
}

// NewAdapterWithExecutor creates a new CLIAdapter with a custom executor (for testing).
func NewAdapterWithExecutor(config Config, executor CommandExecutor) *CLIAdapter {
	return &CLIAdapter{
		config:   config,
		executor: executor,
	}
}

// ticketKeyPattern matches Jira ticket keys like PROJ-123.
var ticketKeyPattern = regexp.MustCompile(`[A-Z]+-\d+`)

// CheckInstalled verifies jira-cli is available in PATH.
func (a *CLIAdapter) CheckInstalled() error {
	_, err := a.executor.LookPath("jira")
	if err != nil {
		return errors.New("jira-cli not found in PATH. Install with: brew install jira-cli")
	}
	return nil
}

// CreateEpic creates an Epic and returns the ticket key.
func (a *CLIAdapter) CreateEpic(summary, priority, assignee string) (*TicketResult, error) {
	args := []string{
		"issue", "create",
		"-t", "Epic",
		"-s", summary,
		"-P", a.config.ProjectKey,
	}

	// Add optional priority flag
	if priority != "" {
		args = append(args, "-y", priority)
	}

	// Add optional assignee flag
	if assignee != "" {
		args = append(args, "-a", assignee)
	}

	output, err := a.executor.Run("jira", args...)
	if err != nil {
		return &TicketResult{
			Error: fmt.Errorf("failed to create Epic '%s': %w", summary, err),
		}, nil
	}

	key := extractTicketKey(string(output))
	if key == "" {
		return &TicketResult{
			Error: fmt.Errorf("failed to extract ticket key from jira-cli output: %s", string(output)),
		}, nil
	}

	return &TicketResult{
		Key:     key,
		Created: true,
	}, nil
}

// CreateStory creates a Story linked to an Epic.
func (a *CLIAdapter) CreateStory(summary, description, epicKey, priority, assignee string) (*TicketResult, error) {
	args := []string{
		"issue", "create",
		"-t", "Story",
		"-s", summary,
		"-P", a.config.ProjectKey,
		"--parent", epicKey,
	}

	// Add description if provided
	if description != "" {
		args = append(args, "-b", description)
	}

	// Add optional priority flag
	if priority != "" {
		args = append(args, "-y", priority)
	}

	// Add optional assignee flag
	if assignee != "" {
		args = append(args, "-a", assignee)
	}

	output, err := a.executor.Run("jira", args...)
	if err != nil {
		return &TicketResult{
			Error: fmt.Errorf("failed to create Story '%s': %w", summary, err),
		}, nil
	}

	key := extractTicketKey(string(output))
	if key == "" {
		return &TicketResult{
			Error: fmt.Errorf("failed to extract ticket key from jira-cli output: %s", string(output)),
		}, nil
	}

	return &TicketResult{
		Key:     key,
		Created: true,
	}, nil
}

// SearchEpic finds an existing Epic by summary.
func (a *CLIAdapter) SearchEpic(summary string) (*SearchResult, error) {
	// Build JQL query to search for Epic by summary
	jql := fmt.Sprintf(`summary ~ "%s"`, escapeJQL(summary))

	args := []string{
		"issue", "list",
		"-t", "Epic",
		"-q", jql,
		"-P", a.config.ProjectKey,
		"--plain",
	}

	output, err := a.executor.Run("jira", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search for Epic '%s': %w", summary, err)
	}

	return parseSearchOutput(string(output), "Epic"), nil
}

// SearchStory finds an existing Story by summary under an Epic.
func (a *CLIAdapter) SearchStory(summary, epicKey string) (*SearchResult, error) {
	// Build JQL query to search for Story by summary and parent Epic
	jql := fmt.Sprintf(`summary ~ "%s" AND parent = %s`, escapeJQL(summary), epicKey)

	args := []string{
		"issue", "list",
		"-t", "Story",
		"-q", jql,
		"-P", a.config.ProjectKey,
		"--plain",
	}

	output, err := a.executor.Run("jira", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search for Story '%s' under Epic '%s': %w", summary, epicKey, err)
	}

	return parseSearchOutput(string(output), "Story"), nil
}

// extractTicketKey extracts the first Jira ticket key from output.
func extractTicketKey(output string) string {
	match := ticketKeyPattern.FindString(output)
	return match
}

// parseSearchOutput parses jira-cli list output and returns the first matching result.
func parseSearchOutput(output string, issueType string) *SearchResult {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	// jira-cli --plain output format is typically:
	// KEY      SUMMARY                    STATUS
	// PROJ-123 Some summary text          To Do
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract ticket key from the line
		key := ticketKeyPattern.FindString(line)
		if key != "" {
			// Extract summary (everything after the key until status columns)
			// This is a simplified extraction - the summary follows the key
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Find where the key ends and extract remaining as summary
				keyIdx := strings.Index(line, key)
				if keyIdx >= 0 {
					afterKey := strings.TrimSpace(line[keyIdx+len(key):])
					// Summary is typically followed by status, but we take what we can
					summary := afterKey
					// Remove trailing status words if present (common statuses)
					for _, status := range []string{"To Do", "In Progress", "Done", "Open", "Closed"} {
						if strings.HasSuffix(summary, status) {
							summary = strings.TrimSpace(strings.TrimSuffix(summary, status))
							break
						}
					}
					return &SearchResult{
						Key:     key,
						Summary: summary,
						Type:    issueType,
					}
				}
			}
		}
	}

	return nil
}

// escapeJQL escapes special characters in JQL string values.
func escapeJQL(s string) string {
	// Escape double quotes and backslashes for JQL
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
