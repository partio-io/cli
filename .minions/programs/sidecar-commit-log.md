---
id: sidecar-commit-log
target_repos:
  - cli
acceptance_criteria:
  - Each checkpoint includes a commits.jsonl file listing all commits made during the session up to that point
  - Each entry in commits.jsonl contains commit hash, short hash, subject, timestamp, checkpoint ID, and session ID
  - The log is cumulative — checkpoint N contains entries for all N commits in the session
  - The commit log is built in memory before checkpoint creation and included in the checkpoint tree
  - The commit log is also persisted to .partio/state/ so the next checkpoint picks up the full history
  - Reading a single checkpoint provides the complete session commit timeline without cross-referencing git history
  - Unit tests verify log accumulation across multiple checkpoints in a session
pr_labels:
  - minion
---

# Add sidecar commit log for session timeline reconstruction

Include a cumulative `commits.jsonl` file in each checkpoint that records all commits made during the session, enabling session timeline reconstruction from any single checkpoint without git log scanning.

## Context

Today, reconstructing "what commits happened during this session" requires cross-referencing checkpoint IDs across git history, session metadata, and the checkpoint branch. For a timeline view where you pick checkpoint 4 and want to see the full session history including checkpoints 1-3, this requires expensive backward lookups across potentially weeks of git history.

## Approach

Each time post-commit creates a checkpoint, append a one-line JSON entry to a session-scoped commit log:

```jsonl
{"hash":"abc123...","short_hash":"abc123d","subject":"Add login","timestamp":"2026-04-16T12:11:06Z","checkpoint_id":"a3b2c4d5e6f7","session_id":"..."}
```

This file accumulates across the session. When checkpoint N is created, its `commits.jsonl` contains entries for all N commits — no backward lookup needed.

The log should be:
- **Built in memory before checkpoint creation** so it's included in the checkpoint tree it belongs to
- **Persisted to `.partio/state/`** so the next checkpoint picks up the full timeline
- **Cleaned up** when the session ends or state is cleared

This enables future features like session timeline UIs and orphaned checkpoint detection without requiring any changes to the existing checkpoint structure or git trailers.
