---
id: clean-smudge-content-filter
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint transcripts can be filtered through a configurable clean/smudge pipeline before storage"
  - "Path-based filter rules allow excluding or transforming content from specific file paths in transcripts"
  - "Filters apply consistently across checkpoint write and read operations"
  - "Filter configuration is defined in .partio/settings.json under a dedicated key"
  - "Default behavior (no filters configured) is unchanged"
pr_labels:
  - minion
---

# Add clean/smudge content filter pipeline for checkpoint transcripts

## Description

Implement a git-inspired clean/smudge filter pipeline that processes checkpoint transcript content before storage (clean) and after retrieval (smudge). This allows teams to:

- Strip or transform content from specific file paths referenced in transcripts
- Apply consistent content policies (e.g., removing internal paths, normalizing environment-specific data)
- Layer path-based rules that determine which filters apply to which content

The pipeline should integrate with the existing checkpoint write path in `internal/checkpoint/` and apply before any redaction rules.

## Why

Current redaction rules operate at the secret/pattern level. Teams need broader content filtering — for example, stripping large tool-call outputs for specific file types, or normalizing absolute paths in transcripts. A clean/smudge model is well-understood by Git users and provides a composable, rule-based approach.

## Context hints

- `internal/checkpoint/` — checkpoint storage where clean filters would apply on write
- `internal/config/` — where filter configuration would be loaded
- `internal/agent/claude/parse_jsonl.go` — JSONL parsing that filter pipeline would wrap
