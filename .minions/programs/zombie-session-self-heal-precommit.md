---
id: zombie-session-self-heal-precommit
target_repos:
  - cli
acceptance_criteria:
  - The pre-commit hook calls session.Manager.CleanupStale before calling session.Manager.RecordActive, so a stale session from a crashed agent is resolved before a new one is recorded.
  - When a stale session is swept in pre-commit, a Warn-level log line identifies the session ID and agent.
  - The cleanup uses the configured stale_session_threshold, defaulting to the same value used by partio cleanup and partio status.
  - A test verifies that a stale session (old file, dead PID) is transitioned to ENDED state when pre-commit runs.
  - No change to the hook's external behaviour when no stale session exists.
pr_labels:
  - enhancement
---

## Self-heal zombie sessions on pre-commit

Partio records a new active session in the pre-commit hook (`internal/hooks/precommit.go`). When an agent crashes without running a stop hook, it leaves a session in ACTIVE or IDLE state indefinitely. Today the only way to resolve it is to run `partio cleanup` or `partio status` manually.

The fix is to call `session.Manager.CleanupStale` at the start of pre-commit, before `RecordActive`. This mirrors what `partio status` already does (see `cmd/partio/status.go:53`). Because pre-commit runs on every commit, zombie sessions are resolved automatically the next time the developer commits — no manual step needed.

The pre-commit hook is the right place because it already owns the `session.Manager` for `RecordActive`, it runs in the repo root context where the stale session file lives, and adding the sweep before `RecordActive` keeps the session state consistent for the rest of the hook.

### Implementation sketch

In `internal/hooks/precommit.go`, add a call to `mgr.CleanupStale` before the existing `mgr.RecordActive` call (around line 127):

```go
mgr := session.NewManager(filepath.Join(repoRoot, config.PartioDir))
if result, err := mgr.CleanupStale(cfg.StaleSessionThreshold.Duration()); err != nil {
    slog.Debug("pre-commit: stale session cleanup failed", "error", err)
} else if result.Cleaned {
    slog.Warn("pre-commit: swept stale session", "id", result.Session.ID, "agent", result.Session.Agent)
}
if recErr := mgr.RecordActive(detector.Name(), branch, repoRoot, pid); recErr != nil {
    slog.Debug("could not record active session", "error", recErr)
}
```

The `mgr` variable is already needed for `RecordActive`; only the cleanup call is new. The configured threshold comes from `cfg.StaleSessionThreshold.Duration()`, which defaults to 10 minutes — the same value used by `partio cleanup` and `partio status`.
