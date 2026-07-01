---
id: ci-headless-agent-detection
target_repos:
  - cli
acceptance_criteria:
  - "Agent detection works in CI environments (GitHub Actions, etc.) where agents run as subprocesses"
  - "Detection checks for CI-specific session paths and environment variables"
  - "Claude Code detector handles non-interactive/headless session layouts"
  - "New `--ci` flag or `PARTIO_CI` env var opts into CI-aware detection mode"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Support agent detection in CI/headless environments

Partio's agent detection currently relies on process detection patterns designed for interactive terminal sessions. When agents like Claude Code or Copilot run in CI environments (e.g., GitHub Actions with Copilot Coding Agent), the session directory layout and process tree may differ, causing detection to silently fail.

## Desired behavior

- Detect agent sessions in CI environments where the agent runs as a subprocess or background service
- Handle CI-specific session paths (e.g., `/home/runner/.copilot/session-state/`)
- Check for CI environment variables (`GITHUB_ACTIONS`, `CI`) to adjust detection strategy
- Support a `PARTIO_CI=true` env var or `--ci` flag on `partio enable` to opt into CI-aware mode

## Implementation notes

- The detector interface is in `internal/agent/detector.go`
- Claude Code detection is in `internal/agent/claude/`
- Session discovery in `find_session_dir.go` walks up from repo root — in CI the session dir may be in a completely different tree
- Add CI-specific path candidates to the session directory search
- Follow the existing pattern of error resilience: if CI detection fails, fall back to standard detection
