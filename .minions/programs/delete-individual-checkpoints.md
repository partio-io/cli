---
id: delete-individual-checkpoints
target_repos:
  - cli
acceptance_criteria:
  - "partio checkpoint delete <checkpoint-id> removes the specified checkpoint from the orphan branch"
  - "Deleting a checkpoint that does not exist returns a clear error"
  - "The command asks for confirmation before deleting (unless --force is passed)"
  - "After deletion, the checkpoint branch is rewritten without the deleted checkpoint's tree entry"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add command to delete individual checkpoints

## Summary

Add a `partio checkpoint delete <checkpoint-id>` command that removes a specific checkpoint from the orphan branch by ID. This complements the existing `partio prune` command (which deletes old checkpoints by age) with targeted deletion of specific checkpoints.

## Background

Inspired by entireio/cli changelog 0.7.7 which added a "delete subcommand" for trails/checkpoints. Users may need to remove specific checkpoints — for example, if a checkpoint captured sensitive data that got past redaction, or to clean up test checkpoints. Currently `partio prune` only supports age-based bulk cleanup.

## Implementation notes

- Add a new `checkpoint delete` subcommand (or `partio checkpoint delete`) in `cmd/partio/`
- Use git plumbing operations in `internal/checkpoint/` to remove the checkpoint's tree entry from the orphan branch
- Require confirmation by default; support `--force` to skip
- Reuse the existing checkpoint store's read operations to locate the checkpoint by ID before deletion
- The orphan branch commit history should be rewritten to exclude the deleted checkpoint's subtree
