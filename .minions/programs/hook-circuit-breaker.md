---
id: hook-circuit-breaker
target_repos:
  - cli
acceptance_criteria:
  - After a hook invocation fails (timeout or error), subsequent hook calls within the same git process skip execution and return success immediately
  - The circuit breaker state is scoped to the current process and does not persist across git commands
  - A debug-level log message is emitted when the circuit breaker trips and when it skips a hook invocation
  - Normal hook execution resumes on the next git command after a failure
  - The circuit breaker does not interfere with hook re-entry prevention for post-commit amend flows
pr_labels:
  - minion
---

# Add process-scoped circuit breaker for hook failures

## Problem

When Partio's git hooks encounter a failure — for example, the agent process is unresponsive and the hook times out — the same failure can repeat on every subsequent hook invocation within a single git operation. A `git commit` triggers pre-commit and then post-commit; if pre-commit times out after 30 seconds, post-commit will also time out, doubling the delay. Rebases and merges that touch many commits amplify this further.

## Desired Behavior

Add a process-scoped circuit breaker to hook execution. After the first hook failure within a git process lifetime:

1. Record the failure in a short-lived state file (e.g., `.partio/state/circuit-open.<pid>`)
2. Subsequent hook invocations check for the circuit file; if present and the PID matches, skip Partio logic and exit 0 immediately
3. The state file is cleaned up on next successful hook run or by a stale-file sweep (e.g., if the PID no longer exists)

This ensures that a single timeout or failure does not cascade into multiple delays within the same git operation, while still retrying on the next independent git command.

## Implementation Hints

- The circuit state file should use the parent process PID (the git process) to scope correctly
- Check for the circuit file early in the hook entry path, before any agent detection or session work
- Clean up stale circuit files (where the PID no longer exists) during hook entry or in `partio doctor`
- This interacts with the existing hook state passing (pre-commit.json → post-commit); ensure the circuit breaker check happens before state file reads

## Source

Inspired by entireio/cli#1218: "OPF: process-scoped circuit breaker after first failure"
