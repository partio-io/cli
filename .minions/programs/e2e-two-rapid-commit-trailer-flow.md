---
id: e2e-two-rapid-commit-trailer-flow
target_repos:
  - partio-io/cli
acceptance_criteria:
  - A test in internal/hooks/ (or a dedicated e2e_test.go) initialises a bare git repo, installs partio hooks via the hook-install path, makes two commits in rapid succession during a simulated active-agent session, and asserts that BOTH resulting commits carry a Partio-Checkpoint trailer in their commit message.
  - The test must drive the real pre-commit and post-commit hook logic (not mocks of readStateWithRetry) so that the full state-write → retry-read → checkpoint → amend cycle is exercised for each commit.
  - The test must not rely on wall-clock sleeps longer than 2 seconds total, and must be deterministic (no flaky timing windows).
pr_labels:
  - minion-proposal
---

## Context

PR partio-io/cli#612 adds `readStateWithRetry` to `internal/hooks/postcommit.go` to tolerate a race where the second rapid commit's post-commit hook fires before pre-commit's state write is visible on the filesystem.

The PR ships three unit tests for the retry helper (`internal/hooks/state_retry_test.go`) that verify:
- The helper picks up a late-appearing file.
- It gives up after a bounded number of attempts.
- It fast-fails on non-`ErrNotExist` errors.

None of these tests drive the full hook chain. Each half of the mechanism (pre-commit state write, post-commit state read + trailer amend) is verified in isolation. It is possible for both halves to pass their unit tests while the whole two-commit flow still fails — for example if the state file path is resolved differently inside the hook runner, or if the timing constants (5 attempts × 100 ms) are insufficient for the real git commit sequence.

## What to build

Write an end-to-end test (suggested location: `internal/hooks/e2e_rapid_commit_test.go`, build tag `//go:build integration` or plain package test using `t.TempDir`) that:

1. Creates a real git repository (`git init`) in a temp directory.
2. Installs the partio pre-commit and post-commit hooks by invoking the hook-install logic (or by writing the hook scripts that call `partio _hook pre-commit` / `partio _hook post-commit`).
3. Stubs or simulates an active agent session (e.g. set `PARTIO_ENABLED=true` and arrange for the pre-commit detector to return `agent_active: true` — look at how `precommit_test.go` does this with `PARTIO_AGENT` env var and a fake detector).
4. Makes commit #1: stage a file, run `git commit -m "first"`.
5. Makes commit #2 immediately after: stage another file, run `git commit -m "second"`.
6. Reads the git log for both commits and asserts each has a `Partio-Checkpoint:` trailer line.

The test should exercise the real `runPreCommit` and `runPostCommit` functions (via the hook runner or direct call) rather than substituting `readStateWithRetry` with a mock.

## Key files

- `internal/hooks/postcommit.go` — `readStateWithRetry` call at line ~33
- `internal/hooks/precommit.go` — state file write
- `internal/hooks/state_retry.go` — retry helper
- `internal/hooks/precommit_test.go` — existing test patterns to follow
- `internal/hooks/runner.go` — `Runner` struct wiring
