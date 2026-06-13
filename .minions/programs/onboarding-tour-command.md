---
id: onboarding-tour-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio tour` command exists and is discoverable via `partio --help`"
  - "Tour walks through core concepts: enable, hooks, checkpoints, sessions, rewind"
  - "Each tour step displays embedded markdown content without requiring network access"
  - "Tour can be exited at any point without side effects"
  - "Tour detects whether current directory has partio enabled and adjusts guidance accordingly"
pr_labels:
  - minion
---

# Add `partio tour` command for interactive onboarding walkthrough

Add a `partio tour` command that walks new users through Partio's core concepts using embedded markdown content displayed directly in the terminal.

## Motivation

New users need to understand several concepts to use Partio effectively: enabling repos, how git hooks capture sessions, what checkpoints contain, and how to browse history with rewind. Currently this requires reading external documentation. An in-CLI tour reduces friction for first-time users and can be suggested after `partio enable`.

Inspired by entireio/cli#1146 which adds an embedded tour command.

## Desired behavior

- `partio tour` launches a sequential walkthrough covering:
  1. What Partio does (captures the *why* behind code changes)
  2. Enabling a repository (`partio enable`)
  3. How hooks work (pre-commit detection, post-commit checkpoint creation)
  4. What checkpoints contain (session transcripts, attribution, metadata)
  5. Browsing history (`partio status`, `partio rewind`)
- Each step is a self-contained markdown block embedded in the Go binary (no network required)
- User advances with Enter, exits with q/Ctrl+C
- If the current directory is a git repo, the tour can note whether partio is already enabled
- The tour content lives in a dedicated file (e.g., `cmd/partio/tour_content.go`) for easy updates

## Context hints

- `cmd/partio/` — CLI command definitions
- `cmd/partio/status.go` — example of a command that reads repo state
