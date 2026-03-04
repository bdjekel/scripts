// Package parser handles markdown parsing for requirements files.
package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Parser handles markdown parsing operations.
type Parser struct{}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a requirements.md file and returns structured data.
func (p *Parser) Parse(filePath string) (*RequirementsDoc, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ParseError{
				Message: fmt.Sprintf("file not found: %s", filePath),
			}
		}
		return nil, &ParseError{
			Message: fmt.Sprintf("failed to read file: %s: %v", filePath, err),
		}
	}
	return p.ParseContent(string(content))
}

// ParseContent parses markdown content directly.
func (p *Parser) ParseContent(content string) (*RequirementsDoc, error) {
	doc := &RequirementsDoc{
		Epics: []Epic{},
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	var currentEpic *Epic
	var currentStory *Story
	var inAcceptanceCriteria bool
	var storyContentLines []string

	// Regex patterns for field extraction
	priorityRegex := regexp.MustCompile(`^\*\*Priority:\*\*\s*(.+)$`)
	assigneeRegex := regexp.MustCompile(`^\*\*Assignee:\*\*\s*(.+)$`)
	bulletRegex := regexp.MustCompile(`^[-*]\s+(.+)$`)
	acHeadingRegex := regexp.MustCompile(`(?i)^###?\s*Acceptance\s+Criteria\s*$`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for Epic heading (# but not ##)
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "##") {
			// Save current story if exists
			if currentStory != nil && currentEpic != nil {
				currentStory.Description = buildDescription(storyContentLines)
				currentEpic.Stories = append(currentEpic.Stories, *currentStory)
				currentStory = nil
				storyContentLines = nil
			}

			// Save current epic if exists
			if currentEpic != nil {
				doc.Epics = append(doc.Epics, *currentEpic)
			}

			// Start new epic
			summary := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if summary == "" {
				return nil, &ParseError{
					Line:    lineNum,
					Message: "Epic heading cannot be empty",
				}
			}
			currentEpic = &Epic{
				Summary: summary,
				Stories: []Story{},
			}
			inAcceptanceCriteria = false
			continue
		}

		// Check for Story heading (##)
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "###") {
			// Save current story if exists
			if currentStory != nil && currentEpic != nil {
				currentStory.Description = buildDescription(storyContentLines)
				currentEpic.Stories = append(currentEpic.Stories, *currentStory)
				storyContentLines = nil
			}

			if currentEpic == nil {
				return nil, &ParseError{
					Line:    lineNum,
					Message: "Story found without parent Epic",
				}
			}

			// Start new story
			summary := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if summary == "" {
				return nil, &ParseError{
					Line:    lineNum,
					Message: "Story heading cannot be empty",
				}
			}
			currentStory = &Story{
				Summary:            summary,
				AcceptanceCriteria: []string{},
			}
			inAcceptanceCriteria = false
			continue
		}

		// Check for Acceptance Criteria heading
		if acHeadingRegex.MatchString(trimmedLine) {
			inAcceptanceCriteria = true
			continue
		}

		// Check for another ### heading (ends acceptance criteria section)
		if strings.HasPrefix(trimmedLine, "###") && !acHeadingRegex.MatchString(trimmedLine) {
			inAcceptanceCriteria = false
			// Add to story content
			if currentStory != nil {
				storyContentLines = append(storyContentLines, line)
			}
			continue
		}

		// Extract Priority field
		if matches := priorityRegex.FindStringSubmatch(trimmedLine); matches != nil {
			priority := strings.TrimSpace(matches[1])
			if currentStory != nil {
				currentStory.Priority = priority
			} else if currentEpic != nil {
				currentEpic.Priority = priority
			}
			continue
		}

		// Extract Assignee field
		if matches := assigneeRegex.FindStringSubmatch(trimmedLine); matches != nil {
			assignee := strings.TrimSpace(matches[1])
			if currentStory != nil {
				currentStory.Assignee = assignee
			} else if currentEpic != nil {
				currentEpic.Assignee = assignee
			}
			continue
		}

		// Extract acceptance criteria bullet points
		if inAcceptanceCriteria && currentStory != nil {
			if matches := bulletRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentStory.AcceptanceCriteria = append(currentStory.AcceptanceCriteria, strings.TrimSpace(matches[1]))
			}
			continue
		}

		// Add to story content (for description)
		if currentStory != nil {
			storyContentLines = append(storyContentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &ParseError{
			Message: fmt.Sprintf("error reading content: %v", err),
		}
	}

	// Save final story if exists
	if currentStory != nil && currentEpic != nil {
		currentStory.Description = buildDescription(storyContentLines)
		currentEpic.Stories = append(currentEpic.Stories, *currentStory)
	}

	// Save final epic if exists
	if currentEpic != nil {
		doc.Epics = append(doc.Epics, *currentEpic)
	}

	return doc, nil
}

// buildDescription creates a description from content lines.
// It trims leading/trailing empty lines and joins the rest.
func buildDescription(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// Trim leading empty lines
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	// Trim trailing empty lines
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

// Format converts structured data back to markdown for round-trip testing.
func (p *Parser) Format(doc *RequirementsDoc) string {
	if doc == nil {
		return ""
	}

	var sb strings.Builder

	for i, epic := range doc.Epics {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Write Epic heading
		sb.WriteString("# ")
		sb.WriteString(epic.Summary)
		sb.WriteString("\n")

		// Write Epic optional fields
		if epic.Priority != "" {
			sb.WriteString("\n**Priority:** ")
			sb.WriteString(epic.Priority)
			sb.WriteString("\n")
		}
		if epic.Assignee != "" {
			sb.WriteString("\n**Assignee:** ")
			sb.WriteString(epic.Assignee)
			sb.WriteString("\n")
		}

		// Write Stories
		for _, story := range epic.Stories {
			sb.WriteString("\n## ")
			sb.WriteString(story.Summary)
			sb.WriteString("\n")

			// Write Story description
			if story.Description != "" {
				sb.WriteString("\n")
				sb.WriteString(story.Description)
				sb.WriteString("\n")
			}

			// Write Story optional fields
			if story.Priority != "" {
				sb.WriteString("\n**Priority:** ")
				sb.WriteString(story.Priority)
				sb.WriteString("\n")
			}
			if story.Assignee != "" {
				sb.WriteString("\n**Assignee:** ")
				sb.WriteString(story.Assignee)
				sb.WriteString("\n")
			}

			// Write Acceptance Criteria
			if len(story.AcceptanceCriteria) > 0 {
				sb.WriteString("\n### Acceptance Criteria\n\n")
				for _, ac := range story.AcceptanceCriteria {
					sb.WriteString("- ")
					sb.WriteString(ac)
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}
