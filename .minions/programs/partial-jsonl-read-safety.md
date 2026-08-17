---
id: partial-jsonl-read-safety
target_repos:
  - cli
acceptance_criteria:
  - JSONL parser skips any trailing line that is not valid JSON (incomplete write) without returning an error
  - A warning is logged when an incomplete trailing line is detected, including the byte offset
  - The checkpoint is still created from the successfully-parsed lines
  - A unit test exercises a JSONL file with a truncated final line (simulate mid-write) and confirms a valid checkpoint is produced with a warning
  - Existing tests for complete JSONL files continue to pass
pr_labels:
  - minion
---

# JSONL parser: handle partial trailing lines from in-progress Claude sessions

## Problem

Partio's post-commit hook parses the Claude Code JSONL session file immediately
after a commit. A commit can happen mid-session — the user (or an agent) runs
`git commit` while Claude Code is still active and still appending to the JSONL
file. In that case the JSONL file may end with a partially-written line: a
valid JSON object that Claude began writing but had not yet flushed with a
newline terminator.

The current `parse_jsonl.go` reads the file line by line. The last line of an
in-progress session may be:

1. A complete JSON object missing its trailing newline (no flush yet).
2. A truncated JSON object (flush happened mid-write).

In case 1, Go's `bufio.Scanner` with default line splitting will silently drop
the line (no newline = scanner considers it an incomplete final token). In case
2, `json.Unmarshal` returns an error that propagates and can prevent checkpoint
creation entirely, even though all prior messages parsed successfully.

The result: a mid-session commit either silently loses the last conversation
turn from the checkpoint, or fails checkpoint creation altogether.

## Desired behavior

The JSONL parser should treat a malformed trailing line as a soft warning, not
a hard error:

1. Attempt to parse each line as a JSON message.
2. If the very last line of the file fails to parse (JSON syntax error or
   empty), log a warning at `WARN` level that includes the file path and byte
   offset of the incomplete line.
3. Continue with all successfully-parsed messages and create the checkpoint
   normally.

This matches the "error resilience in hooks" principle in CLAUDE.md — hooks
should not block git operations on non-critical failures. A partially-written
last JSONL line is non-critical: the session is still active and future commits
will capture the remaining turns.

## Context hints

- `internal/agent/claude/parse_jsonl.go` — JSONL parsing logic
- `internal/agent/claude/` — session discovery and file reading

## Source

Inspired by entireio/cli#2001: "Preserve user prompts when message events
arrive in sequence" — a bug where event-stream ordering caused prompts to be
dropped. The root pattern — partial/out-of-order writes producing silent data
loss — applies directly to Partio's line-by-line JSONL reading of an actively
written session file.
