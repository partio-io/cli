---
id: json-output-for-status
target_repos:
  - cli
acceptance_criteria:
  - "`partio status --json` outputs valid JSON with fields: repo_root, branch, enabled, hooks_installed, checkpoint_branch_exists, strategy, agent"
  - "`partio status` without --json still outputs the existing human-readable format unchanged"
  - "JSON output exits 0 even when Partio is not enabled (enabled field is false)"
  - "A unit test covers JSON serialization of the status struct"
pr_labels:
  - minion
---

# Add `--json` flag to `partio status` command

Add a `--json` flag to `partio status` that emits a machine-readable JSON object containing: repo root, branch, enabled state, hooks installation status, checkpoint branch existence, strategy, and agent name. Human-readable output remains the default.

## Why

Enables tooling, CI scripts, and future dashboard/app integrations to query Partio state programmatically without parsing human-readable text.

## User Relevance

Power users and automation scripts can check Partio health and state without screen-scraping. Opens the door for the Partio app (Next.js dashboard) to query local CLI state.

## Context Hints

- `cmd/partio/status.go`
- `internal/config/config.go`
- `internal/git/repo.go`

## Source

Inspired by entireio/cli#975
