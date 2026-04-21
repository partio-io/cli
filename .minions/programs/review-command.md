---
id: review-command
target_repos:
  - cli
acceptance_criteria:
  - A `partio review` command exists and is registered in the CLI
  - The command captures the review session as a checkpoint on the orphan branch, linked to the current HEAD commit
  - The checkpoint metadata includes a `kind: review` field distinguishing it from regular commit checkpoints
  - Running `partio status` after a review shows that a review checkpoint exists for the current commit
  - A table-driven test verifies the command creates a review checkpoint
pr_labels:
  - minion
---

# Add `partio review` command for AI-assisted code review with checkpoint capture

## Problem

Partio captures the *why* behind code changes during commits, but the review process — where an AI agent analyzes code for issues, suggests improvements, and the author iterates — produces equally valuable reasoning context that is currently lost. There is no way to checkpoint a review session or distinguish it from a commit session.

## Desired behavior

Add a `partio review` command that:

1. Detects the active Claude Code session (using the existing detector interface)
2. Creates a checkpoint on the orphan branch tagged with `kind: review` in its metadata, linked to the current HEAD commit
3. Stores the session transcript and attribution data just like post-commit checkpoints do

This gives teams a Git-native audit trail of AI-assisted reviews alongside the commit checkpoints, without requiring any external review platform integration.

## Context hints

- `cmd/partio/` — existing command structure to follow
- `internal/checkpoint/` — checkpoint creation and storage
- `internal/agent/` — agent detection interface
- `internal/session/` — session state management
- `internal/hooks/post_commit.go` — reference for how commit checkpoints are created
