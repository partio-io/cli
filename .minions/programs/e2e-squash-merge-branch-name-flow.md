---
id: e2e-squash-merge-branch-name-flow
target_repos:
  - cli
acceptance_criteria:
  - A test repo is initialised with partio hooks enabled
  - One or more commits are made on a feature branch while a simulated agent session is active, producing checkpoints with Branch set to the feature branch name
  - The feature branch is squash-merged into main (making the original feature-branch commits unreachable via git cat-file)
  - `partio rewind --list --branch <feature-branch>` lists the checkpoints for that branch and no others
  - `partio resume --branch-name <feature-branch>` selects the most recent checkpoint (by CreatedAt) and exits without error
  - `partio rewind --to <id>` for a checkpoint whose commit is unreachable prints a "Warning:" line and exits 0, without creating a rewind branch
  - All assertions are made against the real compiled CLI binary, not internal Go functions
pr_labels:
  - minion
  - test
---

# e2e: squash-merge branch-name retrieval flow

## Context

PR partio-io/cli#620 added `--branch-name` to `partio resume` and `--branch` to `partio rewind --list` so users can find sessions from a feature branch that was squash-merged into main.

The PR's unit tests call internal Go functions (`runRewindList`, `runRewindTo`, `FindByBranch`, `pickMostRecentCheckpoint`) with manually-written checkpoints. They do not drive the complete chain:

1. `git commit` on a named branch → pre-commit hook captures `git.CurrentBranch()` into state → post-commit hook writes checkpoint with `Branch` field set
2. Squash merge → original commits become unreachable
3. `partio rewind --list --branch <name>` → CLI finds and lists those checkpoints
4. `partio resume --branch-name <name>` → CLI finds the most-recent checkpoint and opens it
5. `partio rewind --to <id>` with an unreachable commit → prints warning, exits 0, no branch created

Each component passes independently; the integration may still break (e.g. if `git.CurrentBranch()` returns an unexpected value in a detached HEAD after amend, or if the orphan-branch layout differs from what `FindByBranch` expects).

## What to build

Write a Go integration test in `cmd/partio/` (package `main_test` or a dedicated `e2e/` directory matching existing conventions) that:

1. Compiles the `partio` binary with `go build` into a temp directory.
2. Initialises a git repo in a `t.TempDir()`, installs partio hooks via `partio enable`, and configures `PARTIO_ENABLED=true`.
3. Simulates an agent being active (set `PARTIO_AGENT` to a stub or use the existing test-double pattern; alternatively write a pre-commit state file directly as the hook would, with `agent_active: true` and `branch: feature/squashed`).
4. Makes one commit on `feature/squashed` so the post-commit hook fires and writes a checkpoint.
5. Makes a second commit on the same branch to produce a second checkpoint (validates most-recent selection).
6. Checks out `main` and performs a squash merge: `git merge --squash feature/squashed && git commit -m "squash"`. This leaves the original two commits unreachable.
7. Runs `partio rewind --list --branch feature/squashed` and asserts:
   - exit code 0
   - both checkpoint IDs appear in stdout
   - no other checkpoint IDs appear
8. Runs `partio resume --branch-name feature/squashed --print` (use `--print` or `--copy` if available, or capture stdout) and asserts exit code 0 and that output references the most-recent checkpoint.
9. Takes the checkpoint ID for the second (most-recent) commit and runs `partio rewind --to <id>`, asserting:
   - exit code 0
   - stdout contains "Warning:"
   - no `partio/rewind/<id>` branch is created

## File location

`cmd/partio/squash_merge_e2e_test.go` (or `e2e/squash_merge_test.go` if an `e2e/` package exists).

## Key source locations

- `internal/hooks/precommit.go` — captures `git.CurrentBranch()` into the pre-commit state `Branch` field
- `internal/hooks/postcommit.go` — reads `state.Branch` and writes it into `cp.Branch`
- `internal/checkpoint/find_by_branch.go` — scans the orphan branch for matching checkpoints
- `internal/git/commit_reachable.go` — `git cat-file -e` reachability check
- `cmd/partio/resume.go:49` — `runResumeByBranch` orchestrates FindByBranch → pick → runResume
- `cmd/partio/rewind.go:99` — `runRewindListByBranch` orchestrates FindByBranch → print
