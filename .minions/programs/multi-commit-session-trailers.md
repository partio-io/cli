---
id: multi-commit-session-trailers
target_repos:
  - cli
acceptance_criteria:
  - "All commits made during an active agent session receive the Partio-Checkpoint trailer"
  - "Second and subsequent commits in the same session are linked to the session checkpoint"
  - "Session continuity is detected by checking if the same agent process is still active"
  - "The post-commit hook correctly chains multiple commits to the same session"
  - "Tests verify trailer presence across multiple sequential commits in one session"
pr_labels:
  - minion
---

# Ensure all commits in an active session receive checkpoint trailers

## Summary

Fix the checkpoint trailer logic so that every commit made during an active agent session gets a `Partio-Checkpoint` trailer, not just the first one. Currently, if an agent session spans multiple commits, only the first commit reliably receives the trailer — subsequent commits may silently skip it.

## Motivation

When an agent session involves multiple commits (e.g., implementing a feature across several incremental commits), losing the checkpoint trailer on later commits breaks the session-to-commit linkage. Users browsing git history can't trace all commits back to the originating session, and tools that rely on trailers for attribution miss commits.

## Behavior

1. Pre-commit hook detects the active agent session and saves state (already works)
2. Post-commit hook reads state, creates/updates checkpoint, and adds trailer (already works for first commit)
3. For subsequent commits in the same session: the hook should detect session continuity and link to the existing or updated checkpoint
4. Session identity is determined by the running agent process, not by the state file lifecycle

## Context

- Inspired by `entireio/cli` issue #784 (only first commit per session gets trailer)
- `internal/hooks/post_commit.go` for trailer injection logic
- `internal/hooks/pre_commit.go` for session state detection
- `internal/session/` for session lifecycle
