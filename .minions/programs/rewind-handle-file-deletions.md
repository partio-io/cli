---
id: rewind-handle-file-deletions
target_repos:
  - cli
acceptance_criteria:
  - "partio rewind correctly restores files that were deleted in a subsequent commit"
  - "partio rewind to a checkpoint that includes file deletions applies those deletions to the working tree"
  - "Rewind handles mixed changes (additions, modifications, and deletions) in the same checkpoint"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Handle file deletions correctly in `partio rewind`

Ensure `partio rewind` properly handles checkpoints where files were deleted, not just added or modified.

## Context

When rewinding to a checkpoint, the current implementation may not correctly handle cases where the checkpoint records file deletions. This was identified as a real bug in entireio/cli (PR #1408) where rewind failed to properly restore the working tree state when tracked files had been deleted between checkpoints.

## Implementation guidance

- When rewinding to a checkpoint, compare the file tree at the target checkpoint with the current working tree
- Files present in the working tree but absent in the target checkpoint's tree should be removed
- Files absent in the working tree but present in the target checkpoint should be restored
- Test with scenarios: delete-only commits, mixed add/delete/modify commits, consecutive deletions

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
