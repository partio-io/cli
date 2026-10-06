## What Happens

When `partio enable` is run more than once in a repository that already has Partio hooks installed, the installer can silently overwrite an existing backup file.

The scenario:

1. Repo has a third-party `pre-commit` hook. User runs `partio enable`.
2. Installer renames `pre-commit` → `pre-commit.partio-backup`, writes Partio's hook as `pre-commit`.
3. User manually replaces `pre-commit` with a different third-party hook (e.g. updating hook manager config).
4. User runs `partio enable` again.
5. Installer sees the current `pre-commit` is not a Partio hook → renames it to `pre-commit.partio-backup`, silently overwriting the backup from step 2.
6. The original hook from step 1 is permanently lost.

This mirrors the bug described in entireio/cli PR #2641, adapted to Partio's hook installation logic.

## What to Build

In `installHooks` (`internal/git/hooks/install.go`), before calling `os.Rename(hookPath, backupPath)`, check whether `backupPath` already exists. If it does, return an error explaining that a backup already exists and the user must remove or rename it before re-installing:

```go
if _, statErr := os.Stat(backupPath); statErr == nil {
    return fmt.Errorf("backup %s already exists; remove it before re-running partio enable", backupPath)
}
```

This makes re-installation explicit and safe: the user knows they must handle the existing backup rather than having it silently overwritten.

## Acceptance Criteria

- [ ] When `partio enable` is run and `<hook>.partio-backup` already exists and the current hook is not a Partio hook, the command returns a clear error rather than overwriting the backup.
- [ ] When `partio enable` is run and no backup exists, hook installation succeeds as before.
- [ ] When `partio enable` is run and the current hook is already a Partio hook (idempotent re-install), no backup operation occurs and the install succeeds.
- [ ] `make test` passes.
- [ ] `make lint` passes.

## Premise

<!-- partio:premise:v1 -->

- Partio's hook installer calls `os.Rename(hookPath, backupPath)` without first checking whether `backupPath` already exists, so an existing backup can be silently overwritten on re-install. [evidence: `internal/git/hooks/install.go`]

## Gathered Evidence

**Claim:** Partio's hook installer calls `os.Rename(hookPath, backupPath)` without checking whether `backupPath` already exists.
**Evidence:** `internal/git/hooks/install.go`
**Verdict:** holds
**Excerpt:**
```go
// install.go:38–44
if data, err := os.ReadFile(hookPath); err == nil {
    content := string(data)
    if !isPartioHook(content) {
        if err := os.Rename(hookPath, backupPath); err != nil {
            return fmt.Errorf("backing up %s hook: %w", name, err)
        }
    }
}
```
No existence check for `backupPath` before the rename. `os.Rename` on Linux replaces the destination atomically; on Windows it fails if the destination exists — but on Linux a pre-existing backup is silently overwritten.

Proposal id: hook-backup-overwrite-protection
