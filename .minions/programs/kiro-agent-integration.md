---
id: kiro-agent-integration
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - A new internal/agent/kiro/ package implements the Detector and SessionParser interfaces
  - Kiro process detection works on Linux and macOS via the existing process-scanning approach
  - Session directory discovery finds Kiro session files following the walk-up-from-repo-root pattern
  - Transcript parsing extracts session ID, context, and prompt from Kiro's session format
  - The detector self-registers in the agent registry via init() following the existing codex/gemini pattern
  - DetectActive() correctly discovers running Kiro instances
  - partio enable --agent kiro configures hooks for Kiro sessions
  - Table-driven tests cover detection, session discovery, and transcript parsing
---

# Add Kiro AI agent integration

Implement a Kiro (Amazon's AI coding agent) detector and session parser, following the
existing pattern established by the Claude Code, Codex, and Gemini agent implementations.

## Context

Kiro is Amazon's AI coding agent that is gaining adoption. The Entire CLI project has
received a feature request for Kiro support (entireio/cli#1054), indicating community
demand for this integration.

Partio's agent detection architecture is already designed for pluggable agents via the
`Detector` interface in `internal/agent/detector.go`. Adding Kiro support follows the
same pattern as the existing Claude Code implementation in `internal/agent/claude/`.

## Implementation Notes

- Create `internal/agent/kiro/` package with detector, session discovery, and transcript
  parsing files following the one-file-per-concern convention
- Implement the `Detector` interface (DetectActive, ParseSession)
- Register the detector in the agent registry via `init()`
- Research Kiro's session file format and location to implement accurate discovery
- Add the agent name to the `PARTIO_AGENT` environment variable documentation
