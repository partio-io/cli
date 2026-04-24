---
id: link-checkpoint-to-branch
target_repos:
  - cli
acceptance_criteria:
  - "partio link <checkpoint-id> --branch <branch-name> associates an orphaned checkpoint with the specified branch"
  - "The command updates the checkpoint metadata on the orphan branch to include the branch reference"
  - "If the checkpoint is already linked to a branch, the command warns and requires --force to overwrite"
  - "partio status shows unlinked checkpoints with a visual indicator"
pr_labels:
  - minion
---

# Link orphaned checkpoints to a branch

Checkpoints can become orphaned (not associated with any branch) when commits are made in detached HEAD state, during interactive rebases, or when a branch is created after the checkpoint was already stored. Currently there is no way to retroactively associate these checkpoints with a branch.

## What to implement

1. Add a `partio link <checkpoint-id> [--branch <name>]` subcommand that updates a checkpoint's metadata to associate it with a git branch.

2. If `--branch` is omitted, default to the current branch (`HEAD`).

3. Update `partio status` to indicate when checkpoints exist that are not linked to any branch, so users know to run `partio link`.

## Context hints

- `internal/checkpoint/` - Checkpoint domain type and storage
- `cmd/partio/` - CLI command implementations
