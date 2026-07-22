---
id: warn-empty-transcript-checkpoint
target_repos:
  - cli
acceptance_criteria:
  - "When post-commit creates a checkpoint but the parsed session has zero messages, a warning is printed to stderr"
  - "The warning identifies the session ID (or path) that produced no messages"
  - "The checkpoint is still written (with empty transcript); the warning is advisory only"
  - "When the session has at least one message, no warning is emitted"
  - "The warning is suppressed when `PARTIO_LOG_LEVEL=error`"
pr_labels:
  - minion
  - enhancement
---

# Warn when post-commit creates a checkpoint with no session messages

## Problem

Partio's post-commit hook writes a checkpoint even when JSONL parsing yields zero messages — for example when Claude Code is detected running but has not yet written any transcript entries, or when the session file is found but contains no parseable assistant/user turns. The checkpoint is created successfully but carries an empty transcript. The user sees no indication that session data was missing.

This silent empty checkpoint is indistinguishable from a successful capture from the user's perspective, making it hard to diagnose capture failures.

## Proposed behavior

After parsing the Claude Code JSONL session in the post-commit hook, check whether the resulting `SessionData` has any messages. If not, emit a warning to stderr before writing the checkpoint:

```
partio: warning: checkpoint created but no session messages were captured (session: <id>)
```

The checkpoint should still be written — this is advisory, not fatal. The warning gives users a signal that something may need investigation (e.g., the session file wasn't ready yet, or session discovery picked the wrong directory).

This matches the robustness improvement in entireio/cli 0.8.0 where `entire attach` was updated to warn on an empty transcript, note when an amend fails, and capture the session footer.

## Why this matters

Empty checkpoints are a silent failure mode. Users who see no warning assume capture succeeded. Adding a visible warning lets users know when session data was not captured, so they can investigate without having to inspect the checkpoint branch manually.

## Source

entireio/cli changelog 0.8.0 — "`entire attach` warns on an empty transcript, captures the session footer, and notes when an amend fails."

## Context hints

- `internal/hooks/` — post-commit hook implementation
- `internal/agent/claude/` — JSONL parsing that produces `SessionData`
- `internal/session/` — session data types
