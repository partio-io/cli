---
id: time-range-checkpoint-filters
target_repos:
  - cli
acceptance_criteria:
  - "`partio rewind --since 24h` returns only checkpoints created within the last 24 hours"
  - "`partio rewind --until 2024-01-01T00:00:00Z` returns only checkpoints created before that timestamp"
  - Both flags can be combined to express a closed time range
  - Relative durations (`24h`, `7d`, `30d`) and RFC3339 absolute timestamps are both accepted
  - Invalid flag values produce a descriptive error and non-zero exit code
  - Unit tests cover filtering logic with a set of checkpoints at known timestamps
pr_labels:
  - minion
---

# Time-range filters for checkpoint listing

## Description

Add `--since` and `--until` flags to `partio rewind` (and any future checkpoint list command) that accept RFC3339 or relative time strings (e.g. `24h`, `7d`). The checkpoint read path should filter entries by comparing `created_at` in the metadata against the provided bounds before returning results.

## Why

As checkpoints accumulate, users need a way to scope queries to a relevant time window. This is a foundational feature for usability that pairs naturally with the retention-prune feature and makes `partio rewind` tractable in active repositories.

## User Relevance

Users can quickly find checkpoints from a specific sprint, day, or session window without scrolling through an unbounded list, making the rewind workflow practical for real codebases.

## Context Hints

- `cmd/partio/rewind.go`
- `internal/checkpoint/read.go`
- `internal/checkpoint/checkpoint.go`

## Source

entireio/cli#1241
