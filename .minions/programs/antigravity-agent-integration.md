---
id: antigravity-agent-integration
target_repos:
  - cli
acceptance_criteria:
  - "Partio detects running Antigravity (agy) CLI processes via the agent detector interface"
  - "Session transcripts from Antigravity sessions are parsed and stored in checkpoints"
  - "`partio enable --agent antigravity` configures Antigravity-specific hooks"
  - "`partio status` shows Antigravity as an active agent when detected"
  - "Antigravity hook lifecycle events (session start, tool use, session end) are captured"
pr_labels:
  - minion
---

# Add Antigravity (agy) CLI agent integration

Add first-class support for Google's Antigravity CLI agent (`agy`) in Partio's agent detection and session capture pipeline.

## Context

Antigravity is Google's successor AI coding CLI agent. As the AI coding agent ecosystem expands beyond Claude Code, Codex, and Cursor, Partio should support emerging agents to remain the universal session capture tool. The Entire CLI added Antigravity support in PR #1287.

## Desired behavior

- Implement an `antigravity` detector in `internal/agent/` following the existing detector interface pattern.
- Detect running `agy` processes and locate Antigravity's session/transcript directory.
- Parse Antigravity session transcripts into Partio's checkpoint format.
- `partio enable --agent antigravity` installs the appropriate hook configuration for Antigravity.
- `partio status` detects and displays Antigravity sessions alongside other agents.
- Hook lifecycle integration captures session start, tool invocations, and session end.

## Why

Partio's value grows with the number of agents it supports. Developers increasingly use multiple AI coding agents, and Antigravity is a significant new entrant from Google. Supporting it keeps Partio relevant as the agent-agnostic session capture layer.
