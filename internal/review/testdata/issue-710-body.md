## Description

`installHooks` in `internal/git/hooks/install.go` backs up any non-partio hook by renaming it to `<hook>.partio-backup`. The rename is unconditional: no check is made for an existing backup at the destination. On Linux, `os.Rename` atomically replaces an existing destination, so a second `partio enable` after a hook manager has reinstalled its hook silently overwrites whatever the first install backed up.

**Triggering sequence:**

1. User has `pre-commit` (their original hook).
2. `partio enable`: `pre-commit` → `pre-commit.partio-backup`; partio hook written.
3. Hook manager (Lefthook, Husky, …) reinstalls: overwrites `pre-commit` with its own script.
4. `partio enable` again: sees a non-partio `pre-commit` → `os.Rename(hookPath, backupPath)` — atomically replaces `pre-commit.partio-backup` with the hook manager's script — partio hook written.
5. `partio disable`: restores `pre-commit.partio-backup`, which now contains the hook manager's script, not the user's original.

## Premise

<!-- partio:premise:v1 -->

- `installHooks` in `internal/git/hooks/install.go` calls `os.Rename(hookPath, backupPath)` to back up a non-partio hook without first checking whether `backupPath` already exists; on Linux `os.Rename` atomically replaces an existing destination, so the prior backup is silently overwritten [evidence: `internal/git/hooks/install.go`]

## Gathered evidence

**Claim**: `installHooks` calls `os.Rename(hookPath, backupPath)` with no prior existence check on `backupPath`.

**Evidence**: `internal/git/hooks/install.go` lines 38–45:

```go
if data, err := os.ReadFile(hookPath); err == nil {
    content := string(data)
    if !isPartioHook(content) {
        if err := os.Rename(hookPath, backupPath); err != nil {
            return fmt.Errorf("backing up %s hook: %w", name, err)
        }
    }
}
```

**Verdict**: `holds` — `os.Rename(hookPath, backupPath)` at line 41 runs without any `os.Stat(backupPath)` guard. On Linux, `os.Rename` is specified to atomically replace the destination if it exists (`rename(2)`: "If newpath already exists, it will be atomically replaced"), so any previously backed-up file is silently discarded.

**Corroborating evidence**: `internal/git/hooks/uninstall.go` lines 32–36 confirm that `Uninstall` restores `backupPath` to `hookPath`, meaning whatever ends up in the backup after a reinstall is what `partio disable` returns to the user.

## Acceptance criteria

- When `installHooks` encounters a non-partio hook and `backupPath` already exists, it does not overwrite the backup
- It emits a warning naming the existing backup and the current foreign hook, then writes the partio hook to `hookPath`
- `partio disable` on a repository that has gone through multiple enable cycles restores the file backed up on the first install
- A unit test in `internal/git/hooks/` covers the reinstall-with-existing-backup scenario and asserts the original backup content is preserved

## Source

Inspired by entireio/cli#2237 (Re-install discards another tool's newer hook and chains to a stale backup).

## Program file

`.minions/programs/preserve-hook-backup-on-reinstall.md`

<!-- program: .minions/programs/preserve-hook-backup-on-reinstall.md -->
