---
id: omit-zero-token-count-in-checkpoint-metadata
target_repos:
  - cli
acceptance_criteria:
  - "`SessionMetadata.TotalTokens` is changed from `int` to `*int` so zero vs unknown can be distinguished"
  - "When token data is not available (nil pointer), `total_tokens` is omitted from the JSON output entirely (`omitempty`)"
  - "When token data is zero but explicitly set, it serializes as `0` (e.g. a session that used 0 tokens)"
  - "Existing checkpoint metadata files with `total_tokens: 0` deserialize without error (nil pointer, not a panic)"
  - "The post-commit hook only sets `TotalTokens` when the parsed session actually reports token usage"
  - "Unit tests in `internal/checkpoint/` or `internal/hooks/` cover the nil case"
pr_labels:
  - minion
---

# Omit token count in checkpoint metadata when no token data exists

## Source

entireio/cli PR #854: "fix: omit token count when no token data exists" — tokens are only calculated during checkpoint creation or condensation. Sessions with no file changes never get token data, so showing "tokens 0" is misleading — the real count is unknown.

## Problem

`internal/checkpoint/metadata.go` defines `TotalTokens int`, which defaults to `0` for any session where token tracking didn't occur (e.g. the session produced no new content, or the JSONL parsing found no token usage entries). This is indistinguishable from a session that genuinely used 0 tokens, making the displayed count misleading.

## Proposed Change

Change `SessionMetadata.TotalTokens` from `int` to `*int` in `internal/checkpoint/metadata.go`:

```go
type SessionMetadata struct {
    Agent       string `json:"agent"`
    TotalTokens *int   `json:"total_tokens,omitempty"`
    Duration    string `json:"duration"`
}
```

Update `internal/hooks/postcommit.go` (and any other callers that set `TotalTokens`) to only assign the field when token data is actually present from the JSONL parse result. When `internal/agent/claude/parse_jsonl.go` returns a token count of 0 because no token events were found, leave `TotalTokens` as nil instead of setting it to 0.

Any code that reads `TotalTokens` for display (e.g. `partio rewind --list`) should treat `nil` as "unknown" and skip the token field rather than showing "0 tokens".

## Why This Matters

Displaying "0 tokens" implies the agent ran but consumed no tokens, which is confusing and inaccurate. Omitting the field when unknown gives users correct signal: if they see a token count it's real data, if they see nothing it means the data wasn't captured.

## Context

- `internal/checkpoint/metadata.go` — `SessionMetadata` struct with `TotalTokens int`
- `internal/hooks/postcommit.go` — sets `TotalTokens` during checkpoint creation
- `internal/agent/claude/parse_jsonl.go` — JSONL parsing that extracts token usage
- `cmd/partio/rewind.go` — displays checkpoint metadata to users
