---
id: incremental-jsonl-parsing
target_repos:
  - cli
acceptance_criteria:
  - "Post-commit hook only reads and parses JSONL lines written since the previous checkpoint, not the entire transcript"
  - "A byte-offset or line-count cursor is persisted (in pre-commit state or checkpoint metadata) so subsequent checkpoints know where to resume"
  - "Checkpoint data is still correct — prompts and context extracted from the new window only"
  - "`make test` passes"
  - "`make lint` passes"
pr_labels:
  - minion
  - performance
---

# Incremental JSONL parsing: parse only new transcript lines per checkpoint

## Background

Inspired by entireio/cli#1836.

Partio's post-commit hook reads and parses the Claude Code JSONL session file to extract prompts and context for the checkpoint. As a session grows (many turns, many commits), the JSONL file grows proportionally. If post-commit always re-parses the full file from the start, the total work across a session is O(N²) in the number of turns — each of N commits re-parses all N×(lines-per-turn) lines.

## What to implement

Track the JSONL byte offset (or line number) at checkpoint creation time and persist it so the next post-commit run only reads lines appended since the last checkpoint.

- In `pre-commit`, record the current JSONL file size (byte offset) alongside the session state saved to `.partio/state/pre-commit.json`
- In `post-commit`, read only lines from that byte offset to the end of file
- Pass the new-content slice to the existing `parse_jsonl.go` parser
- Store the final byte offset in the checkpoint or state so the next turn can use it

## Why this matters

On long agent sessions (dozens of turns), each individual commit's hook invocation grows slower as the JSONL file grows. A session with 50 turns could make the 50th commit's hook 50× slower than the 1st. Incremental parsing keeps hook latency constant per turn regardless of session length.

## Context hints

- `internal/agent/claude/parse_jsonl.go`
- `internal/agent/claude/find_latest_session.go`
- `internal/hooks/` (pre-commit, post-commit implementations)
- `.partio/state/pre-commit.json` state format
