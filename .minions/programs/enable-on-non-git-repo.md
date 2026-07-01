---
id: enable-on-non-git-repo
target_repos:
  - cli
acceptance_criteria:
  - "When `partio enable` is run outside a git repo, it prints: 'This directory is not a git repository. Initialize one now? [y/N]'"
  - "If the user answers y, `git init` is run and enable proceeds normally"
  - "If the user answers N (or hits enter), enable exits with message: 'partio enable requires a git repository. Run git init first.'"
  - "A test covers the non-repo detection branch"
pr_labels:
  - minion
---

# Improve `partio enable` flow when run outside a git repository

When `partio enable` is invoked in a directory that is not yet a git repository, instead of returning a bare error, offer to initialize the git repo first (`git init`) and then proceed with enabling Partio. Present a y/N prompt before running `git init`. If the user declines, exit with a clear message explaining the requirement.

## Why

New project setup often follows the pattern of creating a directory, installing tools, then initializing git. The current hard error stops the flow cold and provides no path forward.

## User Relevance

Developers setting up a new project can run `partio enable` as part of their initialization checklist without needing to remember to `git init` first.

## Context Hints

- `cmd/partio/enable.go`
- `internal/git/repo.go`

## Source

Inspired by entireio/cli#978
