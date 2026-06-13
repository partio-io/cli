---
id: session-handoff-generation
target_repos:
  - cli
acceptance_criteria:
  - "partio handoff command generates a structured handoff document from the current branch's checkpoint data"
  - "Handoff document includes: summary of changes, key decisions made, current state, and suggested next steps"
  - "Output defaults to stdout in Markdown format"
  - "Optional --output flag writes to a file"
  - "Handoff extracts context from session transcripts stored in checkpoints"
  - "Works with partial data gracefully (missing transcripts produce a changes-only handoff)"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add partio handoff command for structured session handoff documents

## Summary

Add a `partio handoff` command that generates structured handoff documents from checkpoint session data. When a developer (human or AI agent) needs to hand off in-progress work to another person or agent, the handoff document provides the context needed to continue: what was done, why, what decisions were made, and what remains.

This leverages Partio's core value proposition — capturing the *why* behind code changes — and makes it actionable for work continuity.

## Context

- `internal/checkpoint/` — checkpoint domain type and storage
- `internal/session/` — session data and transcript access
- `cmd/partio/` — CLI command implementations

## Implementation notes

- Add `cmd/partio/handoff.go` with a new Cobra command
- Read checkpoints for the current branch from the orphan branch
- Extract session transcripts and parse key information: goals, decisions, blockers, file changes
- Generate a Markdown document with sections: Summary, Changes Made, Key Decisions, Current State, Next Steps
- Support `--output <file>` flag for file output, default to stdout
- Handle the case where transcripts are empty or missing gracefully
