# Requirements Document

## Introduction

This feature provides a command-line interface for creating Outlook Calendar events with preset parameters, designed to track work sessions automatically through Git hooks. The system extracts Jira issue keys from Git branch names and commit messages, manages session state, and creates calendar events to log time spent on tasks.

## Glossary

- **CLI Tool**: The command-line interface application that creates Outlook Calendar events
- **Outlook Calendar**: Microsoft Outlook's calendar application on macOS
- **Git Hook**: Automated scripts that run at specific points in the Git workflow
- **Jira Key**: A unique identifier for Jira issues extracted from branch names (e.g., PROJ-123)
- **Session**: A tracked work period between starting work (branch checkout) and ending work (commit)
- **Tracked Ticket**: A ticket (identified by Jira Key) that has been registered for automatic time tracking
- **Linked Branch**: A Git branch associated with a Tracked Ticket
- **State File**: A JSON file that persists information about active work sessions and tracked tickets
- **AppleScript**: macOS scripting language for controlling applications
- **Calendar Event**: An entry in Outlook Calendar representing a work session
- **Log Directory**: A directory in the project root containing stdout and stderr logs

## Requirements

### Requirement 1

**User Story:** As a developer, I want to create Outlook Calendar events via CLI with preset parameters, so that I can quickly log work sessions without manual calendar entry.

#### Acceptance Criteria

1. WHEN a user invokes the CLI with default parameters THEN the CLI Tool SHALL create a calendar event in Outlook Calendar with preset values
2. WHEN a user provides command-line flags THEN the CLI Tool SHALL override default parameters with the provided values using the following flags:
   - `-b, --busy-status <free|tentative|busy|out-of-office>`: Availability status (default: busy)
   - `-c, --category <name>`: Category name (default: none)
   - `-d, --duration <minutes>`: Duration in minutes, alternative to end-time (default: 60)
   - `--dry-run`: Validate parameters and display what would be created without creating the event
   - `-e, --end-time <time>`: End time in ISO 8601 format (default: 1 hour after start time)
   - `-h, --help`: Display usage information, available commands, and flag descriptions
   - `-n, --notes <text>`: Additional notes for the event body
   - `-p, --private <true|false>`: Mark event as private (default: true)
   - `-r, --reminder <minutes>`: Reminder time in minutes before event, 0 for none (default: 0)
   - `-s, --subject <text>`: Calendar event subject (default: "[Jira Key] descriptive-subject" extracted from branch name)
   - `--sensitivity <normal|personal|private|confidential>`: Sensitivity level (default: normal)
   - `-t, --start-time <time>`: Start time in ISO 8601 format (default: current time)
   - `-v, --version`: Display the version number
3. WHEN the CLI is invoked with `-h` or `--help` THEN the CLI Tool SHALL display usage information, available commands, and flag descriptions
4. WHEN the CLI is invoked with `-v` or `--version` THEN the CLI Tool SHALL display the version number
5. WHEN the CLI is invoked with `--dry-run` THEN the CLI Tool SHALL validate parameters and display what would be created without creating the event
6. WHEN creating a calendar event THEN the CLI Tool SHALL apply the following default values: subject ("[Jira Key] descriptive-subject"), start time (current time), end time (1 hour after start time), is private (true), reminder (none), busy status (busy), category (none), sensitivity (normal)
7. WHEN an error occurs THEN the CLI Tool SHALL log the timestamp, command executed, and error message to std_err.csv in the Log Directory
8. WHEN Outlook is not running or accessible THEN the CLI Tool SHALL log an error to std_err.csv and return a non-zero exit code
9. WHEN the Log Directory does not exist THEN the CLI Tool SHALL create it in the project root

### Requirement 2

**User Story:** As a developer, I want to manage ticket tracking and view tracking status, so that I can control which branches are monitored for automatic time logging.

#### Acceptance Criteria

1. WHEN a user invokes the track command THEN the CLI Tool SHALL extract the Jira Key from the current Git branch name using regex pattern matching
2. WHEN a Jira Key is successfully extracted THEN the CLI Tool SHALL register the ticket as a Tracked Ticket in the State File
3. WHEN a Jira Key cannot be extracted from the branch name THEN the CLI Tool SHALL report an error to std_err.csv and not create a Tracked Ticket
4. WHEN a user invokes the track command THEN the CLI Tool SHALL associate the current Git branch as the Linked Branch for the Tracked Ticket
5. WHEN a user invokes the stop-tracking command THEN the CLI Tool SHALL remove the Tracked Ticket from the State File
6. WHEN a user invokes the track command for an already Tracked Ticket THEN the CLI Tool SHALL update the Linked Branch association
7. WHEN a user invokes the list-tracked command THEN the CLI Tool SHALL display all currently Tracked Tickets with their Linked Branches
8. WHEN a user invokes the status command THEN the CLI Tool SHALL display the current Session information including branch name, Jira Key, and start time
9. WHEN a user invokes the status command and no Session is active THEN the CLI Tool SHALL display a message indicating no active session
10. WHEN a user invokes the reset command THEN the CLI Tool SHALL clear all Tracked Tickets and active Sessions from the State File
11. THE State File SHALL persist the mapping between Tracked Tickets, their Linked Branches, and active Sessions

### Requirement 3

**User Story:** As a developer, I want to install and manage Git hooks, so that calendar events are automatically created and updated for tracked branches.

#### Acceptance Criteria

1. WHEN a user invokes the install-hooks command THEN the CLI Tool SHALL create post-checkout and post-commit hook scripts in the .git/hooks directory
2. WHEN a user invokes the uninstall-hooks command THEN the CLI Tool SHALL remove the post-checkout and post-commit hook scripts from the .git/hooks directory
3. WHEN a user invokes the install-hooks command and hooks already exist THEN the CLI Tool SHALL prompt for confirmation before overwriting
4. WHEN a user checks out a Linked Branch THEN the Git Hook SHALL invoke the CLI Tool to create a new Calendar Event with a 1-hour duration
5. WHEN a user commits changes on a Linked Branch THEN the Git Hook SHALL invoke the CLI Tool to update the existing Calendar Event end time to the current time
6. WHEN a user commits changes on a Linked Branch multiple times THEN the Git Hook SHALL extend the existing Calendar Event end time to the current time
7. WHEN a Git hook executes THEN the Git Hook SHALL pass the branch name and operation type to the CLI Tool
8. WHEN a Git hook fails THEN the Git Hook SHALL not prevent the Git operation from completing
9. WHEN a user checks out a non-tracked branch THEN the Git Hook SHALL not create a Calendar Event

### Requirement 4

**User Story:** As a developer, I want the system to handle duplicate checkouts gracefully, so that I don't create redundant calendar events.

#### Acceptance Criteria

1. WHEN a Linked Branch is checked out and a Session for that branch is already active THEN the CLI Tool SHALL not create a new Calendar Event
2. WHEN a Linked Branch is checked out and a Session for that branch is already active THEN the CLI Tool SHALL log the duplicate checkout to std_err.csv
3. THE State File SHALL track whether a Session is currently active for each Tracked Ticket

### Requirement 5

**User Story:** As a developer, I want Jira ticket IDs and descriptive subjects extracted from branch names, so that calendar events have meaningful titles.

#### Acceptance Criteria

1. THE CLI Tool SHALL recognize Jira key patterns matching the format: exactly three uppercase letters, hyphen, exactly four digits
2. WHEN creating a Calendar Event for a Tracked Ticket THEN the CLI Tool SHALL extract the descriptive subject from the branch name by removing the Jira Key pattern
3. WHEN the branch name contains a prefix (e.g., "feature/", "bugfix/") THEN the CLI Tool SHALL preserve the prefix in the extracted subject
4. WHEN creating a Calendar Event THEN the CLI Tool SHALL use the format "[Jira Key] subject" as the Calendar Event subject
5. WHEN a user provides the `--subject` flag THEN the CLI Tool SHALL use the provided subject instead of extracting from the branch name

### Requirement 6

**User Story:** As a developer, I want the CLI tool to be easily installable and accessible, so that I can use it from any directory in my terminal.

#### Acceptance Criteria

1. THE CLI Tool SHALL be executable from any directory in the terminal
2. THE CLI Tool SHALL provide installation instructions for adding to system PATH
3. WHEN the CLI is invoked with a help flag THEN the CLI Tool SHALL display usage information and available options
4. THE CLI Tool SHALL provide clear error messages when invoked incorrectly

### Requirement 7

**User Story:** As a developer, I want the system to handle edge cases gracefully, so that unexpected situations don't cause data loss or system errors.

#### Acceptance Criteria

1. WHEN the State File is corrupted or invalid THEN the CLI Tool SHALL handle the error gracefully and allow recovery
2. WHEN Outlook Calendar API calls fail THEN the CLI Tool SHALL retry once before reporting failure
3. WHEN concurrent sessions are detected THEN the CLI Tool SHALL resolve conflicts by prioritizing the most recent session
4. WHEN system time changes occur THEN the CLI Tool SHALL handle time calculations correctly
5. WHEN the State File directory does not exist THEN the CLI Tool SHALL create it automatically
6. WHEN a commit occurs without an active Session THEN the CLI Tool SHALL not attempt to update any Calendar Event
