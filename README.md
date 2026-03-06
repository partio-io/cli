# partio

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://go.dev/)
[![Release](https://img.shields.io/badge/release-v0.1.0-orange.svg)](https://github.com/partio-io/cli/releases/latest)

**Capture the *why* behind your code changes.**

partio hooks into Git to capture AI agent sessions (Claude Code) alongside your commits. When you commit, partio snapshots the full conversation — prompts, plans, tool calls — so your team can understand *why* code changed, not just *what* changed.

Everything stays in your repo as Git objects. Nothing leaves your machine.

![partio demo](assets/demo.gif)

## Install

### Homebrew

```bash
brew install partio-io/tap/partio
```

### Go

Requires Go 1.25+.

```bash
go install github.com/partio-io/cli/cmd/partio@latest
```

## Quick Start

```bash
# 1. Enable partio in your repo
cd your-project
partio enable

# 2. Code with Claude Code, then commit as usual
git commit -m "add new feature"
# partio automatically captures the AI session as a checkpoint

# 3. Check status
partio status

# 4. List all checkpoints
partio rewind --list

# 5. Resume a previous session
partio resume <checkpoint-id> --print
```

## Commands

| Command | Description |
|---------|-------------|
| `partio enable` | Set up partio in the current repo |
| `partio disable` | Remove hooks (preserves checkpoint data) |
| `partio status` | Show current status |
| `partio rewind --list` | List all checkpoints |
| `partio rewind --to <id>` | Restore repo to a checkpoint's commit |
| `partio resume <id>` | Launch Claude Code with checkpoint context |
| `partio resume <id> --print` | Print the composed context to stdout |
| `partio resume <id> --copy` | Copy the composed context to clipboard |
| `partio resume <id> --branch` | Create a branch at the checkpoint's commit before launching |
| `partio doctor` | Check installation health |
| `partio reset` | Reset the checkpoint branch |
| `partio clean` | Remove orphaned checkpoint data |
| `partio version` | Print version |

## How It Works

1. `partio enable` installs Git hooks (`pre-commit`, `post-commit`, `pre-push`)
2. When you commit, the pre-commit hook detects if Claude Code is running
3. If active, the post-commit hook captures the JSONL transcript, calculates attribution, and creates a checkpoint
4. Checkpoints are stored on an orphan branch (`partio/checkpoints/v1`) using Git plumbing — no working tree changes
5. Commits are annotated with `Partio-Checkpoint` and `Partio-Attribution` trailers
6. On push, the checkpoint branch is pushed alongside your code

<details>
<summary><strong>Checkpoint Data Structure</strong></summary>

Checkpoints are stored on the `partio/checkpoints/v1` orphan branch:

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

Inspect checkpoint data directly with Git:

```bash
# List all checkpoint files
git ls-tree -r --name-only partio/checkpoints/v1

# View checkpoint metadata
git show partio/checkpoints/v1:<shard>/<id>/metadata.json

# View the full Claude session
git show partio/checkpoints/v1:<shard>/<id>/0/full.jsonl
```

</details>

<details>
<summary><strong>Configuration</strong></summary>

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

</details>

<details>
<summary><strong>Git Worktrees</strong></summary>

partio fully supports Git worktrees. Hooks are installed to the shared git directory (`git rev-parse --git-common-dir`) so they work across all worktrees. Session discovery walks up from the repo root to find the Claude Code session directory, which may be keyed to a parent workspace directory.

</details>

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](LICENSE)
