---
id: fix-rewind-checkpoint-not-fetched
target_repos:
  - cli
acceptance_criteria:
  - "`partio rewind --list` shows a helpful fetch suggestion when the checkpoint branch exists on the remote but is not fetched locally"
  - "When neither local nor remote has the branch, the original 'no checkpoint branch found' error is shown"
  - "The fetch hint includes the exact git command to run"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Fix `partio rewind` when checkpoint branch isn't fetched locally

`partio rewind --list` returns `"no checkpoint branch found"` when `partio/checkpoints/v1` doesn't exist locally. If the user is working in a cloned repo or on a new machine where checkpoints were pushed from elsewhere, the branch may exist on the remote but not be fetched. The current error message is confusing because it implies checkpoints don't exist at all.

## Fix

In `runRewindList()` (`cmd/partio/rewind.go`), when `git rev-parse --verify partio/checkpoints/v1` fails:

1. Check if the branch exists on the remote:
   ```
   git ls-remote --heads origin partio/checkpoints/v1
   ```
2. If the remote has the branch, print a helpful message instead of the generic error:
   ```
   Checkpoint branch not fetched locally.
   Run: git fetch origin partio/checkpoints/v1:partio/checkpoints/v1
   ```
3. If neither local nor remote has the branch, keep the current error.

Also apply the same check in `runRewindTo()` for consistency.

## Key files

- `cmd/partio/rewind.go` — `runRewindList()` at line ~50 and `runRewindTo()` at line ~94

**Inspired by:** entireio/cli changelog v0.5.3 — "Resume failing when checkpoints aren't fetched locally yet"
