---
id: fix-enable-re-enable-when-disabled
target_repos:
  - cli
acceptance_criteria:
  - Running `partio enable` when `.partio/` exists but `settings.json` has `enabled: false` re-enables Partio and reinstalls hooks
  - Running `partio enable` when `.partio/` exists and is already enabled prints the current "already enabled" message unchanged
  - A table-driven test covers the disabled-config-exists scenario
pr_labels:
  - minion
---

# Fix `partio enable` to re-enable when config already exists in disabled state

## Problem

`partio enable` checks for the existence of `.partio/` and short-circuits with "partio is already enabled" even when `settings.json` has `enabled: false`. Users who ran `partio disable` (which may leave `.partio/` on disk) or whose config ended up in a disabled state must manually run `partio disable && partio enable` to recover.

## Desired behavior

When `.partio/` exists, `partio enable` should read `settings.json` and check the `enabled` field:
- If already enabled → print "partio is already enabled" (current behavior)
- If disabled → flip `enabled` to `true`, reinstall hooks, and print a confirmation message

This removes the known limitation documented in CLAUDE.md and matches user expectations from `--help` which says the command "enables partio."

## Context hints

- `cmd/partio/enable.go` — the short-circuit check at line ~39
- `internal/config/` — config loading and defaults
- `cmd/partio/disable.go` — reference for the inverse flow
