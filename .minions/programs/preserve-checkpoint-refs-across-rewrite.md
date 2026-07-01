---
id: preserve-checkpoint-refs-across-rewrite
target_repos:
  - cli
acceptance_criteria:
  - When a commit with a Partio-Checkpoint trailer is rebased, the checkpoint on the orphan branch remains discoverable from the new commit SHA
  - When a commit is amended, the checkpoint reference is updated to point to the new commit SHA
  - When git reset rewrites history, orphaned checkpoint refs are detected and a warning is logged
  - Existing checkpoints are not lost or corrupted during rewrite operations
pr_labels:
  - minion
---

# Preserve checkpoint references across commit rewrites

## Problem

When users rebase, amend, or reset commits that have `Partio-Checkpoint` trailers, the linkage between the user commit and the checkpoint on the orphan branch (`partio/checkpoints/v1`) can break. The checkpoint stores the original commit SHA, but after a rewrite that SHA no longer exists in the branch history. This makes checkpoints undiscoverable when browsing rewritten history.

## Desired Behavior

Partio should detect when commits with checkpoint trailers are rewritten and update the checkpoint metadata to maintain the link. This could be implemented via:

1. A `post-rewrite` hook (fired by `git rebase` and `git commit --amend`) that reads the old→new SHA mapping from stdin and updates checkpoint references on the orphan branch.
2. A fallback discovery mechanism that can find checkpoints by searching checkpoint commit messages or metadata for the original SHA, even when the trailer points to a now-dead commit.

## Context Hints

- `internal/checkpoint/` — checkpoint storage and metadata
- `internal/git/hooks/` — hook script generation
- `internal/hooks/` — hook implementations (add post-rewrite)
- `cmd/partio/` — CLI commands
