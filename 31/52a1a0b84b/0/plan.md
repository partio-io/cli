# Plan: `partio restart` — Resume sessions from checkpoints

## Context

Partio currently captures Claude Code sessions during git commits and displays them on the web app. The data flow is one-directional: CLI → Git (orphan branch) → GitHub → Web App. There is no way to go back from a viewed session to a running Claude Code session. This is needed when:

- A different developer wants to continue work from someone else's session
- The original branch/worktree no longer exists
- The developer is in a different session and wants the prior context

The checkpoint data (prompt, plan, diff, transcript) already lives in the git repo on the `partio/checkpoints/v1` orphan branch. We'll read it back and feed it into a new Claude Code session.

## Changes Overview

### 1. CLI: New `partio restart <checkpoint-id>` command

**New file: `cli/internal/checkpoint/read.go`**

Add functions to read checkpoint + session data from the orphan branch:

```go
// CheckpointData holds all readable data from a stored checkpoint.
type CheckpointData struct {
    Metadata Metadata
    Prompt   string
    Plan     string
    Diff     string
    Context  string
}

// Read retrieves all checkpoint data from the orphan branch by ID.
func Read(id string) (*CheckpointData, error)
```

Implementation uses `git show partio/checkpoints/v1:<shard>/<rest>/...` pattern already proven in `rewind.go`. Reads:
- `metadata.json` → parsed into `Metadata`
- `0/prompt.txt` → original prompt
- `0/plan.md` → plan (may be empty)
- `0/diff.patch` → diff
- `0/context.md` → context summary

**New file: `cli/cmd/partio/restart.go`**

Command with flags:
- `--print` — Print the composed context prompt to stdout (don't launch claude)
- `--copy` — Copy context prompt to clipboard
- `--branch` — Create a branch at the checkpoint's commit before launching

Default behavior (no flags):
1. Read checkpoint data via `checkpoint.Read(id)`
2. Write a context file to `.partio/restart/<checkpoint-id>.md` containing the composed prompt
3. Launch `claude` interactively with a short prompt referencing the context file

The composed context prompt template:

```markdown
# Previous Session Context

You are continuing work from a previous Partio session (checkpoint {id}).

## Original Request

{prompt}

## Plan

{plan or "No plan was recorded."}

## Changes Made

{diff}

## Session Info

- **Branch:** {branch}
- **Commit:** {commit_hash}
- **Date:** {created_at}
- **Agent:** {agent} ({agent_percent}%)

---

Please review the current state of the repository and continue this work.
```

Launching Claude Code:
- Find `claude` in PATH via `exec.LookPath("claude")`
- Write context to `.partio/restart/<id>.md`
- Use `syscall.Exec` to replace process with: `claude "Read .partio/restart/<id>.md for full context on a previous session, then continue that work."`
- If `claude` not found, fall back to printing the prompt with instructions

Clipboard copy (`--copy`):
- Use `exec.Command("pbcopy")` on macOS, `xclip -selection clipboard` on Linux
- Write the full context prompt to stdin

**Modified file: `cli/cmd/partio/main.go`**

Add `newRestartCmd()` to the root command's `AddCommand()` list.

### 2. Web App: Restart button on checkpoint detail page

**New file: `app/src/components/ui/restart-button.tsx`**

A button component that:
- Displays a terminal-style icon + "Restart" label
- On click, copies `partio restart <checkpoint-id>` to clipboard
- Shows a brief "Copied!" confirmation (swap text for ~2s via state)
- Styled to match existing pills/buttons in the header (uses `accent-orange` theme)

```tsx
interface RestartButtonProps {
  checkpointId: string;
}
```

**Modified file: `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx`**

Add the `RestartButton` in the header section, after the metadata pills row. Placed at the same level as the existing checkpoint info.

## Files to create

| File | Purpose |
|------|---------|
| `cli/internal/checkpoint/read.go` | Read checkpoint data from orphan branch |
| `cli/cmd/partio/restart.go` | The `restart` command |
| `app/src/components/ui/restart-button.tsx` | Copy-to-clipboard restart button |

## Files to modify

| File | Change |
|------|--------|
| `cli/cmd/partio/main.go:58-68` | Add `newRestartCmd()` to `AddCommand()` |
| `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx:130-198` | Import and render `RestartButton` in header |

## Existing code to reuse

- `checkpoint.Shard(id)` / `checkpoint.Rest(id)` from `cli/internal/checkpoint/checkpoint.go:43-56`
- `git.ExecGit(...)` from `cli/internal/git/git.go:21-23`
- `git.RepoRoot()` from `cli/internal/git/` for repo validation
- `git.CheckpointBranch` constant from `cli/internal/git/git.go:8`
- The `git show branch:path` pattern from `cli/cmd/partio/rewind.go:74-120`
- Command structure pattern from `cli/cmd/partio/rewind.go`

## Verification

1. **CLI unit test**: Test `checkpoint.Read()` with a mock git repo that has an orphan branch with test checkpoint data
2. **CLI integration test**:
   - In a repo with checkpoints: `partio restart --print <id>` should output the composed prompt
   - `partio restart --copy <id>` should copy to clipboard
   - `partio restart <nonexistent>` should error gracefully
3. **Web app**: Run `npm run dev`, navigate to a checkpoint detail page, verify the Restart button appears and copies the correct command
4. **End-to-end**: Run `partio restart <id>` in a repo with checkpoints, verify it launches Claude Code with the context
