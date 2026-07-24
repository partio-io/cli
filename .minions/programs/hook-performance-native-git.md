---
id: hook-performance-native-git
target_repos:
  - cli
acceptance_criteria:
  - "Hook hot paths (pre-commit, post-commit) use `git status --porcelain` subprocess instead of go-git's worktree.Status() for detecting changed files"
  - "Performance on large worktrees (many gitignored files) does not regress — hook latency stays under 1s on repos with 10k+ files"
  - "Existing tests still pass: `make test` green"
  - "`make lint` passes"
pr_labels:
  - minion
  - performance
---

# Use native git CLI for worktree status in hook hot paths

## Background

Inspired by entireio/cli#1642 and entireio/cli PR#1643.

On large repos with many gitignored files, go-git's `worktree.Status()` walks every file on disk — including ignored directories — before filtering. This produces 40–55 second latency per hook invocation on large monorepos (git itself takes 0.02s for the same operation).

## What to implement

Replace go-git `worktree.Status()` calls in Partio's hook hot paths (pre-commit, post-commit) with subprocess calls to the native `git status --porcelain` command.

- Identify all uses of go-git worktree status in `internal/hooks/` and related code
- Replace with a subprocess call: `git -C <repoRoot> status --porcelain`
- Parse the output lines to determine which files were modified/added/deleted
- Keep the go-git-based path for non-hook code where performance is less critical

## Why this matters

Partio hooks run synchronously during `git commit`, blocking the user's terminal. A 40s hang on every commit in a large repo makes Partio unusable for those projects. Using the native git binary avoids the full-tree stat walk that go-git performs and matches what git itself uses for incremental status.

## Context hints

- `internal/hooks/`
- `internal/git/`
- Any code calling `worktree.Status()` from `go-git`
