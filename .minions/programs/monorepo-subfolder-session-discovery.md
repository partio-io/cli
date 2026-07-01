---
id: monorepo-subfolder-session-discovery
target_repos:
  - cli
acceptance_criteria:
  - Session discovery correctly finds Claude Code sessions when the agent is launched from a monorepo subfolder rather than the repo root
  - Hook state passing (pre-commit to post-commit) works when the working directory is a subfolder of the git repo root
  - Tests cover subfolder launch scenarios using t.TempDir() with nested directory structures
pr_labels:
  - minion
---

# Ensure session discovery works when agent launches from monorepo subfolders

When Claude Code (or another agent) is launched from a subdirectory within a monorepo, Partio's session discovery may fail to find the active session because Claude Code keys its session directory to the cwd where it was launched, which differs from the git repo root.

Partio's `find_session_dir.go` already walks up from the repo root to parent directories, but it does not walk *down* into subdirectories of the repo. In a monorepo setup where a user runs `claude` from `repo/packages/frontend/`, the session directory is keyed to that subfolder path, not the repo root.

## What to implement

1. Extend `FindSessionDir` to also check subdirectories of the repo root that match the current working directory when the upward walk doesn't find a session.
2. Ensure hook state files in `.partio/state/` correctly capture the agent detection result regardless of which subdirectory the commit was initiated from.
3. Add test cases for the subfolder discovery path.

## Why this matters

Monorepos are increasingly common, and users often launch agents from package subdirectories rather than the repo root. Without this fix, Partio silently misses sessions in these scenarios, producing commits with no checkpoint data.
