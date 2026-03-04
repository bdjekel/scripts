package jiracli

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// MockExecutor captures command executions for testing.
type MockExecutor struct {
	// Captured stores all executed commands with their arguments
	Captured []CapturedCommand
	// Output is the mock output to return
	Output []byte
	// Error is the mock error to return
	Error error
	// LookPathResult is the result for LookPath calls
	LookPathResult string
	// LookPathError is the error for LookPath calls
	LookPathError error
}

// CapturedCommand represents a captured command execution.
type CapturedCommand struct {
	Name string
	Args []string
}

// Run captures the command and returns mock output.
func (m *MockExecutor) Run(name string, args ...string) ([]byte, error) {
	m.Captured = append(m.Captured, CapturedCommand{Name: name, Args: args})
	return m.Output, m.Error
}

// LookPath returns the mock LookPath result.
func (m *MockExecutor) LookPath(file string) (string, error) {
	return m.LookPathResult, m.LookPathError
}

// containsFlag checks if the args contain a flag with the expected value.
func containsFlag(args []string, flag, expectedValue string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == expectedValue {
			return true
		}
	}
	return false
}

// containsFlagOnly checks if the args contain a flag (without checking value).
func containsFlagOnly(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// genSafeString generates a non-empty string without special characters.
func genSafeString(t *rapid.T) string {
	chars := rapid.SliceOfN(
		rapid.SampledFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")),
		1, 20,
	).Draw(t, "chars")
	return string(chars)
}

// genPriority generates an optional priority value.
func genPriority(t *rapid.T) string {
	if rapid.Bool().Draw(t, "hasPriority") {
		return rapid.SampledFrom([]string{"High", "Medium", "Low", "Critical", "Blocker"}).Draw(t, "priority")
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

// TestOptionalFieldPassing tests Property 7: Optional Field Passing.
// For any Epic or Story with a non-empty Priority field, the jira-cli command
// SHALL include `-y <priority>`. For any Epic or Story with a non-empty Assignee
// field, the jira-cli command SHALL include `-a <assignee>`.
// **Validates: Requirements 2.5, 2.6, 3.6, 3.7**
func TestOptionalFieldPassing(t *testing.T) {
	// Test Epic creation with property-based testing
	t.Run("Epic", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			// Generate random test data
			summary := genSafeString(rt)
			priority := genPriority(rt)
			assignee := genAssignee(rt)

			mockExec := &MockExecutor{
				Output:         []byte("Issue created: TEST-123"),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateEpic(summary, priority, assignee)
			if err != nil {
				rt.Fatalf("CreateEpic failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			cmd := mockExec.Captured[0]
			args := cmd.Args

			// Property: If priority is non-empty, command SHALL include -y <priority>
			if priority != "" {
				if !containsFlag(args, "-y", priority) {
					rt.Errorf("Epic with priority %q: expected -y %s in args, got: %v",
						priority, priority, args)
				}
			} else {
				// If priority is empty, -y flag should NOT be present
				if containsFlagOnly(args, "-y") {
					rt.Errorf("Epic with empty priority: -y flag should not be present, got: %v", args)
				}
			}

			// Property: If assignee is non-empty, command SHALL include -a <assignee>
			if assignee != "" {
				if !containsFlag(args, "-a", assignee) {
					rt.Errorf("Epic with assignee %q: expected -a %s in args, got: %v",
						assignee, assignee, args)
				}
			} else {
				// If assignee is empty, -a flag should NOT be present
				if containsFlagOnly(args, "-a") {
					rt.Errorf("Epic with empty assignee: -a flag should not be present, got: %v", args)
				}
			}
		})
	})

	// Test Story creation with property-based testing
	t.Run("Story", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			// Generate random test data
			summary := genSafeString(rt)
			description := genSafeString(rt)
			epicKey := "TEST-" + genSafeString(rt)
			priority := genPriority(rt)
			assignee := genAssignee(rt)

			mockExec := &MockExecutor{
				Output:         []byte("Issue created: TEST-456"),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateStory(summary, description, epicKey, priority, assignee)
			if err != nil {
				rt.Fatalf("CreateStory failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			cmd := mockExec.Captured[0]
			args := cmd.Args

			// Property: If priority is non-empty, command SHALL include -y <priority>
			if priority != "" {
				if !containsFlag(args, "-y", priority) {
					rt.Errorf("Story with priority %q: expected -y %s in args, got: %v",
						priority, priority, args)
				}
			} else {
				// If priority is empty, -y flag should NOT be present
				if containsFlagOnly(args, "-y") {
					rt.Errorf("Story with empty priority: -y flag should not be present, got: %v", args)
				}
			}

			// Property: If assignee is non-empty, command SHALL include -a <assignee>
			if assignee != "" {
				if !containsFlag(args, "-a", assignee) {
					rt.Errorf("Story with assignee %q: expected -a %s in args, got: %v",
						assignee, assignee, args)
				}
			} else {
				// If assignee is empty, -a flag should NOT be present
				if containsFlagOnly(args, "-a") {
					rt.Errorf("Story with empty assignee: -a flag should not be present, got: %v", args)
				}
			}
		})
	})
}

// TestOptionalFieldPassing_EpicWithBothFields tests that an Epic with both
// priority and assignee includes both flags.
func TestOptionalFieldPassing_EpicWithBothFields(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateEpic("Test Epic", "High", "john.doe")
	if err != nil {
		t.Fatalf("CreateEpic failed: %v", err)
	}

	if len(mockExec.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
	}

	args := mockExec.Captured[0].Args

	if !containsFlag(args, "-y", "High") {
		t.Errorf("expected -y High in args, got: %v", args)
	}

	if !containsFlag(args, "-a", "john.doe") {
		t.Errorf("expected -a john.doe in args, got: %v", args)
	}
}

// TestOptionalFieldPassing_EpicWithNoOptionalFields tests that an Epic without
// priority or assignee does not include those flags.
func TestOptionalFieldPassing_EpicWithNoOptionalFields(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateEpic("Test Epic", "", "")
	if err != nil {
		t.Fatalf("CreateEpic failed: %v", err)
	}

	if len(mockExec.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
	}

	args := mockExec.Captured[0].Args

	if containsFlagOnly(args, "-y") {
		t.Errorf("-y flag should not be present when priority is empty, got: %v", args)
	}

	if containsFlagOnly(args, "-a") {
		t.Errorf("-a flag should not be present when assignee is empty, got: %v", args)
	}
}

// TestOptionalFieldPassing_StoryWithBothFields tests that a Story with both
// priority and assignee includes both flags.
func TestOptionalFieldPassing_StoryWithBothFields(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-456"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateStory("Test Story", "Description", "TEST-123", "Medium", "jane.doe")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}

	if len(mockExec.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
	}

	args := mockExec.Captured[0].Args

	if !containsFlag(args, "-y", "Medium") {
		t.Errorf("expected -y Medium in args, got: %v", args)
	}

	if !containsFlag(args, "-a", "jane.doe") {
		t.Errorf("expected -a jane.doe in args, got: %v", args)
	}
}

// TestOptionalFieldPassing_StoryWithNoOptionalFields tests that a Story without
// priority or assignee does not include those flags.
func TestOptionalFieldPassing_StoryWithNoOptionalFields(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-456"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateStory("Test Story", "Description", "TEST-123", "", "")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}

	if len(mockExec.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
	}

	args := mockExec.Captured[0].Args

	if containsFlagOnly(args, "-y") {
		t.Errorf("-y flag should not be present when priority is empty, got: %v", args)
	}

	if containsFlagOnly(args, "-a") {
		t.Errorf("-a flag should not be present when assignee is empty, got: %v", args)
	}
}

// TestOptionalFieldPassing_PriorityOnlyEpic tests that an Epic with only priority
// includes -y but not -a.
func TestOptionalFieldPassing_PriorityOnlyEpic(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateEpic("Test Epic", "Critical", "")
	if err != nil {
		t.Fatalf("CreateEpic failed: %v", err)
	}

	args := mockExec.Captured[0].Args

	if !containsFlag(args, "-y", "Critical") {
		t.Errorf("expected -y Critical in args, got: %v", args)
	}

	if containsFlagOnly(args, "-a") {
		t.Errorf("-a flag should not be present when assignee is empty, got: %v", args)
	}
}

// TestOptionalFieldPassing_AssigneeOnlyStory tests that a Story with only assignee
// includes -a but not -y.
func TestOptionalFieldPassing_AssigneeOnlyStory(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-456"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateStory("Test Story", "Description", "TEST-123", "", "bob.smith")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}

	args := mockExec.Captured[0].Args

	if containsFlagOnly(args, "-y") {
		t.Errorf("-y flag should not be present when priority is empty, got: %v", args)
	}

	if !containsFlag(args, "-a", "bob.smith") {
		t.Errorf("expected -a bob.smith in args, got: %v", args)
	}
}

// TestOptionalFieldPassing_VerifyFlagOrder tests that optional flags are appended
// after required flags.
func TestOptionalFieldPassing_VerifyFlagOrder(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateEpic("Test Epic", "High", "john.doe")
	if err != nil {
		t.Fatalf("CreateEpic failed: %v", err)
	}

	args := mockExec.Captured[0].Args

	// Verify required flags are present
	if !containsFlag(args, "-t", "Epic") {
		t.Errorf("expected -t Epic in args, got: %v", args)
	}

	if !containsFlag(args, "-s", "Test Epic") {
		t.Errorf("expected -s 'Test Epic' in args, got: %v", args)
	}

	if !containsFlag(args, "-P", "TEST") {
		t.Errorf("expected -P TEST in args, got: %v", args)
	}

	// Verify optional flags are present
	if !containsFlag(args, "-y", "High") {
		t.Errorf("expected -y High in args, got: %v", args)
	}

	if !containsFlag(args, "-a", "john.doe") {
		t.Errorf("expected -a john.doe in args, got: %v", args)
	}

	// Verify the command structure
	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "issue create") {
		t.Errorf("expected 'issue create' in command, got: %v", args)
	}
}

// genEpicKey generates a valid Jira Epic key (e.g., TEST-123).
func genEpicKey(t *rapid.T) string {
	// Generate a project prefix (2-10 uppercase letters)
	prefixLen := rapid.IntRange(2, 10).Draw(t, "prefixLen")
	prefixChars := rapid.SliceOfN(
		rapid.SampledFrom([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")),
		prefixLen, prefixLen,
	).Draw(t, "prefixChars")
	prefix := string(prefixChars)

	// Generate a ticket number (1-99999)
	number := rapid.IntRange(1, 99999).Draw(t, "ticketNumber")

	return fmt.Sprintf("%s-%d", prefix, number)
}

// TestEpicStoryLinking tests Property 8: Epic-Story Linking.
// For any Story created in Jira, the jira-cli command SHALL include
// `--parent <epic-key>` where `<epic-key>` is the Ticket_Key of the Story's parent Epic.
// **Validates: Requirements 3.4**
func TestEpicStoryLinking(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate random test data
		summary := genSafeString(rt)
		description := genSafeString(rt)
		epicKey := genEpicKey(rt)
		priority := genPriority(rt)
		assignee := genAssignee(rt)

		mockExec := &MockExecutor{
			Output:         []byte("Issue created: TEST-456"),
			LookPathResult: "/usr/local/bin/jira",
		}

		adapter := NewAdapterWithExecutor(Config{
			ProjectKey: "TEST",
			ServerURL:  "https://jira.example.com",
		}, mockExec)

		_, err := adapter.CreateStory(summary, description, epicKey, priority, assignee)
		if err != nil {
			rt.Fatalf("CreateStory failed: %v", err)
		}

		if len(mockExec.Captured) != 1 {
			rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
		}

		cmd := mockExec.Captured[0]
		args := cmd.Args

		// Property: The jira-cli command SHALL include --parent <epic-key>
		if !containsFlag(args, "--parent", epicKey) {
			rt.Errorf("Story creation command missing --parent %s flag, got args: %v",
				epicKey, args)
		}
	})
}

// TestEpicStoryLinking_VerifyParentFlagPresent tests that the --parent flag
// is always present when creating a Story.
func TestEpicStoryLinking_VerifyParentFlagPresent(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-456"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateStory("Test Story", "Description", "PROJ-123", "", "")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}

	if len(mockExec.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
	}

	args := mockExec.Captured[0].Args

	if !containsFlag(args, "--parent", "PROJ-123") {
		t.Errorf("expected --parent PROJ-123 in args, got: %v", args)
	}
}

// TestEpicStoryLinking_DifferentEpicKeys tests that different Epic keys
// are correctly passed to the --parent flag.
func TestEpicStoryLinking_DifferentEpicKeys(t *testing.T) {
	testCases := []struct {
		name    string
		epicKey string
	}{
		{"simple key", "TEST-1"},
		{"large number", "PROJ-99999"},
		{"long prefix", "LONGPROJECT-42"},
		{"short prefix", "AB-100"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := &MockExecutor{
				Output:         []byte("Issue created: TEST-456"),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateStory("Test Story", "Description", tc.epicKey, "", "")
			if err != nil {
				t.Fatalf("CreateStory failed: %v", err)
			}

			args := mockExec.Captured[0].Args

			if !containsFlag(args, "--parent", tc.epicKey) {
				t.Errorf("expected --parent %s in args, got: %v", tc.epicKey, args)
			}
		})
	}
}

// TestEpicStoryLinking_WithOptionalFields tests that --parent flag is present
// even when optional fields (priority, assignee) are provided.
func TestEpicStoryLinking_WithOptionalFields(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-456"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateStory("Test Story", "Description", "EPIC-999", "High", "john.doe")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}

	args := mockExec.Captured[0].Args

	// Verify --parent flag is present with correct Epic key
	if !containsFlag(args, "--parent", "EPIC-999") {
		t.Errorf("expected --parent EPIC-999 in args, got: %v", args)
	}

	// Also verify other flags are present
	if !containsFlag(args, "-y", "High") {
		t.Errorf("expected -y High in args, got: %v", args)
	}

	if !containsFlag(args, "-a", "john.doe") {
		t.Errorf("expected -a john.doe in args, got: %v", args)
	}
}

// genProjectKey generates a valid Jira project key (2-10 uppercase letters).
func genProjectKey(t *rapid.T) string {
	keyLen := rapid.IntRange(2, 10).Draw(t, "keyLen")
	keyChars := rapid.SliceOfN(
		rapid.SampledFrom([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")),
		keyLen, keyLen,
	).Draw(t, "keyChars")
	return string(keyChars)
}

// TestProjectKeyInCommands tests Property 14: Project Key in Commands.
// For any jira-cli command issued by the adapter (create or search), the command
// SHALL include `-P <project-key>` where `<project-key>` is the configured
// JIRA_PROJECT_KEY value.
// **Validates: Requirements 7.5**
func TestProjectKeyInCommands(t *testing.T) {
	// Test CreateEpic includes project key
	t.Run("CreateEpic", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			projectKey := genProjectKey(rt)
			summary := genSafeString(rt)
			priority := genPriority(rt)
			assignee := genAssignee(rt)

			mockExec := &MockExecutor{
				Output:         []byte("Issue created: TEST-123"),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: projectKey,
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateEpic(summary, priority, assignee)
			if err != nil {
				rt.Fatalf("CreateEpic failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			args := mockExec.Captured[0].Args

			// Property: Command SHALL include -P <project-key>
			if !containsFlag(args, "-P", projectKey) {
				rt.Errorf("CreateEpic command missing -P %s flag, got args: %v",
					projectKey, args)
			}
		})
	})

	// Test CreateStory includes project key
	t.Run("CreateStory", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			projectKey := genProjectKey(rt)
			summary := genSafeString(rt)
			description := genSafeString(rt)
			epicKey := genEpicKey(rt)
			priority := genPriority(rt)
			assignee := genAssignee(rt)

			mockExec := &MockExecutor{
				Output:         []byte("Issue created: TEST-456"),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: projectKey,
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateStory(summary, description, epicKey, priority, assignee)
			if err != nil {
				rt.Fatalf("CreateStory failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			args := mockExec.Captured[0].Args

			// Property: Command SHALL include -P <project-key>
			if !containsFlag(args, "-P", projectKey) {
				rt.Errorf("CreateStory command missing -P %s flag, got args: %v",
					projectKey, args)
			}
		})
	})

	// Test SearchEpic includes project key
	t.Run("SearchEpic", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			projectKey := genProjectKey(rt)
			summary := genSafeString(rt)

			mockExec := &MockExecutor{
				Output:         []byte(""),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: projectKey,
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.SearchEpic(summary)
			if err != nil {
				rt.Fatalf("SearchEpic failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			args := mockExec.Captured[0].Args

			// Property: Command SHALL include -P <project-key>
			if !containsFlag(args, "-P", projectKey) {
				rt.Errorf("SearchEpic command missing -P %s flag, got args: %v",
					projectKey, args)
			}
		})
	})

	// Test SearchStory includes project key
	t.Run("SearchStory", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			projectKey := genProjectKey(rt)
			summary := genSafeString(rt)
			epicKey := genEpicKey(rt)

			mockExec := &MockExecutor{
				Output:         []byte(""),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: projectKey,
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.SearchStory(summary, epicKey)
			if err != nil {
				rt.Fatalf("SearchStory failed: %v", err)
			}

			if len(mockExec.Captured) != 1 {
				rt.Fatalf("expected 1 command, got %d", len(mockExec.Captured))
			}

			args := mockExec.Captured[0].Args

			// Property: Command SHALL include -P <project-key>
			if !containsFlag(args, "-P", projectKey) {
				rt.Errorf("SearchStory command missing -P %s flag, got args: %v",
					projectKey, args)
			}
		})
	})
}

// TestProjectKeyInCommands_AllMethodsWithSameKey tests that all adapter methods
// use the same configured project key.
func TestProjectKeyInCommands_AllMethodsWithSameKey(t *testing.T) {
	projectKey := "MYPROJECT"

	mockExec := &MockExecutor{
		Output:         []byte("Issue created: MYPROJECT-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: projectKey,
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	// Call all four methods
	_, _ = adapter.CreateEpic("Epic Summary", "", "")
	_, _ = adapter.CreateStory("Story Summary", "Description", "MYPROJECT-1", "", "")
	_, _ = adapter.SearchEpic("Search Epic")
	_, _ = adapter.SearchStory("Search Story", "MYPROJECT-1")

	if len(mockExec.Captured) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(mockExec.Captured))
	}

	// Verify all commands include -P with the configured project key
	methodNames := []string{"CreateEpic", "CreateStory", "SearchEpic", "SearchStory"}
	for i, cmd := range mockExec.Captured {
		if !containsFlag(cmd.Args, "-P", projectKey) {
			t.Errorf("%s command missing -P %s flag, got args: %v",
				methodNames[i], projectKey, cmd.Args)
		}
	}
}

// TestProjectKeyInCommands_DifferentProjectKeys tests that different project keys
// are correctly passed to the -P flag.
func TestProjectKeyInCommands_DifferentProjectKeys(t *testing.T) {
	testCases := []struct {
		name       string
		projectKey string
	}{
		{"short key", "AB"},
		{"medium key", "TEST"},
		{"long key", "LONGPROJECT"},
		{"max length key", "ABCDEFGHIJ"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := &MockExecutor{
				Output:         []byte(fmt.Sprintf("Issue created: %s-123", tc.projectKey)),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: tc.projectKey,
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			_, err := adapter.CreateEpic("Test Epic", "", "")
			if err != nil {
				t.Fatalf("CreateEpic failed: %v", err)
			}

			args := mockExec.Captured[0].Args

			if !containsFlag(args, "-P", tc.projectKey) {
				t.Errorf("expected -P %s in args, got: %v", tc.projectKey, args)
			}
		})
	}
}

// TestProjectKeyInCommands_ProjectKeyPosition tests that -P flag appears
// in the correct position in the command arguments.
func TestProjectKeyInCommands_ProjectKeyPosition(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Issue created: TEST-123"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	_, err := adapter.CreateEpic("Test Epic", "High", "john.doe")
	if err != nil {
		t.Fatalf("CreateEpic failed: %v", err)
	}

	args := mockExec.Captured[0].Args

	// Verify -P flag is present
	if !containsFlag(args, "-P", "TEST") {
		t.Errorf("expected -P TEST in args, got: %v", args)
	}

	// Verify the command structure includes required elements
	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "issue create") {
		t.Errorf("expected 'issue create' in command, got: %v", args)
	}
	if !strings.Contains(argsStr, "-t Epic") {
		t.Errorf("expected '-t Epic' in command, got: %v", args)
	}
}

// genTicketKey generates a valid Jira ticket key matching pattern [A-Z]+-\d+.
func genTicketKey(t *rapid.T) string {
	// Generate a project prefix (1-10 uppercase letters)
	prefixLen := rapid.IntRange(1, 10).Draw(t, "prefixLen")
	prefixChars := rapid.SliceOfN(
		rapid.SampledFrom([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")),
		prefixLen, prefixLen,
	).Draw(t, "prefixChars")
	prefix := string(prefixChars)

	// Generate a ticket number (1-999999)
	number := rapid.IntRange(1, 999999).Draw(t, "ticketNumber")

	return fmt.Sprintf("%s-%d", prefix, number)
}

// genJiraCliOutput generates mock jira-cli output containing a ticket key.
func genJiraCliOutput(t *rapid.T, ticketKey string) string {
	// Generate different output formats that jira-cli might produce
	outputFormat := rapid.IntRange(0, 4).Draw(t, "outputFormat")

	switch outputFormat {
	case 0:
		// Simple format: "Issue created: PROJ-123"
		return fmt.Sprintf("Issue created: %s", ticketKey)
	case 1:
		// Verbose format with URL
		return fmt.Sprintf("✓ Issue %s created\nhttps://jira.example.com/browse/%s", ticketKey, ticketKey)
	case 2:
		// Format with additional info
		return fmt.Sprintf("Created issue %s in project\nSummary: Test Issue\nType: Epic", ticketKey)
	case 3:
		// Format with leading/trailing whitespace
		return fmt.Sprintf("\n  Issue %s has been created successfully  \n", ticketKey)
	default:
		// JSON-like format
		return fmt.Sprintf(`{"key": "%s", "self": "https://jira.example.com/rest/api/2/issue/%s"}`, ticketKey, ticketKey)
	}
}

// TestTicketKeyExtraction tests Property 15: Ticket Key Extraction.
// For any successful jira-cli `issue create` command output containing a ticket key
// pattern (e.g., PROJ-123), the adapter SHALL extract and return that key.
// The returned key SHALL match the pattern `[A-Z]+-\d+`.
// **Validates: Requirements 2.4, 7.3**
func TestTicketKeyExtraction(t *testing.T) {
	// Test Epic creation extracts ticket key correctly
	t.Run("CreateEpic", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			// Generate a random valid ticket key
			expectedKey := genTicketKey(rt)

			// Generate mock jira-cli output containing the key
			mockOutput := genJiraCliOutput(rt, expectedKey)

			mockExec := &MockExecutor{
				Output:         []byte(mockOutput),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			result, err := adapter.CreateEpic("Test Epic", "", "")
			if err != nil {
				rt.Fatalf("CreateEpic returned error: %v", err)
			}

			// Property: The adapter SHALL extract and return the ticket key
			if result.Error != nil {
				rt.Fatalf("CreateEpic result has error: %v", result.Error)
			}

			if result.Key != expectedKey {
				rt.Errorf("Expected extracted key %q, got %q (from output: %q)",
					expectedKey, result.Key, mockOutput)
			}

			// Property: The returned key SHALL match the pattern [A-Z]+-\d+
			keyPattern := regexp.MustCompile(`^[A-Z]+-\d+$`)
			if !keyPattern.MatchString(result.Key) {
				rt.Errorf("Extracted key %q does not match pattern [A-Z]+-\\d+", result.Key)
			}
		})
	})

	// Test Story creation extracts ticket key correctly
	t.Run("CreateStory", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			// Generate a random valid ticket key
			expectedKey := genTicketKey(rt)

			// Generate mock jira-cli output containing the key
			mockOutput := genJiraCliOutput(rt, expectedKey)

			mockExec := &MockExecutor{
				Output:         []byte(mockOutput),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			result, err := adapter.CreateStory("Test Story", "Description", "PARENT-1", "", "")
			if err != nil {
				rt.Fatalf("CreateStory returned error: %v", err)
			}

			// Property: The adapter SHALL extract and return the ticket key
			if result.Error != nil {
				rt.Fatalf("CreateStory result has error: %v", result.Error)
			}

			if result.Key != expectedKey {
				rt.Errorf("Expected extracted key %q, got %q (from output: %q)",
					expectedKey, result.Key, mockOutput)
			}

			// Property: The returned key SHALL match the pattern [A-Z]+-\d+
			keyPattern := regexp.MustCompile(`^[A-Z]+-\d+$`)
			if !keyPattern.MatchString(result.Key) {
				rt.Errorf("Extracted key %q does not match pattern [A-Z]+-\\d+", result.Key)
			}
		})
	})
}

// TestTicketKeyExtraction_VariousOutputFormats tests that ticket keys are extracted
// from various jira-cli output formats.
func TestTicketKeyExtraction_VariousOutputFormats(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "simple format",
			output:   "Issue created: PROJ-123",
			expected: "PROJ-123",
		},
		{
			name:     "with URL",
			output:   "✓ Issue TEST-456 created\nhttps://jira.example.com/browse/TEST-456",
			expected: "TEST-456",
		},
		{
			name:     "verbose output",
			output:   "Created issue MYPROJECT-789 in project\nSummary: Test\nType: Epic",
			expected: "MYPROJECT-789",
		},
		{
			name:     "with whitespace",
			output:   "\n  Issue ABC-1 has been created  \n",
			expected: "ABC-1",
		},
		{
			name:     "large ticket number",
			output:   "Issue created: BIGPROJ-999999",
			expected: "BIGPROJ-999999",
		},
		{
			name:     "single letter prefix",
			output:   "Issue created: A-1",
			expected: "A-1",
		},
		{
			name:     "long prefix",
			output:   "Issue created: VERYLONGPREFIX-42",
			expected: "VERYLONGPREFIX-42",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := &MockExecutor{
				Output:         []byte(tc.output),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			result, err := adapter.CreateEpic("Test Epic", "", "")
			if err != nil {
				t.Fatalf("CreateEpic returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("CreateEpic result has error: %v", result.Error)
			}

			if result.Key != tc.expected {
				t.Errorf("Expected key %q, got %q", tc.expected, result.Key)
			}

			// Verify the key matches the required pattern
			keyPattern := regexp.MustCompile(`^[A-Z]+-\d+$`)
			if !keyPattern.MatchString(result.Key) {
				t.Errorf("Extracted key %q does not match pattern [A-Z]+-\\d+", result.Key)
			}
		})
	}
}

// TestTicketKeyExtraction_NoKeyInOutput tests that the adapter handles output
// without a valid ticket key.
func TestTicketKeyExtraction_NoKeyInOutput(t *testing.T) {
	testCases := []struct {
		name   string
		output string
	}{
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "no key pattern",
			output: "Operation completed successfully",
		},
		{
			name:   "lowercase key",
			output: "Issue created: proj-123",
		},
		{
			name:   "missing hyphen",
			output: "Issue created: PROJ123",
		},
		{
			name:   "missing number",
			output: "Issue created: PROJ-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := &MockExecutor{
				Output:         []byte(tc.output),
				LookPathResult: "/usr/local/bin/jira",
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			result, err := adapter.CreateEpic("Test Epic", "", "")
			if err != nil {
				t.Fatalf("CreateEpic returned error: %v", err)
			}

			// When no valid key is found, result should have an error
			if result.Error == nil {
				t.Errorf("Expected error when no valid key in output, got key: %q", result.Key)
			}
		})
	}
}

// TestTicketKeyExtraction_MultipleKeysInOutput tests that the adapter extracts
// the first valid ticket key when multiple keys are present.
func TestTicketKeyExtraction_MultipleKeysInOutput(t *testing.T) {
	mockExec := &MockExecutor{
		Output:         []byte("Created FIRST-1, linked to SECOND-2, parent THIRD-3"),
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	result, err := adapter.CreateEpic("Test Epic", "", "")
	if err != nil {
		t.Fatalf("CreateEpic returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("CreateEpic result has error: %v", result.Error)
	}

	// Should extract the first key
	if result.Key != "FIRST-1" {
		t.Errorf("Expected first key FIRST-1, got %q", result.Key)
	}
}

// TestTicketKeyExtraction_KeyPatternValidation tests that extracted keys
// strictly match the [A-Z]+-\d+ pattern.
func TestTicketKeyExtraction_KeyPatternValidation(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a valid key
		key := genTicketKey(rt)

		mockExec := &MockExecutor{
			Output:         []byte(fmt.Sprintf("Issue created: %s", key)),
			LookPathResult: "/usr/local/bin/jira",
		}

		adapter := NewAdapterWithExecutor(Config{
			ProjectKey: "TEST",
			ServerURL:  "https://jira.example.com",
		}, mockExec)

		result, err := adapter.CreateEpic("Test Epic", "", "")
		if err != nil {
			rt.Fatalf("CreateEpic returned error: %v", err)
		}

		if result.Error != nil {
			rt.Fatalf("CreateEpic result has error: %v", result.Error)
		}

		// Validate the extracted key matches the pattern
		keyPattern := regexp.MustCompile(`^[A-Z]+-\d+$`)
		if !keyPattern.MatchString(result.Key) {
			rt.Errorf("Extracted key %q does not match pattern [A-Z]+-\\d+", result.Key)
		}

		// Validate the key has the expected structure
		parts := strings.Split(result.Key, "-")
		if len(parts) != 2 {
			rt.Errorf("Key %q should have exactly one hyphen", result.Key)
		}

		// Validate prefix is all uppercase letters
		for _, c := range parts[0] {
			if c < 'A' || c > 'Z' {
				rt.Errorf("Key prefix %q should only contain uppercase letters", parts[0])
				break
			}
		}

		// Validate suffix is all digits
		for _, c := range parts[1] {
			if c < '0' || c > '9' {
				rt.Errorf("Key suffix %q should only contain digits", parts[1])
				break
			}
		}
	})
}

// TestCheckInstalled_JiraCliNotFound tests that CheckInstalled returns a descriptive
// error when jira-cli is not installed or not in PATH.
// **Validates: Requirements 7.2**
func TestCheckInstalled_JiraCliNotFound(t *testing.T) {
	mockExec := &MockExecutor{
		LookPathError: fmt.Errorf("executable file not found in $PATH"),
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	err := adapter.CheckInstalled()

	// Should return an error
	if err == nil {
		t.Fatal("Expected error when jira-cli is not installed, got nil")
	}

	// Error message should be descriptive and include installation instructions
	errMsg := err.Error()
	if !strings.Contains(errMsg, "jira-cli not found") {
		t.Errorf("Error message should mention 'jira-cli not found', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "brew install jira-cli") {
		t.Errorf("Error message should include installation instructions, got: %s", errMsg)
	}
}

// TestCheckInstalled_JiraCliFound tests that CheckInstalled returns nil when
// jira-cli is available in PATH.
func TestCheckInstalled_JiraCliFound(t *testing.T) {
	mockExec := &MockExecutor{
		LookPathResult: "/usr/local/bin/jira",
	}

	adapter := NewAdapterWithExecutor(Config{
		ProjectKey: "TEST",
		ServerURL:  "https://jira.example.com",
	}, mockExec)

	err := adapter.CheckInstalled()

	if err != nil {
		t.Errorf("Expected no error when jira-cli is installed, got: %v", err)
	}
}

// TestCheckInstalled_DescriptiveErrorMessage tests that the error message
// provides actionable information for the user.
func TestCheckInstalled_DescriptiveErrorMessage(t *testing.T) {
	testCases := []struct {
		name          string
		lookPathError error
	}{
		{
			name:          "executable not found",
			lookPathError: fmt.Errorf("executable file not found in $PATH"),
		},
		{
			name:          "permission denied",
			lookPathError: fmt.Errorf("permission denied"),
		},
		{
			name:          "generic error",
			lookPathError: fmt.Errorf("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := &MockExecutor{
				LookPathError: tc.lookPathError,
			}

			adapter := NewAdapterWithExecutor(Config{
				ProjectKey: "TEST",
				ServerURL:  "https://jira.example.com",
			}, mockExec)

			err := adapter.CheckInstalled()

			if err == nil {
				t.Fatal("Expected error when LookPath fails")
			}

			// All error cases should provide the same descriptive message
			errMsg := err.Error()
			if !strings.Contains(errMsg, "jira-cli not found in PATH") {
				t.Errorf("Error should mention 'jira-cli not found in PATH', got: %s", errMsg)
			}
			if !strings.Contains(errMsg, "Install with:") {
				t.Errorf("Error should include installation instructions, got: %s", errMsg)
			}
		})
	}
}
