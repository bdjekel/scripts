// Package jiracli provides a wrapper around the jira-cli subprocess.
package jiracli

// TicketResult represents the outcome of a ticket creation attempt.
type TicketResult struct {
	Key     string // Jira ticket key (e.g., PROJ-123)
	Created bool   // true if newly created, false if existing
	Skipped bool   // true if skipped due to duplicate
	Error   error  // non-nil if creation failed
}

// SearchResult represents a found existing ticket.
type SearchResult struct {
	Key     string // Jira ticket key (e.g., PROJ-123)
	Summary string // The ticket summary/title
	Type    string // Issue type: "Epic" or "Story"
}

// Config holds Jira connection settings.
type Config struct {
	ProjectKey string // JIRA_PROJECT_KEY - the Jira project identifier
	ServerURL  string // JIRA_SERVER_URL - the Jira server URL
}
