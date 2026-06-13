---
id: interactive-checkpoint-browser
target_repos:
  - cli
acceptance_criteria:
  - "New `partio summary` command exists and is registered in the root command"
  - "Command reads checkpoint data from the orphan branch using git plumbing"
  - "Output lists checkpoints with commit hash, date, agent status, and prompt snippet"
  - "Supports `--since` flag to filter by time period (e.g., 24h, 7d, 30d)"
  - "Supports `--branch` flag to filter by branch (defaults to current branch)"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add `partio summary` command for browsing checkpoint history

Add a `partio summary` command that lets users browse their checkpoint history with filtering options.

## Context

Users currently have no way to browse their checkpoint history from the CLI. The checkpoint data exists on the orphan branch but requires manual git commands to inspect. A summary command would make this data accessible.

## Implementation

- Add `cmd/partio/summary.go` with a new Cobra command
- Read checkpoint commits from the `partio/checkpoints/v1` branch using git log/plumbing
- Parse checkpoint metadata (commit hash, timestamp, agent detection state, prompt)
- Display as a formatted table or list to stdout
- Add `--since` flag accepting durations like `24h`, `7d`, `30d` (default: `7d`)
- Add `--branch` flag to filter checkpoints linked to a specific branch
- Keep output simple and non-interactive for v1 (no TUI dependency)
