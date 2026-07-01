---
id: token-usage-metrics
target_repos:
  - cli
acceptance_criteria:
  - "Token usage (input and output tokens) is extracted from Claude Code session JSONL when available"
  - "Token counts are stored in checkpoint metadata"
  - "partio status or checkpoint listing surfaces token usage when present"
  - "Missing or zero token data is handled gracefully without errors"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Track token usage metrics in checkpoint metadata

Extract and store token usage statistics (input tokens, output tokens) from agent session transcripts as part of checkpoint metadata. This gives users visibility into the computational cost of each session alongside the code changes it produced.

## Context

Claude Code session JSONL files contain token usage information in API response entries. Currently Partio captures the transcript content but does not extract or surface these metrics. Storing token counts in checkpoint metadata enables cost tracking, usage analysis, and helps teams understand the resource footprint of AI-assisted development.

## Implementation hints

- In the JSONL parser (`internal/agent/claude/parse_jsonl.go`), look for usage fields in assistant response entries (typically `usage.input_tokens` and `usage.output_tokens`) and accumulate totals.
- Add `InputTokens` and `OutputTokens` fields to the `SessionData` type or checkpoint metadata.
- Surface token counts in `partio status` output when checkpoint data is available.
- Handle sessions where token data is missing or zero (older transcript formats, non-Claude agents).
