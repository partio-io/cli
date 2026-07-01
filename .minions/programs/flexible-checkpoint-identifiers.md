---
id: flexible-checkpoint-identifiers
target_repos:
  - cli
acceptance_criteria:
  - "`partio rewind --to` accepts both a checkpoint ID and a commit SHA, resolving the appropriate checkpoint"
  - "When given a commit SHA, it finds the checkpoint associated with that commit via the Entire-Checkpoint trailer or checkpoint metadata"
  - "When given an ambiguous identifier, checkpoint ID takes precedence"
  - "Clear error message when no checkpoint is found for the given identifier"
  - "Works with both full and abbreviated SHAs (minimum 7 characters)"
pr_labels:
  - minion
  - enhancement
---

# Accept commit SHA or checkpoint ID as flexible identifiers in rewind

## Summary

Allow `partio rewind --to` to accept either a checkpoint ID or a commit SHA, automatically resolving the correct checkpoint. This reduces friction when users want to rewind but only have the commit hash from `git log`.

## Motivation

Users commonly have a commit SHA from `git log` or GitHub but need to look up the corresponding checkpoint ID before they can rewind. This extra step is unnecessary — Partio already stores the mapping between commits and checkpoints via trailers. Accepting either identifier makes the command more intuitive.

Source: entireio/cli changelog 0.5.6 — "`entire explain` accepts checkpoint ID or commit SHA as positional argument"

## Implementation Notes

- In `runRewindTo`, attempt to parse the identifier:
  1. First, try to resolve as a checkpoint ID directly on the orphan branch
  2. If not found, try to resolve as a commit SHA by reading the commit's `Partio-Checkpoint` trailer
  3. If the commit exists but has no trailer, return a clear error
- Consider also accepting identifiers as a positional argument (in addition to `--to` flag) for ergonomics
- Reuse existing `checkpoint.Store` methods for resolution
- Keep backward compatibility — existing checkpoint ID usage must continue to work

## Context

- `cmd/partio/rewind.go` — current rewind command (uses `--to` flag)
- `internal/checkpoint/` — checkpoint storage and resolution
- `internal/git/` — git operations for reading commit trailers
