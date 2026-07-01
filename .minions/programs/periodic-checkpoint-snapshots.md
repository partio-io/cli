---
id: periodic-checkpoint-snapshots
target_repos:
  - cli
acceptance_criteria:
  - A background timer or event-based mechanism captures session state at configurable intervals during active agent sessions
  - Snapshots are stored on the checkpoint orphan branch using the existing git plumbing storage
  - The interval is configurable via `strategy_options` in settings (e.g., `snapshot_interval: 5m`)
  - Periodic snapshots do not interfere with normal commit-triggered checkpoints
  - When no agent session is active, no snapshots are created
  - `partio status` shows the last snapshot timestamp when periodic snapshots are enabled
pr_labels:
  - minion
---

# Periodic checkpoint snapshots during active sessions

## Summary

Add a periodic snapshot mechanism that captures agent session state at configurable intervals during active sessions, independent of git commits.

## Problem

Partio currently only captures checkpoints when a git commit occurs (via the post-commit hook). Long agent sessions that involve extensive research, planning, or iterating before committing can lose significant context if the session ends unexpectedly. The entire reasoning chain between commits is only preserved if the session JSONL happens to still be on disk.

## Proposed Solution

Add a lightweight snapshot mechanism that periodically writes the current session state to the checkpoint branch:

1. During `post-commit`, if an active session is detected and `snapshot_interval` is configured, start a background goroutine (or use the existing hook lifecycle) that periodically:
   - Reads the current session JSONL
   - Creates a snapshot checkpoint on the orphan branch with a `type: snapshot` marker in metadata
   - Updates a timestamp file in `.partio/state/last_snapshot.json`

2. Snapshots should be lightweight — only the new JSONL content since the last snapshot/checkpoint needs to be stored.

3. The `pre-commit` hook can stop any pending snapshot timer and let the normal checkpoint flow take over.

## Configuration

```json
{
  "strategy_options": {
    "snapshot_interval": "5m"
  }
}
```

Setting to `"0"` or omitting disables periodic snapshots (default behavior).

## Context

- `internal/checkpoint/` — checkpoint storage
- `internal/session/` — session state management
- `internal/hooks/` — hook implementations
- `internal/config/` — configuration
