---
id: validate-checkpoint-tree-paths
target_repos:
  - cli
acceptance_criteria:
  - mktree helper in internal/checkpoint/store.go validates that every tree entry name is non-empty, relative, and contains no empty path segments before formatting the mktree input
  - absolute paths are rejected or converted to repo-relative paths before tree construction
  - a unit test confirms that a treeEntry with an absolute path (e.g., /home/user/repo/file.go) is rejected or normalized
  - a unit test confirms that a treeEntry with an empty name produces an error
  - existing checkpoint write tests continue to pass
pr_labels:
  - minion
---

# Validate git tree entry paths before checkpoint writes

## Description

Add path validation to the `mktree` helper in `internal/checkpoint/store.go` to reject or normalize malformed tree entry names before writing checkpoint tree objects. Currently, tree entry names are formatted into `git mktree` input without any validation. If an absolute path or a path with empty segments (e.g., from Windows path splitting) leaks into a tree entry name, `git mktree` silently creates a corrupted tree object with empty-filename entries. This causes `git fsck` to report `badTree` errors and can break operations like `git bundle create --all`.

The fix should validate each `treeEntry.name` in the `mktree` function:
- Reject empty names with an error
- Reject or strip leading `/` (absolute paths)
- Reject names containing empty path segments (e.g., `foo//bar`)
- Reject names containing `.` or `..` segments

## Why

Corrupted tree objects on the checkpoint branch can cascade into repository-wide git failures. Since checkpoints are written on every commit via git hooks, a single bad path can silently corrupt the checkpoint branch. This was discovered in entireio/cli#886 where Windows absolute paths produced empty-named tree entries, and fixed in entireio/cli#902.

## Source

- **Origin:** entireio/cli#886, entireio/cli#902
- **Detected from:** `entireio-cli-issues`, `entireio-cli-pulls`
