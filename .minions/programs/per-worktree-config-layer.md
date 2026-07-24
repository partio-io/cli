---
id: per-worktree-config-layer
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - "A `.partio/settings.worktree.json` file in the git worktree directory is loaded between the repo config and the local config in the precedence chain"
  - "The worktree config only affects the worktree it lives in — other worktrees sharing the same common git dir are unaffected"
  - "The file is optional: if absent, config resolution behaves exactly as before"
  - "`partio status` shows which config files are active (including the worktree layer when present)"
  - "`partio enable` documents the worktree config path in its output or help text"
  - "Tests use `t.TempDir()` to create multi-worktree setups and verify per-worktree isolation"
---

# Add per-worktree config layer (`.partio/settings.worktree.json`)

## Description

Partio already has a layered config system (defaults → global → repo → local → env), but all worktrees sharing a common git dir read the same repo-level `.partio/settings.json`. Teams using multiple worktrees often want different Partio behaviour per worktree — for example, disabling checkpointing in a dedicated build/test worktree while keeping it enabled in the main development worktree.

Add a `.partio/settings.worktree.json` file that is read from the _worktree's_ git dir (not the common dir), inserting it between repo and local in the precedence chain:

```
defaults → global (~/.config/partio/settings.json)
         → repo   (.partio/settings.json in common git dir)
         → worktree (.partio/settings.worktree.json in per-worktree git dir)
         → local  (.partio/settings.local.json)
         → env    (PARTIO_*)
```

This is analogous to the `.worktreeinclude` file introduced in entireio 0.7.8, adapted to Partio's flat JSON config format.

### Desired behaviour

```
# In worktree-b, disable checkpointing without touching the main worktree:
echo '{"enabled": false}' > .git/worktrees/worktree-b/partio/settings.worktree.json

# worktree-b: partio is disabled
cd worktree-b && partio status
# Partio: disabled (worktree override)

# main worktree: unaffected
cd main && partio status
# Partio: enabled
```

### Implementation notes

- The per-worktree config path resolves via `git rev-parse --git-dir` (the per-worktree git dir, e.g. `.git/worktrees/<name>/`) not `--git-common-dir`. This is already threaded through `internal/git/` for other operations.
- The `.partio/` subdirectory in the worktree git dir mirrors the repo-level convention; the file name suffix `.worktree.json` distinguishes it visually from the shared settings.
- `internal/config/` is where the layered load lives — insert the new layer after `LoadRepoConfig`.
- Gitignore: the worktree git dir is not in the working tree, so no `.gitignore` entry is needed.

<!-- program: .minions/programs/per-worktree-config-layer.md -->
