---
id: agent-neutral-cli-messaging
target_repos:
  - cli
acceptance_criteria:
  - "All user-facing CLI output uses generic agent terminology instead of hardcoded 'Claude Code'"
  - "Agent-specific names only appear when displaying detected agent info (e.g., status output)"
  - "Error messages, help text, and empty-state messages are agent-neutral"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Use agent-neutral wording in all CLI output

Audit and update all user-facing CLI messages (error messages, help text, status output, empty-state messages) to use generic agent terminology instead of hardcoding "Claude Code".

Agent-specific names should only appear when reporting detected agent information (e.g., `partio status` showing which agent is active). All other messaging should use terms like "agent", "AI agent", or "coding agent" instead of "Claude Code".

## Context hints

- `cmd/partio/` — all CLI command implementations
- `internal/hooks/` — hook implementations that may produce user-facing output
- `internal/agent/` — agent detection that may have hardcoded references

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
