---
id: enable-non-git-folders
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable` in a non-git directory offers to initialize a git repo first"
  - "If the user declines git init, the command exits with a clear message explaining git is required"
  - "After git init, the normal enable flow continues (hook installation, config creation)"
  - "Non-interactive mode (e.g., CI) skips the prompt and fails with a clear error"
pr_labels:
  - minion
  - enhancement
---

# Improve `partio enable` flow for non-git directories

## Description

When a user runs `partio enable` in a directory that is not a git repository, the command currently fails with an error. Improve the flow to detect this case and offer to run `git init` first, then continue with the normal enable process.

This reduces friction for users who are starting new projects — they can run `partio enable` before or after `git init` and get the same result.

## Implementation Notes

- In the enable command, check if the current directory is a git repo before proceeding
- If not a git repo and running interactively, prompt: "This directory is not a git repository. Initialize one? [Y/n]"
- If yes, run `git init` and continue with hook installation
- If no or non-interactive, exit with a clear error message
- Consider also handling the case where the user is in a subdirectory of a git repo (already works, but worth verifying)

## Context Hints

- `cmd/partio/enable.go` — enable command implementation
- `internal/git/` — git repository detection
