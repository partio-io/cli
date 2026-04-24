---
id: recap-activity-summary
target_repos:
  - cli
acceptance_criteria:
  - "`partio recap` command exists and is registered in the root command"
  - "Reads checkpoint data from the orphan branch to compute activity stats"
  - "Displays summary of recent agent activity: number of checkpoints, commits linked, agents detected"
  - "Supports `--days` flag to control the lookback window (default 7)"
  - "Supports `--json` flag for machine-readable output"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add `partio recap` command for agent activity summary

Add a `partio recap` command that reads checkpoint data from the `partio/checkpoints/v1` branch and displays a summary of recent agent activity in the repository.

## What it should show

- Total checkpoints created in the time window
- Number of commits linked to agent sessions
- Breakdown by detected agent (e.g., "claude-code: 12 checkpoints")
- Time range covered

## Implementation notes

- Read checkpoint commits from the orphan branch using git log on `partio/checkpoints/v1`
- Parse checkpoint metadata to extract agent info and timestamps
- Follow the existing command pattern in `cmd/partio/` (one file per command)
- Default to 7-day lookback; `--days N` overrides
- `--json` outputs structured JSON instead of human-readable text
- No external dependencies beyond what's already used
