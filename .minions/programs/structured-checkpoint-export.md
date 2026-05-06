---
id: structured-checkpoint-export
target_repos:
  - cli
acceptance_criteria:
  - "`partio export --session <id> --format json` outputs a complete session record as structured JSON"
  - "`partio export --checkpoint <id> --format json` outputs a single checkpoint record as structured JSON"
  - "Export includes: repo path, session ID, checkpoint IDs, branch/commit refs, timestamps, files touched, agent/model metadata, and prompt/transcript references"
  - "Output schema is documented and stable for downstream consumption"
  - "Export works without requiring a running agent session"
pr_labels:
  - minion
---

# Add `partio export` command for structured checkpoint/session data

## Description

Add a `partio export` command that outputs checkpoint and session data in a stable, machine-readable JSON format. This enables downstream tools — CI pipelines, review automation, handoff workflows, eval frameworks — to consume Partio's captured context without parsing internal storage formats.

The export should read from the checkpoint orphan branch and session state, assembling a complete record that includes:

- Repository and branch context
- Session metadata (ID, agent, model, timestamps)
- Checkpoint records (IDs, commit SHAs, attribution)
- File change summaries
- Transcript references (paths, not full content — to keep exports lightweight)

## Why

Partio captures rich context about AI-assisted code changes, but this data is currently only accessible through Partio's own CLI commands or by reading the internal git plumbing storage directly. A structured export surface makes this data a building block for broader workflows: cross-tool handoffs, session evaluation, audit reporting, and integration with review systems.

## Context hints

- `internal/checkpoint/` — checkpoint storage and retrieval
- `internal/session/` — session state and metadata
- `cmd/partio/` — CLI command registration
