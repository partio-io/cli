---
id: checks-fix
target_repos:
  - cli
acceptance_criteria:
  - "Every change in the exported patch addresses a specific finding in $MINION_AUDIT_DIR/verdict.json; no unrelated changes"
  - "No test file is modified unless its path is listed in $MINION_AUDIT_DIR/added-tests.txt"
  - "make lint and make test pass in the patched tree before the patch is exported"
  - "The session runs no git write commands: no commit, no push, no branches, no PRs, no comments"
  - "A patch is exported only when at least one finding was safely fixable and checks pass; otherwise nothing is exported"
  - "The skip marker file is written as the session's final act on every path"
---

# Repair the lint and test failures a check found

The lint and test check just failed on a pull request of the cli repo.
You get exactly one attempt to repair what it found: fix the findings
in your worktree, prove the tree passes the repo's own checks, and
export the result as a patch. You do not commit or push — the workflow
applies your patch, re-runs the checks, and pushes the fix commit
deterministically. The next run judges the result, not you. There is no
second attempt inside this session, so a fix you cannot verify is worse
than no fix at all.

The PR — its reference (`org/repo#number`), title, body, and full diff
— is provided in the **## Pull Request** section of this prompt. The
findings are in `$MINION_AUDIT_DIR/verdict.json`: a `findings` list
where each entry has a `location` (path:line, or the package that
failed) and `reasoning` carrying what the command reported. The
reasoning is complete — you do not need to re-run lint or the tests to
learn what failed, only to prove your repair.

## Agents

### checks-fix

```capabilities
max_turns: 50
skip_marker: .minion-fix-done
```

Work through the repair in this order:

1. **Read the findings.** Parse `$MINION_AUDIT_DIR/verdict.json`. If it
   is missing, malformed, or its `findings` list is empty, write the
   skip marker (step 7) and end without touching the repo.

2. **Read your boundary.** `$MINION_AUDIT_DIR/added-tests.txt` lists,
   one path per line, the test files this pull request added. Those are
   the only test files you may modify. If the file is missing, no test
   file is in bounds — treat the list as empty and repair
   implementation code only. An empty list is a real answer, not an
   error.

3. **Move to the PR head — your worktree is based on origin's default
   branch, not the PR.** Resolve the head and detach onto it before
   editing anything:

   ```
   gh pr view <pr-ref> --json headRefName,headRefOid
   git fetch origin <head-ref-name>
   git checkout --detach <head-ref-oid>
   ```

4. **Repair what the findings name.** For each finding, apply the
   smallest change that makes the reported failure go away. Rules:

   - Implementation code is yours to change, bounded by the findings.
   - A test file is yours to change only if step 2 listed it. A test
     that predates this pull request is out of bounds, whatever it
     asserts and however wrong it looks. Leave that finding unfixed: it
     survives, the check stays red, and a person decides. This is the
     one rule that separates a repair from a bot that reaches green by
     deleting an assertion.
   - Repair the implementation first. Change a listed test only when
     the failure shows the test itself is wrong — it asserts a
     behaviour nobody agreed to, or it was written against an interface
     that changed in the same pull request. Never weaken a test to make
     it pass: do not delete an assertion, do not skip a case, and do
     not relax a comparison, in a listed test or any other.
   - Do not guess. If a repair needs a decision the finding does not
     settle — which of two behaviours is correct, whether a linter rule
     should be suppressed — leave that finding unfixed. A surviving
     finding is the correct outcome for a fix that needs a human.
   - No opportunistic edits. Style, naming, or refactors beyond the
     findings are out of scope, however tempting.
   - Keep each repair coherent: a signature you change is a signature
     every caller must follow — the checks below must pass.

5. **Prove the tree with the repo's own checks.** Run `make lint` and
   `make test` from the worktree root. Both must pass. If they fail
   because of your edits, repair your edits and rerun. If you cannot
   get both green, restore the tree (`git checkout -- .` and delete
   files you added) and skip step 6 — exporting nothing is the correct
   failed-round outcome.

6. **Export the patch — only when checks pass and something was
   fixed.**

   ```
   git add -A
   git diff --cached --binary > "$MINION_AUDIT_DIR/fix.patch"
   ```

   Also write one plain-text line describing the repair to
   `$MINION_AUDIT_DIR/fix-summary.txt` — it becomes the fix commit's
   body.

7. **Write the skip marker — always your final act.** Create the empty
   file `.minion-fix-done` at the worktree root. This tells the runtime
   the session's output is the exported patch, not a branch: without
   it, leftover edits would be auto-committed to a stray `minion/*`
   pull request.

Never run `git commit`, `git push`, `gh pr create`, or post comments,
on any path. Write outside the worktree only under `$MINION_AUDIT_DIR`.
