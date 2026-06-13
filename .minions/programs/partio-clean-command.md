---
id: partio-clean-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio clean` removes orphaned state files from `.partio/state/` (pre-commit state left by interrupted hooks)"
  - "`partio clean --all` removes all session state files, including idle sessions"
  - "Command is idempotent — safe to run when nothing needs cleaning, exits 0"
  - "Command prints a summary of what was removed (or 'nothing to clean' when clean)"
  - "Existing active session state is preserved by default (only orphaned/ended state is removed)"
pr_labels:
  - minion
---

# Add `partio clean` command to remove stale session state

## What

Add a `partio clean` command that removes orphaned and stale state files from `.partio/state/`. When git hooks are interrupted (e.g., a crash during commit, a failed amend, or aborted operations), pre-commit state files can be left behind. Over time these accumulate and can cause spurious checkpoint creation on the next commit.

Default behavior:
- Remove pre-commit state files that have no corresponding live commit operation in progress
- Remove session state for sessions that ended but whose state files were not cleaned up

With `--all` flag:
- Remove all session state files, regardless of session status

## Why

Partio saves detection state to `.partio/state/pre-commit.json` before each commit and deletes it in post-commit. If the commit is interrupted before post-commit runs, this file lingers. On the next commit, Partio may pick up the stale state and create an incorrect checkpoint linkage. A `partio clean` command lets users recover from this situation without having to run `partio disable && partio enable`.

## Source

Inspired by `entireio/cli` PR #846 which fixed `entire clean --all` to clean all sessions (not just orphaned ones) — changelog 0.5.3.

## Implementation hints

- `cmd/partio/clean.go` — new Cobra subcommand
- `internal/session/` — for listing and identifying stale session state
- State files are in `.partio/state/` relative to the git worktree root
- Use `internal/git` to get the worktree root path
- Check if `.partio/state/pre-commit.json` exists and whether a commit is in progress (presence of `.git/COMMIT_EDITMSG` being written) before removing
<!-- program: .minions/programs/partio-clean-command.md -->
