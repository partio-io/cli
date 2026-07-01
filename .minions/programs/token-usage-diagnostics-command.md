---
id: token-usage-diagnostics-command
target_repos:
  - cli
acceptance_criteria:
  - partio session tokens displays token usage breakdown for a given session including input/output/cache tokens
  - partio checkpoint tokens displays committed token metrics for a given checkpoint
  - Output includes per-turn token counts when available in the JSONL transcript
  - JSON output mode is supported via --json flag
  - Command gracefully handles sessions/checkpoints where token data is unavailable
pr_labels:
  - minion
  - minion-proposal
---

# Token usage diagnostics command

## Problem

Partio captures session transcripts that contain token usage information in the JSONL data, but there is no way for users to inspect token consumption patterns. Users want to understand how many tokens their AI agent sessions are consuming, which turns are most expensive, and whether there are optimization opportunities (e.g., large context windows, repeated tool calls).

## Proposed solution

Add two new CLI commands:

1. **`partio session tokens [session-id]`** - Reports token usage for a live or recent session:
   - Total input/output/cache tokens
   - Per-turn breakdown showing which turns consumed the most tokens
   - Identifies likely cost contributors (large file reads, repeated tool calls)
   - Notes any limitations in the data (e.g., missing fields in older transcript formats)

2. **`partio checkpoint tokens <checkpoint-id>`** - Reports committed token metrics from checkpoint metadata:
   - Token counts stored in checkpoint metadata
   - Comparison across checkpoints in the same session to show consumption over time

Both commands should support `--json` for machine-readable output.

## Why this matters

Token usage directly correlates with cost and latency. Giving users visibility into token consumption patterns helps them make informed decisions about their AI coding workflows and identify sessions that are unexpectedly expensive.

## Context hints

- `internal/agent/claude/parse_jsonl.go` - JSONL transcript parsing
- `internal/checkpoint/` - Checkpoint domain types and storage
- `internal/session/` - Session lifecycle management
- `cmd/partio/` - CLI command definitions
