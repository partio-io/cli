---
id: checkpoint-retention-prune
target_repos:
  - cli
acceptance_criteria:
  - A `partio clean --prune` subcommand removes checkpoint tree entries older than `retention_days` from the orphan branch
  - "`retention_days` is configurable via `.partio/settings.json` and defaults to 0 (disabled)"
  - Pruning uses git plumbing (no checkout) consistent with the existing store implementation
  - A dry-run flag (`--dry-run`) prints which checkpoints would be removed without modifying the branch
  - If no checkpoints fall outside the retention window, the command exits cleanly with no changes
  - Unit tests verify that only out-of-window entries are removed from a constructed test tree
pr_labels:
  - minion
---

# Retention-based auto-prune for the checkpoint branch

## Description

Add a `partio clean --prune` command (or extend the existing clean command) that removes checkpoints older than a configurable retention window from the orphan branch using git plumbing. Add a `retention_days` field to config. When pruning, rebuild the checkpoint branch tree excluding entries whose `created_at` falls outside the retention window, then force-update the ref.

## Why

The checkpoint branch grows unboundedly with every commit. Without pruning, long-lived repos accumulate large orphan branches that slow pushes and bloat remote storage — especially since full JSONL session data is stored per checkpoint.

## User Relevance

Users working in active repos can keep checkpoint storage under control without manually deleting the branch or disabling partio. A sane default retention (e.g. 90 days) prevents surprise storage growth.

## Context Hints

- `cmd/partio/clean.go`
- `internal/checkpoint/store.go`
- `internal/checkpoint/read.go`
- `internal/config/config.go`
- `internal/config/defaults.go`

## Source

entireio/cli#1260, entireio/cli#1160
