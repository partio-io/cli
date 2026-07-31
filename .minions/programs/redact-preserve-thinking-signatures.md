---
id: redact-preserve-thinking-signatures
target_repos: [cli]
acceptance_criteria:
  - Known signature field names (e.g. "signature", "thinkingSignature") are skipped by the entropy scanner during transcript redaction
  - A JSONL transcript containing extended-thinking blocks with high-entropy base64 signature values passes redaction without those fields being modified
  - Existing redaction behaviour for genuine secrets is unchanged
  - Unit test covers a thinking block with a high-entropy base64 "thinkingSignature" field and asserts it is not redacted
pr_labels:
  - minion
---

# Preserve extended-thinking signature fields during transcript redaction

When Partio redacts session transcripts before storing checkpoints, the entropy scanner should skip known signature field names used by AI agents in extended-thinking blocks. Currently, if only the exact field name `"signature"` is skipped, other agents that store the extended-thinking signature under a different key (such as `"thinkingSignature"`) fall through to entropy scanning. Because signature values are high-entropy base64, the scanner replaces segments with `REDACTED`, corrupting the cryptographic signature. A corrupted thinking-block signature causes the agent to reject the transcript on replay with a 400-class error.

Adapt the redaction skip-list in `internal/session/` (or wherever `shouldSkipJSONLField` / transcript redaction lives) to include all known signature field names used by supported agents.

## Why

Corrupt thinking-block signatures break session replay and `partio rewind`, destroying the very history Partio is meant to preserve. The fix is low-risk: extending a known-safe skip-list to cover additional field names that are definitionally not secrets.

## User Relevance

Users running Claude Code or other agents with extended thinking enabled will find that `partio rewind` or any feature that replays the stored transcript fails with a cryptic API error. This fix ensures transcripts remain valid after redaction.

## Source

entireio/cli PR #1866 "redact: preserve thinking-block signatures (fix omp replay 400)" and CHANGELOG 0.9.0.

## Acceptance Criteria

- Known signature field names (e.g. "signature", "thinkingSignature") are skipped by the entropy scanner during transcript redaction
- A JSONL transcript containing extended-thinking blocks with high-entropy base64 signature values passes redaction without those fields being modified
- Existing redaction behaviour for genuine secrets is unchanged
- Unit test covers a thinking block with a high-entropy base64 "thinkingSignature" field and asserts it is not redacted
