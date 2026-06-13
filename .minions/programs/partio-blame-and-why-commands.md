---
id: partio-blame-and-why-commands
target_repos:
  - cli
acceptance_criteria:
  - "`partio blame <file>` resolves lines through `git blame` and labels each as AI-authored or human-authored"
  - "`partio blame` enriches blamed commits with checkpoint/session metadata when Partio-Checkpoint trailer is present"
  - "`partio why <file:line>` shows the prompt, session context, and reasoning that produced a specific line"
  - "Both commands support `--json` output for machine consumption"
  - "Lines from uncommitted changes are labeled as unknown attribution"
  - "Commands work correctly in git worktree setups"
pr_labels:
  - minion
  - minion-proposal
---

# Add `partio blame` and `partio why` commands for AI-aware line attribution

## Description

Add first-class `partio blame` and `partio why` commands that bridge git's native `git blame` with Partio's checkpoint metadata to provide AI-aware line attribution.

`partio blame <file> [--line N|N-M] [--json]` resolves current file lines through `git blame --line-porcelain`, then enriches each blamed commit with checkpoint and session metadata from the `partio/checkpoints/v1` branch. Lines are labeled as `[AI]`, `[HU]` (human), or `[??]` (uncommitted).

`partio why <file[:line]> [--json]` goes deeper: for AI-attributed lines, it extracts and displays the prompt, session context, and relevant transcript excerpt that led to that code being written. This answers the question "why was this line written?" directly from the checkpoint data.

## Why

Partio already captures attribution data (binary 0%/100%) and stores rich session context in checkpoints, but there's no ergonomic way to query this at the line level. Developers reviewing AI-generated code currently have to manually cross-reference `git blame` output with checkpoint data. These commands close that gap and make Partio's stored context immediately actionable during code review.

## Context hints

- `internal/attribution/` — existing attribution calculation logic
- `internal/checkpoint/` — checkpoint storage and retrieval
- `cmd/partio/` — CLI command definitions
