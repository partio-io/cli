---
id: partio-why-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio why <file>` parses `git blame --porcelain` output and enriches each line with checkpoint metadata"
  - "Lines authored during an agent session display the agent name and session prompt"
  - "Output renders in a code-centric overview with syntax highlighting and aligned agent labels"
  - "When a checkpoint ID is found for a blame hunk, the corresponding session context (prompt, model) is shown"
  - "Gracefully falls back to plain blame output when no checkpoint data exists for a hunk"
  - "Unit tests cover porcelain parsing, checkpoint enrichment, and missing-checkpoint fallback"
pr_labels:
  - minion
---

# Add `partio why` command for checkpoint-enriched blame

## Description

Add a `partio why <file>` command that enriches `git blame` output with checkpoint context from the orphan branch. For each blame hunk, look up the commit's `Partio-Checkpoint` trailer, fetch the corresponding checkpoint metadata (agent, model, prompt, attribution), and render a code-centric overview that shows *why* each line was written — not just *who* and *when*.

The command should:
- Parse `git blame --porcelain` for the specified file
- For each commit in the blame output, check for a `Partio-Checkpoint` trailer
- When found, read the checkpoint metadata from `partio/checkpoints/v1` to extract session context
- Render output with syntax highlighting, agent labels, and session prompts aligned alongside code lines
- Support both static output (for piping) and a TUI overview mode

## Why

Partio's core mission is capturing the *why* behind code changes. Today, that context is stored in checkpoints but there's no way to see it *in the context of the code itself*. `partio why` bridges this gap by answering "why was this line written?" directly from the terminal, making checkpoint data actionable during code review, debugging, and onboarding.

## Source

Inspired by entireio/cli PR #1074 (`entire why` command).
