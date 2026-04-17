---
id: tool-level-event-capture
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint metadata includes a list of tool invocations (tool name, file path if applicable) extracted from the session transcript"
  - "Tool events are parsed from Claude Code JSONL session data during post-commit processing"
  - "partio status or checkpoint inspection commands can display tool event summaries"
  - "Tool event extraction is best-effort and does not block checkpoint creation on parse failures"
  - "Existing checkpoints without tool data continue to work without errors"
pr_labels:
  - minion
---

# Capture tool-level events in checkpoint metadata

## Summary

Enrich checkpoint metadata with tool-level event data extracted from agent session transcripts. This provides finer-grained insight into what the agent actually did during a commit — which files it read, which tools it invoked, and what edits it made — beyond the current prompt/response capture.

## Motivation

Current checkpoints capture the session transcript as a whole, but don't extract structured data about individual tool invocations. Tool-level events would enable:

- Better attribution: knowing exactly which files the agent touched and how
- Richer checkpoint browsing: users can see a timeline of tool calls alongside the diff
- Foundation for analytics: aggregating tool usage patterns across sessions
- More accurate "files touched" tracking for multi-file agent sessions

## Implementation Notes

- Extend the JSONL parser in `internal/agent/claude/` to extract tool use events (tool name, input summary, file paths)
- Store tool events as a structured field in checkpoint metadata (e.g., `tools_used` array)
- Keep extraction best-effort: log warnings on parse failures, never block checkpoint creation
- Consider preTool/postTool lifecycle hooks for future agent integrations that support them
- Tool events should be lightweight summaries, not full tool input/output (to keep checkpoint size reasonable)

## Source

Inspired by entireio/cli#860 — adds preTool/postTool hook support to capture tool-level events during agent sessions.
