---
id: doctor-checkpoint-integrity
target_repos:
  - cli
acceptance_criteria:
  - "partio doctor validates checkpoint branch tree objects are readable"
  - "partio doctor reports when checkpoint commits reference missing blobs"
  - "partio doctor detects orphaned checkpoint refs not linked to any commit"
  - "partio doctor supports --fix flag to prune corrupted checkpoint entries"
  - "Existing doctor checks continue to work unchanged"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Extend `partio doctor` with checkpoint integrity validation and repair

Enhance the existing `partio doctor` command to validate checkpoint branch integrity and offer repair for common corruption scenarios.

## Context

The current `partio doctor` only checks whether the checkpoint branch exists, hooks are installed, and the `.partio/` directory is present. It does not validate the actual checkpoint data integrity. If checkpoint tree objects become corrupted (e.g., due to interrupted writes, gc, or manual ref manipulation), users have no way to detect or fix this.

## Implementation

- Extend `cmd/partio/doctor.go` to walk checkpoint commits on `partio/checkpoints/v1`
- For each checkpoint commit, verify the tree object is readable via `git cat-file -t`
- Verify blobs referenced in the tree exist via `git cat-file -e`
- Report corrupted or missing objects as `[WARN]` with the affected checkpoint hash
- Add `--fix` flag that prunes checkpoint commits with unreadable trees by rewriting the ref to skip them
- Keep the existing checks intact; add new checks after them
- Limit traversal depth to last 100 checkpoints by default (configurable via `--depth`)
