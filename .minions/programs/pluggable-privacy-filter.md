---
id: pluggable-privacy-filter
target_repos:
  - cli
acceptance_criteria:
  - "A privacy filter interface is defined that accepts transcript content and returns redacted content"
  - "At least one built-in filter is implemented (e.g., regex-based secret pattern matching)"
  - "Filters are configurable via partio settings"
  - "Filters run automatically before transcript data is written to checkpoint storage"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add pluggable privacy filter pipeline for checkpoint transcripts

Add a configurable pipeline of privacy/redaction filters that process transcript content before it is stored in checkpoints. This goes beyond the existing user-defined redaction rules by supporting multiple composable filter layers that can be enabled or disabled independently.

## Context

Agent session transcripts may inadvertently contain secrets, API keys, credentials, or other sensitive data. A pluggable filter pipeline allows users to configure detection layers (regex patterns, entropy-based detection, allow/deny lists) that automatically scrub sensitive content before it reaches the checkpoint branch.

## Implementation hints

- Define a `TranscriptFilter` interface in `internal/checkpoint/` (or a new `internal/filter/` package) with a `Filter(content []byte) ([]byte, error)` method.
- Implement a `RegexFilter` that matches configurable patterns (env vars, API key formats, high-entropy strings).
- Wire filters into the checkpoint creation path, after JSONL parsing but before `git hash-object`.
- Add a `privacy_filters` section to the config schema listing enabled filters and their settings.
- Ensure filters are tested with table-driven tests covering match, no-match, and edge cases.
