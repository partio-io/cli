---
id: tag-thinking-blocks-in-transcripts
target_repos:
  - cli
acceptance_criteria:
  - JSONL parser identifies extended thinking / reasoning blocks in Claude Code transcripts
  - Thinking blocks are tagged with a structured marker in checkpoint transcript data
  - Tagged thinking blocks are distinguishable from regular assistant messages and tool calls
  - Existing transcripts without thinking blocks are handled gracefully (no errors)
pr_labels:
  - minion
---

# Tag reasoning/thinking blocks in checkpoint transcripts

## Problem

When browsing checkpoint transcripts, all content is treated uniformly — agent reasoning (extended thinking blocks) is mixed in with regular assistant messages, tool calls, and user prompts. This makes it harder to understand the agent's decision-making process vs. its actions.

## Solution

Extend the JSONL parser in `internal/agent/claude/` to detect and tag extended thinking blocks in Claude Code transcripts. When writing checkpoint transcript data, annotate thinking blocks with a structured marker (e.g., a `"type": "thinking"` field) so downstream consumers (CLI display, web UI) can render them distinctly.

### Implementation hints

- Claude Code JSONL transcripts contain thinking blocks as separate entries or nested within assistant messages
- The parser in `internal/agent/claude/parse_jsonl.go` already extracts prompt and context; extend it to recognize thinking content
- Add a field to the checkpoint transcript metadata to carry the thinking tag
- This is a non-breaking addition — existing checkpoints without thinking tags remain valid

## Source

Inspired by entireio/cli#973 (Tag "thinking" lines for compact transcript).
