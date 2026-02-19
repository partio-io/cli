# partio

Capture the *why* behind your code changes.

**partio** hooks into Git workflows to capture AI agent sessions (Claude Code), preserving the reasoning behind code changes alongside the *what* that Git already tracks.

The "partial" version of [entire.io](https://entire.io).

## Install

```bash
go install github.com/partio-io/cli/cmd/partio@latest
```

Or with Homebrew:

```bash
brew install partio-io/tap/partio
```

## Quick Start

```bash
# Enable in your repo
cd your-project
partio enable

# Code with Claude Code as usual, then commit
git commit -m "add new feature"
# partio automatically captures the AI session as a checkpoint

# View checkpoints
partio rewind --list

# Inspect a checkpoint
git show partio/checkpoints/v1:<shard>/<id>/0/full.jsonl

# Rewind to a checkpoint
partio rewind --to <id>

# Check status
partio status
```

## Commands

| Command | Description |
|---------|-------------|
| `partio enable` | Set up partio in the current repo |
| `partio disable` | Remove hooks (preserves data) |
| `partio status` | Show current status |
| `partio rewind --list` | List all checkpoints |
| `partio rewind --to <id>` | Restore to a checkpoint |
| `partio doctor` | Check installation health |
| `partio reset` | Reset the checkpoint branch |
| `partio clean` | Remove orphaned data |
| `partio version` | Print version |

## How It Works

1. `partio enable` installs git hooks (`pre-commit`, `post-commit`, `pre-push`)
2. When you commit, hooks detect if Claude Code is running
3. If active, it captures the JSONL transcript, calculates attribution, and creates a checkpoint
4. Checkpoints are stored on an orphan branch (`partio/checkpoints/v1`) using git plumbing
5. Commits are annotated with `Partio-Checkpoint` and `Partio-Attribution` trailers
6. On push, the checkpoint branch is pushed alongside your code

## Git Worktrees

partio fully supports git worktrees. Hooks are installed to the shared git directory (`git rev-parse --git-common-dir`) so they work across all worktrees. Session discovery walks up from the repo root to find the Claude Code session directory, which may be keyed to a parent workspace directory.

## Checkpoint Data

Checkpoints are stored on the `partio/checkpoints/v1` orphan branch with this structure:

```
<shard>/<checkpoint-id>/
  metadata.json          # Checkpoint metadata (commit, branch, agent %, timestamps)
  0/
    metadata.json        # Session metadata (agent, tokens, duration)
    context.md           # First 200 chars of the initial prompt
    prompt.txt           # Full initial human message
    full.jsonl           # Complete Claude Code transcript
    content_hash.txt     # Commit hash reference
```

You can inspect checkpoint data directly with git:

```bash
# List all checkpoint files
git ls-tree -r --name-only partio/checkpoints/v1

# View checkpoint metadata
git show partio/checkpoints/v1:<shard>/<id>/metadata.json

# View the full Claude session
git show partio/checkpoints/v1:<shard>/<id>/0/full.jsonl
```

## Configuration

Config files (highest priority wins):
- Environment variables (`PARTIO_ENABLED`, `PARTIO_STRATEGY`, `PARTIO_LOG_LEVEL`)
- `.partio/settings.local.json` (git-ignored)
- `.partio/settings.json`
- `~/.config/partio/settings.json`

```json
{
  "enabled": true,
  "strategy": "manual-commit",
  "agent": "claude-code",
  "log_level": "info",
  "strategy_options": { "push_sessions": true }
}
```

## License

MIT
