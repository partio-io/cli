---
id: type-safe-transcript-redaction
target_repos:
  - cli
acceptance_criteria:
  - Transcript data flows through typed wrapper structs that enforce redaction state at compile time
  - Raw transcript content cannot be written to checkpoint storage without passing through the redaction pipeline
  - The redaction pipeline is applied during JSONL parsing before checkpoint creation
  - Existing tests continue to pass with the new type boundaries in place
  - A unit test verifies that creating a checkpoint with unredacted transcript data produces a compile error or is prevented by the type system
pr_labels:
  - minion
---

# Type-safe transcript redaction pipeline

## Summary

Add Go type-level enforcement for transcript redaction boundaries so that sensitive content in session transcripts cannot accidentally bypass redaction before checkpoint storage.

## Problem

Partio parses Claude Code JSONL session files and stores them in checkpoints. Currently, the raw JSONL content flows through string/byte types with no compile-time guarantee that redaction (e.g., stripping API keys, tokens, or PII patterns) has been applied before the data reaches the checkpoint storage layer.

As Partio grows to handle more agents and larger transcripts, the risk of accidentally writing unredacted content increases. The entireio/cli project recently refactored their condensation logic with "type-enforced redaction boundaries" to address this exact risk.

## Proposed Solution

Introduce typed wrapper structs that track redaction state:

```go
// RawTranscript represents unprocessed transcript content that may contain sensitive data.
type RawTranscript struct {
    content []byte
}

// RedactedTranscript represents transcript content that has passed through the redaction pipeline.
type RedactedTranscript struct {
    content []byte
}

// Redact processes a RawTranscript through the redaction pipeline and returns a RedactedTranscript.
func Redact(raw RawTranscript, rules []RedactionRule) RedactedTranscript { ... }
```

The checkpoint storage functions should accept only `RedactedTranscript`, making it a compile error to pass raw content directly.

## Context

- `internal/agent/claude/parse_jsonl.go` — JSONL parsing
- `internal/checkpoint/` — checkpoint storage
- `internal/session/` — session data types
