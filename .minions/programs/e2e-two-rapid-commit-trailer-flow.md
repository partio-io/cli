---
id: e2e-two-rapid-commit-trailer-flow
target_repos:
  - cli
acceptance_criteria:
  - "A test initialises a real git repo in a temp directory and installs partio pre-commit and post-commit hooks"
  - "It simulates an active agent session (e.g. via env vars and the existing fake-detector pattern)"
  - "It stages and commits two files in rapid succession with no artificial sleep between them"
  - "Both resulting commits carry a Partio-Checkpoint: trailer line in their git log output"
  - "The test exercises the real runPreCommit / runPostCommit code paths, not mocks of readStateWithRetry"
  - "Total test runtime does not exceed 5 seconds"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Build an end-to-end test: two rapid commits both receive Partio-Checkpoint trailers

## Background

PR partio-io/cli#612 fixes a race where a second rapid commit could lose its
`Partio-Checkpoint` trailer. The fix adds `readStateWithRetry` in
`internal/hooks/postcommit.go` (5 attempts, 100 ms apart) so post-commit
tolerates a brief lag before the state file written by pre-commit is visible.

The PR's unit tests (`internal/hooks/state_retry_test.go`) verify the retry
helper in isolation — they prove the function picks up a late file, bounds
retries, and fast-fails on unexpected errors. No test drives the full hook
chain for two commits in a row. Each half can pass unit tests while the whole
flow fails (e.g. wrong path resolution inside the real runner, or the timing
constants being insufficient for real git commit latency).

## What to build

Add a test file `internal/hooks/e2e_two_commit_test.go` (package `hooks`)
that drives the full flow through `runPreCommit` and `runPostCommit` directly
(not via shell-installed hooks, but the Go functions) in a real temp git repo.

### Flow to exercise

1. Create a temporary directory, initialise a git repo in it (`git init`,
   set `user.email` and `user.name`), and write an initial commit so HEAD exists.
2. Create the `.partio/state/` directory structure inside the repo.
3. Call `runPreCommit(repoRoot, cfg)` with `PARTIO_ENABLED=true` and a config
   that has `AgentActive = true` (simulate an active agent session by writing
   the pre-commit state file directly, the same way `runPreCommit` does, since
   actual process detection won't find a real Claude Code process in CI).
   Alternatively, write the state JSON file manually before each call to
   `runPostCommit` to simulate what pre-commit would write.
4. Stage a file and create a real commit via `git commit` (using `exec.Command`
   so git fires post-commit hooks — OR call `runPostCommit` directly after
   staging/committing with `--no-verify` to skip hooks, then supply the state
   file as pre-commit would have).
5. Repeat for a second file committed immediately after the first, with no
   sleep between them.
6. After both commits, read `git log --format=%B` for each commit and assert
   that both contain a `Partio-Checkpoint:` trailer line.

### Preferred approach (direct function calls, no shell hooks)

The simplest approach that still exercises the real code path:

```go
func TestTwoRapidCommitsBothGetTrailers(t *testing.T) {
    // 1. Init real git repo
    dir := t.TempDir()
    runGit(t, dir, "init")
    runGit(t, dir, "config", "user.email", "test@example.com")
    runGit(t, dir, "config", "user.name", "Test")

    // 2. Build config pointing at the temp repo
    cfg := config.Defaults()
    cfg.Agent = "claude-code"

    stateDir := filepath.Join(dir, ".partio", "state")
    os.MkdirAll(stateDir, 0o755)
    stateFile := filepath.Join(stateDir, "pre-commit.json")

    // Helper: write state as pre-commit would, commit a file, call runPostCommit
    commitAndCapture := func(filename, content string) {
        // Write pre-commit state (simulate what runPreCommit writes)
        stateJSON := `{"agent_active":true,"branch":"main","agent_name":"claude-code"}`
        os.WriteFile(stateFile, []byte(stateJSON), 0o644)

        // Stage and commit (--no-verify so no real hooks fire)
        f := filepath.Join(dir, filename)
        os.WriteFile(f, []byte(content), 0o644)
        runGit(t, dir, "add", filename)
        runGit(t, dir, "commit", "--no-verify", "-m", "add "+filename)

        // Run post-commit directly
        if err := runPostCommit(dir, cfg); err != nil {
            t.Fatalf("runPostCommit: %v", err)
        }
    }

    commitAndCapture("a.txt", "hello")
    commitAndCapture("b.txt", "world")

    // Assert both commits have Partio-Checkpoint trailer
    log := runGitOutput(t, dir, "log", "--format=%B", "-2")
    count := strings.Count(log, "Partio-Checkpoint:")
    if count < 2 {
        t.Fatalf("expected 2 Partio-Checkpoint trailers, got %d\nlog:\n%s", count, log)
    }
}
```

### What to assert

- `strings.Count(gitLog, "Partio-Checkpoint:")` >= 2
- Optionally: `strings.Count(gitLog, "Partio-Attribution:")` >= 2

### Where the test lives

`internal/hooks/e2e_two_commit_test.go`, package `hooks` (same package as
`postcommit.go` so it can call `runPostCommit` directly).

### Notes

- `runGit` and `runGitOutput` are small helpers that call `exec.Command("git",
  args...).CombinedOutput()` in the given dir and call `t.Fatal` on error.
- The checkpoint store writes to the orphan branch inside the temp repo — that
  is fine; just ensure the temp dir is a real git repo with at least one commit
  before the test calls `runPostCommit`.
- If the checkpoint store fails because the orphan branch machinery can't find
  a remote, that is acceptable to ignore in the test (the trailer is written
  before the store write). Check whether the assertion holds even if
  `store.Write` returns an error — the trailer amend happens before the store
  write in `postcommit.go`.
- Do not add sleeps. The retry logic in `readStateWithRetry` is what this test
  validates implicitly; the state file is written immediately before each call,
  so no retry should be needed — but the code path is the same as production.
