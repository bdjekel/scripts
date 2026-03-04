# jira-ticket

A Go CLI tool that creates Jira Epics and Stories from structured `requirements.md` files using `jira-cli`.

## Prerequisites

1. **jira-cli** must be installed and authenticated:
   ```bash
   brew install jira-cli
   jira init
   ```

2. **Environment variables** (or `.env` file):
   ```bash
   export JIRA_PROJECT_KEY=PROJ
   export JIRA_SERVER_URL=https://your-company.atlassian.net
   ```

## Installation

```bash
go install github.com/user/jira-ticket/cmd/jira-ticket@latest
```

Or build from source:
```bash
go build -o jira-ticket ./cmd/jira-ticket
```

## Usage

```bash
jira-ticket [options] <requirements.md>
```

### Options

- `--help` - Display usage information
- `--dry-run` - Parse only, don't create tickets
- `--verbose` - Enable detailed logging
- `--on-duplicate` - Action on duplicate: `skip` (default) or `fail`

### Examples

```bash
# Preview what would be created
jira-ticket --dry-run requirements.md

# Create tickets with verbose output
jira-ticket --verbose requirements.md

# Fail on duplicates instead of skipping
jira-ticket --on-duplicate=fail requirements.md
```

## Requirements File Format

The tool parses markdown files with this structure:

```markdown
# Epic Title

Optional epic description.

**Priority:** High
**Assignee:** username

## Story Title

Story description and context.

**Priority:** Medium
**Assignee:** another-user

### Acceptance Criteria

- Criterion 1
- Criterion 2
- Criterion 3

## Another Story

...

# Another Epic

...
```

### Hierarchy Mapping

- `#` headings → Jira Epics
- `##` headings → Jira Stories (linked to parent Epic)
- `### Acceptance Criteria` → Included in Story description

### Optional Fields

- `**Priority:**` - Maps to Jira priority (Highest, High, Medium, Low, Lowest)
- `**Assignee:**` - Jira username for assignment

## Exit Codes

- `0` - All tickets created or skipped successfully
- `1` - One or more tickets failed to create

## Development

```bash
# Run tests
go test ./...

# Build
go build ./cmd/jira-ticket

# Run with dry-run
./jira-ticket --dry-run example-requirements.md
```
