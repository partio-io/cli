---
id: transcript-compaction
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint transcript data is compacted before storage, reducing size while preserving essential content
  - Compaction removes redundant whitespace, truncates large tool outputs, and deduplicates repeated content
  - Original prompt and key assistant responses are preserved intact
  - Compacted transcripts are readable by existing checkpoint display commands without errors
  - Checkpoint storage size is measurably reduced for typical Claude Code sessions
pr_labels:
  - minion
---

# Compact checkpoint transcripts for storage efficiency

## Problem

Claude Code JSONL session transcripts can grow large — long sessions with many tool calls, file reads, and verbose outputs generate substantial data. Storing the full raw transcript in every checkpoint inflates the orphan branch size over time, slowing fetches and increasing storage costs.

## Solution

Add a compaction step to the checkpoint creation pipeline that produces a condensed transcript representation. The compacted form retains the essential session narrative (prompts, key decisions, tool names, outcomes) while stripping verbose intermediary content (full file contents from reads, large command outputs, repeated context).

### Implementation hints

- Add compaction logic in `internal/checkpoint/` or `internal/agent/claude/` that processes raw JSONL before checkpoint storage
- Truncate tool result payloads beyond a configurable threshold (e.g., 500 chars)
- Collapse consecutive tool_result entries that contain identical or near-identical content
- Preserve the first and last N lines of large outputs, replacing the middle with a `[truncated]` marker
- Make compaction opt-in via config (`compaction: true` in settings) to avoid surprising users
- Consider a `partio status` field showing checkpoint storage size to make the benefit visible

## Source

Inspired by entireio/cli#980 (external agent transcript compaction support) and changelog entries for compact transcript support in v0.5.3–0.5.5.
