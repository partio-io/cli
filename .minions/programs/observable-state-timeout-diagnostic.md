---
id: observable-state-timeout-diagnostic
target_repos:
  - cli
acceptance_criteria:
  - Long-running hook operations (checkpoint push, session parsing) emit real-time phase events on TTY
  - When an operation times out, the error message reflects the actual phase (e.g. "timed out writing checkpoint objects" vs generic timeout)
  - Non-TTY environments get one line per phase event for CI/log compatibility
  - Timeout is opt-in via config or flag, not a hard default cap
pr_labels:
  - minion
  - enhancement
---

# Add live progress and observable-state timeout diagnostics

## Problem

Partio's post-commit hook performs several potentially long-running operations (session parsing, checkpoint object creation, branch update, push). When these operations are slow or hang, the user sees no output — just a frozen terminal. If a timeout fires, the error message is generic and doesn't help diagnose whether the issue is network, disk I/O, or a stuck subprocess.

## Desired Behavior

1. **Live progress on TTY**: Real-time phase events during checkpoint creation (detecting session → parsing JSONL → writing objects → updating ref → done). In-place line updates on TTY; one line per event when piped.

2. **Observable-state timeout diagnostic**: When a configurable timeout fires, the error message reflects what the operation was actually doing — distinguishing "stuck parsing large JSONL" from "network timeout during push" from "waiting for git lock".

3. **Opt-in timeout**: Operations don't have a hard cap by default. Users can set `checkpoint_timeout_seconds` in settings to enable timeout with diagnostics.

## Context Hints

- `internal/hooks/post_commit.go` — main operation sequence
- `internal/checkpoint/` — checkpoint creation pipeline
- `internal/session/` — session parsing
- `internal/log/` — existing slog-based logging

## Source

Inspired by entireio/cli#964 (changelog 0.6.3) — live progress and observable-state timeout diagnostic for explain --generate.
