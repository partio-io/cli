---
id: deadcode-audit-fix
target_repos:
  - cli
acceptance_criteria:
  - "Every change in the exported patch addresses a specific finding in $MINION_AUDIT_DIR/verdict.json; no unrelated changes"
  - "make lint and make test pass in the patched tree before the patch is exported"
  - "The session runs no git write commands: no commit, no push, no branches, no PRs, no comments"
  - "A patch is exported only when at least one finding was safely fixable and checks pass; otherwise nothing is exported"
  - "The skip marker file is written as the session's final act on every path"
---

# Fix the dead code a minion audit found

A dead-code audit just failed on a pull request of the cli repo. You
get exactly one attempt to repair what it found: fix the findings in
your worktree, prove the tree still passes the repo's own checks, and
export the result as a patch. You do not commit or push — the
workflow applies your patch, re-runs the checks, and pushes the fix
commit deterministically. A re-audit with a fresh session judges the
result — not you. There is no second attempt, so a fix you cannot
verify is worse than no fix at all.

The PR — its reference (`org/repo#number`), title, body, and full
diff — is provided in the **## Pull Request** section of this prompt.
The audit's findings are in `$MINION_AUDIT_DIR/verdict.json`: a
`findings` list where each entry has a `location` (path:line at the
PR head) and `reasoning` naming the chain that proves the code inert.

## Agents

### deadcode-audit-fix

```capabilities
max_turns: 50
skip_marker: .minion-fix-done
```

Work through the repair in this order:

1. **Read the findings.** Parse `$MINION_AUDIT_DIR/verdict.json`. If
   it is missing, malformed, or its `findings` list is empty, write
   the skip marker (step 6) and end without touching the repo.

2. **Move to the PR head — your worktree is based on origin's
   default branch, not the PR.** Resolve the head and detach onto
   it before editing anything:

   ```
   gh pr view <pr-ref> --json headRefName,headRefOid
   git fetch origin <head-ref-name>
   git checkout --detach <head-ref-oid>
   ```

3. **Fix only what the reasoning proves.** For each finding, apply
   the minimal repair that follows from its reasoning — usually
   deleting the unreachable branch, the constant parameter, or the
   never-called path, and simplifying the call sites the reasoning
   names. Rules:

   - Do not guess. If a repair requires a decision the finding does
     not settle — choosing between two plausible behaviors,
     resurrecting the dead path instead of removing it — leave that
     finding unfixed. A surviving finding is the correct outcome
     for a fix that needs a human.
   - An exported identifier is fair game. Every Go file in this
     module lives under `internal/` or `cmd/`, and Go forbids
     importing an `internal/` package from outside the module that
     declares it — so no package here is importable and a capital
     letter is not a public contract. Repair a proven-inert
     exported function, method, type, field, or signature exactly
     as you would an unexported one. What bounds you is the proof
     in the finding's reasoning, not the case of the first letter.
   - No opportunistic edits. Style, naming, or refactors beyond the
     findings are out of scope, however tempting.
   - Keep each repair coherent: removing a branch that orphans a
     variable means removing the variable too — the checks below
     must pass.

4. **Prove the tree with the repo's own checks.** Run `make lint`
   and `make test` from the worktree root. Both must pass. If they
   fail because of your edits, repair your edits and rerun. If you
   cannot get both green, restore the tree (`git checkout -- .` and
   delete files you added) and skip step 5 — exporting nothing is
   the correct failed-round outcome.

5. **Export the patch — only when checks pass and something was
   fixed.**

   ```
   git add -A
   git diff --cached --binary > "$MINION_AUDIT_DIR/fix.patch"
   ```

   Also write one plain-text line describing the repair to
   `$MINION_AUDIT_DIR/fix-summary.txt` — it becomes the fix
   commit's body.

6. **Write the skip marker — always your final act.** Create the
   empty file `.minion-fix-done` at the worktree root. This tells
   the runtime the session's output is the exported patch, not a
   branch: without it, leftover edits would be auto-committed to a
   stray `minion/*` pull request.

Never run `git commit`, `git push`, `gh pr create`, or post
comments, on any path. Write outside the worktree only under
`$MINION_AUDIT_DIR`.
