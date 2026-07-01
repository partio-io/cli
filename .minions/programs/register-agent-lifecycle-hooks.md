---
id: register-agent-lifecycle-hooks
target_repos:
  - cli
acceptance_criteria:
  - "partio enable --agent claude-code installs SessionStart and Stop hooks in .claude/settings.json that invoke partio _hook"
  - "Hook entries include a timeout (e.g., 10s) so the agent fails open if partio stalls"
  - "partio disable removes the agent-native hook entries it installed"
  - "Hook installation is idempotent — does not duplicate entries on repeated enable"
  - "Existing user-defined hooks in the agent config are preserved"
  - "partio doctor checks whether agent-native hooks are correctly registered"
pr_labels:
  - minion
---

# Register agent-native lifecycle hooks during enable

When `partio enable` runs, install hooks into the agent's own hook system (e.g., Claude Code's `.claude/settings.json` hooks) in addition to git hooks. This gives Partio precise session lifecycle boundaries (start/stop) rather than relying solely on process detection at commit time.

## Why

Partio currently detects agent sessions via process detection during git hooks. This approach has blind spots:
- It only fires at commit time, missing session start/end events
- Process detection can be fragile (race conditions, stale PIDs)
- No timeout protection — if detection stalls, the git hook blocks

Agent-native hooks (like Claude Code's `SessionStart`/`Stop`) provide explicit lifecycle signals with built-in timeout support, enabling more reliable session tracking and richer metadata capture.

## What to implement

- Extend the `enable` command to register hooks in the agent's native hook config file
- For Claude Code: add `SessionStart` and `Stop` hook entries in `.claude/settings.json` that invoke `partio _hook claude-code session-start` and `partio _hook claude-code session-stop`
- Include a configurable timeout (default 10s) in the hook entry so the agent fails open
- Extend `disable` to remove only the Partio-managed hook entries
- Extend `doctor` to verify agent-native hooks are registered and point to a valid `partio` binary

## Source

Inspired by entireio/cli PR#1237 (claude-code session-start hook timeout), issue #1065 (CLI hangs blocking agent hooks), and issues #957/#1072 (repeated hook timeouts).
