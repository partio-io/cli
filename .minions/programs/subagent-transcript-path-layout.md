---
id: subagent-transcript-path-layout
target_repos:
  - cli
acceptance_criteria:
  - Session parsing tries the nested path `<transcriptDir>/<sessionID>/subagents/agent-<agentID>.jsonl` first
  - Falls back to the old flat sibling path `<transcriptDir>/agent-<agentID>.jsonl` for backward compatibility with older Claude Code versions
  - Tests cover both path layouts using `t.TempDir()` to construct mock transcript directories
pr_labels:
  - minion
  - enhancement
---

# Update subagent transcript discovery to Claude Code's nested path layout

## Problem

Claude Code 2.1.221+ changed where it stores subagent JSONL transcripts. Previously, subagent transcripts were written as a sibling of the main transcript:

```
<transcriptDir>/agent-<agentID>.jsonl
```

Since Claude Code 2.1.221 they are nested one level deeper under a per-session subdirectory:

```
<transcriptDir>/<sessionID>/subagents/agent-<agentID>.jsonl
<transcriptDir>/<sessionID>/subagents/agent-<agentID>.meta.json
```

If Partio's session parsing in `internal/agent/claude/` resolves the old flat path, it will always get a path that does not exist on current Claude Code versions, silently dropping any data that would have been read from the subagent transcript.

## What to implement

In `internal/agent/claude/` (likely `parse_jsonl.go` or any code that constructs paths into the Claude transcript directory), update the subagent transcript path resolution to:

1. Try the new nested path: `<transcriptDir>/<sessionID>/subagents/agent-<agentID>.jsonl`
2. If that file does not exist, fall back to the old flat path: `<transcriptDir>/agent-<agentID>.jsonl`

This two-step fallback ensures Partio works with both current and older Claude Code installations.

## Source

entireio/cli#1935 — fix(subagent): resolve subagent transcripts in Claude Code's current layout

## Context hints

- `cli/internal/agent/claude/` — session discovery and JSONL parsing
- `cli/internal/agent/claude/parse_jsonl.go` — primary location for JSONL path construction

<!-- program: .minions/programs/subagent-transcript-path-layout.md -->
