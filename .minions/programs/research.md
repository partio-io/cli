---
id: research
target_repos:
  - cli
---

# Research minion (skeleton)

Tracer-bullet stub for the research minion path. Slice 1 of the
multi-slice rollout described in
`partio-minions/docs/2026-05-05-research-minion/issues/01-workflow-skeleton.md`.

The single agent below posts a one-line acknowledgement comment on
the parent issue to confirm that label → workflow → minions runtime
→ `gh` side effect all wire up correctly. The real researcher,
persona, prd-writer, slicer, and publisher agents land in slices 2
through 4.

## Agents

### stub

```capabilities
tools: Bash
max_turns: 5
```

Post a one-line "research started" comment on the parent issue, then
exit cleanly without modifying the worktree.

The parent issue number is available in the env var
`$MINION_ISSUE_NUMBER`. The workflow run id is in `$GITHUB_RUN_ID`.

Run, in this exact order:

1. Compute a 7-character run identifier from the workflow run id:

   ```bash
   RUN_SHA7=$(printf '%s' "$GITHUB_RUN_ID" | sha1sum | head -c 7)
   ```

2. Post the comment:

   ```bash
   gh issue comment "$MINION_ISSUE_NUMBER" \
     --repo partio-io/cli \
     --body "research started — run-id \`$RUN_SHA7\`"
   ```

3. Exit. Do not write any files. Do not modify the worktree. Do not
   call any other tools.
