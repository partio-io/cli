---
id: add-sessions-subcommands
target_repos:
  - cli
acceptance_criteria:
  - "`partio sessions list` displays all sessions (active and ended) with session ID, agent, status, start time, and linked commit count"
  - "`partio sessions info <session-id>` shows detailed session metadata including prompts, file touches, and checkpoint IDs"
  - "`partio sessions stop <session-id>` transitions an ACTIVE session to ENDED state and cleans up state files"
  - "All subcommands support `--json` flag for machine-readable output"
  - "Commands work correctly in git worktree setups"
pr_labels:
  - minion
---

# Add `partio sessions` subcommands for session lifecycle management

## Background

Partio currently manages sessions internally through the hook lifecycle (pre-commit/post-commit), but provides no user-facing commands for inspecting or managing sessions. Users cannot list active sessions, inspect session details, or manually stop a stuck session.

Inspired by entireio/cli's `entire sessions` subcommands (list, info, stop) added in v0.5.3.

## What to implement

Add a `partio sessions` command group with three subcommands:

### `partio sessions list`
- List all known sessions (active and ended) from `.partio/state/`
- Show: session ID, agent type, status (ACTIVE/ENDED), start time, number of linked checkpoints
- Support `--json` flag for scripting
- Flag sessions that appear stale (ACTIVE but no activity for a configurable threshold)

### `partio sessions info <session-id>`
- Show full session metadata: agent, status, start/end time, prompts, files touched, checkpoint IDs
- Support `--json` flag

### `partio sessions stop <session-id>`
- Transition a session from ACTIVE to ENDED
- Clean up any leftover state files in `.partio/state/`
- Useful for recovering from stuck sessions (e.g., agent crashed without cleanup)

## Key files to examine
- `internal/session/` — existing session domain types and state management
- `internal/session/manager.go` — session lifecycle operations
- `cmd/partio/` — existing command structure for reference
