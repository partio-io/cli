---
id: strip-ide-context-from-session-titles
target_repos:
  - cli
acceptance_criteria:
  - "Session JSONL parsing strips IDE-injected context blocks (e.g. <ide_opened_file>) from the user prompt before extracting session titles"
  - "Checkpoint metadata title and prompt fields contain clean user text without IDE context tags"
  - "Full transcript data is preserved unmodified — only derived fields (title, prompt summary) are cleaned"
  - "Stripping handles multiple IDE context tag formats gracefully"
  - "make test passes with new test cases covering IDE context stripping"
  - "make lint passes"
pr_labels:
  - minion
---

# Strip IDE-injected context tags from session titles and prompts

When Claude Code runs inside IDE extensions (VS Code, Cursor, JetBrains), the IDE prepends context blocks like `<ide_opened_file>...</ide_opened_file>` to the user's first prompt. Partio currently surfaces these injected blocks verbatim in checkpoint metadata titles and prompt fields, making session titles unreadable (e.g. a title that starts with XML tags instead of the user's actual request).

## What to implement

Add a sanitization step in the JSONL transcript parsing pipeline that strips known IDE-injected context tags from the user's prompt before using it to derive:
- The session/checkpoint title
- The `prompt` field in checkpoint metadata

The full transcript (`full.jsonl` / raw JSONL data) must remain unmodified — the stripping only applies to derived display fields.

### Known IDE context tag patterns to handle
- `<ide_opened_file>...</ide_opened_file>`
- Other `<ide_*>...</ide_*>` tags that IDEs may inject

### Where to look
- `internal/agent/claude/parse_jsonl.go` — JSONL parsing and field extraction
- `internal/checkpoint/` — checkpoint metadata construction

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
