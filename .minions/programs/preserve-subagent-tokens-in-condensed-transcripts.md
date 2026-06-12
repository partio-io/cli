---
id: preserve-subagent-tokens-in-condensed-transcripts
target_repos:
  - cli
acceptance_criteria:
  - "When parsing a condensed Claude Code JSONL transcript, token usage from sub-agent tool calls is aggregated even if the sub-agent result entries were removed during condensation"
  - "Checkpoint metadata token_usage field includes tokens from sub-agent sessions when available"
  - "Non-condensed transcripts continue to work identically"
  - "Unit tests cover condensed transcripts with missing sub-agent tool_result entries"
pr_labels:
  - minion
---

# Preserve sub-agent token totals when parsing condensed transcripts

When Claude Code condenses its JSONL transcript (removing older messages to stay within context limits), token usage information from sub-agent tool calls can be lost because the `tool_result` entries containing token summaries are removed. Partio should detect and handle this gap when extracting checkpoint metadata from transcripts.

## Implementation guidance

- During JSONL transcript parsing, track `tool_use` entries that invoke sub-agents (e.g., `Agent` tool calls)
- If a corresponding `tool_result` with token usage data exists, aggregate those tokens into the checkpoint's total
- If the `tool_result` is missing (condensed away), look for any remaining token summary entries or accept that sub-agent tokens for that segment are unavailable
- Store a `condensed: true` flag or similar indicator in metadata when condensation gaps are detected, so downstream consumers know the token count may be a lower bound

## Context Hints

- `internal/agent/claude/` — JSONL parsing and session discovery
- `internal/checkpoint/` — checkpoint metadata construction

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
