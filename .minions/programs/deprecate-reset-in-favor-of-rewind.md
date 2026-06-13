---
id: deprecate-reset-in-favor-of-rewind
target_repos:
  - cli
acceptance_criteria:
  - "`partio reset` prints a deprecation warning to stderr before executing: `Warning: 'partio reset' is deprecated and will be removed in a future release. Use 'partio rewind' instead.`"
  - "The command still functions after the warning so existing scripts are not broken"
  - "`partio reset` is hidden from the default help output (marked as deprecated in cobra)"
  - "The deprecation warning is tested in `cmd/partio/reset_test.go`"
pr_labels:
  - minion
---

# Deprecate `partio reset` in favor of `partio rewind`

## Source

entireio/cli changelog 0.5.3: "Deprecated `entire reset` command in favor of `entire rewind`"

## Problem

`partio reset` is a destructive command that deletes and recreates the entire checkpoint branch, wiping all stored checkpoint data. `partio rewind` is the safer alternative — it lists checkpoints and restores to a specific one without discarding history. Having both commands creates confusion about which to use, and `reset` is easy to run accidentally with severe consequences.

## Proposed Change

Mark `partio reset` as deprecated in Cobra so it:

1. Is hidden from `partio --help` output (users discover it only if they explicitly run it or know about it)
2. Prints a deprecation warning to stderr when invoked: `Warning: 'partio reset' is deprecated and will be removed in a future release. Use 'partio rewind' instead.`
3. Still executes its current logic so existing scripts continue to work during the transition period

The implementation touches only `cmd/partio/reset.go`. Cobra supports `Deprecated` field on commands which automatically prints the warning and hides from help.

## Why This Matters

Surfacing the deprecation now sets expectations before `reset` is eventually removed. Users who discover `reset` (e.g. from old blog posts or AI suggestions) get a clear redirect to `rewind`, reducing the risk of accidental checkpoint data loss.

## Context

- `cmd/partio/reset.go` — current reset implementation
- `cmd/partio/rewind.go` — the preferred replacement command
