---
id: hook-invocation-isolation
target_repos:
  - cli
acceptance_criteria:
  - Pre-commit and post-commit hooks identify the active agent before executing agent-specific logic
  - Hooks configured for Agent A do not fire when Agent B is the active agent
  - When multiple agents are configured, only the hooks for the detected active agent execute
  - Hook state files include the detected agent identity to prevent cross-agent state leakage
pr_labels:
  - minion
---

# Isolate hook invocations per detected agent

When multiple AI coding agents are installed (e.g., Claude Code and Cursor), Partio's git hooks can be incorrectly triggered by the wrong agent. For example, if Cursor does not have its own hook configuration, it may fall through to Claude Code's hooks, causing Partio to misattribute the session or create checkpoints with incorrect agent metadata.

## What to implement

Update the hook implementations (`internal/hooks/`) to add agent-scoped hook invocation:

1. At the start of pre-commit hook execution, detect which agent is currently active using the existing detector interface.
2. Compare the detected agent against the agent(s) configured in Partio settings.
3. Only proceed with checkpoint state capture if the detected agent matches a configured agent.
4. Write the detected agent identity into the pre-commit state file (`.partio/state/pre-commit.json`) so post-commit can verify consistency.
5. Log a debug message when a hook invocation is skipped due to agent mismatch.

## Why this matters

As Partio adds support for more agents (Cursor, Copilot, Codex, etc.), the risk of cross-agent hook triggering increases. Without isolation, users get incorrect session metadata, wrong agent attribution, and potentially corrupted checkpoint data. This is especially problematic in environments where developers switch between agents frequently.
