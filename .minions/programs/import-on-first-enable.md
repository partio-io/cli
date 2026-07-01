---
id: import-on-first-enable
target_repos:
  - cli
acceptance_criteria:
  - During first-time partio enable (no prior .partio/ directory), the CLI checks for pre-existing Claude Code sessions in ~/.claude/projects/
  - If discoverable sessions exist, the user is prompted to import them (interactive mode) or they are auto-imported (non-interactive/--yes mode)
  - The import uses the same session discovery and checkpoint creation logic as the existing checkpoint flow
  - Re-running partio enable after initial setup does not re-offer import
  - Import failures are logged as warnings but never fail the enable command
  - Sessions are filtered to those relevant to the current repository
pr_labels:
  - minion
---

# Offer session import during first-time enable

## What

When a user runs `partio enable` for the first time in a repository, proactively detect pre-existing Claude Code sessions (from `~/.claude/projects/`) and offer to import them as checkpoints. This brings historical AI session context into Partio immediately, rather than requiring users to discover a separate import command.

## Why

Users who install Partio in a repository where they've already been using Claude Code have valuable session history that Partio can't see. Today, if a `partio import` command existed (see proposal #427), users would need to know about it and run it separately. Most users won't — they'll enable Partio and wonder why their history starts empty.

By offering import during first-time setup, Partio captures the maximum amount of context from day one, making features like `partio rewind` and checkpoint browsing immediately useful instead of starting from a blank slate.

## How

In the `enable` command (`cmd/partio/enable.go`), after creating `.partio/` and installing hooks:

1. Check if this is a first-time enable (no pre-existing `.partio/` directory — capture this flag before creating it)
2. Use the existing session discovery logic (`internal/agent/claude/find_session_dir.go`) to find Claude Code sessions for the current repository
3. Filter to sessions with at least one entry (skip empty/corrupt sessions)
4. If sessions are found:
   - Interactive mode: prompt the user ("Found N existing Claude Code sessions. Import them? [y/N]")
   - Non-interactive mode (`--yes` or no TTY): auto-import all
5. For each session, create a checkpoint using the standard checkpoint creation flow
6. Log warnings for any individual import failures but continue with remaining sessions

The first-run gate prevents re-offering on subsequent `partio enable` calls (e.g., after disable/re-enable).

## Source

Inspired by entireio/cli#1595 and changelog 0.7.8 — Entire added first-run import offer during `entire enable`, gated on first-time setup, with interactive multi-select and non-interactive auto-import.
