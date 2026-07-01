---
id: guard-post-commit-rebase-reentry
target_repos:
  - cli
acceptance_criteria:
  - "Post-commit hook detects an in-progress interactive rebase (presence of `.git/rebase-merge/` or `.git/rebase-apply/`) and exits 0 immediately without creating a checkpoint"
  - "A debug-level log message is emitted when the rebase guard triggers: 'skipping checkpoint creation: interactive rebase in progress'"
  - "Normal commits (no rebase) are unaffected"
  - "The guard does not prevent Partio from operating on commits made *after* the rebase completes"
  - "Unit test covers the rebase-detection logic with a fake git dir structure"
pr_labels:
  - minion
---

# Guard post-commit hook against spurious checkpoints during interactive rebase

## What

During `git rebase -i`, git replays each selected commit and fires the post-commit hook for each one. Partio's post-commit hook will attempt to create a new checkpoint for every replayed commit — even though those commits already have checkpoints and no agent session is running.

Add a guard to `internal/hooks/post_commit.go` (or wherever the post-commit logic lives) that detects an in-progress rebase and returns early with a no-op. Detection: check for the presence of `.git/rebase-merge/` or `.git/rebase-apply/` directories (these exist for the duration of an interactive or am-style rebase).

Note: Partio already guards against `--amend` re-entry by deleting the state file before the amend. The rebase case is separate — the state file won't be present (so that guard won't trigger), but the hook still fires.

## Why

Without this guard, `git rebase -i` on a branch with many commits will:
1. Attempt to create a new checkpoint for each replayed commit (most will fail gracefully since there's no state file, but it adds unnecessary overhead)
2. Potentially create a spurious checkpoint if a pre-commit state file happens to be present from a prior interrupted operation
3. Slow down rebases with checkpoint overhead

This matches a pattern seen in `entireio/cli` PR #824 which added a guard against duplicate `session.created` events being processed.

## Source

Inspired by `entireio/cli` PR #824 (guard session-start hook on duplicate session.created) and the general pattern of making hooks resilient to git operations that replay commits.

## Implementation hints

- `internal/hooks/` — post-commit hook implementation
- `internal/git/` — add a `IsRebaseInProgress(gitDir string) bool` helper
- Check `filepath.Join(gitDir, "rebase-merge")` and `filepath.Join(gitDir, "rebase-apply")`
- Use `internal/log` for the debug-level log message
- The git dir is available via `git rev-parse --git-dir` (already used elsewhere)
<!-- program: .minions/programs/guard-post-commit-rebase-reentry.md -->
