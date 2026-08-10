---
id: nested-agent-session-isolation
target_repos:
  - cli
acceptance_criteria:
  - Process scanner skips `claude` processes that are subagent children (e.g. carry a `--no-session` flag, or whose parent PID is itself a `claude` process)
  - Top-level interactive Claude Code session is still correctly detected and captured
  - Detection logic is tested with a table-driven test that covers: main process only, main + nested child, and nested child only
pr_labels:
  - minion
  - enhancement
---

# Ignore nested Claude Code subagent processes during session detection

## Problem

Claude Code spawns child `claude` processes when the user's session invokes the Task tool or similar agent-mode features. These child processes run non-interactively (typically with `--no-session` or equivalent flags). Partio's process scanner in `internal/agent/claude/` currently detects any running `claude` process. If a subagent child process is running at commit time, Partio may associate the checkpoint with the wrong session, record a corrupted session state, or report false positives for "agent is active".

The manifestation in entireio/cli#1936 showed that nested Pi agent processes loaded the Entire extension (because the extension was project-local) and forwarded lifecycle events as if they were the user's session — overwriting the parent's last prompt and opening spurious turn IDs. Partio does not use lifecycle events, but its PID-based process scan faces the same ambiguity: a nested `claude` subagent process is a valid `claude` binary running in the same working directory.

## What to implement

In `internal/agent/claude/` (likely the process-detection code that enumerates running `claude` processes), add a filter that excludes nested subagent processes. Two detection heuristics, either sufficient:

1. **Command-line flag check**: if a `claude` process was started with `--no-session` (or any other flag that Claude Code uses to mark non-interactive subagents), skip it.
2. **Parent PID check**: if a `claude` process's parent PID is itself a `claude` process, it is a subagent child — skip it.

Apply whichever heuristic is cheapest given the information already available from `/proc/<pid>/cmdline` and `/proc/<pid>/status` (Linux) or the equivalent on other platforms.

## Why

A user running a long Claude Code session with the Task tool active will have one or more nested `claude` subagent processes running at commit time. Without this filter, Partio's session detection is non-deterministic — the "detected session" may flip between the parent and a child on each commit depending on process enumeration order, producing checkpoints that mix sessions or silently capture the wrong JSONL file.

## Source

entireio/cli#1936 — fix(pi): stop nested subagent processes from claiming the parent session

## Context hints

- `cli/internal/agent/claude/` — process detection and session discovery
- `cli/internal/agent/` — `Detector` interface and detection contract

<!-- program: .minions/programs/nested-agent-session-isolation.md -->
