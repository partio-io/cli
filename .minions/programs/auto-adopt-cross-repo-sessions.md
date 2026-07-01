---
id: auto-adopt-cross-repo-sessions
target_repos:
  - cli
acceptance_criteria:
  - Pre-commit hook discovers active agent sessions from other enabled repos on the same machine when no local session is found
  - A commit made in repo B by an agent whose session started in repo A gets a checkpoint trailer
  - Session discovery across repos only activates when no local session matches (no performance penalty for normal workflows)
  - Cross-repo session matching respects the agent process's working directory history, not just the session start directory
  - Works correctly when the source repo and target repo have different git common dirs
pr_labels:
  - minion
---

# Auto-adopt agent sessions that move across git repositories

## Problem

When a long-running agent session (e.g., Claude Code) starts in one git repository and then makes commits in a different enabled repository, the second repo's hooks cannot find the session state because it lives in the first repo's directory. The commits in the second repo get no checkpoint trailer, silently losing session context.

This is different from worktree discovery (same git common dir) and from manual cross-repo attach (explicit CLI command). The gap is that hooks have no fallback discovery path when the agent process is active but its session state lives in a sibling repo.

## Desired behavior

When the pre-commit hook detects an active agent process (e.g., Claude Code is running) but finds no matching session in the current repo's session directory:

1. Check a shared session registry (e.g., `~/.config/partio/active-sessions.json`) that tracks which repos have active sessions
2. If a matching session is found in another repo, use that session's data for checkpoint creation
3. The registry is updated by hooks in the originating repo when a session starts/stops

## Context hints

- `internal/agent/claude/find_session_dir.go` — current session discovery logic
- `internal/hooks/pre_commit.go` — where session detection happens
- `internal/hooks/post_commit.go` — where checkpoint creation reads session state
- `internal/session/` — session lifecycle management
