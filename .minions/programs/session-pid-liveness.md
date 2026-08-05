---
id: session-pid-liveness
target_repos: [cli]
acceptance_criteria:
  - "Session-end and condensed-session skip decisions consult the liveness of the agent PID recorded in the session state, not a machine-global process-name grep"
  - "When the session record carries no PID, the decision falls back to the current global detection, so capture never regresses"
  - "Unit tests cover the liveness decision with a live PID, a dead PID, and an absent PID"
  - "A second commit during a still-live session still receives its Partio-Checkpoint trailer (the PR #586 behavior is preserved)"
  - "The pre-commit condensed-session block is either revived by the real liveness signal or deleted outright, and the PR states which and why"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Use the recorded agent PID for session-end and skip decisions

Hook liveness currently asks "is any process with `claude` in its command line
running anywhere on this machine?" — `(*claude.Detector).IsRunning()` in
`internal/agent/claude/process.go` shells out to `pgrep -f claude`. On a
workstation where some Claude session is nearly always running (another repo,
another terminal), that answer is effectively always yes. Session-end is never
observed, the condensed-session skip never engages, and a manual commit made
after a session ended is rewritten with agent-attribution trailers and a
redundant checkpoint — false data in the product's core promise.

The information needed for a correct answer is already recorded: pre-commit
resolves the owning agent's PID through the `agent.PIDProvider` seam
(`internal/agent/detector.go`) and persists it as `agent_pid` in
`.partio/sessions/current.json` via `session.Manager.RecordActive`
(`internal/session/record.go`; the `Session` struct with `AgentPID` lives in
`internal/session/session.go`). Make session-end and skip decisions consult
that PID's liveness instead of the global grep. Strictly the session-end/skip
decision: agent detection for attribution percentages is out of scope and
unchanged.

## Agents

### session-pid-liveness

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

Work through the change in this order:

1. **Lift the existing liveness seam.** `internal/session/cleanup.go`
   already implements exactly this decision for stale-session cleanup:
   `isStale` checks `Session.AgentPID` with `isProcessAlive(pid)`
   (`syscall.Kill(pid, 0)`). Expose it as a reusable method on
   `session.Manager` (e.g. `AgentAlive() bool`) instead of
   re-implementing it. The method answers "is the PID that owns this
   repo's session alive?" and must fall back to the caller's global
   detection when the record carries no PID (`agent_pid` 0 or missing —
   an older record, a detector without `PIDProvider`, a pgrep miss at
   record time). Prefer today's over-capturing to new under-capturing.

2. **Switch the decision sites in the hooks.** In
   `internal/hooks/precommit.go` and `internal/hooks/postcommit.go`,
   wherever session end or the condensed-session skip is decided by the
   detector's machine-global `IsRunning()` — the `shouldSkipSession`
   predicate and the `MarkCondensed` call — consult the recorded PID's
   liveness instead, with the no-PID fallback from step 1.

3. **Resolve the pre-commit condensed-session block.** PR #586 fixed a
   live session's second commit losing its trailer by threading an
   `agentRunning` flag through `MarkCondensed` and `shouldSkipSession`;
   pre-commit passes `agentRunning=true`, which leaves the pre-commit
   condensed-session block (`internal/hooks/precommit.go`, the
   `// Check for condensed sessions` block) structurally unable to
   fire. With a real per-session liveness signal available, either
   revive that block by passing the actual PID liveness instead of a
   literal `true`, or delete it if the post-commit skip alone suffices.
   State in the PR description which you chose and why. Whichever way,
   the #586 guarantee stays: a second commit in a still-live session
   gets its trailer. If #586 is not yet merged when you build, apply
   the same PID-liveness replacement to the current call sites and say
   so in the PR.

4. **Test the liveness decision.** Unit tests must cover: a live PID
   (use the test process's own PID), a dead PID (a spawned-and-waited
   child), and an absent PID (record with `agent_pid` 0 falls back to
   global detection). Mirror the table-driven patterns in
   `internal/session/cleanup_test.go`. Add a hook-level test proving a
   second commit during a still-live session still receives its
   Partio-Checkpoint trailer.

5. **Run `make test` and `make lint`** and fix what they surface.
