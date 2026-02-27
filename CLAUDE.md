# CLAUDE.md — ai-workflow-core (Partio)

## Project Overview

Partio captures the *why* behind code changes by hooking into Git workflows to preserve AI agent sessions (currently Claude Code) alongside the code changes tracked by Git. It stores checkpoints on an orphan branch (`partio/checkpoints/v1`) using git plumbing commands.

**Module:** `github.com/partio-io/cli`
**Go version:** 1.25.0
**CLI framework:** Cobra

## Project Structure

```
cmd/partio/          CLI commands (enable, disable, status, rewind, resume, doctor, hook, etc.)
internal/
  agent/             Agent detection interface + Claude Code implementation
    claude/          Claude-specific: process detection, JSONL parsing, session discovery
  attribution/       Code attribution calculation (agent vs human line counts)
  checkpoint/        Checkpoint domain type + git plumbing-based storage
  config/            Layered configuration (defaults → global → repo → local → env)
  git/               Git operations (repo info, diff, branch, hooks/)
    hooks/           Hook script generation, install/uninstall
  hooks/             Git hook implementations (pre-commit, post-commit, pre-push)
  session/           Session lifecycle management and state persistence
  log/               slog-based logging setup
```

## Build & Test

```bash
make build       # Compile binary (embeds version from git tags)
make test        # Run all tests (go test -v ./...)
make lint        # Run golangci-lint
make install     # Build and install to $GOPATH/bin
make clean       # Remove compiled binary
```

## Architecture

- **DDD-influenced**: Domain types (`Checkpoint`, `SessionData`, `Config`, `Session`) live alongside their logic in focused packages.
- **Detector interface** (`agent/detector.go`): Pluggable agent detection — currently only Claude Code is implemented.
- **Git plumbing storage**: Checkpoints are written directly to git object database (hash-object, mktree, commit-tree, update-ref) without checkout.
- **Hook state passing**: Pre-commit saves detection state to `.partio/state/pre-commit.json`; post-commit reads and deletes it immediately to create checkpoints.
- **Layered config**: Defaults → `~/.config/partio/settings.json` → `.partio/settings.json` → `.partio/settings.local.json` → environment variables.

## Key Patterns & Conventions

- **One primary concern per file** — e.g., `find_session_dir.go`, `find_latest_session.go`, `parse_jsonl.go`.
- **Table-driven tests** with `t.TempDir()` for filesystem isolation and `t.Setenv()` for env vars.
- **No external test frameworks** — standard library `testing` only.
- **Minimal dependencies** — only `cobra` (CLI) and `google/uuid`.
- **Error resilience in hooks** — hooks log warnings but don't block git operations on non-critical failures.
- **Binary attribution** — currently 0% or 100% based on agent detection, not complexity analysis.

## Data Flow

```
git commit
  → pre-commit:  detect if Claude Code is running, save state to .partio/state/
  → commit completes
  → post-commit: read + delete state → calculate attribution → parse JSONL session
                  → create checkpoint on orphan branch → amend commit with trailers
git push
  → pre-push:    push checkpoint branch to origin (if push_sessions enabled)
```

## Git Worktree Support

Partio supports git worktrees:

- **Hooks** are installed to `git rev-parse --git-common-dir` (shared across worktrees), not the per-worktree git dir.
- **Session discovery** (`find_session_dir.go`) walks up from the repo root to parent directories, since Claude Code keys sessions to the cwd where it was launched (which may be a parent of the git repo root in worktree setups).
- **Hook re-entry prevention**: Post-commit deletes the state file immediately (not via defer) before amending the commit, since `git commit --amend` re-triggers post-commit (`--no-verify` only skips pre-commit/commit-msg). Pre-push uses `--no-verify` on its internal push to avoid recursion.
- **Hook scripts** use `git rev-parse --git-common-dir` for backup chaining and end with `exit 0` to prevent `[ -f ... ]` false-returns from failing the hook.

## Environment Variables

| Variable | Purpose |
|---|---|
| `PARTIO_ENABLED` | Enable/disable partio |
| `PARTIO_STRATEGY` | Capture strategy (default: `manual-commit`) |
| `PARTIO_LOG_LEVEL` | Log level: debug, info, warn, error |
| `PARTIO_AGENT` | Agent name (default: `claude-code`) |

## Known Limitations

- `partio enable` skips hook installation if `.partio/` already exists. To reinstall hooks, run `partio disable && partio enable`.
- The `_hook` command accepts extra args (`cobra.MinimumNArgs(1)`) because git passes additional arguments to hooks (e.g., remote name and URL for pre-push).
- Session JSONL parsing may produce empty context/prompt fields depending on the Claude session format.
