---
id: cross-repo-session-browsing
target_repos:
  - cli
acceptance_criteria:
  - "`partio sessions list --all-repos` discovers and lists sessions from all git repos under a configurable search root"
  - "Output includes repo path, session ID, agent name, status, and timestamp for each session"
  - "Search root defaults to the current directory and can be overridden with `--root <path>` or a global config setting"
  - "`partio sessions info <session-id> --repo <path>` shows details for a session in a specific repo"
  - "Repos without partio enabled are silently skipped"
  - "Performance is acceptable for directories containing up to 50 git repos"
pr_labels:
  - minion
---

# Cross-repo session browsing

## Summary

Add a `--all-repos` flag to session-related commands that discovers and aggregates checkpoint/session data across multiple git repositories on the local machine. Users working across many repos currently have no way to get a unified view of their AI agent sessions without navigating to each repo individually.

## Motivation

Users who work across multiple repositories want to compare AI agent sessions, review outcomes across projects, and find specific sessions without remembering which repo they were in. This is especially valuable for running evals with different prompts and comparing how sessions played out (as reported by community feedback on entireio/cli#985).

## Design Notes

- Walk directories from a search root, identify git repos (look for `.git` or `.partio/`), and collect session data from each
- Reuse existing session discovery logic per-repo
- Consider a global config setting (`session_search_roots: ["/home/user/projects"]`) to avoid slow recursive walks
- Output should be tabular with repo path as a column, sortable by timestamp
- This builds on top of the `partio sessions` subcommands (list, info, stop) if/when they are implemented

## Context Hints

- `internal/session/` — session discovery and state
- `internal/checkpoint/` — checkpoint storage and retrieval
- `cmd/partio/` — CLI command definitions
