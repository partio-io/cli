---
id: cross-agent-session-synthesis
target_repos:
  - cli
acceptance_criteria:
  - When multiple agents contribute to the same commit, all agent sessions are captured in the checkpoint
  - `partio status` shows all active agents and their session attribution percentages
  - Checkpoint metadata includes a per-agent breakdown when multiple agents are detected
  - Attribution calculation correctly handles overlapping file changes from different agents
  - The detector interface supports enumerating all running agents, not just the first match
pr_labels:
  - minion
---

# Add cross-agent session synthesis for multi-agent attribution

## Problem

Partio's current detector interface (`agent/detector.go`) finds and returns a single active agent. In practice, developers increasingly use multiple AI agents in the same repository simultaneously (e.g., Claude Code for architecture, Codex for tests, Cursor for quick edits). When multiple agents contribute to the same commit, Partio only captures one agent's session, losing the others' context and producing inaccurate attribution.

## Solution

Extend the detector interface and checkpoint metadata to support multiple concurrent agent sessions per commit.

### Key behaviors

- `DetectAll()` method on the detector interface returns all active agent sessions, not just the first match
- Pre-commit hook saves state for all detected agents
- Post-commit creates a checkpoint that includes session data from all contributing agents
- Checkpoint metadata includes a `contributors` array with per-agent session references and attribution
- Attribution calculation splits credit across agents based on which files each agent's session touched
- `partio status` shows all active agents with their current session info

### Implementation hints

- Add `DetectAll() ([]Detection, error)` to the `Detector` interface in `internal/agent/detector.go`
- Pre-commit state in `.partio/state/pre-commit.json` becomes an array of detections
- Post-commit iterates all detected sessions and merges their data into the checkpoint
- Attribution in `internal/attribution/` needs a multi-agent mode that maps staged files to the agent whose session touched them
- Falls back gracefully to single-agent behavior when only one agent is detected

## Context

Inspired by entireio/cli's cross-agent synthesis in `entire labs review` (changelog 0.6.1) and multi-agent detection patterns. Adapted for Partio's pluggable detector architecture.
