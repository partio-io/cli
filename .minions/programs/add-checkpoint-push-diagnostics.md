---
id: add-checkpoint-push-diagnostics
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint push operations log each attempt with branch name, target, elapsed time, and outcome classification at INFO level
  - Push errors are classified into distinct types (timeout, fetch failure, auth error, network error)
  - Timeout errors display a clear message instead of hanging silently
  - Push subprocess cancellation kills the entire process group to prevent orphaned git-remote-https processes
  - Table-driven tests verify error classification and log output for each failure type
pr_labels:
  - minion
---

# Add structured diagnostics for checkpoint push operations

## Summary

Make checkpoint push failures in the pre-push hook diagnosable and bounded by adding structured logging, typed error classification, and proper process group termination. Currently, when `partio`'s pre-push hook pushes the checkpoint branch and the push hangs or fails, users see no diagnostic output and may experience indefinite hangs.

## Motivation

Partio's pre-push hook pushes `partio/checkpoints/v1` to origin. When this push encounters network issues, credential problems, or remote-side delays, the hook can hang indefinitely because:
- `exec.CommandContext` only kills the direct child process, not spawned helpers like `git-remote-https`
- There are no structured logs of push attempts, so `partio doctor` cannot show push history
- Push errors lack classification, making it unclear whether the issue is transient (retry) or permanent (fix config)

## Implementation Notes

- Add process group killing on context cancellation (Unix: `Setpgid` + kill process group; consider platform differences)
- Classify push errors into typed constants: `errPushTimedOut`, `errFetchFailed`, `errAuthFailed`, `errNetworkError`
- Log each push attempt at INFO level with branch, remote, elapsed time, and classification
- When a push times out, display `" timed out"` on the progress line instead of trailing silently
- Skip retry cascade on timeout (same network condition would trip again)
- Ensure outer-context cancellation (Ctrl-C / hook deadline) bails cleanly at every stage

## Source

Inspired by entireio/cli#1193 which addresses stuck checkpoint pushes with process group killing, typed error classification, and structured push logging.
