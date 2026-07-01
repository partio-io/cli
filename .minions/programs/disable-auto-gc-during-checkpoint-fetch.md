---
id: disable-auto-gc-during-checkpoint-fetch
target_repos:
  - cli
acceptance_criteria:
  - "All git fetch operations for checkpoint refs pass --no-auto-gc to prevent concurrent garbage collection"
  - "Pack file corruption from gc/fetch races is eliminated in normal operation"
  - "Existing fetch behavior is otherwise unchanged"
pr_labels:
  - minion
---

# Disable auto-gc during checkpoint fetch operations

When Partio fetches checkpoint refs from a remote, git may trigger automatic garbage collection (`git gc --auto`) concurrently. This can cause a race condition where gc repacks or deletes pack files that the fetch is actively reading, leading to corrupt or missing objects.

## What to implement

1. Add `--no-auto-gc` to all `git fetch` invocations used for checkpoint sync (both push and pull directions).

2. This is a targeted, low-risk fix: it only defers gc to the next user-initiated git operation rather than disabling it entirely.

## Context hints

- `internal/checkpoint/` - Checkpoint storage, which includes fetch/push operations
- `internal/git/` - Git operation wrappers
