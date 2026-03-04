// Package generator orchestrates ticket creation from parsed requirements.
package generator

import (
	"fmt"
	"io"
	"os"

	"github.com/user/jira-ticket/internal/jiracli"
	"github.com/user/jira-ticket/internal/parser"
)

// Options configures generator behavior.
type Options struct {
	DryRun      bool   // Parse only, don't create tickets
	Verbose     bool   // Enable detailed logging
	OnDuplicate string // "skip" or "fail"
}

// Result summarizes the generation outcome.
type Result struct {
	Created []string       // Keys of newly created tickets
	Skipped []string       // Keys of skipped duplicates
	Failed  []FailedTicket // Failed creation attempts
}

// FailedTicket records a failed creation attempt.
type FailedTicket struct {
	Summary string // The ticket summary that failed
	Type    string // "Epic" or "Story"
	Error   string // Error message describing the failure
}

// Generator orchestrates ticket creation.
type Generator interface {
	// Generate parses the file and creates tickets.
	Generate(filePath string, opts Options) (*Result, error)
}

// DefaultGenerator implements the Generator interface.
type DefaultGenerator struct {
	parser  *parser.Parser
	adapter jiracli.Adapter
	output  io.Writer
}

// NewGenerator creates a new DefaultGenerator with the given adapter.
func NewGenerator(adapter jiracli.Adapter) *DefaultGenerator {
	return &DefaultGenerator{
		parser:  parser.NewParser(),
		adapter: adapter,
		output:  os.Stdout,
	}
}

// NewGeneratorWithOutput creates a new DefaultGenerator with a custom output writer.
func NewGeneratorWithOutput(adapter jiracli.Adapter, output io.Writer) *DefaultGenerator {
	return &DefaultGenerator{
		parser:  parser.NewParser(),
		adapter: adapter,
		output:  output,
	}
}

// Generate parses the file and creates tickets.
func (g *DefaultGenerator) Generate(filePath string, opts Options) (*Result, error) {
	// Set default OnDuplicate mode
	if opts.OnDuplicate == "" {
		opts.OnDuplicate = "skip"
	}

	// Parse the requirements file
	doc, err := g.parser.Parse(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse requirements file: %w", err)
	}

	result := &Result{
		Created: []string{},
		Skipped: []string{},
		Failed:  []FailedTicket{},
	}

	// In dry-run mode, just return the parsed structure without creating tickets
	if opts.DryRun {
		g.log(opts.Verbose, "Dry-run mode: parsed %d epics\n", len(doc.Epics))
		for _, epic := range doc.Epics {
			g.log(opts.Verbose, "  Epic: %s (%d stories)\n", epic.Summary, len(epic.Stories))
			for _, story := range epic.Stories {
				g.log(opts.Verbose, "    Story: %s\n", story.Summary)
			}
		}
		return result, nil
	}

	// Process each Epic and its Stories
	for _, epic := range doc.Epics {
		epicKey, epicCreated, epicSkipped, epicErr := g.processEpic(epic, opts)

		if epicErr != nil {
			result.Failed = append(result.Failed, FailedTicket{
				Summary: epic.Summary,
				Type:    "Epic",
				Error:   epicErr.Error(),
			})
			g.log(opts.Verbose, "Failed to create Epic '%s': %s\n", epic.Summary, epicErr.Error())
			// Continue processing remaining epics and stories
			// Skip stories for this epic since we don't have a parent key
			continue
		}

		if epicCreated {
			result.Created = append(result.Created, epicKey)
			g.log(opts.Verbose, "Created Epic '%s': %s\n", epic.Summary, epicKey)
		} else if epicSkipped {
			result.Skipped = append(result.Skipped, epicKey)
			g.log(opts.Verbose, "Skipped Epic '%s' (duplicate): %s\n", epic.Summary, epicKey)
		}

		// Process Stories under this Epic
		for _, story := range epic.Stories {
			storyKey, storyCreated, storySkipped, storyErr := g.processStory(story, epicKey, opts)

			if storyErr != nil {
				result.Failed = append(result.Failed, FailedTicket{
					Summary: story.Summary,
					Type:    "Story",
					Error:   storyErr.Error(),
				})
				g.log(opts.Verbose, "Failed to create Story '%s': %s\n", story.Summary, storyErr.Error())
				// Continue processing remaining stories
				continue
			}

			if storyCreated {
				result.Created = append(result.Created, storyKey)
				g.log(opts.Verbose, "Created Story '%s': %s\n", story.Summary, storyKey)
			} else if storySkipped {
				result.Skipped = append(result.Skipped, storyKey)
				g.log(opts.Verbose, "Skipped Story '%s' (duplicate): %s\n", story.Summary, storyKey)
			}
		}
	}

	return result, nil
}

// processEpic handles duplicate checking and creation for an Epic.
// Returns: (key, created, skipped, error)
func (g *DefaultGenerator) processEpic(epic parser.Epic, opts Options) (string, bool, bool, error) {
	// Check for existing Epic
	existing, err := g.adapter.SearchEpic(epic.Summary)
	if err != nil {
		return "", false, false, fmt.Errorf("failed to search for existing Epic: %w", err)
	}

	if existing != nil {
		// Duplicate found
		if opts.OnDuplicate == "skip" {
			return existing.Key, false, true, nil
		}
		// Mode is "fail"
		return "", false, false, fmt.Errorf("duplicate Epic found: %s", existing.Key)
	}

	// Create new Epic
	result, err := g.adapter.CreateEpic(epic.Summary, epic.Priority, epic.Assignee)
	if err != nil {
		return "", false, false, err
	}

	if result.Error != nil {
		return "", false, false, result.Error
	}

	return result.Key, true, false, nil
}

// processStory handles duplicate checking and creation for a Story.
// Returns: (key, created, skipped, error)
func (g *DefaultGenerator) processStory(story parser.Story, epicKey string, opts Options) (string, bool, bool, error) {
	// Check for existing Story under this Epic
	existing, err := g.adapter.SearchStory(story.Summary, epicKey)
	if err != nil {
		return "", false, false, fmt.Errorf("failed to search for existing Story: %w", err)
	}

	if existing != nil {
		// Duplicate found
		if opts.OnDuplicate == "skip" {
			return existing.Key, false, true, nil
		}
		// Mode is "fail"
		return "", false, false, fmt.Errorf("duplicate Story found: %s", existing.Key)
	}

	// Build description including acceptance criteria
	description := buildStoryDescription(story)

	// Create new Story
	result, err := g.adapter.CreateStory(story.Summary, description, epicKey, story.Priority, story.Assignee)
	if err != nil {
		return "", false, false, err
	}

	if result.Error != nil {
		return "", false, false, result.Error
	}

	return result.Key, true, false, nil
}

// buildStoryDescription creates the full description including acceptance criteria.
func buildStoryDescription(story parser.Story) string {
	description := story.Description

	if len(story.AcceptanceCriteria) > 0 {
		if description != "" {
			description += "\n\n"
		}
		description += "h3. Acceptance Criteria\n"
		for _, ac := range story.AcceptanceCriteria {
			description += "* " + ac + "\n"
		}
	}

	return description
}

// log writes a message if verbose mode is enabled.
func (g *DefaultGenerator) log(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(g.output, format, args...)
	}
}
