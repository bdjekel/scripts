// Package parser handles markdown parsing for requirements files.
package parser

// Epic represents a top-level Jira Epic parsed from a # heading.
type Epic struct {
	Summary  string  // The heading text (without #)
	Priority string  // Optional: extracted from **Priority:** field
	Assignee string  // Optional: extracted from **Assignee:** field
	Stories  []Story // Child stories under this epic
}

// Story represents a Jira Story parsed from a ## heading.
type Story struct {
	Summary            string   // The heading text (without ##)
	Description        string   // Full content between this heading and next
	AcceptanceCriteria []string // Extracted acceptance criteria items
	Priority           string   // Optional: extracted from **Priority:** field
	Assignee           string   // Optional: extracted from **Assignee:** field
}

// RequirementsDoc represents the full parsed document.
type RequirementsDoc struct {
	Epics []Epic
}

// ParseError provides detailed error information for parsing failures.
type ParseError struct {
	Line    int
	Message string
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	if e.Line > 0 {
		return "requirements.md:" + itoa(e.Line) + ": " + e.Message
	}
	return e.Message
}

// itoa converts an integer to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
