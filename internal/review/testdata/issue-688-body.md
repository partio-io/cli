## What

Add Goose (Agentic AI Foundation, Apache-2.0) as a supported agent in Partio's detector registry, following the same pattern as `internal/agent/codex/`.

## Why

Partio captures AI agent sessions at commit time. The more agents it supports natively, the broader the user base that gets automatic session capture without configuration. Goose is an actively maintained open-source coding agent with a hook lifecycle and a first-class session export command.

## Description

Create a `goose` package under `internal/agent/goose/` that implements `agent.Detector`, `agent.SessionParser`, and `agent.PIDProvider`. Register it via `agent.Register("goose", ...)` in an `init()` function and import the package from the hook runner (alongside the existing codex import).

Goose stores sessions in SQLite at `~/.local/share/goose/sessions/sessions.db`. The integration must not open that database directly — use `goose session export --format json <id>` to read session content and discover sessions via `goose session list --format json` (or equivalent). This avoids coupling to a schema the vendor migrates in-place.

## Premise

<!-- partio:premise:v1 -->

- Partio's agent system supports registering new detector implementations via `agent.Register` in `internal/agent/registry.go` [evidence: `internal/agent/registry.go`]
- Partio has no Goose agent implementation — no package or `Register` call for "goose" exists under `internal/agent/` [evidence: `grep -rn "goose" internal/agent/`]

## Gathered evidence

**Claim 1:** Partio's agent system supports registering new detector implementations via `agent.Register` in `internal/agent/registry.go`
- Evidence: `internal/agent/registry.go`
- Verdict: `holds`
- Excerpt: Lines 12–17 define `registry = map[string]NewDetectorFunc{}` and `func Register(name string, fn NewDetectorFunc)`. Both claude-code and codex register themselves via `init()` calls in their respective `register.go` files.

**Claim 2:** Partio has no Goose agent implementation — no package or `Register` call for "goose" exists under `internal/agent/`
- Evidence: `grep -rn "goose" internal/agent/` → no output
- Verdict: `holds`
- Excerpt: The `internal/agent/` directory contains only `claude/` and `codex/` subdirectories. No file in the tree mentions "goose".

**Block verdict:** `holds`

## Source

Inspired by `entireio/cli` PRs #2075 and #2077 (feat(agent): add Goose coding agent integration).

## Program file

<!-- program: .minions/programs/goose-agent-integration.md -->
