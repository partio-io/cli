---
id: user-prompt-submit-context-injection
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable --agent claude-code` installs a `user-prompt-submit` hook entry in `.claude/settings.json` that invokes `partio _hook claude-code user-prompt-submit`"
  - "The hook injects a brief Partio behavioral invariant block into the agent context on the first prompt of each new session in a Partio-enabled repo"
  - "Injection is skipped on subsequent prompts in the same session (tracked via session state in `.partio/state/`)"
  - "`partio disable` removes the Partio-managed `user-prompt-submit` hook entry without touching user-defined entries"
  - "`partio doctor` reports whether the `user-prompt-submit` hook is registered and points to a valid `partio` binary"
  - "The injected invariant block is minimal: states that commits auto-capture sessions and warns against manually invoking partio during agent runs"
  - "Unit tests in `internal/agent/claude/` cover hook entry generation and idempotency"
pr_labels:
  - minion
---

# Inject Partio usage invariants via Claude Code user-prompt-submit hook

## Problem

Agents running in Partio-enabled repos have no guaranteed awareness of Partio's session-capture semantics unless they explicitly run `partio agent-help` (proposal #501) or discover the CLAUDE.md injection from proposal #455. This leads to agents that:

- Are confused by the `Partio-Checkpoint` trailer appearing in commits they authored
- Attempt to manually trigger checkpointing or call non-existent commands
- Don't know that their session is already being captured and that no special commit ceremony is needed

The gap is that static context files (CLAUDE.md, #455) and on-demand help commands (#501) both require the agent to act first. There is no mechanism that delivers a minimal behavioral contract to the agent at the moment it receives its first user message.

## What to implement

Add a `user-prompt-submit` hook for Claude Code that fires on every prompt submission. On the first prompt of a new session in a Partio-enabled repo, the hook appends a short invariant block to the Claude Code injection context:

```
[Partio] This repository captures AI sessions automatically on each commit.
- You do not need to call any partio command during your session.
- Normal git commits trigger checkpoint creation; this happens transparently.
- The Partio-Checkpoint trailer in commit messages is written by the post-commit hook.
- Use `partio status` to confirm capture is active; `partio rewind --list` to browse history.
```

Implementation steps:

1. **Extend `cmd/partio/enable.go`**: when `--agent claude-code` is selected, write a `user-prompt-submit` hook entry to `.claude/settings.json` alongside existing lifecycle hooks (from proposal #456). Entry invokes `partio _hook claude-code user-prompt-submit`.

2. **Add hook handler in `internal/hooks/`**: implement `UserPromptSubmit` handler for the claude-code agent. Read session state from `.partio/state/` to check whether context has already been injected this session. If not, write the invariant block to the hook's injection file and mark the state as injected.

3. **Track injection state**: add a per-session flag in `.partio/state/session-context-injected.json` (keyed by session ID) so the injection fires exactly once per session, not on every prompt.

4. **Hook cleanup**: `partio disable` removes the `user-prompt-submit` entry from `.claude/settings.json`. `partio doctor` checks for its presence.

## Why

Reliable behavioral invariant delivery eliminates the ambiguity agents encounter when they see checkpoint trailers or Partio output. A first-prompt injection ensures the agent has the essential contract before it starts working — without requiring static file changes (which may be overwritten) or explicit help commands (which agents may not run).

Distinct from #455 (passive CLAUDE.md write) and #456 (session-start/stop hooks): this targets the user-prompt-submit lifecycle event specifically to inject context at the moment the agent is about to reason, and is limited to the first prompt to avoid noise.

## Source

Inspired by [entireio/cli PR #1821](https://github.com/entireio/cli/pull/1821) — "feat(agent-help): teach agents entire usage via examples + injected invariant" — which identified that first-turn pointer-only injections fail to convey behavioral invariants agents need even if they never run the help command.
