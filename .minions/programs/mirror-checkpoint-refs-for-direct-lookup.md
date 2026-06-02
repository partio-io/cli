---
id: mirror-checkpoint-refs-for-direct-lookup
target_repos:
  - cli
acceptance_criteria:
  - Partio creates secondary refs (e.g., refs/partio/by-commit/<sha>) pointing to the checkpoint for each commit
  - Checkpoint lookup by commit SHA is O(1) via git show-ref instead of walking the orphan branch
  - Mirror refs are created atomically alongside the primary checkpoint commit on the orphan branch
  - Mirror refs are pushed alongside the checkpoint branch when push_sessions is enabled
  - Existing checkpoint storage on the orphan branch is unchanged (mirrors are additive)
pr_labels:
  - minion
---

# Mirror checkpoint refs for direct commit/session lookup

## Summary

Add secondary git refs that mirror checkpoint data, enabling O(1) lookup of checkpoints by commit SHA or session ID without walking the orphan branch.

## Problem

Partio currently stores all checkpoints as commits on a single orphan branch (`partio/checkpoints/v1`). To find the checkpoint associated with a specific commit, you must walk the branch history and match commit metadata. This is O(n) in the number of checkpoints and becomes slow as checkpoint history grows.

## Solution

After creating a checkpoint commit on the orphan branch, also create lightweight refs that point to it:

- `refs/partio/by-commit/<commit-sha>` — maps a source commit to its checkpoint
- `refs/partio/by-session/<session-id>` — maps a session ID to its latest checkpoint

These are created using `git update-ref` (consistent with existing plumbing approach) and pushed alongside the checkpoint branch.

## Why this matters

As repositories accumulate hundreds or thousands of checkpoints, direct lookup becomes essential for:
- Fast `partio rewind` to a specific commit's checkpoint
- Integration with external tools that need checkpoint data for a known commit
- Efficient cross-referencing between commits and their captured context

## Source

Inspired by entireio/cli v1.1 checkpoint mirroring to custom refs (changelog 0.7.0, #1300, #1315, #1311).
