# Implementation Plan: Jira Ticket Generator

## Overview

This plan implements a Go CLI tool that parses `requirements.md` files and creates Jira tickets using `jira-cli`. The implementation follows a 3-layer architecture (CLI → Core → Adapter) with property-based testing using the `rapid` library.

## Tasks

- [x] 1. Set up project structure and dependencies
  - Initialize Go module with `go mod init`
  - Create directory structure: `cmd/jira-ticket/`, `internal/parser/`, `internal/generator/`, `internal/jiracli/`, `pkg/config/`
  - Add dependencies: `github.com/flyingmutant/rapid`, `github.com/joho/godotenv`
  - _Requirements: 4.4_

- [x] 2. Implement configuration management
  - [x] 2.1 Create config types and loader in `pkg/config/config.go`
    - Define `Config` struct with `ProjectKey` and `ServerURL` fields
    - Implement `Load()` function to read from environment and `.env` file
    - Implement `Validate()` to check required variables
    - System env vars take precedence over `.env` file
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [x] 2.2 Write unit tests for config loading
    - Test missing required variables return descriptive errors
    - Test `.env` file loading
    - Test system env var precedence over `.env`
    - _Requirements: 4.3, 4.5_

- [x] 3. Implement parser component
  - [x] 3.1 Create parser types in `internal/parser/types.go`
    - Define `Epic` struct with Summary, Priority, Assignee, Stories fields
    - Define `Story` struct with Summary, Description, AcceptanceCriteria, Priority, Assignee fields
    - Define `RequirementsDoc` struct with Epics slice
    - Define `ParseError` struct with Line and Message fields
    - _Requirements: 1.1, 1.2, 1.8, 1.9_
  
  - [x] 3.2 Implement parser logic in `internal/parser/parser.go`
    - Implement `Parse(filePath string)` to read and parse file
    - Implement `ParseContent(content string)` for direct content parsing
    - Extract `#` headings as Epics
    - Extract `##` headings as Stories under their parent Epic
    - Extract `**Priority:**` and `**Assignee:**` fields
    - Extract acceptance criteria from bullet points under "Acceptance Criteria" heading
    - Implement `Format(doc *RequirementsDoc)` for round-trip testing
    - Return descriptive errors for file-not-found and invalid markdown
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.8, 1.9_
  
  - [x] 3.3 Write property test for round-trip consistency
    - **Property 1: Parse-Format Round Trip**
    - **Validates: Requirements 1.7**
  
  - [x] 3.4 Write property test for hierarchical preservation
    - **Property 2: Hierarchical Preservation**
    - **Validates: Requirements 1.6**
  
  - [x] 3.5 Write property test for heading extraction
    - **Property 3: Heading Extraction**
    - **Validates: Requirements 1.1, 1.2, 2.2, 3.2**
  
  - [x] 3.6 Write property test for optional field extraction
    - **Property 4: Optional Field Extraction**
    - **Validates: Requirements 1.8, 1.9**
  
  - [x] 3.7 Write property test for acceptance criteria inclusion
    - **Property 5: Acceptance Criteria Inclusion**
    - **Validates: Requirements 1.3, 3.3**
  
  - [x] 3.8 Write unit tests for parser error cases
    - Test file not found returns descriptive error
    - Test invalid markdown returns parse error with line number
    - _Requirements: 1.4, 1.5_

- [x] 4. Checkpoint - Parser complete
  - Ensure all parser tests pass, ask the user if questions arise.

- [x] 5. Implement Jira CLI adapter
  - [x] 5.1 Create adapter types in `internal/jiracli/types.go`
    - Define `TicketResult` struct with Key, Created, Skipped, Error fields
    - Define `SearchResult` struct with Key, Summary, Type fields
    - Define `Config` struct with ProjectKey and ServerURL fields
    - _Requirements: 2.4, 7.3_
  
  - [x] 5.2 Implement adapter logic in `internal/jiracli/adapter.go`
    - Implement `Adapter` interface with subprocess execution
    - Implement `CheckInstalled()` to verify jira-cli is in PATH
    - Implement `CreateEpic(summary, priority, assignee)` using `jira issue create -t Epic`
    - Implement `CreateStory(summary, description, epicKey, priority, assignee)` using `jira issue create -t Story --parent`
    - Implement `SearchEpic(summary)` using `jira issue list -t Epic`
    - Implement `SearchStory(summary, epicKey)` using `jira issue list -t Story`
    - Parse jira-cli output to extract ticket keys matching `[A-Z]+-\d+`
    - Include `-P <project-key>` in all commands
    - Include `-y <priority>` when priority is non-empty
    - Include `-a <assignee>` when assignee is non-empty
    - _Requirements: 2.1, 2.2, 2.4, 2.5, 2.6, 3.1, 3.2, 3.3, 3.4, 3.6, 3.7, 7.1, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 5.3 Write property test for optional field passing
    - **Property 7: Optional Field Passing**
    - **Validates: Requirements 2.5, 2.6, 3.6, 3.7**
  
  - [x] 5.4 Write property test for Epic-Story linking
    - **Property 8: Epic-Story Linking**
    - **Validates: Requirements 3.4**
  
  - [x] 5.5 Write property test for project key in commands
    - **Property 14: Project Key in Commands**
    - **Validates: Requirements 7.5**
  
  - [x] 5.6 Write property test for ticket key extraction
    - **Property 15: Ticket Key Extraction**
    - **Validates: Requirements 2.4, 7.3**
  
  - [x] 5.7 Write unit tests for adapter error cases
    - Test jira-cli not installed returns descriptive error
    - Test various jira-cli output formats are parsed correctly
    - _Requirements: 7.2_

- [x] 6. Checkpoint - Adapter complete
  - Ensure all adapter tests pass, ask the user if questions arise.

- [x] 7. Implement generator component
  - [x] 7.1 Create generator types in `internal/generator/generator.go`
    - Define `Options` struct with DryRun, Verbose, OnDuplicate fields
    - Define `Result` struct with Created, Skipped, Failed slices
    - Define `FailedTicket` struct with Summary, Type, Error fields
    - _Requirements: 6.3, 9.1, 9.2_
  
  - [x] 7.2 Implement generator orchestration logic
    - Implement `Generator` interface with `Generate(filePath, opts)` method
    - Parse requirements file using Parser
    - For each Epic: check for duplicate, create or skip based on mode
    - For each Story: check for duplicate, create with parent Epic key, or skip based on mode
    - Continue processing on individual ticket failures
    - Track created, skipped, and failed tickets
    - Support dry-run mode (parse only, no jira-cli calls)
    - _Requirements: 2.3, 3.5, 5.1, 5.2, 5.4, 5.5, 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7_
  
  - [x] 7.3 Write property test for error resilience
    - **Property 6: Error Resilience**
    - **Validates: Requirements 2.3, 3.5, 5.1**
  
  - [x] 7.4 Write property test for dry-run mode
    - **Property 9: Dry-Run Mode**
    - **Validates: Requirements 6.3**
  
  - [x] 7.5 Write property test for duplicate skip mode
    - **Property 10: Duplicate Skip Mode**
    - **Validates: Requirements 9.3, 9.4**
  
  - [x] 7.6 Write property test for duplicate fail mode
    - **Property 11: Duplicate Fail Mode**
    - **Validates: Requirements 9.5**
  
  - [x] 7.7 Write property test for summary accuracy
    - **Property 12: Summary Accuracy**
    - **Validates: Requirements 5.2, 5.4, 5.5, 9.7**
  
  - [x] 7.8 Write property test for exit code reflects outcome
    - **Property 13: Exit Code Reflects Outcome**
    - **Validates: Requirements 5.3**

- [x] 8. Checkpoint - Generator complete
  - Ensure all generator tests pass, ask the user if questions arise.

- [x] 9. Implement CLI entry point
  - [x] 9.1 Create CLI in `cmd/jira-ticket/main.go`
    - Parse positional argument for requirements.md file path
    - Implement `--help` flag to display usage information
    - Implement `--dry-run` flag for parse-only mode
    - Implement `--verbose` flag for detailed logging
    - Implement `--on-duplicate` flag with `skip` (default) and `fail` values
    - Load and validate configuration
    - Check jira-cli is installed
    - Call generator and handle result
    - Output success summary with created ticket keys
    - Output error summary with failed operations
    - Exit with non-zero code on failures
    - _Requirements: 4.3, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 7.2, 9.1, 9.2_
  
  - [x] 9.2 Write unit tests for CLI argument parsing
    - Test `--help` displays usage
    - Test `--dry-run` flag is passed to generator
    - Test `--verbose` flag enables detailed output
    - Test `--on-duplicate` flag values
    - Test missing file path argument exits with usage error
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 10. Create steering file documentation
  - Create steering file documenting requirements.md format
  - Document `#` headings become Epics
  - Document `##` headings become Stories linked to preceding Epic
  - Document acceptance criteria format
  - Document optional `**Priority:**` and `**Assignee:**` fields
  - Provide examples of valid requirements.md structure
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [x] 11. Final checkpoint - All tests pass
  - Run `go test ./...` to ensure all tests pass
  - Ensure all property tests run minimum 100 iterations
  - Ask the user if questions arise.

## Notes

- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using the `rapid` library
- Unit tests validate specific examples and edge cases
- The `Adapter` interface enables testing the generator with mocks
