---
inclusion: fileMatch
fileMatchPattern: "**/.kiro/specs/**/requirements.md"
description: "Jira integration guidance for spec requirements"
---

# Spec Requirements → Jira Structure

When writing or reviewing requirements in a Kiro spec, structure them for Jira import using the jira-ticket tool.

## Hierarchy Mapping

- `#` top-level headings → Jira Epics
- `##` second-level headings → Jira Stories (linked to parent Epic)
- Acceptance Criteria → Story description content

## Writing Guidelines

- Each `#` heading represents a cohesive Epic (logical grouping of work)
- Keep Epic names concise — they become Epic titles in Jira
- Include an Epic description immediately after the `#` heading to provide context
- `##` headings under an Epic become Stories linked to that Epic
- User Stories should follow "As a [role], I want [goal] so that [benefit]"
- Acceptance Criteria are included in the Story description as bullet points

## Required Structure

```markdown
# Epic Name

## Story Title

**User Story:** As a [role], I want [goal] so that [benefit].

**Priority:** High
**Assignee:** username

### Acceptance Criteria

1. WHEN [condition], THE [Subject] SHALL [action]
2. THE [Subject] SHALL [requirement]
```

## Optional Fields

- `**Priority:**` — Maps to jira-cli `-y` flag. Values: `Highest`, `High`, `Medium`, `Low`, `Lowest`
- `**Assignee:**` — Maps to jira-cli `-a` flag. Use Jira username or `$(jira me)` for self-assignment

## Example

```markdown
# User Authentication

## Login with Email

**User Story:** As a user, I want to log in with my email and password so that I can access my account.

**Priority:** High
**Assignee:** jdoe

### Acceptance Criteria

1. WHEN valid credentials are provided, THE system SHALL authenticate the user
2. WHEN invalid credentials are provided, THE system SHALL display an error message
3. THE system SHALL lock the account after 5 failed attempts

## Password Reset

**User Story:** As a user, I want to reset my password so that I can regain access if I forget it.

**Priority:** Medium

### Acceptance Criteria

1. WHEN a reset is requested, THE system SHALL send a reset link to the registered email
2. THE reset link SHALL expire after 24 hours
```

## Notes

- The jira-ticket tool parses this structure and creates corresponding Jira tickets
- Running the tool multiple times is safe — duplicates are skipped by default
- Use `--dry-run` to preview what tickets would be created without actually creating them

## Related Knowledge Base

For jira-cli commands and authentication setup:
```
/knowledge search coding-context jira
```

Or reference directly: `~/ai_agent/assistant/coding/stack/jira.md`