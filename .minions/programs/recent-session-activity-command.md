---
id: recent-session-activity-command
target_repos:
  - cli
acceptance_criteria:
  - "'partio activity' lists up to 10 recent checkpoints by default"
  - "'partio activity --limit N' shows up to N checkpoints"
  - "Each entry displays: checkpoint ID, short commit hash, branch, agent, attribution percent, and timestamp"
  - "'partio activity --json' outputs a JSON array of checkpoint metadata objects"
  - "Command exits gracefully with a helpful message if no checkpoints exist yet"
  - "Command must be run inside a git repository or returns a clear error"
pr_labels:
  - minion
---

# Add `partio activity` command showing recent session activity

Implement a `partio activity` command that reads the checkpoint orphan branch (`partio/checkpoints/v1`) and lists recent checkpoints in reverse chronological order. Each entry should display: checkpoint ID, commit hash (short), branch, agent name, agent attribution percentage, creation timestamp, and plan slug if present.

Support a `--limit` flag (default 10) to control how many entries are shown. Also add a `--json` flag for machine-readable output.

## Why

Currently there is no way to browse what checkpoints have been captured for a repository. Users have no visibility into recent AI-assisted commits without manually inspecting the orphan branch.

## User relevance

Users can quickly audit which recent commits had AI agent involvement, review attribution percentages, and locate specific checkpoints for rewind operations.

## Context hints

- `cmd/partio/` — new command file
- `internal/checkpoint/` — reading checkpoint data from orphan branch
