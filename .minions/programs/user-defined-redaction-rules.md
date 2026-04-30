---
id: user-defined-redaction-rules
target_repos:
  - cli
acceptance_criteria:
  - Users can define custom regex patterns in .partio/settings.json under a redaction.custom_patterns key
  - Custom patterns are applied to transcript content before checkpoint storage
  - A .partio/redactors/ directory supports shareable rule pack files (YAML or JSON)
  - Personal overrides via .partio/redactors/local/ are gitignored by default
  - Invalid regex patterns produce a clear warning without blocking the commit
  - Built-in patterns (API keys, tokens) remain active alongside custom rules
pr_labels:
  - minion
---

# User-defined redaction rules for checkpoint transcripts

## Problem

Teams have internal credential formats, proprietary token prefixes, and domain-specific secrets that Partio's built-in redaction cannot anticipate. Without user-configurable redaction, sensitive content specific to a team's infrastructure may be stored in checkpoint transcripts and pushed to shared remotes.

## Desired Behavior

Allow teams to define custom redaction patterns that are applied to transcript content before it is written to the checkpoint branch:

1. **Inline patterns** in settings: `redaction.custom_patterns` array of `{name, regex, replacement}` objects in `.partio/settings.json`.
2. **Rule packs** in `.partio/redactors/`: YAML/JSON files with named pattern sets that can be committed and shared across the team.
3. **Local overrides** in `.partio/redactors/local/` for personal patterns (directory added to `.gitignore` by `partio enable`).
4. Patterns are validated at load time — invalid regex logs a warning but does not block the hook.

## Context Hints

- `internal/checkpoint/` — where transcript content is written to git objects
- `internal/config/` — layered configuration loading
- `internal/session/` — session data that includes transcript content
