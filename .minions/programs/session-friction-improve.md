---
id: session-friction-improve
target_repos:
  - cli
acceptance_criteria:
  - "partio improve analyzes recent session transcripts and identifies recurring friction patterns (repeated errors, retries, misunderstandings)"
  - "partio improve generates specific suggestions for CLAUDE.md improvements with evidence excerpts from transcripts"
  - "Suggestions include unified diffs showing proposed changes to context files"
  - "Analysis works offline using locally stored checkpoint data on the orphan branch"
  - "Command exits cleanly with a helpful message when no sessions or insufficient data exist"
pr_labels:
  - minion
---

# Add `partio improve` command for AI-powered context file improvement suggestions

## Description

Add a `partio improve` command that analyzes captured session transcripts from checkpoints to identify recurring friction patterns (repeated errors, tool-call retries, misunderstandings, wasted turns) and generates actionable suggestions for improving context files like `CLAUDE.md`.

The command should use a two-phase approach:
1. **Index phase**: Scan checkpoint session data to identify recurring friction themes (e.g., agent repeatedly hitting the same linting error, misunderstanding project conventions, retrying failed approaches).
2. **Suggest phase**: Read relevant transcript excerpts and generate specific improvement suggestions — with evidence quotes and proposed diffs — for context files that would prevent the friction in future sessions.

This leverages Partio's unique position of having captured session transcripts to close the feedback loop: sessions reveal what the agent struggled with, and those struggles inform better prompts/context for future sessions.

## Why

Partio already captures the full reasoning behind code changes. This feature turns that captured data into direct value by helping users write better CLAUDE.md files and project instructions. Without this, users must manually review session logs to identify patterns — a tedious process that rarely happens. Automated friction analysis makes the captured session data actively useful rather than purely archival.

## Source

- **Origin:** entireio/cli#765 (PR: Agent Improvement Engine)
- **Detected from:** `entireio-cli-pulls`
