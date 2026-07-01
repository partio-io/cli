---
id: per-file-attribution-metadata
target_repos:
  - cli
acceptance_criteria:
  - "`attribution.Result` gains a `Files []FileAttribution` field with per-file added/deleted line counts"
  - "`attribution.Calculate` populates per-file data from the existing `git diff --numstat` output without additional git calls"
  - "`checkpoint.SessionMetadata` stores the per-file attribution slice"
  - Existing aggregate `AgentPercent` field is preserved for backward compatibility
  - Unit tests cover the per-file parsing from numstat output including binary files (which report `-`)
pr_labels:
  - minion
---

# Per-file attribution in checkpoint metadata

## Description

Extend the attribution calculation to produce per-file line counts (agent vs human) and store them in checkpoint metadata. The `attribution.Result` and `SessionMetadata` types should be extended with a `Files` map. `attribution.Calculate` already calls `git.DiffNumstat` which returns per-file data — that data should be preserved rather than summed. `checkpoint.SessionFiles.Metadata` should carry the per-file breakdown.

## Why

The current binary 0%/100% attribution is a known limitation. Per-file granularity provides meaningful data for repos with mixed human/agent contributions across different files in the same commit.

## User Relevance

Teams reviewing AI-assisted code want to know which specific files were agent-written vs human-edited, not just an aggregate percentage for the whole commit.

## Context Hints

- `internal/attribution/calculate.go`
- `internal/attribution/attribution.go`
- `internal/checkpoint/metadata.go`
- `internal/checkpoint/store.go`
- `internal/hooks/postcommit.go`

## Source

entireio/cli#1258, entireio/cli#1103
