---
id: copilot-agent-integration
target_repos:
  - cli
acceptance_criteria:
  - "Copilot CLI process detection works via the detector interface"
  - "Copilot session JSONL files are discovered and parsed into checkpoint data"
  - "Pre-commit hook correctly identifies active Copilot sessions"
  - "Checkpoint metadata includes correct agent name and model for Copilot sessions"
  - "make test passes with new Copilot agent tests"
  - "make lint passes"
pr_labels:
  - minion
---

# Add GitHub Copilot CLI agent integration

Implement a Copilot CLI agent detector following the existing detector interface pattern used by the Claude Code implementation.

## What to implement

Add a `copilot` package under `internal/agent/` that implements the `Detector` interface:

1. **Process detection** — detect when GitHub Copilot CLI (or VS Code Copilot agent) is actively running during a commit
2. **Session discovery** — locate Copilot session data files on disk, following the walk-up-from-repo-root pattern used by Claude Code session discovery
3. **JSONL/transcript parsing** — parse Copilot session logs into the checkpoint data format
4. **Hook payload support** — handle Copilot's hook payload format (VS Code-compatible payloads) for lifecycle events

## Why this matters

GitHub Copilot is one of the most widely-used AI coding agents. Supporting it expands Partio's value to a much larger user base. The entireio/cli project added Copilot support across 0.5.1–0.5.4, validating market demand.

## Context hints

- `internal/agent/detector.go` — the detector interface to implement
- `internal/agent/claude/` — reference implementation to follow
- `cmd/partio/` — CLI commands that may need agent selection updates
