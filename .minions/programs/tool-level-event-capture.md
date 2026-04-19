---
id: tool-level-event-capture
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes a list of tool invocations extracted from the agent session transcript
  - Each tool event records at minimum the tool name and a timestamp or sequence index
  - Tool events are stored as part of checkpoint metadata without breaking existing checkpoint format
  - Sessions with no tool calls produce an empty tool events list (not an error)
pr_labels:
  - minion
---

# Capture tool-level events in checkpoint metadata

## Problem

Checkpoints currently capture session-level context (prompt, transcript summary, attribution) but not individual tool invocations within a session. Knowing which tools an agent used (file reads, writes, shell commands, searches) during a commit provides valuable audit and review context — reviewers can see not just what changed, but how the agent arrived at the change.

## Solution

Extend the JSONL parser to extract tool use events from Claude Code transcripts and include them as structured metadata in checkpoints.

### Implementation hints

- Claude Code JSONL entries with `"type": "tool_use"` or `"type": "tool_result"` contain tool invocation data
- Extract tool name, input summary (truncated), and sequence position
- Store as a `tool_events` array in checkpoint metadata alongside existing fields
- Keep the extraction lightweight — store tool names and counts, not full input/output payloads
- The checkpoint storage in `internal/checkpoint/` writes metadata as a JSON blob; extend the schema

## Source

Inspired by entireio/cli#860 (support preTool and postTool hooks) and entireio/cli#938 (Factory Droid mission mode hook failures).
