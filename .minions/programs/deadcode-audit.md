---
id: deadcode-audit
target_repos:
  - cli
acceptance_criteria:
  - "The verdict file exists at $MINION_AUDIT_DIR/verdict.json and is valid JSON"
  - "Verdict status is pass or fail; a fail carries at least one finding, each with location and reasoning"
  - "The working tree is untouched: no edits, no commits, no branches, no PRs, no comments"
---

# Audit a PR for semantically dead code

A pull request on the cli repo is under audit. Judge whether its diff
introduces or reveals *semantically dead code*, then record your
verdict. You change nothing: the workflow turns your verdict into the
check outcome and the PR comment.

The PR — its reference (`org/repo#number`), title, body, and full
diff — is provided in the **## Pull Request** section of this prompt.

Semantically dead code is code that is still invoked but provably
inert. The canonical exemplar: a callee whose first branch
short-circuits on a parameter that every remaining call site
hardcodes, leaving the rest of the block unreachable. No
deterministic tool flags this; your job is to reason about the diff
and catch it.

## Agents

### deadcode-audit

```capabilities
max_turns: 40
```

Work through the audit in this order:

1. **Read the diff.** For each hunk, ask what the change makes
   unreachable, uncalled, or constant-folded — in the changed code
   itself or in code it calls.

2. **Verify against the PR's head state, not just the diff.** Your
   working directory is a cli checkout, useful for structure and
   grep, but it may lag the PR. For anything decisive — call sites,
   callee bodies, who else passes this parameter — read files at the
   PR head via `gh` (e.g. `gh api repos/<org>/<repo>/contents/<path>?ref=<head-sha>`
   with header `Accept: application/vnd.github.raw+json`). A finding
   must hold against the head state.

3. **Judge conservatively.** Report only provable inertness — where
   you can name the branch, the parameter, and every call site that
   seals it. Style, duplication, and "probably unused" suspicions are
   not findings. When you cannot prove it, it is a pass.

4. **Write the verdict — your final act.** Create the directory and
   write `$MINION_AUDIT_DIR/verdict.json`:

   ```json
   {
     "status": "fail",
     "findings": [
       {
         "location": "internal/foo/bar.go:42",
         "reasoning": "callers X and Y both hardcode force=true, so the branch below the early return is unreachable"
       }
     ]
   }
   ```

   `status` is `pass` or `fail`. A pass has `"findings": []`. A fail
   has at least one finding; every finding needs a non-empty
   `location` (path:line as in the head state) and `reasoning` naming
   the exact chain that proves inertness. A missing or malformed file
   is treated as a failed audit, so write it even when everything
   passes.

Do **not** edit files, run `git` write commands, open PRs, or post
comments. Read-only `git`/`gh` is fine. The verdict file lives
outside your checkout on purpose — leave the working tree exactly as
you found it.
