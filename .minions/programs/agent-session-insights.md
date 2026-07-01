---
id: agent-session-insights
target_repos:
  - cli
acceptance_criteria:
  - "`partio insights` analyzes stored checkpoint transcripts and surfaces patterns"
  - "Output includes metrics like average session length, common error patterns, and tool usage frequency"
  - "Insights are computed locally from checkpoint data on the orphan branch"
  - "Supports --json flag for machine-readable output"
  - "Works with existing checkpoint storage format without migration"
pr_labels:
  - minion
---

# Add `partio insights` command for session pattern analysis

## Summary

Add a `partio insights` command that analyzes stored checkpoint transcripts to surface patterns and learnings from past agent sessions. This enables developers to understand how AI agents are being used in their repositories, identify common workflows, and spot recurring issues.

## Motivation

Partio captures rich session data in checkpoints, but currently there's no way to learn from that accumulated data. Developers don't know which types of tasks agents handle well, where sessions tend to go wrong, or how agent usage patterns evolve over time. An insights command unlocks the value of historical checkpoint data.

## Behavior

1. `partio insights` reads checkpoint transcripts from the orphan branch
2. Analyzes session patterns: average turn count, common tool calls, error frequency
3. Surfaces actionable observations like "sessions involving file X tend to be longer"
4. `partio insights --json` outputs structured data for integration with other tools
5. Optional `--since` flag to scope analysis to recent checkpoints

## Context

- Inspired by `entireio/cli` PR #765 (Agent Improvement Engine with insights/improve/evolve commands)
- `internal/checkpoint/` for reading checkpoint data from the orphan branch
- `internal/agent/claude/parse_jsonl.go` for transcript parsing
