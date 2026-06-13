---
id: dry-run-destructive-operations
target_repos:
  - cli
acceptance_criteria:
  - "`partio rewind --dry-run` prints what would be restored without modifying any files or refs"
  - "Dry-run output includes the target checkpoint ID, affected files, and session summary"
  - "No git objects are written and no refs are updated during dry-run"
  - "Flag is documented in command help text"
  - "Tests verify dry-run produces output without side effects"
pr_labels:
  - minion
---

# Add `--dry-run` flag to destructive checkpoint operations

Add a `--dry-run` flag to `partio rewind` (and future destructive commands) that previews what would happen without actually making changes.

## Motivation

`partio rewind` restores working directory state from a checkpoint, which is a destructive operation that overwrites current files. Users need a way to preview what would change before committing to it. This is especially important when browsing unfamiliar checkpoint history where picking the wrong checkpoint could discard work.

Inspired by entireio/cli#1141 which adds `--dry-run` to migration commands.

## Desired behavior

- `partio rewind --dry-run <checkpoint-id>` outputs:
  - The checkpoint being targeted (ID, timestamp, associated commit)
  - The session context (agent, prompt summary if available)
  - A diff summary of what files would change (similar to `git diff --stat`)
- No files are modified, no refs updated, no git objects written
- Exit code 0 on success, non-zero if the checkpoint is invalid or not found
- The flag follows Go CLI conventions: `--dry-run` with a short alias `-n`

## Context hints

- `cmd/partio/rewind.go` — rewind command implementation
- `internal/checkpoint/` — checkpoint storage and retrieval
