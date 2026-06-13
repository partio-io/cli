---
id: batch-checkpoint-reconciliation
target_repos:
  - cli
acceptance_criteria:
  - A `partio reconcile` (or `partio doctor --reconcile`) command processes commits missing checkpoints
  - Batch mode processes multiple commits in a single pass without repeated ref updates
  - Performance scales sub-linearly with commit count (avoid N sequential ref updates)
  - Commits without active agent sessions are skipped gracefully
  - Dry-run flag shows what would be reconciled without making changes
pr_labels:
  - minion
  - enhancement
---

# Add batch checkpoint reconciliation for missed commits

Implement a command that catches up on commits that were made while Partio hooks were disabled, broken, or timed out — processing them efficiently in batch.

## Motivation

When hooks fail (timeout, misconfiguration, agent not detected) or Partio is temporarily disabled, commits accumulate without checkpoint metadata. Currently there's no way to backfill these gaps. Entireio/cli 0.6.3 fixed a reconciliation performance issue (66s → 7s for 50 commits) by batching operations, showing this is a real user need.

Inspired by entireio/cli changelog 0.6.3 (checkpoint metadata reconciliation performance fix).

## Implementation Notes

- Walk `git log` from HEAD backwards, identify commits lacking `Partio-Checkpoint` trailers
- For each, attempt to find a matching session (by timestamp/tree hash correlation)
- Batch git plumbing operations: write all blobs/trees first, then create checkpoint commits in one pass, then single ref update
- Add `--since` and `--limit` flags to scope the reconciliation window
- Report summary: reconciled N commits, skipped M (no session found)
