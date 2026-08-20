---
id: auto-elect-checkpoint-remote
target_repos:
  - cli
acceptance_criteria:
  - After a successful checkpoint push in pre-push, the remote name is persisted as push_sessions_remote in .partio/settings.local.json
  - On subsequent pre-push invocations, the elected remote is used for checkpoint sync regardless of which remote the user is pushing to
  - If push_sessions_remote is already explicitly set in .partio/settings.json or global config, the auto-election is skipped
  - The auto-election is logged at info level so users can discover the behavior
  - A unit test covers the election write path with a temporary git repo
pr_labels:
  - minion
---

# Auto-elect checkpoint sync remote from observed push behavior

### Problem

Partio's pre-push hook defaults to `origin` when deciding where to sync
`partio/checkpoints/v1`. Users who routinely push to a non-origin remote
(e.g., a team remote, a mirror, or a fork upstream) discover that checkpoints
strand locally because the hook keeps targeting a dormant `origin`.

The current behavior requires manual intervention: the user must discover the
mismatch via `partio status` and then hand-configure `push_sessions_remote` in
`.partio/settings.json`.

### Desired behavior

After the first successful checkpoint sync, Partio should observe the remote
name it used (available from the pre-push hook's `$1` argument) and persist it
as `push_sessions_remote` in `.partio/settings.local.json`. From that point on,
checkpoint syncs follow the elected remote automatically.

If `push_sessions_remote` is already set in any config layer with higher
precedence than `.partio/settings.local.json` (project or global settings),
the auto-election is a no-op — the user's explicit setting wins.

The election should be logged at info level:

```
INFO pre-push: elected "upstream" as checkpoint sync remote; saved to .partio/settings.local.json
```

## Context hints

- `internal/hooks/` — pre-push hook implementation receives remote name as its first argument
- `internal/config/` — layered config system; write to `.partio/settings.local.json` (local layer)
- `internal/git/` — git operations; the remote name from hook args maps to a git remote URL

## Source

Inspired by entireio/cli PR#1991: "feat(strategy): capture the checkpoint sync
remote from tracked pushes" — which solved the same problem by observing the
actual push target rather than guessing from static config.
