# Parse Requirements File

**User Story:** As a developer, I want to parse a requirements.md file into structured data, so that I can create Jira tickets from it.

## Extract Epic Headings

WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Epic headings from top-level `#` markdown headings.

### Acceptance Criteria

- WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Epic headings from top-level `#` markdown headings
- WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Story sections from `##` markdown headings under each Epic

## Extract Story Sections

WHEN a valid requirements.md file path is provided, THE Parser SHALL extract all Story sections from `##` markdown headings under each Epic.

### Acceptance Criteria

- WHEN a Story section contains acceptance criteria, THE Parser SHALL include them in the Story description

## Handle File Not Found

WHEN the requirements.md file does not exist, THE Parser SHALL return a descriptive file-not-found error.

### Acceptance Criteria

- WHEN the requirements.md file does not exist, THE Parser SHALL return a descriptive file-not-found error

## Handle Invalid Markdown

WHEN the requirements.md file contains invalid markdown, THE Parser SHALL return a descriptive parsing error.

### Acceptance Criteria

- WHEN the requirements.md file contains invalid markdown, THE Parser SHALL return a descriptive parsing error

## Preserve Hierarchy

THE Parser SHALL preserve the hierarchical relationship between Epics and their child Stories.

### Acceptance Criteria

- THE Parser SHALL preserve the hierarchical relationship between Epics and their child Stories
- FOR ALL valid Requirements_File content, parsing then formatting back to markdown then parsing SHALL produce equivalent structured data (round-trip property)

## Extract Optional Fields

WHEN an Epic or Story section contains optional fields, THE Parser SHALL extract them.

### Acceptance Criteria

- WHEN an Epic or Story section contains a `**Priority:**` field, THE Parser SHALL extract the priority value
- WHEN an Epic or Story section contains an `**Assignee:**` field, THE Parser SHALL extract the assignee value

# Create Epics in Jira

**User Story:** As a developer, I want to create Epics in Jira from parsed requirements, so that I can organize work at the initiative level.

## Create Epic Issue

WHEN an Epic is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create an Epic issue in the configured Jira project.

### Acceptance Criteria

- WHEN an Epic is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create an Epic issue in the configured Jira project
- WHEN creating an Epic, THE Jira_CLI_Adapter SHALL use the `#` heading text as the Epic summary

## Handle Epic Creation Failure

WHEN the jira-cli command fails for an Epic, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items.

### Acceptance Criteria

- WHEN the jira-cli command fails for an Epic, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items
- WHEN an Epic is successfully created, THE Jira_CLI_Adapter SHALL capture and return the assigned Ticket_Key

## Pass Epic Optional Fields

WHEN an Epic specifies optional fields, THE Jira_CLI_Adapter SHALL pass them to jira-cli.

### Acceptance Criteria

- WHEN an Epic specifies a priority, THE Jira_CLI_Adapter SHALL pass the priority to jira-cli using the `-y` flag
- WHEN an Epic specifies an assignee, THE Jira_CLI_Adapter SHALL pass the assignee to jira-cli using the `-a` flag

# Create Stories in Jira

**User Story:** As a developer, I want to create Stories linked to Epics, so that I can track user-level requirements in Jira.

## Create Story Issue

WHEN a Story is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create a Story issue in the configured Jira project.

### Acceptance Criteria

- WHEN a Story is parsed from the Requirements_File, THE Jira_CLI_Adapter SHALL create a Story issue in the configured Jira project
- WHEN creating a Story, THE Jira_CLI_Adapter SHALL use the `##` heading text as the Story summary
- WHEN creating a Story, THE Jira_CLI_Adapter SHALL include the acceptance criteria in the Story description

## Link Story to Epic

WHEN creating a Story, THE Jira_CLI_Adapter SHALL link it to its parent Epic using the Epic's Ticket_Key.

### Acceptance Criteria

- WHEN creating a Story, THE Jira_CLI_Adapter SHALL link it to its parent Epic using the Epic's Ticket_Key

## Handle Story Creation Failure

WHEN the jira-cli command fails for a Story, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items.

### Acceptance Criteria

- WHEN the jira-cli command fails for a Story, THE Jira_Ticket_Generator SHALL log the error and continue processing remaining items

## Pass Story Optional Fields

WHEN a Story specifies optional fields, THE Jira_CLI_Adapter SHALL pass them to jira-cli.

### Acceptance Criteria

- WHEN a Story specifies a priority, THE Jira_CLI_Adapter SHALL pass the priority to jira-cli using the `-y` flag
- WHEN a Story specifies an assignee, THE Jira_CLI_Adapter SHALL pass the assignee to jira-cli using the `-a` flag

# Configuration Management

**User Story:** As a developer, I want to configure the tool via environment variables, so that I can use it across different Jira projects without code changes.

## Read Environment Variables

THE Jira_Ticket_Generator SHALL read configuration from environment variables.

### Acceptance Criteria

- THE Jira_Ticket_Generator SHALL read the Jira Project_Key from the `JIRA_PROJECT_KEY` environment variable
- THE Jira_Ticket_Generator SHALL read the Jira server URL from the `JIRA_SERVER_URL` environment variable

## Handle Missing Configuration

WHEN a required environment variable is missing, THE Jira_Ticket_Generator SHALL exit with a descriptive configuration error.

### Acceptance Criteria

- WHEN a required environment variable is missing, THE Jira_Ticket_Generator SHALL exit with a descriptive configuration error

## Support .env File

THE Jira_Ticket_Generator SHALL support loading environment variables from a `.env` file.

### Acceptance Criteria

- THE Jira_Ticket_Generator SHALL support loading environment variables from a `.env` file in the current directory
- WHEN both `.env` file and system environment variables define the same variable, THE Jira_Ticket_Generator SHALL prefer the system environment variable

# Error Reporting

**User Story:** As a developer, I want comprehensive error reporting, so that I can diagnose and fix issues with ticket creation.

## Continue on Failure

WHEN one or more ticket creations fail, THE Jira_Ticket_Generator SHALL continue processing all remaining tickets.

### Acceptance Criteria

- WHEN one or more ticket creations fail, THE Jira_Ticket_Generator SHALL continue processing all remaining tickets

## Output Error Summary

WHEN processing completes with errors, THE Jira_Ticket_Generator SHALL output a summary of all failed operations.

### Acceptance Criteria

- WHEN processing completes with errors, THE Jira_Ticket_Generator SHALL output a summary of all failed operations
- WHEN processing completes with errors, THE Jira_Ticket_Generator SHALL exit with a non-zero exit code

## Output Success Summary

WHEN all tickets are created successfully, THE Jira_Ticket_Generator SHALL output a success summary with created Ticket_Keys.

### Acceptance Criteria

- WHEN all tickets are created successfully, THE Jira_Ticket_Generator SHALL output a success summary with created Ticket_Keys
- THE Jira_Ticket_Generator SHALL log each ticket creation attempt with its outcome (created, skipped, or failed)

# CLI Interface

**User Story:** As a developer, I want a simple CLI interface, so that I can easily invoke the tool from scripts and Kiro agents.

## Accept File Path Argument

THE Jira_Ticket_Generator SHALL accept the requirements.md file path as a positional argument.

### Acceptance Criteria

- THE Jira_Ticket_Generator SHALL accept the requirements.md file path as a positional argument
- WHEN invoked without a file path argument, THE Jira_Ticket_Generator SHALL exit with a usage error

## Support Help Flag

WHEN invoked with `--help`, THE Jira_Ticket_Generator SHALL display usage information.

### Acceptance Criteria

- WHEN invoked with `--help`, THE Jira_Ticket_Generator SHALL display usage information

## Support Dry Run Flag

WHEN invoked with `--dry-run`, THE Jira_Ticket_Generator SHALL parse and validate without creating tickets.

### Acceptance Criteria

- WHEN invoked with `--dry-run`, THE Jira_Ticket_Generator SHALL parse and validate without creating tickets

## Support Verbose Flag

WHEN invoked with `--verbose`, THE Jira_Ticket_Generator SHALL output detailed logging of all operations.

### Acceptance Criteria

- WHEN invoked with `--verbose`, THE Jira_Ticket_Generator SHALL output detailed logging of all operations

# Jira CLI Integration

**User Story:** As a developer, I want the tool to use jira-cli for Jira operations, so that I can leverage existing authentication and configuration.

## Execute via Subprocess

THE Jira_CLI_Adapter SHALL execute jira-cli commands via subprocess calls.

### Acceptance Criteria

- THE Jira_CLI_Adapter SHALL execute jira-cli commands via subprocess calls

## Check jira-cli Installation

WHEN jira-cli is not installed or not in PATH, THE Jira_Ticket_Generator SHALL exit with a descriptive dependency error.

### Acceptance Criteria

- WHEN jira-cli is not installed or not in PATH, THE Jira_Ticket_Generator SHALL exit with a descriptive dependency error

## Parse jira-cli Output

THE Jira_CLI_Adapter SHALL parse jira-cli output to extract created Ticket_Keys.

### Acceptance Criteria

- THE Jira_CLI_Adapter SHALL parse jira-cli output to extract created Ticket_Keys
- WHEN checking for existing tickets, THE Jira_CLI_Adapter SHALL use jira-cli search functionality
- THE Jira_CLI_Adapter SHALL pass the configured Project_Key to all jira-cli commands

# Requirements File Format Specification

**User Story:** As a developer, I want a documented requirements.md format, so that I can write Jira-compatible requirements.

## Document Format in Steering File

THE Jira_Ticket_Generator SHALL document the expected requirements.md format in a steering file.

### Acceptance Criteria

- THE Jira_Ticket_Generator SHALL document the expected requirements.md format in a steering file
- THE steering file SHALL specify that `#` headings become Epics
- THE steering file SHALL specify that `##` headings become Stories linked to the preceding Epic
- THE steering file SHALL specify the format for acceptance criteria within Story sections
- THE steering file SHALL provide examples of valid requirements.md structure
- THE steering file SHALL specify the format for optional priority field (e.g., `**Priority:** High`)
- THE steering file SHALL specify the format for optional assignee field (e.g., `**Assignee:** username`)

# Duplicate Handling

**User Story:** As a developer, I want the tool to handle duplicate tickets gracefully, so that I can run the tool multiple times without creating duplicates.

## Support On-Duplicate Parameter

THE Jira_Ticket_Generator SHALL accept a `--on-duplicate` parameter with values `skip` or `fail`.

### Acceptance Criteria

- THE Jira_Ticket_Generator SHALL accept a `--on-duplicate` parameter with values `skip` or `fail`
- WHEN `--on-duplicate` is not specified, THE Jira_Ticket_Generator SHALL default to `skip` mode

## Handle Duplicates in Skip Mode

WHEN a duplicate exists AND mode is `skip`, THE Jira_Ticket_Generator SHALL skip creation and use the existing Ticket_Key.

### Acceptance Criteria

- WHEN an Epic with the same summary already exists in the project AND mode is `skip`, THE Jira_Ticket_Generator SHALL skip creation and use the existing Ticket_Key
- WHEN a Story with the same summary already exists under the same Epic AND mode is `skip`, THE Jira_Ticket_Generator SHALL skip creation and continue processing remaining items

## Handle Duplicates in Fail Mode

WHEN a duplicate exists AND mode is `fail`, THE Jira_Ticket_Generator SHALL log an error and continue processing remaining items.

### Acceptance Criteria

- WHEN a duplicate Epic or Story exists AND mode is `fail`, THE Jira_Ticket_Generator SHALL log an error and continue processing remaining items

## Report Duplicates in Summary

THE Jira_Ticket_Generator SHALL report skipped or failed duplicates in the final summary output.

### Acceptance Criteria

- THE duplicate detection function SHALL be designed to accept additional modes (e.g., `update`) in future implementations
- THE Jira_Ticket_Generator SHALL report skipped or failed duplicates in the final summary output
