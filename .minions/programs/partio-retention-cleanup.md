---
id: partio-retention-cleanup
target_repos:
  - cli
acceptance_criteria:
  - "`partio clean` removes checkpoints older than the configured retention period (default: 90 days)"
  - "`partio clean --dry-run` lists checkpoints that would be removed without modifying the branch"
  - "Retention settings configurable via `retention_days` and `retention_count` in layered config"
  - "At least one checkpoint per session is always preserved regardless of age (safety net)"
  - "The orphan branch ref is updated atomically via update-ref after cleanup"
pr_labels:
  - minion
  - minion-proposal
---

# Retention-based checkpoint cleanup with configurable policies

## Summary

Extend `partio clean` to support retention-based cleanup: remove checkpoints older than a configurable age (e.g., 90 days) or exceeding a count limit, while preserving the most recent N checkpoints per session. The orphan branch (`partio/checkpoints/v1`) grows unboundedly as checkpoints accumulate — without retention cleanup, users must manually delete the branch and lose all history.

## Why

The orphan branch grows unboundedly as checkpoints accumulate. Over time this inflates repo size, slows clones, and wastes storage on remotes. Without retention cleanup, users will eventually disable Partio to control repo size. A retention policy lets users keep recent/relevant checkpoints while reclaiming space.

## User Relevance

Long-lived repositories will accumulate hundreds of checkpoints with full JSONL transcripts. Without cleanup, users will eventually disable Partio to control repo size. Retention-based cleanup makes Partio sustainable for long-term use.

## Context Hints

- `cmd/partio/clean.go` currently only lists entries — it needs to be extended to read metadata from each checkpoint, filter by age/count, and rebuild the tree without expired entries.
- Checkpoint metadata includes `created_at` in `internal/checkpoint/checkpoint.go`.
- Config should gain `retention_days` and `retention_count` fields in `config.Config`.
- Cleanup should rewrite the orphan branch tree (similar to how `write.go` builds trees with `addToTree`) minus the expired shard entries, then `commit-tree` and `update-ref`.
- Consider a `--dry-run` flag that reports what would be removed without modifying anything.

## Source

Inspired by entireio/cli PR #972 (retention-based cleanup for v2 transcript generations) and the `entire clean` command replacing the deprecated `entire reset`.
