---
id: prevent-concurrent-checkpoint-write-conflicts
target_repos:
  - cli
acceptance_criteria:
  - concurrent checkpoint writes from parallel sessions do not silently overwrite each other
  - git update-ref uses compare-and-swap (expected old value) to detect conflicts
  - on conflict, the write retries with the updated parent commit
  - all checkpoint data is preserved even when multiple sessions commit simultaneously
pr_labels:
  - minion
---

# Prevent concurrent checkpoint write conflicts

When multiple AI agent sessions are active in the same repository and commit simultaneously, their post-commit hooks race to write checkpoints to the same `partio/checkpoints/v1` ref. The current implementation reads the current tree, builds a new commit, and calls `update-ref` without any atomicity guarantee — a classic lost-update race condition.

## What to implement

Use `git update-ref`'s compare-and-swap mode to detect when the ref has been updated between read and write. On conflict, re-read the current tree and retry the checkpoint write with the updated parent. A small number of retries (e.g., 3) with the retry loop is sufficient since checkpoint writes are fast.

Specifically:
1. In `Store.Write()`, capture the current ref value before building the new commit
2. Pass the expected old ref value to `git update-ref` (the third positional arg)
3. If `update-ref` fails due to ref mismatch, re-read the current tree and rebuild the commit on top of the new parent
4. Retry up to 3 times before logging a warning and giving up (don't fail the git operation)

## Context hints

- `internal/checkpoint/write.go` — `Store.Write()` with the non-atomic read-modify-write sequence
- `internal/checkpoint/store.go` — helper methods for git plumbing operations

## Why this matters

Users running multiple agent sessions in parallel (e.g., in different worktrees or terminal tabs) can silently lose checkpoint data. The lost checkpoint is never recoverable because it was never committed to the branch — the ref was overwritten before it could be read.
