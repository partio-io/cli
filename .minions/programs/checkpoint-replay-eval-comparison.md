---
id: checkpoint-replay-eval-comparison
target_repos:
  - cli
acceptance_criteria:
  - partio replay eval runs a checkpoint prompt against multiple agents in isolated worktrees
  - Each agent run produces a structured report with pass/fail results, diff stats, and execution metadata
  - Results are ranked by a configurable scoring rubric (e.g., test pass rate, diff size, execution time)
  - Reports are saved locally for later comparison
  - JSON output mode is supported for CI integration
pr_labels:
  - minion
  - minion-proposal
---

# Checkpoint replay eval comparison

## Problem

Partio captures the full context of AI agent sessions (prompts, transcripts, tool calls) in checkpoints. Users who want to evaluate whether a different agent or model would produce better results for their coding tasks have no structured way to replay a checkpoint's prompt against multiple agents and compare outcomes.

## Proposed solution

Extend the replay concept (see existing proposal #410) with a multi-agent evaluation mode:

1. **`partio replay eval <checkpoint-id> --agents claude-code,<other>`** - Replays the original prompt from a checkpoint against each specified agent in isolated git worktrees (one per agent). Each worktree starts from the pre-checkpoint commit state.

2. **Structured reports**: Each replay produces a report containing:
   - Agent name and model used
   - Whether the resulting code compiles/passes tests (if a test command is configured)
   - Diff stats (lines added/removed/modified)
   - Token usage and wall-clock time
   - Pass/fail on the checkpoint's original acceptance criteria (if available)

3. **Ranking and comparison**: After all agents complete, produce a summary table ranking results by a configurable scoring rubric. Default rubric weights test pass rate highest, then diff minimality, then token efficiency.

4. **Report persistence**: Save reports to `.partio/eval/<checkpoint-id>/` for later review and trend analysis across evaluations.

## Why this matters

As more AI coding agents become available, developers need data-driven ways to evaluate which agent works best for their specific codebase and task types. Replay eval turns Partio's captured checkpoints into a reusable benchmark suite.

## Context hints

- `internal/checkpoint/` - Checkpoint domain types and storage
- `internal/agent/` - Agent detection interface (detector.go)
- `internal/session/` - Session data with prompt/transcript fields
- `cmd/partio/` - CLI command definitions
