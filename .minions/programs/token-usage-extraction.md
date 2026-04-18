---
id: token-usage-extraction
target_repos:
  - cli
acceptance_criteria:
  - "ParseJSONL sums input_tokens + output_tokens from all assistant message usage blocks into TotalTokens"
  - "SessionMetadata stores InputTokens and OutputTokens as separate fields in addition to TotalTokens"
  - "A unit test with a synthetic JSONL fixture containing known usage values verifies the correct totals are returned"
  - "Sessions with no usage fields still parse successfully with TotalTokens = 0"
pr_labels:
  - minion
---

# Extract token usage from Claude JSONL transcripts

The SessionData.TotalTokens field may not properly accumulate token usage from Claude Code JSONL entries. Claude Code JSONL includes usage objects (input_tokens, output_tokens, cache_read_input_tokens, cache_creation_input_tokens) on assistant messages and summary entries. ParseJSONL should accumulate these into TotalTokens, and SessionMetadata should also expose input/output token counts separately for richer checkpoint metadata.

## Why

Token usage is a key signal for understanding AI contribution cost and session complexity. Having it always at 0 renders the field meaningless in the checkpoint metadata and any downstream dashboard display.

## User Relevance

The app dashboard can show meaningful token usage per checkpoint, allowing teams to track AI compute costs per commit and session.

## Context Hints

- `internal/agent/claude/parse_jsonl.go`
- `internal/agent/claude/jsonl.go`
- `internal/checkpoint/metadata.go`

## Source

Inspired by entireio/cli#956
