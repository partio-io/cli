---
id: transcript-condensation
target_repos:
  - cli
acceptance_criteria:
  - Long transcripts above a configurable threshold are condensed by summarizing older portions while keeping recent content verbatim
  - Condensed transcripts are stored alongside or replace the full transcript in the checkpoint
  - Original full transcript is preserved in a separate blob if condensation is enabled
  - Condensation runs during checkpoint creation, not retroactively
  - A `condense_threshold` setting controls when condensation activates (e.g., transcript > 100KB)
  - Unit tests verify condensation preserves recent content and summarizes old content
pr_labels:
  - minion
---

# Transcript condensation for long sessions

## Summary

Add a condensation mechanism that summarizes older parts of long session transcripts while preserving recent content in full detail, reducing checkpoint storage size and improving browsability.

## Problem

Long Claude Code sessions can produce very large JSONL transcripts (hundreds of KB to several MB). Storing the full transcript in every checkpoint leads to:

1. **Storage bloat** on the checkpoint orphan branch — each checkpoint contains a full copy of the growing transcript
2. **Slow browsing** — reviewing a checkpoint with a massive raw transcript is unwieldy
3. **Redundancy** — early parts of a long session are often captured in previous checkpoints already

The entireio/cli project addresses this with their checkpoints v2 "compact transcript" format and condensation logic.

## Proposed Solution

1. During checkpoint creation, if the transcript exceeds `condense_threshold`:
   - Split the transcript into "old" (already captured in previous checkpoints) and "recent" (since last checkpoint) segments
   - Store the recent segment verbatim in the checkpoint
   - Generate a compact summary of the old segment (structured metadata: turn count, tools used, files touched — not AI-generated prose)
   - Write both as separate blobs in the checkpoint tree

2. Configuration:
```json
{
  "strategy_options": {
    "condense_threshold": "100KB"
  }
}
```

3. The full transcript is still available via the session JSONL on disk; condensation only affects what's stored in checkpoint blobs.

## Context

- `internal/checkpoint/` — checkpoint storage and tree construction
- `internal/agent/claude/parse_jsonl.go` — JSONL parsing
- `internal/session/` — session data
- `internal/config/` — configuration
