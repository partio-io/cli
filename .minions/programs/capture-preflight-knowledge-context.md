---
id: capture-preflight-knowledge-context
target_repos:
  - cli
acceptance_criteria:
  - "Pre-commit hook captures CLAUDE.md and .claude/settings.json content hashes at session start"
  - "Checkpoint metadata includes a knowledge_context field with file paths and content hashes"
  - "partio status shows whether knowledge context capture is enabled"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Capture pre-flight knowledge context in checkpoint metadata

Add an optional metadata field to checkpoint records that captures the structured knowledge context available to the agent at session start — CLAUDE.md files, project instructions, and configuration that shaped the agent's behavior.

## Why

Checkpoints currently record what the agent decided and did (transcripts, tool calls, reasoning) but not what the agent knew before execution. For audit, reproducibility, and debugging, the complete record should answer: "What knowledge was the agent operating from?" A checkpoint that includes knowledge state becomes a complete audit trail.

## What to implement

1. During pre-commit detection, collect paths and SHA-256 content hashes of knowledge context files present in the repo (e.g., `CLAUDE.md`, `.claude/settings.json`, `.claude/settings.local.json`).
2. Store this as a `knowledge_context` field in the checkpoint metadata JSON — an array of `{path, hash}` objects.
3. Add a config option `capture_knowledge_context` (default: `true`) to enable/disable this capture.
4. Keep the implementation minimal — store hashes only (not file contents) to avoid bloating checkpoint data.

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
