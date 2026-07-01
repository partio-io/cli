---
id: agent-detection-disambiguation
target_repos:
  - cli
acceptance_criteria:
  - "Agent detection correctly distinguishes between agents with similar process characteristics"
  - "Detection uses multiple signals (process name, parent process, environment variables) for disambiguation"
  - "Misidentification of one agent as another is prevented"
  - "Tests cover disambiguation scenarios"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Improve agent detection disambiguation

Improve the agent detection logic to correctly distinguish between agents that share similar process characteristics. For example, agents launched from IDE extensions (e.g., Cursor) may have process trees that resemble Claude Code's standalone process, leading to misidentification.

The detector should use multiple signals for disambiguation:
- Process name and command-line arguments
- Parent process chain
- Environment variables set by specific agents
- Session directory structure differences

This prevents incorrect attribution when multiple agents could match the same process heuristics.

## Context hints

- `internal/agent/detector.go` — detector interface
- `internal/agent/claude/` — Claude-specific detection logic
- `internal/hooks/` — where detection is invoked

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
