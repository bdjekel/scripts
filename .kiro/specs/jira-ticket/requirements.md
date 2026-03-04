# Requirements Document

## Introduction

The Jira Ticket Generator (jira-ticket) is a Go-based CLI tool that parses a `requirements.md` file and creates corresponding Jira tickets (Epics and Stories) using the `jira-cli` tool. The tool enables Kiro agents to automate Jira ticket creation directly from structured requirements documents, supporting idempotent operations and graceful error handling.

## Prerequisites

### jira-cli Authentication

The tool requires [jira-cli](https://github.com/ankitpokhrel/jira-cli) to be installed and authenticated:

```bash
# Install jira-cli
brew install jira-cli

# Authenticate (interactive setup)
jira init
```

Follow the prompts to configure your Jira server URL and authentication token.

### Knowledge Base Requirement

For AI agent integration, create a knowledge base file at `~/ai_agent/assistant/coding/stack/jira.md` containing:
- jira-cli common commands and usage patterns
- Project-specific Jira workflows
- Custom field mappings if applicable

This enables agents to query Jira context via `/knowledge search coding-context jira`.

## Glossary

- **Jira_Ticket_Generator**: The Go CLI tool that parses requirements and creates Jira tickets
- **Requirements_File**: A markdown file following the defined structure containing Epics and Stories
- **Parser**: The component that extracts Epic and Story data from the Requirements_File
- **Jira_CLI_Adapter**: The component that executes jira-cli commands via subprocess
- **Epic**: A top-level Jira issue type representing a major feature or initiative
- **Story**: A Jira issue type representing a user story, linked to a parent Epic
- **Ticket_Key**: The unique identifier assigned by Jira to each issue (e.g., PROJ-123)
- **Project_Key**: The Jira project identifier used when creating tickets

## Requirements

### Requirement 1: Parse Requirements File

**User Story:** As a developer, I want to parse a requirements.md file into structured data, so that I can create Jira tickets from it.

#### Acceptance Criteria

1. WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Epic headings from top-level `#` markdown headings
2. WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Story sections from `##` markdown headings under each Epic
3. WHEN a Story section contains acceptance criteria, THE Parser SHALL include them in the Story description
4. WHEN the requirements.md file does not exist, THE Parser SHALL return a descriptive file-not-found error
5. WHEN the requirements.md file contains invalid markdown, THE Parser SHALL return a descriptive parsing error
6. THE Parser SHALL preserve the hierarchical relationship between Epics and their child Stories
7. FOR ALL valid Requirements_File content, parsing then formatting back to markdown then parsing SHALL produce equivalent structured data (round-trip property)
8. WHEN an Epic or Story section contains a `**Priority:**` field, THE Parser SHALL extract the priority value
9. WHEN an Epic or Story section contains an `**Assignee:**` field, THE Parser SHALL extract the assignee value

### Requirement 2: Create Epics in Jira

**User Story:** As a developer, I want to create Epics in Jira from parsed requirements, so that I can organize work at the initiative level.

#### Acceptance Criteria

1. WHEN an Epic is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create an Epic issue in the configured Jira project
2. WHEN creating an Epic, THE Jira_CLI_Adapter SHALL use the `#` heading text as the Epic summary
3. WHEN the jira-cli command fails for an Epic, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items
4. WHEN an Epic is successfully created, THE Jira_CLI_Adapter SHALL capture and return the assigned Ticket_Key
5. WHEN an Epic specifies a priority, THE Jira_CLI_Adapter SHALL pass the priority to jira-cli using the `-y` flag
6. WHEN an Epic specifies an assignee, THE Jira_CLI_Adapter SHALL pass the assignee to jira-cli using the `-a` flag

### Requirement 3: Create Stories in Jira

**User Story:** As a developer, I want to create Stories linked to Epics, so that I can track user-level requirements in Jira.

#### Acceptance Criteria

1. WHEN a Story is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create a Story issue in the configured Jira project
2. WHEN creating a Story, THE Jira_CLI_Adapter SHALL use the `##` heading text as the Story summary
3. WHEN creating a Story, THE Jira_CLI_Adapter SHALL include the acceptance criteria in the Story description
4. WHEN creating a Story, THE Jira_CLI_Adapter SHALL link it to its parent Epic using the Epic's Ticket_Key
5. WHEN the jira-cli command fails for a Story, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items
6. WHEN a Story specifies a priority, THE Jira_CLI_Adapter SHALL pass the priority to jira-cli using the `-y` flag
7. WHEN a Story specifies an assignee, THE Jira_CLI_Adapter SHALL pass the assignee to jira-cli using the `-a` flag

### Requirement 4: Configuration Management

**User Story:** As a developer, I want to configure the tool via environment variables, so that I can use it across different Jira projects without code changes.

#### Acceptance Criteria

1. THE Jira_Ticket_Generator SHALL read the Jira Project_Key from the `JIRA_PROJECT_KEY` environment variable
2. THE Jira_Ticket_Generator SHALL read the Jira server URL from the `JIRA_SERVER_URL` environment variable
3. WHEN a required environment variable is missing, THE Jira_Ticket_Generator SHALL exit with a descriptive configuration error
4. THE Jira_Ticket_Generator SHALL support loading environment variables from a `.env` file in the current directory
5. WHEN both `.env` file and system environment variables define the same variable, THE Jira_Ticket_Generator SHALL prefer the system environment variable

### Requirement 5: Error Reporting

**User Story:** As a developer, I want comprehensive error reporting, so that I can diagnose and fix issues with ticket creation.

#### Acceptance Criteria

1. WHEN one or more ticket creations fail, THE Jira_Ticket_Generator SHALL continue processing all remaining tickets
2. WHEN processing completes with errors, THE Jira_Ticket_Generator SHALL output a summary of all failed operations
3. WHEN processing completes with errors, THE Jira_Ticket_Generator SHALL exit with a non-zero exit code
4. WHEN all tickets are created successfully, THE Jira_Ticket_Generator SHALL output a success summary with created Ticket_Keys
5. THE Jira_Ticket_Generator SHALL log each ticket creation attempt with its outcome (created, skipped, or failed)

### Requirement 6: CLI Interface

**User Story:** As a developer, I want a simple CLI interface, so that I can easily invoke the tool from scripts and Kiro agents.

#### Acceptance Criteria

1. THE Jira_Ticket_Generator SHALL accept the requirements.md file path as a positional argument
2. WHEN invoked with `--help`, THE Jira_Ticket_Generator SHALL display usage information
3. WHEN invoked with `--dry-run`, THE Jira_Ticket_Generator SHALL parse and validate without creating tickets
4. WHEN invoked with `--verbose`, THE Jira_Ticket_Generator SHALL output detailed logging of all operations
5. WHEN invoked without a file path argument, THE Jira_Ticket_Generator SHALL exit with a usage error

### Requirement 7: Jira CLI Integration

**User Story:** As a developer, I want the tool to use jira-cli for Jira operations, so that I can leverage existing authentication and configuration.

#### Acceptance Criteria

1. THE Jira_CLI_Adapter SHALL execute jira-cli commands via subprocess calls
2. WHEN jira-cli is not installed or not in PATH, THE Jira_Ticket_Generator SHALL exit with a descriptive dependency error
3. THE Jira_CLI_Adapter SHALL parse jira-cli output to extract created Ticket_Keys
4. WHEN checking for existing tickets, THE Jira_CLI_Adapter SHALL use jira-cli search functionality
5. THE Jira_CLI_Adapter SHALL pass the configured Project_Key to all jira-cli commands

### Requirement 8: Requirements File Format Specification

**User Story:** As a developer, I want a documented requirements.md format, so that I can write Jira-compatible requirements.

#### Acceptance Criteria

1. THE Jira_Ticket_Generator SHALL document the expected requirements.md format in a steering file
2. THE steering file SHALL specify that `#` headings become Epics
3. THE steering file SHALL specify that `##` headings become Stories linked to the preceding Epic
4. THE steering file SHALL specify the format for acceptance criteria within Story sections
5. THE steering file SHALL provide examples of valid requirements.md structure
6. THE steering file SHALL specify the format for optional priority field (e.g., `**Priority:** High`)
7. THE steering file SHALL specify the format for optional assignee field (e.g., `**Assignee:** username`)
### Requirement 9: Duplicate Handling

**User Story:** As a developer, I want the tool to handle duplicate tickets gracefully, so that I can run the tool multiple times without creating duplicates.

#### Acceptance Criteria

1. THE Jira_Ticket_Generator SHALL accept a `--on-duplicate` parameter with values `skip` or `fail`
2. WHEN `--on-duplicate` is not specified, THE Jira_Ticket_Generator SHALL default to `skip` mode
3. WHEN an Epic with the same summary already exists in the project AND mode is `skip`, THE Jira_Ticket_Generator SHALL skip creation and use the existing Ticket_Key
4. WHEN a Story with the same summary already exists under the same Epic AND mode is `skip`, THE Jira_Ticket_Generator SHALL skip creation and continue processing remaining items
5. WHEN a duplicate Epic or Story exists AND mode is `fail`, THE Jira_Ticket_Generator SHALL log an error and continue processing remaining items
6. THE duplicate detection function SHALL be designed to accept additional modes (e.g., `update`) in future implementations
7. THE Jira_Ticket_Generator SHALL report skipped or failed duplicates in the final summary output

## Future Enhancements

The following features are out of scope for the initial implementation but should be considered for future versions:

### Agent Integration
- **Kiro Hook Integration** — Auto-run `jira-ticket --dry-run` when a `requirements.md` file is saved, providing immediate validation feedback on structure
- **`--output json` flag** — Machine-readable output format for agent parsing and logging integration

### Duplicate Handling
- **`--on-duplicate update` mode** — Extend duplicate handling to support updating existing tickets with changed content

### Extended Issue Types
- **Task and Bug support** — Support additional issue types beyond Epic and Story (e.g., `### [Bug] Login fails` or `### [Task] Setup CI`)
- **Sub-task support** — Allow `###` headings to create Sub-tasks linked to parent Stories

### Extended Fields
- **Labels** — Support `-l` flag for labels (e.g., `**Labels:** backend, urgent`)
- **Components** — Support `-C` flag for components (e.g., `**Component:** Backend`)
- **Fix versions** — Support `--fix-version` flag (e.g., `**Fix Version:** v2.0`)

### Sprint Integration
- **Sprint assignment** — Auto-assign created issues to a sprint using `jira sprint add`
- **`--sprint` flag** — Accept sprint ID or `current` keyword to assign all created issues

### Issue Linking
- **Arbitrary issue links** — Support linking issues beyond Epic→Story parent relationships (e.g., `Blocks`, `Relates to`)

### Front-matter Defaults
Support YAML front-matter in requirements.md for project-wide defaults:
```yaml
---
priority: High
labels: [backend, q1-2026]
sprint: current
assignee: $(jira me)
---
```

### Architecture Note
The `Jira_CLI_Adapter` interface is intentionally narrow for MVP but designed to wrap additional jira-cli commands as needed. New operations should follow the same subprocess pattern with structured output parsing.
