---
id: claude-token-usage-tracking
target_repos:
  - cli
acceptance_criteria:
  - ParseJSONL extracts input_tokens and output_tokens from the usage field inside assistant message entries in Claude JSONL transcripts
  - TotalTokens in the returned SessionData equals the sum of all input and output tokens across assistant turns
  - Sessions with no usage data return TotalTokens of 0 (backward-compatible)
  - Checkpoint metadata stored on the orphan branch reflects the actual token count for Claude Code sessions, not always 0
  - The assistantMessage struct in jsonl.go carries a Usage field that is populated by json.Unmarshal
pr_labels:
  - enhancement
---

## Summary

Partio's Claude JSONL parser (`internal/agent/claude/parse_jsonl.go`) declares a `totalTokens` variable but never assigns it from parsed data, so every Claude Code session records 0 tokens in checkpoint metadata. The Codex parser (`internal/agent/codex/parse_jsonl.go:117`) already extracts token counts correctly. This proposal closes the gap for Claude Code.

## What to implement

Extend `assistantMessage` in `internal/agent/claude/jsonl.go` to include a `Usage` struct:

```go
type assistantMessage struct {
    Model string       `json:"model"`
    Usage usageMetrics `json:"usage"`
}

type usageMetrics struct {
    InputTokens              int `json:"input_tokens"`
    OutputTokens             int `json:"output_tokens"`
    CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
    CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}
```

In the scan loop in `parse_jsonl.go`, when an assistant entry is decoded, accumulate its token counts:

```go
if json.Unmarshal(entry.Message, &am) == nil {
    if am.Model != "" {
        model = am.Model
    }
    totalTokens += am.Usage.InputTokens + am.Usage.OutputTokens +
        am.Usage.CacheCreationInputTokens + am.Usage.CacheReadInputTokens
}
```

No other files need to change. The populated `SessionData.TotalTokens` flows through `postcommit.go:170` into `checkpoint.SessionMetadata.TotalTokens` and is written to the orphan branch as part of the existing checkpoint write path.

## Why this matters

Checkpoint metadata is the primary per-commit record of AI usage. With `TotalTokens` always 0 for Claude Code sessions, no downstream feature that wants to reason about cost or usage can be built on that data. Several open proposals (#490, #643, #477, #447) depend on accurate token counts already being captured. This fix is the missing foundation.

<!-- program: .minions/programs/claude-token-usage-tracking.md -->
