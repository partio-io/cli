---
id: skip-empty-sessions
target_repos:
  - cli
acceptance_criteria:
  - "Post-commit skips checkpoint creation when sessionData.Context and sessionData.Prompt are both empty strings"
  - "A debug log line 'skipping empty session' is emitted when the skip occurs"
  - "store.Write returns an error (rather than silently creating malformed tree entries) if cp.ID is empty or shorter than 3 characters"
  - "Unit tests cover the empty-session skip path and the empty-ID guard"
pr_labels:
  - minion
---

# Skip empty sessions and prevent phantom checkpoint paths

Add a guard in the post-commit hook: before writing a checkpoint, verify that the parsed session data contains at least one meaningful message (non-empty context or prompt). If the session is empty (e.g., a freshly started Claude session with no conversation), skip checkpoint creation entirely and log a debug message. Also validate that the checkpoint shard/rest path components are non-empty strings before calling store.Write to prevent empty path entries on the orphan branch.

## Why

Empty or near-empty Claude sessions produce useless checkpoint entries with blank context and prompt fields. Phantom paths on the orphan branch corrupt the checkpoint tree and waste storage.

## User Relevance

Users browsing checkpoints with `partio rewind --list` see only meaningful entries, not noise from sessions where the agent was technically running but had not yet done any work.

## Context Hints

- `internal/hooks/postcommit.go`
- `internal/checkpoint/store.go`
- `internal/checkpoint/write.go`
- `internal/agent/claude/parse_jsonl.go`

## Source

Inspired by entireio/cli#958
