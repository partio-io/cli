---
id: enable-import-progress
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable` prints at least one progress line per session imported when stdout is a TTY"
  - "Output appears incrementally (not buffered until end) so the terminal is never silent for more than a few seconds during import"
  - "Non-TTY (piped/CI) mode still works and prints a final summary line after import completes"
  - "Progress output goes to stderr so it does not corrupt stdout pipes"
pr_labels:
  - minion
  - enhancement
---

# Show live progress during `partio enable` session import

## Source

Inspired by entireio/cli#1847 (issue) and entireio/cli#1848 (PR: feat(enable): show progress while importing existing sessions).

## Description

When `partio enable` is run for the first time in a repo that already has Claude Code session history under `~/.claude/projects/`, the import phase currently produces no output — the terminal appears hung. On repos with long or many Claude Code sessions, this silence can last for minutes or more.

Add incremental progress output during session discovery and checkpoint creation:

- **Interactive (TTY):** Print a live-updating line (or spinner) as each session is processed, e.g.:
  ```
  [partio] Importing session abc123... done (47 turns)
  [partio] Importing session def456... done (12 turns)
  [partio] Imported 2 sessions (59 turns total)
  ```
- **Non-interactive (no TTY, piped, CI):** Buffer is flushed after each session; a summary line is always printed at the end so automated scripts can confirm completion.

All progress output goes to stderr so stdout pipes remain unaffected.

## Why This Matters

First-run import is the first impression Partio makes after installation. Silent operation during a long import looks like a hang, erodes trust, and causes users to kill the process — aborting the import midway and leaving the checkpoint branch incomplete. This is the same problem that drove entireio/cli to add the same feature in their v0.8.1+ cycle.

## User Relevance

Any user running `partio enable` for the first time on a repo with pre-existing Claude Code sessions will experience the silent hang. This is the typical adoption path for new Partio users.

## Context Hints

- `cmd/partio/` — `enable` command implementation
- `internal/session/` — session lifecycle, discovery
- `internal/checkpoint/` — checkpoint creation from sessions
- `internal/agent/claude/` — JSONL parsing, session enumeration
