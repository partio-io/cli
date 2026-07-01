---
id: stale-session-inheritance-prevention
target_repos:
  - cli
acceptance_criteria:
  - "post-commit computes the set of files changed in the current commit and the set of files mentioned in the candidate session transcript"
  - "If file overlap is zero AND session age exceeds the staleness threshold, the session is not attached to the checkpoint and a warning is logged"
  - "The staleness threshold defaults to 2 hours and is configurable via config"
  - "A unit test in internal/hooks/ covers the stale-session-skip path using a synthetic preCommitState and session"
  - "Existing tests that verify normal session attachment continue to pass"
pr_labels:
  - minion
---

# Prevent stale ACTIVE sessions from inheriting another session's committed files

When post-commit looks up the latest agent session via `FindLatestSession`, it can incorrectly attach a stale session that is still marked ACTIVE from a previous coding task to a new commit. Add a file-overlap check in `internal/hooks/postcommit.go`: after finding the candidate session, compare the files changed in the current commit (from `git.DiffNameOnly`) against the files mentioned in the session transcript.

If there is zero file overlap and the session started more than a configurable threshold (default 2 hours) before the commit, skip session attachment and log a warning. The threshold should be configurable via config.

## Why

Partio's session logic only skips sessions marked Condensed. A long-running or abandoned ACTIVE session can be incorrectly linked to unrelated commits made hours later, corrupting the checkpoint's attribution and context.

## User relevance

Users who leave Claude Code running while doing manual commits will have their manual work incorrectly attributed to the AI session and stored with misleading context in the checkpoint.

## Context hints

- `internal/hooks/postcommit.go`
- `internal/agent/claude/find_latest_session.go`
- `internal/git/diff_name_only.go`
- `internal/session/session.go`
- `internal/config/config.go`
