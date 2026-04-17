---
id: detect-session-divergence-on-reset
target_repos:
  - cli
acceptance_criteria:
  - "`partio status` detects when an active session's BaseCommit no longer matches HEAD"
  - "When HEAD has the same Partio-Checkpoint trailer as the session's LastCheckpointID, BaseCommit is silently updated to match HEAD"
  - "When HEAD has a different or no checkpoint trailer, a divergence warning is displayed without mutating session state"
  - "Worktree paths are normalized before comparison so symlinked and temp paths reconcile correctly"
  - "Unit tests cover both the auto-reconciliation and warning paths"
pr_labels:
  - minion
---

# Detect session divergence after `git reset` in `partio status`

## Background

When a user runs `git reset` (soft, mixed, or hard) during an active Partio session, HEAD moves to a different commit but the session's `BaseCommit` still references the old commit. This can lead to incorrect attribution calculations or confusing status output, since Partio doesn't know the commit graph has changed underneath.

Currently `partio status` does not check whether the session's tracked commit still matches HEAD, so the divergence is silent.

Inspired by entireio/cli PR #948 which adds lightweight reset detection to the status path.

## What to implement

### 1. Divergence detection in `partio status`
- After loading active sessions, resolve the current HEAD commit
- For each active session in the current worktree, compare `BaseCommit` against HEAD
- If they match, no action needed

### 2. Safe auto-reconciliation
- If HEAD has moved but contains the same `Partio-Checkpoint` trailer value as the session's `LastCheckpointID`:
  - Update `BaseCommit` and `AttributionBaseCommit` to the current HEAD
  - This handles the common case of `git reset` to a commit that was already checkpointed

### 3. Divergence warning
- If HEAD has moved and does NOT have a matching checkpoint trailer:
  - Display a warning in `partio status` output (e.g., "Session <id> was tracking commit <old>, but HEAD is now at <new>")
  - Do NOT mutate session state — the user may be doing something intentional

### 4. Path normalization
- Normalize worktree paths before comparing session worktree against current worktree
- Handle symlinks and temp directory paths that may resolve differently

## Key files to examine
- `cmd/partio/status.go` — status command implementation
- `internal/session/state.go` — session state with BaseCommit field
- `internal/git/` — git operations for reading HEAD, parsing trailers
- `internal/checkpoint/` — checkpoint ID from commit trailers
