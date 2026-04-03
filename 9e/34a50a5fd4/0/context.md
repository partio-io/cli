# Session Context

## User Prompts

### Prompt 1

# Minion Task: Detect external hook managers during enable

You are a coding agent executing a task autonomously. Complete the task in a single session without human interaction.

## Implementation Plan

Follow this plan that was reviewed and approved:

The plan is ready for review. Key design decisions:

- **Detection lives in `internal/git/hooks/detect.go`** — follows the "one concern per file" convention and keeps it close to the hook installation code it relates to.
- **Warning placement ...

