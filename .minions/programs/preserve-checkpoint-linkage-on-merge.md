---
id: preserve-checkpoint-linkage-on-merge
target_repos:
  - cli
acceptance_criteria:
  - After a fast-forward merge, checkpoints linked to feature branch commits remain discoverable from the target branch
  - partio status on the target branch shows checkpoint history from merged feature branches
  - No checkpoint data is lost or orphaned during fast-forward or regular merges
  - A post-merge hook or equivalent mechanism ensures checkpoint refs are updated
pr_labels:
  - minion
---

# Preserve checkpoint linkage across fast-forward merges

## Problem

When a feature branch with checkpoint trailers is merged via fast-forward into the target branch (e.g., main), the checkpoint references on the orphan branch (`partio/checkpoints/v1`) may not be easily discoverable from the perspective of the target branch. Users report that after a fast-forward merge, session references and agent labels appear lost from the target branch context.

This happens because checkpoint discovery may be scoped to the branch where checkpoints were originally created, and fast-forward merges don't create a merge commit that could trigger checkpoint reference updates.

## Desired Behavior

1. Checkpoint discovery should follow commit ancestry — if a commit with a `Partio-Checkpoint` trailer is reachable from the current branch (regardless of which branch it was originally committed on), the checkpoint should be discoverable.
2. Optionally, a `post-merge` hook could update an index or ref that maps the target branch to all reachable checkpoints.
3. `partio status` and checkpoint browsing commands should show all checkpoints reachable from HEAD, not just those created on the current branch name.

## Context Hints

- `internal/checkpoint/` — checkpoint storage and discovery
- `internal/git/` — branch and ref operations
- `internal/hooks/` — hook implementations (consider post-merge)
- `cmd/partio/` — status command
