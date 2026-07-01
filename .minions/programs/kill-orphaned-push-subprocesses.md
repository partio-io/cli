---
id: kill-orphaned-push-subprocesses
target_repos:
  - cli
acceptance_criteria:
  - Pre-push hook push of partio/checkpoints/v1 has a configurable timeout (default 120s)
  - When the timeout fires, the entire process group (including child git-remote-https and credential helpers) is terminated
  - On Unix, use Setpgid + process group kill; on Windows, use CREATE_NEW_PROCESS_GROUP or equivalent
  - Push timeout produces a clear user-visible message instead of hanging silently
  - Tests verify timeout behavior using a mock slow git remote
pr_labels:
  - minion
---

# Terminate orphaned git subprocesses on checkpoint push timeout

Partio's pre-push hook pushes the `partio/checkpoints/v1` branch to origin. When this push hangs (slow network, unresponsive remote, credential helper stuck), the `exec.CommandContext` cancellation only SIGKILLs the direct `git push` child. Spawned subprocesses like `git-remote-https` and credential helpers inherit the stdio pipes and keep them open, causing `cmd.CombinedOutput()` to block indefinitely even after the context deadline.

## What to implement

1. Create a helper (e.g., `execx.KillOnCancel`) that sets up process group isolation for subprocess execution:
   - Unix: `Setpgid: true` in `SysProcAttr`, kill the negative PID (process group) on cancel
   - Windows: create a new process group and use `taskkill /F /T /PID` for descendant tree termination
   - Set `WaitDelay` as a backstop for pipe drain
2. Apply this helper to the git push subprocess in the pre-push hook's checkpoint push path.
3. When the push times out, emit a clear message (e.g., `[partio] Checkpoint push timed out after 120s`) instead of hanging with no output.
4. Skip any retry/sync cascade on timeout since the same network condition would trip again.

## Why this matters

A hung checkpoint push blocks `git push` entirely, which is the worst possible failure mode for a non-critical background operation. Users must Ctrl-C to recover, and the resulting error message is empty or misleading.
