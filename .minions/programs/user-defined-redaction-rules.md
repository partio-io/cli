---
id: user-defined-redaction-rules
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - Users can define custom redaction rules in .partio/settings.json under a "redaction_rules" key
  - Each rule specifies a regex pattern and an optional replacement label (e.g., "[REDACTED:API_KEY]")
  - Redaction is applied to session transcript content before checkpoint creation
  - Built-in default rules cover common secret patterns (API keys, tokens, passwords in env vars)
  - Redaction rules are documented in partio doctor output when enabled
  - Unit tests verify that custom patterns are applied and that default patterns catch common secrets
---

# User-defined redaction rules for session transcripts

Allow users to define custom regex-based redaction rules that strip sensitive content from agent session transcripts before they are stored in checkpoints.

## Context

Entireio/cli v0.6.2 added user-defined redaction rules and rule packs, and v0.7.4/PR #1214 added an OpenAI Privacy Filter layer. Partio stores full Claude Code JSONL transcripts in checkpoints on the orphan branch. These transcripts may contain secrets, API keys, or other sensitive data that users don't want persisted in git history — especially when `push_sessions` is enabled.

## What to implement

1. Add a `redaction_rules` configuration field to `internal/config/` that accepts an array of `{pattern, label}` objects.
2. Include sensible built-in defaults (e.g., patterns for `*_API_KEY`, `*_SECRET`, `*_TOKEN`, bearer tokens, base64-encoded credentials).
3. Apply redaction to JSONL transcript content in `internal/agent/claude/parse_jsonl.go` before the content is passed to checkpoint creation.
4. Report active redaction rules in `partio doctor` output so users can verify their configuration.
