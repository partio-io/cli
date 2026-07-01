---
id: stream-driven-session-summaries
target_repos:
  - cli
acceptance_criteria:
  - Partio can generate incremental session summaries from JSONL transcript data as it grows
  - Summaries update as new entries are appended to the session transcript during an active session
  - Summary generation works for Claude Code sessions and is extensible to other agents via the detector interface
  - A timeout configuration controls how long summary generation waits before giving up
pr_labels:
  - minion
---

# Add stream-driven session summaries from live transcripts

## Problem

Currently, Partio captures session transcripts as static JSONL snapshots at checkpoint time. Understanding what happened in a session requires reading through the raw transcript. Post-hoc summary generation (if added) would need to process the entire transcript after the session ends, adding latency and requiring an external LLM call.

## Solution

Add the ability to generate incremental session summaries by processing the JSONL transcript as it grows during an active agent session. This enables:

- Real-time progress visibility in `partio status` showing what the agent is currently working on
- Lightweight summaries stored in checkpoint metadata without requiring a separate LLM call
- Faster post-session understanding since the summary is already available when the session ends

### Implementation hints

- In `internal/agent/claude/`, add a transcript watcher that tails the active session JSONL file
- Extract key events (tool uses, file edits, user messages) to build a running summary
- Store the current summary in `.partio/state/` alongside the session state
- Include the summary in checkpoint metadata when a checkpoint is created
- Add a `summary_timeout_seconds` config option (default: 300s) to control how long the watcher runs
- Expose the current summary via `partio status` when a session is active
- Design the summary interface in `internal/agent/detector.go` so other agents can implement their own summary extraction

## Inspiration

Adapted from entireio/cli PR #1230 (stream-driven summaries for multiple agents) and changelog 0.6.2 (`entire explain --generate` honoring `summary_timeout_seconds`).
