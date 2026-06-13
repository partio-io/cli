---
id: fix-multi-commit-session-checkpointing
target_repos:
  - cli
acceptance_criteria:
  - "When Claude Code is active and three consecutive commits are made, all three commits receive a Partio-Checkpoint trailer"
  - "The post-commit hook does not skip a commit silently when the pre-commit state file was already consumed by a prior commit"
  - "A test covering multiple sequential commits in the same session verifies checkpoint creation for each commit"
pr_labels:
  - minion
---

# Checkpoint all commits within a session, not just the first

Currently Partio's post-commit hook may only attach a checkpoint trailer and create a checkpoint for the first commit in a session turn. Subsequent commits within the same session are silently skipped — no trailer is added, no checkpoint is created, and no error is logged.

The fix should ensure that every commit made while a Claude Code session is active gets a checkpoint entry, linking each commit back to the session. The pre-commit state detection should not be blocked by a previously consumed state file, and the post-commit hook should correctly handle multiple sequential commits within one agent turn.

## Why

The core value proposition of Partio is that every commit is linked to the agent session that produced it. If only the first commit per session gets a checkpoint, the chain of reasoning is broken for multi-step agent workflows.

## User relevance

Users who let Claude Code make several commits in a row (e.g., feature + tests + fix) will find that only the first commit has a checkpoint trailer, making the audit trail incomplete and rewind unreliable for the later commits.

## Context hints

- `internal/hooks/` — pre-commit and post-commit hook logic
- `internal/session/` — session state management
- `internal/checkpoint/` — checkpoint creation
