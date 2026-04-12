---
id: transcript-growth-bounding
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint storage respects a configurable maximum transcript size"
  - "When transcript exceeds the limit, older generations are rotated out while preserving the latest content"
  - "Generation rotation is transparent — checkpoint reading still works with rotated transcripts"
  - "Default behavior is unchanged for small transcripts (no rotation unless threshold exceeded)"
  - "make test passes with rotation tests"
  - "make lint passes"
pr_labels:
  - minion
---

# Bound checkpoint transcript growth via generation rotation

Add generation rotation to limit unbounded growth of transcript data stored in checkpoints.

## What to implement

1. **Size tracking** — track the cumulative size of transcript JSONL data per session as checkpoints are created
2. **Generation rotation** — when transcript data exceeds a configurable threshold, rotate older content out of the active checkpoint while preserving the most recent generation
3. **Configuration** — add a `max_transcript_size` setting (with a sensible default, e.g. 10MB) to the layered config system
4. **Backward compatibility** — existing checkpoints with large transcripts should remain readable; rotation only applies to new checkpoint writes

## Why this matters

Long-running agent sessions produce large JSONL transcript files that grow unboundedly. This inflates the git object database, slows checkpoint operations, and can cause performance issues during push. Generation rotation keeps storage bounded while preserving the most valuable recent context.

## Context hints

- `internal/checkpoint/` — checkpoint storage and creation
- `internal/agent/claude/parse_jsonl.go` — JSONL parsing
- `internal/config/` — layered configuration system
