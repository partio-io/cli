---
id: linear-checkpoint-history
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint branch alignment with remote uses rebase instead of merge"
  - "Checkpoint history remains linear (no merge commits on the checkpoint branch)"
  - "Pre-push hook handles rebase conflicts gracefully without blocking the user's push"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Use rebase for linear checkpoint branch history

## Problem

When pushing the checkpoint branch (`partio/checkpoints/v1`) to a remote that has diverged, the current approach may create merge commits. This produces a non-linear history on the checkpoint branch, making it harder to browse and reason about checkpoint data.

## Solution

Replace merge-based alignment with rebase when syncing the local checkpoint branch with its remote counterpart. This keeps the checkpoint history linear and clean. Handle rebase conflicts gracefully — if a conflict occurs, abort the rebase and warn the user rather than blocking their push operation.

## Context

- Inspired by entireio/cli#863
- Relevant code: `internal/hooks/` (pre-push implementation), `internal/git/` (branch operations)

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
