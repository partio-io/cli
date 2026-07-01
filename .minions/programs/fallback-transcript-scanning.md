---
id: fallback-transcript-scanning
target_repos:
  - cli
acceptance_criteria:
  - When the primary session directory path yields no transcript files, the system falls back to scanning known agent session storage locations
  - Fallback scanning covers Claude Code worktree session directories (~/.claude/projects/...)
  - The fallback uses the commit timestamp and repo identity to narrow candidate sessions
  - A warning is logged when fallback scanning is used instead of direct path resolution
  - Existing direct-path discovery still works and is tried first
  - Tests cover the fallback path with a session directory that doesn't match the expected derived path
pr_labels:
  - minion
---

# Add fallback transcript scanning when expected session path is missing

## Summary

When the session directory derived from the detected agent process doesn't contain the expected transcript files, fall back to scanning known session storage locations (e.g., `~/.claude/projects/`) to find the correct session by matching against repo identity and timing.

## Why

Claude Code can be launched from various working directories — worktree subdirectories, parent directories, or relocated workspaces. When the agent runs from inside its own `.claude/worktrees/<branch>/` directory, the transcript path gets encoded relative to that worktree CWD rather than the actual repo root. This causes Partio's current session discovery to miss the transcript entirely, resulting in checkpoints without session context. This was identified and fixed in entireio/cli PR #1159.

## What to implement

1. **Keep existing discovery as primary**: The current `find_session_dir.go` logic that walks up from the repo root remains the first attempt.

2. **Add fallback scanner**: If the primary path yields no transcript files, scan `~/.claude/projects/` for session directories whose project path component matches the current repository (by normalized path or repo name).

3. **Time-window filtering**: Among candidate sessions found by scanning, select the one whose last-modified timestamp is closest to (and before) the current commit time.

4. **Log fallback usage**: Emit a debug/warning log when the fallback path is used so users can diagnose session discovery issues.

5. **Agent-specific implementation**: The fallback scanner should be part of the Claude agent detector implementation, not the generic session discovery, since storage locations are agent-specific.

## Context hints

- `internal/agent/claude/find_session_dir.go` — current session directory discovery
- `internal/agent/claude/find_latest_session.go` — session selection logic
- `internal/hooks/post_commit.go` — where session discovery is invoked
