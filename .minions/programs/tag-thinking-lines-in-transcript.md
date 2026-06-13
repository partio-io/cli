---
id: tag-thinking-lines-in-transcript
target_repos:
  - cli
acceptance_criteria:
  - "jsonlEntry.ContentBlocks recognises blocks with type 'thinking' and does not mix them into regular text extraction"
  - "agent.SessionData gains a ThinkingContent field populated by ParseJSONL when thinking blocks are present"
  - "checkpoint.SessionFiles stores thinking content separately from the main transcript/context fields"
  - "A unit test in internal/agent/claude/ covers a JSONL fixture containing thinking blocks and asserts they are isolated from the regular transcript"
  - "Existing tests continue to pass with no regression on non-thinking transcripts"
pr_labels:
  - minion
---

# Tag 'thinking' blocks in JSONL transcript for compact storage

When parsing Claude Code JSONL transcripts, identify and tag content blocks with type `thinking` (Claude's extended thinking / reasoning tokens) separately from regular assistant text. Store a flag or separate field in the checkpoint's session files so that the full reasoning trace can be omitted from the compact stored transcript while still being accessible on demand.

Concretely: extend `jsonlEntry` and `contentBlock` in `internal/agent/claude/jsonl.go` to recognise `type: "thinking"`, and in `ParseJSONL` accumulate thinking blocks into a separate slice on `agent.SessionData`. When writing checkpoint session files, store thinking content under a distinct key so it can be stripped for space-efficient storage without losing the data.

## Why

Claude's extended-thinking sessions produce very large JSONL files. Tagging thinking lines lets Partio store compact transcripts by default without permanently discarding reasoning context, keeping checkpoint branch size manageable.

## User relevance

Users working with Claude's extended thinking mode currently have their entire reasoning trace stored verbatim in every checkpoint, bloating the orphan branch. Separating thinking content gives them smaller checkpoints and the option to view or discard the reasoning trace.

## Context hints

- `internal/agent/claude/jsonl.go`
- `internal/agent/claude/parse_jsonl.go`
- `internal/agent/types.go`
- `internal/checkpoint/write.go`
