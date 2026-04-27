---
id: display-active-agents-in-status
target_repos:
  - cli
acceptance_criteria:
  - "`partio status` output includes an 'Active agents' section listing currently detected running agents"
  - "When no agents are running, the section shows 'none' or is omitted"
  - "Detection reuses the existing agent detector interface without adding new dependencies"
  - "Output works correctly in both TTY and --json modes"
pr_labels:
  - minion
  - enhancement
---

# Display active agents in `partio status` output

## Summary

Enhance `partio status` to show which AI agents are currently detected as running in the repository's working directory. This gives users immediate visibility into whether Partio will capture their next commit.

## Motivation

Currently, `partio status` shows configuration and session state but doesn't tell the user whether an agent is actively running right now. Users have to infer this or check manually. Showing active agents directly in status output provides immediate feedback about capture readiness — "yes, your next commit will be captured" vs "no agent detected, commits won't have checkpoints."

Source: entireio/cli changelog 0.5.4 — "`entire status` now shows active agents"

## Implementation Notes

- Use the existing `agent.Detector` interface to check for running agents
- Add an "Active agents" line/section to the status output
- In `--json` mode (if/when implemented), include as an `active_agents` array field
- Keep detection fast — it should not noticeably slow down `partio status`
- Handle the case where detection fails gracefully (log warning, don't block output)

## Context

- `cmd/partio/status.go` — current status command implementation
- `internal/agent/detector.go` — agent detection interface
- `internal/agent/claude/` — Claude Code detection implementation
