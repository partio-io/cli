---
id: configurable-hook-timeout
target_repos:
  - cli
acceptance_criteria:
  - Hook execution respects a configurable timeout value from settings (e.g., `hook_timeout_seconds`)
  - Default timeout is 30 seconds to match current implicit behavior
  - When a hook times out, it logs a warning with the elapsed time and exits cleanly without blocking the git operation
  - The timeout value can be set via layered config (global, repo, local settings) and PARTIO_HOOK_TIMEOUT env var
  - Unit tests verify timeout behavior using a mock long-running operation
pr_labels:
  - minion
---

# Add configurable hook execution timeout with graceful degradation

## Summary

Add a configurable timeout for Partio hook execution (pre-commit, post-commit, pre-push) so that slow session detection, checkpoint creation, or push operations don't block git workflows indefinitely.

## Motivation

Users of similar tools (entireio/cli#957, entireio/cli#956) report repeated hook timeouts when agent detection or checkpoint operations take longer than expected — particularly with agents like Codex that have heavier session state. Partio's hooks currently have no explicit timeout control; if session discovery or checkpoint writes stall (e.g., large JSONL parsing, slow git plumbing on large repos), the hook blocks the entire git operation with no user feedback.

A configurable timeout lets users tune the tradeoff between checkpoint completeness and git responsiveness, while graceful degradation ensures hooks never break the developer workflow.

## Implementation hints

- Add `hook_timeout_seconds` to the config schema in `internal/config/` with a default of 30
- Add `PARTIO_HOOK_TIMEOUT` environment variable override
- In each hook implementation (`internal/hooks/`), wrap the main operation in a context with deadline
- On timeout, log a warning (slog) with the hook name and elapsed time, then exit 0 to avoid blocking git
- Consider showing the timeout value in `partio doctor` output for debuggability
