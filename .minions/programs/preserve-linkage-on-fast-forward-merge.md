---
id: preserve-linkage-on-fast-forward-merge
target_repos:
  - cli
acceptance_criteria:
  - When a fast-forward merge moves HEAD to a commit with a Partio-Checkpoint trailer, the session linkage is preserved and visible in `partio status` on the target branch
  - A post-merge hook detects fast-forward merges and updates session state to reflect the new branch context
  - Unit tests verify that checkpoint trailers from a merged branch remain resolvable after fast-forward merge
pr_labels:
  - minion
---

# Preserve session linkage across fast-forward merges

## Problem

When a feature branch with Partio-linked commits is fast-forward merged into another branch (e.g., `git merge --ff-only feature`), the session references from the feature branch are not automatically tracked on the target branch. Users report that after a fast-forward merge, `partio status` and checkpoint browsing on the target branch lose visibility of the sessions that were linked on the source branch.

This happens because Partio's session state tracks commits by branch, and a fast-forward merge doesn't create a new commit — it just moves the branch pointer. The existing post-commit hook doesn't fire, so no session state update occurs.

## Proposed solution

1. **Add post-merge hook support**: Install a `post-merge` hook that detects when HEAD has moved and checks whether the new HEAD commit(s) carry `Partio-Checkpoint` trailers.
2. **Update session state**: When a fast-forward merge brings in commits with checkpoint trailers, update the local session index to associate those checkpoints with the current branch.
3. **Hook installation**: Extend `partio enable` and the hook installation logic to include `post-merge` alongside the existing `pre-commit`, `post-commit`, and `pre-push` hooks.

## Why this matters

Fast-forward merges are the default merge strategy for many workflows (especially with linear history requirements). If Partio's session linkage silently breaks on the most common merge type, users lose the ability to trace code reasoning on their main branch — exactly where it matters most.

## Context hints

- `internal/hooks/` — hook implementations (pre-commit, post-commit, pre-push)
- `internal/git/hooks/` — hook script generation and installation
- `internal/session/` — session state and lifecycle management
- `cmd/partio/` — CLI command registration for hook wiring
