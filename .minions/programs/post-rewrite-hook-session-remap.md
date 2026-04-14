---
id: post-rewrite-hook-session-remap
target_repos:
  - cli
acceptance_criteria:
  - "A managed `post-rewrite` hook script is installed alongside existing hooks by `partio enable`"
  - "After `git commit --amend`, the active session's BaseCommit is updated to the new commit hash"
  - "After `git rebase`, all affected sessions have their BaseCommit remapped using the old-to-new commit mapping from stdin"
  - "LastCheckpointID remains unchanged so checkpoint linkage follows the rewritten commit"
  - "The post-rewrite hook respects the same re-entry prevention and error resilience patterns as existing hooks"
  - "Hook chaining works correctly (backs up and calls previous post-rewrite hook if present)"
pr_labels:
  - minion
---

# Add `post-rewrite` hook to remap session state after git history rewrites

## Background

When users run `git commit --amend` or `git rebase`, the commit hashes change but Partio's session state still references the old hashes. This means `BaseCommit` and `AttributionBaseCommit` in the session become stale, potentially breaking attribution calculations and session-to-commit linkage.

Currently, Partio has no `post-rewrite` hook. The existing `pre-commit` and `post-commit` hooks handle normal commit flow, and `pre-push` handles checkpoint pushing. But history rewrites silently orphan session references.

Inspired by entireio/cli PR #947 which adds post-rewrite tracking to preserve local session linkage.

## What to implement

### 1. Hook script generation
- Add `post-rewrite` to the hook types in `internal/git/hooks/`
- Generate a hook script that calls `partio _hook post-rewrite <rewrite-type>`
- Follow existing patterns: `git rev-parse --git-common-dir` for installation, backup chaining, `exit 0` at end

### 2. Hook implementation
- Add `post-rewrite` handler in `internal/hooks/`
- Git passes `amend` or `rebase` as the rewrite type argument
- Read old-new commit mappings from stdin (git's standard format: `<old-hash> <new-hash>` per line)
- For each active session, check if its `BaseCommit` or `AttributionBaseCommit` matches any old hash
- Update to the corresponding new hash
- Leave `LastCheckpointID` unchanged

### 3. Hook installation
- Update `partio enable` to install the `post-rewrite` hook
- Update `partio disable` to uninstall it

## Key files to examine
- `internal/git/hooks/` — hook script generation and install/uninstall
- `internal/hooks/` — hook implementations (pre-commit, post-commit, pre-push patterns)
- `internal/session/state.go` — session state persistence
- `cmd/partio/hook.go` — hook command routing
