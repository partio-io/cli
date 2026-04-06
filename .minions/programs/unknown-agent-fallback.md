---
id: unknown-agent-fallback
target_repos:
  - cli
acceptance_criteria:
  - "When agent detection returns false (agent not running), checkpoints record agent as `Unknown` not the configured agent name"
  - "The configured agent name is used only as a detection hint, not as an attribution fallback"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Use "Unknown" as fallback agent type for unidentified sessions

When Partio's pre-commit hook runs and `Detector.IsRunning()` returns false (no agent process found), the current flow still stores the configured agent name (e.g., `claude-code`) in the session state. The post-commit hook then creates a checkpoint attributing the work to that agent, even though detection failed.

This is misleading: if the agent wasn't detected, the attribution is a guess. The agent field should be `"Unknown"` in this case rather than the configured default.

## Change

In the hook flow, when `IsRunning()` returns false:
- Record the agent as `"Unknown"` in the pre-commit state file (`.partio/state/pre-commit.json`)
- Post-commit reads this and writes it into the checkpoint metadata

The configured agent name (from `cfg.Agent`) should remain the detector selector — i.e., which agent to look for — but should not appear in checkpoint metadata when detection failed.

## Key files

- `internal/hooks/precommit.go` — where detection state is saved; change the agent field to `"Unknown"` when `IsRunning()` returns false
- `internal/hooks/postcommit.go` — reads pre-commit state and creates checkpoints
- `internal/session/state.go` — the state struct that carries agent information between hooks

**Inspired by:** entireio/cli#838 (fix: use Unknown instead of Claude Code for unidentified agent fallback)
