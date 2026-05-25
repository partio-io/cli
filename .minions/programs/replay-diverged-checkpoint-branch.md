---
id: replay-diverged-checkpoint-branch
target_repos:
  - cli
acceptance_criteria:
  - When a checkpoint push fails due to remote divergence, Partio fetches the remote state and replays local checkpoint commits on top
  - No checkpoint data is lost during replay — both local and remote checkpoints are preserved
  - If replay fails (e.g., due to conflicting refs), the operation falls back to a clear error message with manual resolution instructions
  - Concurrent checkpoint pushes from different machines or worktrees converge without force-push
pr_labels:
  - minion
---

# Replay local checkpoints when remote checkpoint branch diverges

## Problem

When multiple machines, worktrees, or CI jobs create checkpoints for the same repository, the checkpoint orphan branch (`partio/checkpoints/v1`) can diverge between local and remote. Currently, a non-fast-forward push would fail, potentially requiring manual intervention or risking data loss if force-pushed.

## Solution

When a checkpoint push detects that the remote branch has diverged (non-fast-forward), Partio should:

1. Fetch the latest remote checkpoint branch state
2. Replay local checkpoint commits that aren't on the remote onto the fetched remote tip
3. Push the reconciled branch

This is analogous to `git pull --rebase` but for the checkpoint orphan branch, ensuring checkpoint data from all sources is preserved.

### Implementation hints

- In `internal/checkpoint/storage.go` (or the push path), detect non-fast-forward push failures
- Fetch the remote checkpoint ref
- Walk local commits not reachable from the remote tip
- Replay those commits (re-creating tree objects and commit objects via git plumbing) on top of the remote tip
- Retry the push; if it fails again (another concurrent push), retry with backoff up to a limit
- Log warnings if replay is needed so users are aware of the reconciliation

## Inspiration

Adapted from entireio/cli PR #1251 which adds strategy to replay local checkpoints when fetch finds a diverged remote.
