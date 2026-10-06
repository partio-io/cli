## Summary

Goose (Agentic AI Foundation, Apache-2.0) fires `session-start`, `turn-start`, and `turn-end` lifecycle events as subprocess hooks, writes a JSONL session transcript on disk, and exposes `session_id` on hook stdin. This maps directly onto the three capabilities Partio needs from any agent: process detection, session directory discovery, and transcript parsing.

Partio already ships two agents (`claude-code`, `codex`) built on the same `agent.Detector` / `agent.SessionParser` / `agent.Register` seam. Adding Goose follows the same pattern without touching the hook infrastructure.

Inspired by: entireio/cli PR #2075, #2076, #2077 (Goose agent integration).

## What to build

Add an `internal/agent/goose/` package implementing the `agent.Detector` and `agent.SessionParser` interfaces, registered via an `init()` function so `partio enable --agent goose` works without further changes.

Files:
- `goose.go` — `Detector` struct with `Name() string`
- `process.go` — `IsRunning()` via `pgrep -f goose`; `AgentPID()` via `agent.PgrepFirst`
- `find_session_dir.go` — locate `~/.config/goose/sessions/`
- `parse_jsonl.go` — walk session files, match by CWD, parse into `*agent.SessionData`
- `register.go` — `agent.Register("goose", ...)` in `init()`

Blank-import the package alongside `claude` and `codex` so registration runs.

## Premise

<!-- partio:premise:v1 -->

- Partio exposes a `Detector` interface with `Name()`, `IsRunning()`, and `FindSessionDir()` methods that new agents implement [evidence: `internal/agent/detector.go`]
- New agents are registered via `agent.Register()` in package `init()` functions [evidence: `internal/agent/claude/register.go`, `internal/agent/codex/register.go`]

## Gathered evidence

**Claim 1:** Partio exposes a `Detector` interface with `Name()`, `IsRunning()`, and `FindSessionDir()` methods that new agents implement.
- Evidence: `internal/agent/detector.go`
- Verdict: `holds`
- Excerpt:
  ```go
  type Detector interface {
      Name() string
      IsRunning() (bool, error)
      FindSessionDir(repoRoot string) (string, error)
  }
  ```

**Claim 2:** New agents are registered via `agent.Register()` in package `init()` functions.
- Evidence: `internal/agent/claude/register.go`, `internal/agent/codex/register.go`
- Verdict: `holds`
- Excerpt from `codex/register.go`:
  ```go
  func init() {
      agent.Register("codex", func() agent.Detector { return New() })
  }
  ```

<!-- program: .minions/programs/goose-agent-integration.md -->
