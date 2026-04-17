---
id: scrub-pii-from-transcripts
target_repos:
  - cli
acceptance_criteria:
  - Partio redacts detected PII (emails, phone numbers) and API key patterns from session transcripts before writing checkpoint data
  - Redaction runs on transcript content during checkpoint creation, before any git object is written
  - A configuration option `redact_sensitive_data` (default true) in settings allows users to disable redaction
  - Redacted values are replaced with a placeholder like `[REDACTED]` that preserves line structure
  - Unit tests confirm known PII and API key patterns are scrubbed and that non-sensitive content passes through unchanged
pr_labels:
  - minion
---

# Scrub PII and API keys from session transcripts before checkpoint storage

## Problem

Partio captures agent session transcripts and stores them as checkpoint data in git. Users may inadvertently paste PII (emails, phone numbers, addresses) or API keys/secrets into their prompts. These sensitive values then get committed to the checkpoint branch and potentially pushed to remote repositories, creating a security and compliance risk.

## Proposed solution

Add a redaction pass during checkpoint creation that scans transcript content for common sensitive data patterns before writing git objects:

1. **Pattern-based detection**: Use regex patterns to detect common secret formats (AWS keys, GitHub tokens, generic API keys with high-entropy strings) and PII formats (email addresses, phone numbers).
2. **Integration point**: Run redaction in the post-commit hook path, after JSONL parsing but before `hash-object` writes the transcript blob.
3. **Configuration**: Add `redact_sensitive_data` boolean to the config layer (default: `true`). Users in environments where transcripts are intentionally sensitive (e.g., security research) can opt out.
4. **Placeholder format**: Replace matches with `[REDACTED:<type>]` (e.g., `[REDACTED:api-key]`, `[REDACTED:email]`) to preserve structure and make redaction visible during review.

## Why this matters

Partio's value proposition is preserving the *why* behind code changes. But if that preservation inadvertently leaks secrets or PII into git history, it becomes a liability. Proactive redaction reduces the risk of accidental credential exposure through checkpoint data, which is especially important when `push_sessions` is enabled and checkpoint branches are pushed to shared remotes.

## Context hints

- `internal/checkpoint/` — checkpoint creation and storage
- `internal/agent/claude/parse_jsonl.go` — JSONL transcript parsing
- `internal/config/` — configuration layer for the new setting
