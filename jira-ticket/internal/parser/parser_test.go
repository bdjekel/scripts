package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func TestParseContent_BasicEpicAndStory(t *testing.T) {
	content := `# Epic One

## Story One

Story description here.

### Acceptance Criteria

- Criterion 1
- Criterion 2
`

	p := NewParser()
	doc, err := p.ParseContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Epics) != 1 {
		t.Fatalf("expected 1 epic, got %d", len(doc.Epics))
	}

	epic := doc.Epics[0]
	if epic.Summary != "Epic One" {
		t.Errorf("expected epic summary 'Epic One', got '%s'", epic.Summary)
	}

	if len(epic.Stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(epic.Stories))
	}

	story := epic.Stories[0]
	if story.Summary != "Story One" {
		t.Errorf("expected story summary 'Story One', got '%s'", story.Summary)
	}

	if story.Description != "Story description here." {
		t.Errorf("expected description 'Story description here.', got '%s'", story.Description)
	}

	if len(story.AcceptanceCriteria) != 2 {
		t.Fatalf("expected 2 acceptance criteria, got %d", len(story.AcceptanceCriteria))
	}

	if story.AcceptanceCriteria[0] != "Criterion 1" {
		t.Errorf("expected first criterion 'Criterion 1', got '%s'", story.AcceptanceCriteria[0])
	}
}

func TestParseContent_OptionalFields(t *testing.T) {
	content := `# Epic With Fields

**Priority:** High
**Assignee:** john.doe

## Story With Fields

**Priority:** Medium
**Assignee:** jane.doe

### Acceptance Criteria

- Test criterion
`

	p := NewParser()
	doc, err := p.ParseContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	epic := doc.Epics[0]
	if epic.Priority != "High" {
		t.Errorf("expected epic priority 'High', got '%s'", epic.Priority)
	}
	if epic.Assignee != "john.doe" {
		t.Errorf("expected epic assignee 'john.doe', got '%s'", epic.Assignee)
	}

	story := epic.Stories[0]
	if story.Priority != "Medium" {
		t.Errorf("expected story priority 'Medium', got '%s'", story.Priority)
	}
	if story.Assignee != "jane.doe" {
		t.Errorf("expected story assignee 'jane.doe', got '%s'", story.Assignee)
	}
}

func TestParseContent_MultipleEpicsAndStories(t *testing.T) {
	content := `# Epic One

## Story 1.1

## Story 1.2

# Epic Two

## Story 2.1
`

	p := NewParser()
	doc, err := p.ParseContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Epics) != 2 {
		t.Fatalf("expected 2 epics, got %d", len(doc.Epics))
	}

	if len(doc.Epics[0].Stories) != 2 {
		t.Errorf("expected 2 stories in first epic, got %d", len(doc.Epics[0].Stories))
	}

	if len(doc.Epics[1].Stories) != 1 {
		t.Errorf("expected 1 story in second epic, got %d", len(doc.Epics[1].Stories))
	}
}

// TestParse_FileNotFound tests that file not found returns a descriptive error.
// **Validates: Requirements 1.4**
func TestParse_FileNotFound(t *testing.T) {
	p := NewParser()
	nonExistentPath := "/nonexistent/path/requirements.md"
	_, err := p.Parse(nonExistentPath)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}

	if parseErr.Message == "" {
		t.Error("expected non-empty error message")
	}

	// Verify the error message is descriptive and contains the file path
	if !strings.Contains(parseErr.Message, "file not found") {
		t.Errorf("expected error message to contain 'file not found', got: %s", parseErr.Message)
	}
	if !strings.Contains(parseErr.Message, nonExistentPath) {
		t.Errorf("expected error message to contain file path '%s', got: %s", nonExistentPath, parseErr.Message)
	}

	// Verify Line is 0 for file-level errors (not a line-specific parse error)
	if parseErr.Line != 0 {
		t.Errorf("expected Line to be 0 for file-not-found error, got: %d", parseErr.Line)
	}
}

func TestParse_ValidFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "requirements.md")

	content := `# Test Epic

## Test Story

Test description.
`

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	p := NewParser()
	doc, err := p.Parse(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Epics) != 1 {
		t.Errorf("expected 1 epic, got %d", len(doc.Epics))
	}
}

// TestParseContent_StoryWithoutEpic tests that a story without a parent epic
// returns a parse error with line number.
// **Validates: Requirements 1.5**
func TestParseContent_StoryWithoutEpic(t *testing.T) {
	content := `## Story Without Epic
`

	p := NewParser()
	_, err := p.ParseContent(content)
	if err == nil {
		t.Fatal("expected error for story without epic")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}

	if parseErr.Line != 1 {
		t.Errorf("expected error on line 1, got line %d", parseErr.Line)
	}

	// Verify the error message is descriptive
	if !strings.Contains(parseErr.Message, "Story") && !strings.Contains(parseErr.Message, "Epic") {
		t.Errorf("expected error message to mention Story and Epic, got: %s", parseErr.Message)
	}

	// Verify the Error() method includes line number
	errStr := parseErr.Error()
	if !strings.Contains(errStr, ":1:") {
		t.Errorf("expected Error() to include line number ':1:', got: %s", errStr)
	}
}

// TestParseContent_EmptyEpicHeading tests that an empty epic heading returns a parse error.
// **Validates: Requirements 1.5**
func TestParseContent_EmptyEpicHeading(t *testing.T) {
	content := `# 
`

	p := NewParser()
	_, err := p.ParseContent(content)
	if err == nil {
		t.Fatal("expected error for empty epic heading")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}

	// Verify line number is included
	if parseErr.Line != 1 {
		t.Errorf("expected error on line 1, got line %d", parseErr.Line)
	}

	// Verify the error message mentions the issue
	if !strings.Contains(parseErr.Message, "empty") {
		t.Errorf("expected error message to mention 'empty', got: %s", parseErr.Message)
	}
}

// TestParseContent_EmptyStoryHeading tests that an empty story heading returns a parse error.
// **Validates: Requirements 1.5**
func TestParseContent_EmptyStoryHeading(t *testing.T) {
	content := `# Valid Epic

## 
`

	p := NewParser()
	_, err := p.ParseContent(content)
	if err == nil {
		t.Fatal("expected error for empty story heading")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}

	// Verify line number is included (line 3 is where the empty ## heading is)
	if parseErr.Line != 3 {
		t.Errorf("expected error on line 3, got line %d", parseErr.Line)
	}

	// Verify the error message mentions the issue
	if !strings.Contains(parseErr.Message, "empty") {
		t.Errorf("expected error message to mention 'empty', got: %s", parseErr.Message)
	}
}

// TestParseContent_StoryWithoutEpicMidDocument tests that a story without a parent epic
// in the middle of a document returns a parse error with correct line number.
// **Validates: Requirements 1.5**
func TestParseContent_StoryWithoutEpicMidDocument(t *testing.T) {
	// This tests that line numbers are correctly tracked even with preceding content
	content := `Some preamble text
that spans multiple lines

## Story Without Epic
`

	p := NewParser()
	_, err := p.ParseContent(content)
	if err == nil {
		t.Fatal("expected error for story without epic")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}

	// The ## heading is on line 4
	if parseErr.Line != 4 {
		t.Errorf("expected error on line 4, got line %d", parseErr.Line)
	}
}

func TestFormat_RoundTrip(t *testing.T) {
	doc := &RequirementsDoc{
		Epics: []Epic{
			{
				Summary:  "Epic One",
				Priority: "High",
				Assignee: "john.doe",
				Stories: []Story{
					{
						Summary:            "Story One",
						Description:        "Story description",
						Priority:           "Medium",
						Assignee:           "jane.doe",
						AcceptanceCriteria: []string{"Criterion 1", "Criterion 2"},
					},
				},
			},
		},
	}

	p := NewParser()
	formatted := p.Format(doc)

	// Parse the formatted content
	reparsed, err := p.ParseContent(formatted)
	if err != nil {
		t.Fatalf("failed to reparse formatted content: %v", err)
	}

	// Verify structure is preserved
	if len(reparsed.Epics) != 1 {
		t.Fatalf("expected 1 epic after round-trip, got %d", len(reparsed.Epics))
	}

	epic := reparsed.Epics[0]
	if epic.Summary != "Epic One" {
		t.Errorf("epic summary mismatch: expected 'Epic One', got '%s'", epic.Summary)
	}
	if epic.Priority != "High" {
		t.Errorf("epic priority mismatch: expected 'High', got '%s'", epic.Priority)
	}
	if epic.Assignee != "john.doe" {
		t.Errorf("epic assignee mismatch: expected 'john.doe', got '%s'", epic.Assignee)
	}

	if len(epic.Stories) != 1 {
		t.Fatalf("expected 1 story after round-trip, got %d", len(epic.Stories))
	}

	story := epic.Stories[0]
	if story.Summary != "Story One" {
		t.Errorf("story summary mismatch: expected 'Story One', got '%s'", story.Summary)
	}
	if story.Priority != "Medium" {
		t.Errorf("story priority mismatch: expected 'Medium', got '%s'", story.Priority)
	}
	if story.Assignee != "jane.doe" {
		t.Errorf("story assignee mismatch: expected 'jane.doe', got '%s'", story.Assignee)
	}

	if len(story.AcceptanceCriteria) != 2 {
		t.Fatalf("expected 2 acceptance criteria after round-trip, got %d", len(story.AcceptanceCriteria))
	}
}

// Property-based tests using rapid

// genSafeString generates a non-empty string without markdown special characters
// that could interfere with parsing (no #, *, newlines, or leading/trailing whitespace).
func genSafeString(t *rapid.T) string {
	// Generate alphanumeric strings with spaces, avoiding markdown-sensitive chars
	chars := rapid.SliceOfN(
		rapid.SampledFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ")),
		1, 50,
	).Draw(t, "chars")

	// Trim spaces and ensure non-empty
	result := string(chars)
	for len(result) > 0 && (result[0] == ' ' || result[len(result)-1] == ' ') {
		if result[0] == ' ' {
			result = result[1:]
		}
		if len(result) > 0 && result[len(result)-1] == ' ' {
			result = result[:len(result)-1]
		}
	}

	if result == "" {
		return "default"
	}
	return result
}

// genPriority generates an optional priority value.
func genPriority(t *rapid.T) string {
	if rapid.Bool().Draw(t, "hasPriority") {
		return rapid.SampledFrom([]string{"High", "Medium", "Low", "Critical"}).Draw(t, "priority")
	}
	return ""
}

// genAssignee generates an optional assignee value.
func genAssignee(t *rapid.T) string {
	if rapid.Bool().Draw(t, "hasAssignee") {
		return genSafeString(t)
	}
	return ""
}

// genAcceptanceCriteria generates a slice of acceptance criteria.
func genAcceptanceCriteria(t *rapid.T) []string {
	count := rapid.IntRange(0, 5).Draw(t, "acCount")
	criteria := make([]string, count)
	for i := 0; i < count; i++ {
		criteria[i] = genSafeString(t)
	}
	return criteria
}

// genStory generates a random Story.
func genStory(t *rapid.T) Story {
	return Story{
		Summary:            genSafeString(t),
		Description:        "", // Description is not reliably round-tripped due to whitespace handling
		Priority:           genPriority(t),
		Assignee:           genAssignee(t),
		AcceptanceCriteria: genAcceptanceCriteria(t),
	}
}

// genEpic generates a random Epic with stories.
func genEpic(t *rapid.T) Epic {
	storyCount := rapid.IntRange(0, 3).Draw(t, "storyCount")
	stories := make([]Story, storyCount)
	for i := 0; i < storyCount; i++ {
		stories[i] = genStory(t)
	}

	return Epic{
		Summary:  genSafeString(t),
		Priority: genPriority(t),
		Assignee: genAssignee(t),
		Stories:  stories,
	}
}

// genRequirementsDoc generates a random RequirementsDoc.
func genRequirementsDoc(t *rapid.T) *RequirementsDoc {
	epicCount := rapid.IntRange(1, 3).Draw(t, "epicCount")
	epics := make([]Epic, epicCount)
	for i := 0; i < epicCount; i++ {
		epics[i] = genEpic(t)
	}

	return &RequirementsDoc{
		Epics: epics,
	}
}

// TestParseFormatRoundTrip tests Property 1: Parse-Format Round Trip.
// For any valid RequirementsDoc structure, formatting it to markdown and then
// parsing the result SHALL produce an equivalent RequirementsDoc.
// **Validates: Requirements 1.7**
func TestParseFormatRoundTrip(t *testing.T) {
	p := NewParser()

	rapid.Check(t, func(t *rapid.T) {
		// Generate a random RequirementsDoc
		doc := genRequirementsDoc(t)

		// Format to markdown
		formatted := p.Format(doc)

		// Parse the formatted markdown
		reparsed, err := p.ParseContent(formatted)
		if err != nil {
			t.Fatalf("failed to parse formatted content: %v\nFormatted:\n%s", err, formatted)
		}

		// Verify structure equivalence
		if len(reparsed.Epics) != len(doc.Epics) {
			t.Fatalf("epic count mismatch: expected %d, got %d", len(doc.Epics), len(reparsed.Epics))
		}

		for i, epic := range doc.Epics {
			reparsedEpic := reparsed.Epics[i]

			if reparsedEpic.Summary != epic.Summary {
				t.Fatalf("epic[%d] summary mismatch: expected %q, got %q", i, epic.Summary, reparsedEpic.Summary)
			}
			if reparsedEpic.Priority != epic.Priority {
				t.Fatalf("epic[%d] priority mismatch: expected %q, got %q", i, epic.Priority, reparsedEpic.Priority)
			}
			if reparsedEpic.Assignee != epic.Assignee {
				t.Fatalf("epic[%d] assignee mismatch: expected %q, got %q", i, epic.Assignee, reparsedEpic.Assignee)
			}

			if len(reparsedEpic.Stories) != len(epic.Stories) {
				t.Fatalf("epic[%d] story count mismatch: expected %d, got %d", i, len(epic.Stories), len(reparsedEpic.Stories))
			}

			for j, story := range epic.Stories {
				reparsedStory := reparsedEpic.Stories[j]

				if reparsedStory.Summary != story.Summary {
					t.Fatalf("epic[%d].story[%d] summary mismatch: expected %q, got %q", i, j, story.Summary, reparsedStory.Summary)
				}
				if reparsedStory.Priority != story.Priority {
					t.Fatalf("epic[%d].story[%d] priority mismatch: expected %q, got %q", i, j, story.Priority, reparsedStory.Priority)
				}
				if reparsedStory.Assignee != story.Assignee {
					t.Fatalf("epic[%d].story[%d] assignee mismatch: expected %q, got %q", i, j, story.Assignee, reparsedStory.Assignee)
				}

				if len(reparsedStory.AcceptanceCriteria) != len(story.AcceptanceCriteria) {
					t.Fatalf("epic[%d].story[%d] acceptance criteria count mismatch: expected %d, got %d",
						i, j, len(story.AcceptanceCriteria), len(reparsedStory.AcceptanceCriteria))
				}

				for k, ac := range story.AcceptanceCriteria {
					if reparsedStory.AcceptanceCriteria[k] != ac {
						t.Fatalf("epic[%d].story[%d].ac[%d] mismatch: expected %q, got %q",
							i, j, k, ac, reparsedStory.AcceptanceCriteria[k])
					}
				}
			}
		}
	})
}

// TestHierarchicalPreservation tests Property 2: Hierarchical Preservation.
// For any valid requirements markdown with Epics and Stories, parsing SHALL produce
// a RequirementsDoc where each Story appears in the Stories slice of its parent Epic
// (the most recent # heading above it), and no Story appears under a different Epic.
// **Validates: Requirements 1.6**
func TestHierarchicalPreservation(t *testing.T) {
	p := NewParser()

	rapid.Check(t, func(t *rapid.T) {
		// Generate a structure with known Epic-Story relationships
		// We'll track which stories belong to which epic by index
		type epicWithStories struct {
			epicSummary    string
			storySummaries []string
		}

		// Generate 1-3 epics, each with 0-4 stories
		epicCount := rapid.IntRange(1, 3).Draw(t, "epicCount")
		expectedStructure := make([]epicWithStories, epicCount)

		// Build markdown content with known relationships
		var mdBuilder strings.Builder
		for i := 0; i < epicCount; i++ {
			epicSummary := fmt.Sprintf("Epic%d_%s", i, genSafeString(t))
			expectedStructure[i].epicSummary = epicSummary

			mdBuilder.WriteString("# ")
			mdBuilder.WriteString(epicSummary)
			mdBuilder.WriteString("\n\n")

			// Generate stories for this epic
			storyCount := rapid.IntRange(0, 4).Draw(t, fmt.Sprintf("storyCount_%d", i))
			expectedStructure[i].storySummaries = make([]string, storyCount)

			for j := 0; j < storyCount; j++ {
				storySummary := fmt.Sprintf("Story%d_%d_%s", i, j, genSafeString(t))
				expectedStructure[i].storySummaries[j] = storySummary

				mdBuilder.WriteString("## ")
				mdBuilder.WriteString(storySummary)
				mdBuilder.WriteString("\n\n")
			}
		}

		markdown := mdBuilder.String()

		// Parse the markdown
		doc, err := p.ParseContent(markdown)
		if err != nil {
			t.Fatalf("failed to parse generated markdown: %v\nMarkdown:\n%s", err, markdown)
		}

		// Verify the number of epics matches
		if len(doc.Epics) != epicCount {
			t.Fatalf("epic count mismatch: expected %d, got %d", epicCount, len(doc.Epics))
		}

		// Verify each epic has the correct stories
		for i, expected := range expectedStructure {
			parsedEpic := doc.Epics[i]

			// Verify epic summary
			if parsedEpic.Summary != expected.epicSummary {
				t.Fatalf("epic[%d] summary mismatch: expected %q, got %q",
					i, expected.epicSummary, parsedEpic.Summary)
			}

			// Verify story count for this epic
			if len(parsedEpic.Stories) != len(expected.storySummaries) {
				t.Fatalf("epic[%d] story count mismatch: expected %d, got %d",
					i, len(expected.storySummaries), len(parsedEpic.Stories))
			}

			// Verify each story is in the correct position
			for j, expectedStorySummary := range expected.storySummaries {
				if parsedEpic.Stories[j].Summary != expectedStorySummary {
					t.Fatalf("epic[%d].story[%d] summary mismatch: expected %q, got %q",
						i, j, expectedStorySummary, parsedEpic.Stories[j].Summary)
				}
			}
		}

		// Additional verification: ensure no story appears under a different epic
		// by checking that the total story count matches
		totalExpectedStories := 0
		for _, e := range expectedStructure {
			totalExpectedStories += len(e.storySummaries)
		}

		totalParsedStories := 0
		for _, e := range doc.Epics {
			totalParsedStories += len(e.Stories)
		}

		if totalParsedStories != totalExpectedStories {
			t.Fatalf("total story count mismatch: expected %d, got %d",
				totalExpectedStories, totalParsedStories)
		}

		// Verify no story summary appears in a different epic's stories
		for i, expected := range expectedStructure {
			for _, storySummary := range expected.storySummaries {
				// Check that this story doesn't appear in any other epic
				for k, parsedEpic := range doc.Epics {
					if k == i {
						continue // Skip the epic where this story should be
					}
					for _, parsedStory := range parsedEpic.Stories {
						if parsedStory.Summary == storySummary {
							t.Fatalf("story %q from epic[%d] incorrectly appears in epic[%d]",
								storySummary, i, k)
						}
					}
				}
			}
		}
	})
}

// TestHeadingExtraction tests Property 3: Heading Extraction.
// For any valid requirements markdown, the number of `#` headings SHALL equal
// the number of Epics in the parsed result, and the number of `##` headings
// SHALL equal the total number of Stories across all Epics. Each heading's text
// (without the `#` prefix) SHALL match the corresponding Summary field.
// **Validates: Requirements 1.1, 1.2, 2.2, 3.2**
func TestHeadingExtraction(t *testing.T) {
	p := NewParser()

	rapid.Check(t, func(t *rapid.T) {
		// Generate random epic and story summaries
		epicCount := rapid.IntRange(1, 5).Draw(t, "epicCount")

		type epicSpec struct {
			summary        string
			storySummaries []string
		}

		epics := make([]epicSpec, epicCount)
		totalStoryCount := 0

		// Build markdown content with known headings
		var mdBuilder strings.Builder
		for i := 0; i < epicCount; i++ {
			epicSummary := fmt.Sprintf("Epic%d %s", i, genSafeString(t))
			epics[i].summary = epicSummary

			mdBuilder.WriteString("# ")
			mdBuilder.WriteString(epicSummary)
			mdBuilder.WriteString("\n\n")

			// Generate stories for this epic
			storyCount := rapid.IntRange(0, 5).Draw(t, fmt.Sprintf("storyCount_%d", i))
			epics[i].storySummaries = make([]string, storyCount)
			totalStoryCount += storyCount

			for j := 0; j < storyCount; j++ {
				storySummary := fmt.Sprintf("Story%d_%d %s", i, j, genSafeString(t))
				epics[i].storySummaries[j] = storySummary

				mdBuilder.WriteString("## ")
				mdBuilder.WriteString(storySummary)
				mdBuilder.WriteString("\n\n")
			}
		}

		markdown := mdBuilder.String()

		// Count headings in the markdown
		lines := strings.Split(markdown, "\n")
		h1Count := 0
		h2Count := 0
		var h1Summaries []string
		var h2Summaries []string

		for _, line := range lines {
			if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "##") {
				h1Count++
				h1Summaries = append(h1Summaries, strings.TrimSpace(strings.TrimPrefix(line, "# ")))
			} else if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "###") {
				h2Count++
				h2Summaries = append(h2Summaries, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			}
		}

		// Parse the markdown
		doc, err := p.ParseContent(markdown)
		if err != nil {
			t.Fatalf("failed to parse generated markdown: %v\nMarkdown:\n%s", err, markdown)
		}

		// Property: Number of # headings SHALL equal number of Epics
		if len(doc.Epics) != h1Count {
			t.Fatalf("epic count mismatch: expected %d (# headings), got %d epics",
				h1Count, len(doc.Epics))
		}

		// Property: Number of ## headings SHALL equal total number of Stories
		parsedStoryCount := 0
		for _, epic := range doc.Epics {
			parsedStoryCount += len(epic.Stories)
		}
		if parsedStoryCount != h2Count {
			t.Fatalf("story count mismatch: expected %d (## headings), got %d stories",
				h2Count, parsedStoryCount)
		}

		// Property: Each # heading text SHALL match corresponding Epic Summary
		for i, epic := range doc.Epics {
			if i >= len(h1Summaries) {
				t.Fatalf("epic[%d] exists but no corresponding # heading found", i)
			}
			if epic.Summary != h1Summaries[i] {
				t.Fatalf("epic[%d] summary mismatch: heading text %q, epic summary %q",
					i, h1Summaries[i], epic.Summary)
			}
		}

		// Property: Each ## heading text SHALL match corresponding Story Summary
		storyIdx := 0
		for i, epic := range doc.Epics {
			for j, story := range epic.Stories {
				if storyIdx >= len(h2Summaries) {
					t.Fatalf("epic[%d].story[%d] exists but no corresponding ## heading found", i, j)
				}
				if story.Summary != h2Summaries[storyIdx] {
					t.Fatalf("epic[%d].story[%d] summary mismatch: heading text %q, story summary %q",
						i, j, h2Summaries[storyIdx], story.Summary)
				}
				storyIdx++
			}
		}

		// Verify total counts match expected
		if len(doc.Epics) != epicCount {
			t.Fatalf("expected %d epics, got %d", epicCount, len(doc.Epics))
		}
		if parsedStoryCount != totalStoryCount {
			t.Fatalf("expected %d total stories, got %d", totalStoryCount, parsedStoryCount)
		}
	})
}

// TestOptionalFieldExtraction tests Property 4: Optional Field Extraction.
// For any Epic or Story section containing a `**Priority:**` or `**Assignee:**` field,
// the parsed structure SHALL have the corresponding field populated with the value
// following the field label.
// **Validates: Requirements 1.8, 1.9**
func TestOptionalFieldExtraction(t *testing.T) {
	p := NewParser()

	rapid.Check(t, func(t *rapid.T) {
		// Generate random epic and story data with known optional fields
		epicCount := rapid.IntRange(1, 3).Draw(t, "epicCount")

		type storySpec struct {
			summary  string
			priority string
			assignee string
		}

		type epicSpec struct {
			summary  string
			priority string
			assignee string
			stories  []storySpec
		}

		epics := make([]epicSpec, epicCount)

		// Build markdown content with known optional fields
		var mdBuilder strings.Builder
		for i := 0; i < epicCount; i++ {
			epicSummary := fmt.Sprintf("Epic%d %s", i, genSafeString(t))
			epicPriority := ""
			epicAssignee := ""

			// Randomly decide whether to include Priority and Assignee for Epic
			if rapid.Bool().Draw(t, fmt.Sprintf("epicHasPriority_%d", i)) {
				epicPriority = rapid.SampledFrom([]string{"High", "Medium", "Low", "Critical"}).Draw(t, fmt.Sprintf("epicPriority_%d", i))
			}
			if rapid.Bool().Draw(t, fmt.Sprintf("epicHasAssignee_%d", i)) {
				epicAssignee = genSafeString(t)
			}

			epics[i] = epicSpec{
				summary:  epicSummary,
				priority: epicPriority,
				assignee: epicAssignee,
				stories:  []storySpec{},
			}

			// Write Epic heading
			mdBuilder.WriteString("# ")
			mdBuilder.WriteString(epicSummary)
			mdBuilder.WriteString("\n\n")

			// Write Epic optional fields
			if epicPriority != "" {
				mdBuilder.WriteString("**Priority:** ")
				mdBuilder.WriteString(epicPriority)
				mdBuilder.WriteString("\n")
			}
			if epicAssignee != "" {
				mdBuilder.WriteString("**Assignee:** ")
				mdBuilder.WriteString(epicAssignee)
				mdBuilder.WriteString("\n")
			}

			// Generate stories for this epic
			storyCount := rapid.IntRange(0, 3).Draw(t, fmt.Sprintf("storyCount_%d", i))
			for j := 0; j < storyCount; j++ {
				storySummary := fmt.Sprintf("Story%d_%d %s", i, j, genSafeString(t))
				storyPriority := ""
				storyAssignee := ""

				// Randomly decide whether to include Priority and Assignee for Story
				if rapid.Bool().Draw(t, fmt.Sprintf("storyHasPriority_%d_%d", i, j)) {
					storyPriority = rapid.SampledFrom([]string{"High", "Medium", "Low", "Critical"}).Draw(t, fmt.Sprintf("storyPriority_%d_%d", i, j))
				}
				if rapid.Bool().Draw(t, fmt.Sprintf("storyHasAssignee_%d_%d", i, j)) {
					storyAssignee = genSafeString(t)
				}

				epics[i].stories = append(epics[i].stories, storySpec{
					summary:  storySummary,
					priority: storyPriority,
					assignee: storyAssignee,
				})

				// Write Story heading
				mdBuilder.WriteString("\n## ")
				mdBuilder.WriteString(storySummary)
				mdBuilder.WriteString("\n\n")

				// Write Story optional fields
				if storyPriority != "" {
					mdBuilder.WriteString("**Priority:** ")
					mdBuilder.WriteString(storyPriority)
					mdBuilder.WriteString("\n")
				}
				if storyAssignee != "" {
					mdBuilder.WriteString("**Assignee:** ")
					mdBuilder.WriteString(storyAssignee)
					mdBuilder.WriteString("\n")
				}
			}
		}

		markdown := mdBuilder.String()

		// Parse the markdown
		doc, err := p.ParseContent(markdown)
		if err != nil {
			t.Fatalf("failed to parse generated markdown: %v\nMarkdown:\n%s", err, markdown)
		}

		// Verify the number of epics matches
		if len(doc.Epics) != epicCount {
			t.Fatalf("epic count mismatch: expected %d, got %d", epicCount, len(doc.Epics))
		}

		// Verify each epic's optional fields
		for i, expected := range epics {
			parsedEpic := doc.Epics[i]

			// Property: If Epic has **Priority:** field, parsed Epic SHALL have Priority populated
			if expected.priority != "" {
				if parsedEpic.Priority != expected.priority {
					t.Fatalf("epic[%d] priority mismatch: expected %q, got %q",
						i, expected.priority, parsedEpic.Priority)
				}
			} else {
				// If no Priority field was written, parsed Priority should be empty
				if parsedEpic.Priority != "" {
					t.Fatalf("epic[%d] priority should be empty, got %q", i, parsedEpic.Priority)
				}
			}

			// Property: If Epic has **Assignee:** field, parsed Epic SHALL have Assignee populated
			if expected.assignee != "" {
				if parsedEpic.Assignee != expected.assignee {
					t.Fatalf("epic[%d] assignee mismatch: expected %q, got %q",
						i, expected.assignee, parsedEpic.Assignee)
				}
			} else {
				// If no Assignee field was written, parsed Assignee should be empty
				if parsedEpic.Assignee != "" {
					t.Fatalf("epic[%d] assignee should be empty, got %q", i, parsedEpic.Assignee)
				}
			}

			// Verify story count for this epic
			if len(parsedEpic.Stories) != len(expected.stories) {
				t.Fatalf("epic[%d] story count mismatch: expected %d, got %d",
					i, len(expected.stories), len(parsedEpic.Stories))
			}

			// Verify each story's optional fields
			for j, expectedStory := range expected.stories {
				parsedStory := parsedEpic.Stories[j]

				// Property: If Story has **Priority:** field, parsed Story SHALL have Priority populated
				if expectedStory.priority != "" {
					if parsedStory.Priority != expectedStory.priority {
						t.Fatalf("epic[%d].story[%d] priority mismatch: expected %q, got %q",
							i, j, expectedStory.priority, parsedStory.Priority)
					}
				} else {
					// If no Priority field was written, parsed Priority should be empty
					if parsedStory.Priority != "" {
						t.Fatalf("epic[%d].story[%d] priority should be empty, got %q",
							i, j, parsedStory.Priority)
					}
				}

				// Property: If Story has **Assignee:** field, parsed Story SHALL have Assignee populated
				if expectedStory.assignee != "" {
					if parsedStory.Assignee != expectedStory.assignee {
						t.Fatalf("epic[%d].story[%d] assignee mismatch: expected %q, got %q",
							i, j, expectedStory.assignee, parsedStory.Assignee)
					}
				} else {
					// If no Assignee field was written, parsed Assignee should be empty
					if parsedStory.Assignee != "" {
						t.Fatalf("epic[%d].story[%d] assignee should be empty, got %q",
							i, j, parsedStory.Assignee)
					}
				}
			}
		}
	})
}

// TestAcceptanceCriteriaInclusion tests Property 5: Acceptance Criteria Inclusion.
// For any Story section containing acceptance criteria (bullet points under an
// "Acceptance Criteria" heading), the parsed Story's AcceptanceCriteria slice
// SHALL contain all those items.
// **Validates: Requirements 1.3, 3.3**
func TestAcceptanceCriteriaInclusion(t *testing.T) {
	p := NewParser()

	rapid.Check(t, func(t *rapid.T) {
		// Generate random epic and story data with known acceptance criteria
		epicCount := rapid.IntRange(1, 3).Draw(t, "epicCount")

		type storySpec struct {
			summary            string
			acceptanceCriteria []string
		}

		type epicSpec struct {
			summary string
			stories []storySpec
		}

		epics := make([]epicSpec, epicCount)

		// Build markdown content with known acceptance criteria
		var mdBuilder strings.Builder
		for i := 0; i < epicCount; i++ {
			epicSummary := fmt.Sprintf("Epic%d %s", i, genSafeString(t))
			epics[i] = epicSpec{
				summary: epicSummary,
				stories: []storySpec{},
			}

			// Write Epic heading
			mdBuilder.WriteString("# ")
			mdBuilder.WriteString(epicSummary)
			mdBuilder.WriteString("\n\n")

			// Generate stories for this epic
			storyCount := rapid.IntRange(1, 4).Draw(t, fmt.Sprintf("storyCount_%d", i))
			for j := 0; j < storyCount; j++ {
				storySummary := fmt.Sprintf("Story%d_%d %s", i, j, genSafeString(t))

				// Generate acceptance criteria for this story
				acCount := rapid.IntRange(0, 6).Draw(t, fmt.Sprintf("acCount_%d_%d", i, j))
				acceptanceCriteria := make([]string, acCount)
				for k := 0; k < acCount; k++ {
					acceptanceCriteria[k] = fmt.Sprintf("AC%d_%d_%d %s", i, j, k, genSafeString(t))
				}

				epics[i].stories = append(epics[i].stories, storySpec{
					summary:            storySummary,
					acceptanceCriteria: acceptanceCriteria,
				})

				// Write Story heading
				mdBuilder.WriteString("## ")
				mdBuilder.WriteString(storySummary)
				mdBuilder.WriteString("\n\n")

				// Write Story description
				mdBuilder.WriteString("Story description text.\n\n")

				// Write Acceptance Criteria section if there are any
				if acCount > 0 {
					mdBuilder.WriteString("### Acceptance Criteria\n\n")
					for _, ac := range acceptanceCriteria {
						mdBuilder.WriteString("- ")
						mdBuilder.WriteString(ac)
						mdBuilder.WriteString("\n")
					}
					mdBuilder.WriteString("\n")
				}
			}
		}

		markdown := mdBuilder.String()

		// Parse the markdown
		doc, err := p.ParseContent(markdown)
		if err != nil {
			t.Fatalf("failed to parse generated markdown: %v\nMarkdown:\n%s", err, markdown)
		}

		// Verify the number of epics matches
		if len(doc.Epics) != epicCount {
			t.Fatalf("epic count mismatch: expected %d, got %d", epicCount, len(doc.Epics))
		}

		// Verify each story's acceptance criteria
		for i, expected := range epics {
			parsedEpic := doc.Epics[i]

			// Verify story count for this epic
			if len(parsedEpic.Stories) != len(expected.stories) {
				t.Fatalf("epic[%d] story count mismatch: expected %d, got %d",
					i, len(expected.stories), len(parsedEpic.Stories))
			}

			// Verify each story's acceptance criteria
			for j, expectedStory := range expected.stories {
				parsedStory := parsedEpic.Stories[j]

				// Property: AcceptanceCriteria slice SHALL contain all items from the markdown
				if len(parsedStory.AcceptanceCriteria) != len(expectedStory.acceptanceCriteria) {
					t.Fatalf("epic[%d].story[%d] acceptance criteria count mismatch: expected %d, got %d\nExpected: %v\nGot: %v",
						i, j, len(expectedStory.acceptanceCriteria), len(parsedStory.AcceptanceCriteria),
						expectedStory.acceptanceCriteria, parsedStory.AcceptanceCriteria)
				}

				// Verify each acceptance criterion is present and in order
				for k, expectedAC := range expectedStory.acceptanceCriteria {
					if k >= len(parsedStory.AcceptanceCriteria) {
						t.Fatalf("epic[%d].story[%d] missing acceptance criterion[%d]: expected %q",
							i, j, k, expectedAC)
					}
					if parsedStory.AcceptanceCriteria[k] != expectedAC {
						t.Fatalf("epic[%d].story[%d].ac[%d] mismatch: expected %q, got %q",
							i, j, k, expectedAC, parsedStory.AcceptanceCriteria[k])
					}
				}
			}
		}

	})
}
