---
id: concurrent-checkpoint-locking
target_repos:
  - cli
acceptance_criteria:
  - "Concurrent checkpoint writes to the same orphan branch are serialized via file lock"
  - "A second concurrent operation fails fast with a clear error naming the blocking process"
  - "Lock is released on process exit (including crashes)"
  - "Existing single-operation behavior is unchanged"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add file-based locking to prevent concurrent checkpoint corruption

When multiple agent sessions trigger post-commit hooks simultaneously (e.g., concurrent Claude Code and Codex sessions), parallel checkpoint writes to the same orphan branch (`partio/checkpoints/v1`) can race and corrupt ref state. Add a `flock`-based mutual exclusion lock around checkpoint write operations.

## What to implement

1. Acquire an exclusive file lock (e.g., `.partio/checkpoint.lock` or equivalent) before writing checkpoint data to the orphan branch.
2. On contention, fail fast with a clear stderr message identifying the blocking PID rather than queuing or hanging.
3. Ensure the lock is released on normal exit and process crash (use `flock` semantics where the OS releases on fd close).
4. Add test coverage for the locking mechanism.

## Context hints

- `internal/checkpoint/` — checkpoint storage operations
- `internal/hooks/` — post-commit hook that triggers checkpoint writes
- `.partio/state/` — existing state file location (lock file could live nearby)
