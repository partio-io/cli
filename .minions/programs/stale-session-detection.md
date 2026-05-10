---
id: stale-session-detection
target_repos:
  - cli
acceptance_criteria:
  - partio status shows a stale indicator for sessions that have not had interaction within a configurable timeout (default 30 minutes)
  - Stale sessions are visually distinguished from active sessions in status output
  - The stale timeout is configurable via settings (stale_session_timeout)
  - Stale detection uses session LastInteractionTime, not process detection alone
pr_labels:
  - minion
---

# Show stale session indicators in partio status

## Summary

Enhance `partio status` to detect and indicate stale sessions — agent sessions that were started but have had no recent interaction. This helps users understand whether a detected session is genuinely active or was abandoned (e.g., a terminal left open with Claude Code idle).

## Why

The current binary detection (agent running or not) can be misleading. A user may have Claude Code running in a background terminal with no recent interaction while actively working in a different terminal. Showing "active session" in this case creates false confidence that changes are being tracked. Stale detection gives users accurate session state information.

## Context

- Session state is persisted in `.partio/state/` with timestamps
- The upstream project (entireio/cli) added stale session indicators in v0.5.4
- Process detection currently lives in `internal/agent/claude/` and checks if Claude Code is running
- Stale detection should layer on top of process detection: process running + recent interaction = active, process running + no recent interaction = stale

## References

- entireio/cli changelog 0.5.4: "Stale session indicator in `entire status` output"
- entireio/cli#1118: Stale sessions inheriting other sessions' files (related concern)
