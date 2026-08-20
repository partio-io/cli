---
id: session-commit-link-by-pid
target_repos:
  - cli
acceptance_criteria:
  - Pre-commit state file records the detected agent's process PID alongside the worktree path
  - Post-commit reads the recorded PID and selects only sessions whose tracked PID matches
  - When no session matches the PID (e.g., agent restarted between pre-commit and post-commit), post-commit logs a warning and skips session linking instead of misattributing to the wrong session
  - When multiple Claude Code processes are running in the same worktree, the PID is used to select the correct one
  - Existing pre-commit state files without a PID field continue to work (backward-compatible graceful fallback to worktree-path matching)
pr_labels:
  - minion
---

# Link commits to sessions by process identity (PID)

Partio currently matches sessions to commits via worktree path: the pre-commit hook saves the
detected agent's `WorktreePath` to `.partio/state/pre-commit.json`, and post-commit selects
sessions by comparing stored path against known sessions. This causes silent misattribution in
three real scenarios:

1. **Multiple sessions in the same worktree** — nested Claude Code subagent processes or two
   concurrent sessions both match the same worktree path; one gets attributed to the commit at
   random.
2. **Agent restart between hooks** — if Claude Code exits and restarts between pre-commit and
   post-commit, the post-commit hook links the new session instead of the one active at commit
   time.
3. **Stale WorktreePath drift** — if the repo was relocated or the session was started from a
   parent directory, the path comparison fails and no session is linked.

Store the agent PID in the pre-commit state file when the agent is detected. Post-commit
matches by PID first; if the PID is missing or no session carries it, fall back to the current
worktree-path logic so existing state files remain valid.

Inspired by entireio/cli PR #2013 "feat(strategy): link commits to sessions by process
identity", which fixed the same class of misattribution bugs.

## Context hints

- `internal/hooks/` — pre-commit and post-commit hook implementations
- `internal/agent/claude/` — Claude process detection (already reads PID for liveness checks)
- `internal/session/` — SessionState and session matching logic
- `internal/git/` — pre-commit state file read/write helpers

<!-- program: .minions/programs/session-commit-link-by-pid.md -->
