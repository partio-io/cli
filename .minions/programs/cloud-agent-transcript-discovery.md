---
id: cloud-agent-transcript-discovery
target_repos:
  - cli
acceptance_criteria:
  - Detector interface supports a TranscriptSource method indicating local-file or remote-api
  - Checkpoint creation gracefully handles agents with no local transcript (stores metadata without transcript)
  - partio status shows transcript availability status per session
  - Hook execution does not fail or hang when transcript files are absent for a detected agent
pr_labels:
  - minion
---

# Support transcript discovery for cloud-based agents

## Summary

Extend the agent detection and session discovery pipeline to handle agents whose transcripts are not stored as local files. Cloud-hosted agents (e.g., Copilot Cloud Agent, remote coding assistants) may be detected as running but have no local JSONL session files — currently this causes silent failures or empty checkpoint transcripts.

## What to implement

1. Extend `agent.Detector` interface in `internal/agent/detector.go` with a method to indicate transcript source type:
   - `TranscriptSource() string` returning `"local"`, `"remote"`, or `"none"`
   - The existing Claude Code detector returns `"local"`

2. Update `internal/session/` to handle the case where `FindSessionDir` returns empty for a detected agent:
   - Still create the checkpoint with commit metadata and attribution
   - Skip transcript parsing rather than treating it as an error
   - Log a debug message indicating transcript was unavailable

3. Update `internal/hooks/post_commit.go` to check transcript source before attempting JSONL parsing — prevent errors when the agent is detected but session files don't exist locally.

4. Show transcript availability in `partio status` so users know whether their sessions are being captured with full context or metadata-only.

## Context

- `internal/agent/detector.go` — Detector interface
- `internal/agent/claude/` — Claude Code implementation (local transcript)
- `internal/session/` — session lifecycle and state
- `internal/hooks/post_commit.go` — post-commit hook implementation
- `internal/checkpoint/` — checkpoint creation

## Why

As more AI coding agents move to cloud-hosted architectures, Partio's assumption that detected agents always have local JSONL transcripts will increasingly fail. Handling this gracefully ensures checkpoints still capture attribution and commit metadata even when full transcripts aren't available locally. Inspired by entireio/cli issue #1006 (transcript files not found for Copilot Cloud Agent).
