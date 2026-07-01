---
id: pi-agent-integration
target_repos:
  - cli
acceptance_criteria:
  - "Pi coding agent is detected via the detector interface when running"
  - "Pi session transcripts are discovered and parsed from Pi's native session store"
  - "Checkpoints include Pi session data when commits are made during a Pi session"
  - "partio enable --agent pi configures hooks for Pi"
  - "partio status shows Pi as the active agent when detected"
  - "PARTIO_AGENT=pi environment variable selects Pi detection"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add Pi coding agent integration

## Summary

Add built-in support for the Pi coding agent (github.com/earendil-works/pi-mono) alongside the existing Claude Code integration. Pi is a coding agent that stores session transcripts in its own native format. Partio should detect when Pi is running, discover its session files, parse its transcript format, and include session data in checkpoints.

## Context

- `internal/agent/detector.go` — pluggable agent detection interface
- `internal/agent/claude/` — existing Claude Code implementation (reference for new agent)
- `internal/session/` — session lifecycle management
- `cmd/partio/enable.go` — enable command with --agent flag

## Implementation notes

- Create `internal/agent/pi/` package following the same structure as `internal/agent/claude/`
- Implement the Detector interface for Pi: process detection, session directory discovery, transcript parsing
- Pi stores sessions at `<piHome>/sessions/<encoded-repo>/<ts>_<id>.jsonl` where encoded path uses `--` delimiters (e.g., `/Users/foo/repo` becomes `--Users-foo-repo--`)
- Register Pi in the agent registry so `--agent pi` and `PARTIO_AGENT=pi` work
- Honor `PI_CODING_AGENT_DIR` environment variable for custom Pi home directories
