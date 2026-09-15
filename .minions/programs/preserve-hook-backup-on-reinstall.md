---
id: preserve-hook-backup-on-reinstall
target_repos:
  - cli
acceptance_criteria:
  - When `partio enable` encounters a non-partio hook at a hook path and a `.partio-backup` file already exists at the backup path, it does not overwrite the existing backup
  - Instead, it logs a warning naming both the current foreign hook and the intact backup, and writes the partio hook script to the hook path without touching the backup
  - `partio disable` on a repository that has gone through multiple enable cycles restores the file that was originally backed up on the first install, not a later foreign hook
  - A unit test in `internal/git/hooks/` covers the reinstall-with-existing-backup scenario and asserts the original backup content is preserved
pr_labels:
  - bug
---

## Description

`installHooks` in `internal/git/hooks/install.go` backs up any non-partio hook it encounters by renaming it to `<hook>.partio-backup`. The rename runs unconditionally: it does not check whether a backup file already exists at the destination. On Linux, `os.Rename` atomically replaces an existing destination, so a second `partio enable` after an external hook manager has reinstalled its hook silently overwrites whatever was in the backup.

The sequence that triggers the loss:

1. User has `pre-commit` (original hook).
2. `partio enable`: renames `pre-commit` → `pre-commit.partio-backup`; writes partio hook to `pre-commit`.
3. Hook manager (e.g. lefthook) reinstalls: overwrites `pre-commit` with its own script.
4. `partio enable` again: sees `pre-commit` is not a partio hook → renames `pre-commit` (hook manager's script) to `pre-commit.partio-backup`, **overwriting the original user hook**; writes partio hook.
5. `partio disable`: removes partio hook, restores `pre-commit.partio-backup` — which now contains the hook manager's script, not the original.

The fix is to check whether `backupPath` exists before renaming. If it does, skip the rename and emit a warning; the partio hook is still written to `hookPath`, so capture continues working. The backup already preserves whatever it holds — overwriting it silently is the bug.

## Why

Users who run `partio enable` more than once (which the docs encourage as the idempotent repair path) and whose repos also use a hook manager will silently lose their original hook scripts. The loss is invisible at install time and only discovered after `partio disable`, when the restored hook is the wrong one. This is especially painful for users who had carefully crafted commit-msg or post-commit scripts.

## User relevance

Any developer who uses `partio enable` alongside Husky, Lefthook, or Overcommit and later needs to restore their original hooks via `partio disable` is affected. The current behaviour makes `partio disable` untrustworthy as a reversible operation after the second enable.

## Context hints

- `internal/git/hooks/install.go` — `installHooks`, specifically the `os.Rename(hookPath, backupPath)` call
- `internal/git/hooks/uninstall.go` — `Uninstall`, which restores `backupPath` to `hookPath`
- `internal/git/hooks/hooks_test.go` — existing hook install tests to extend
