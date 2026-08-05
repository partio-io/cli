---
id: e2e-audit
target_repos:
  - cli
acceptance_criteria:
  - "The verdict file exists at $MINION_AUDIT_DIR/verdict.json, is valid JSON, and its status is pass whenever the audit ran to completion"
  - "Each distinct e2e-coverage gap has exactly one open minion-proposal issue; pre-existing proposals are reused, never duplicated"
  - "Filed proposals start with the Minion audit — body prefix and carry acceptance criteria plus a program-file reference a fresh session can build from"
  - "The PR itself is untouched: no comments, no edits, no commits to its branch"
---

# Audit a PR for missing end-to-end coverage

A pull request on the cli repo is under audit. Judge whether its
change warrants an end-to-end test that does not exist yet. Findings
never block the PR: each one becomes a `minion-proposal` issue so the
test gets built through the normal minion flow, and your verdict is
`pass` whenever you ran to completion.

The PR — its reference (`org/repo#number`), title, body, and full
diff — is provided in the **## Pull Request** section of this prompt.

The canonical exemplar: a fix whose headline acceptance criterion —
"two rapid commits in one live session both get trailers" — was
verified only via unit tests of the two halves of the mechanism, with
nothing driving both hooks through the real two-commit flow. Each
half can pass while the whole still fails; that gap is what this
audit names.

## Agents

### e2e-audit

```capabilities
max_turns: 50
```

Work through the audit in this order:

1. **Name the headline behavior.** From the PR's title, body, and
   diff, state what the change promises end to end. Ask whether that
   promise spans multiple components run in sequence — git hooks
   firing on real commits, session lifecycle across processes, push
   and fetch against a real remote — or is fully contained in one
   unit.

2. **Check what the PR's tests actually drive.** Verify against the
   PR's head state, not just the diff. Your working directory is a
   cli checkout, useful for structure and grep, but it may lag the
   PR. For anything decisive — which tests exist, what flow they
   exercise — read files at the PR head via `gh` (e.g.
   `gh api repos/<org>/<repo>/contents/<path>?ref=<head-sha>` with
   header `Accept: application/vnd.github.raw+json`). A gap must
   hold against the head state.

3. **Judge conservatively.** A finding is a headline behavior that
   only an end-to-end run can prove, where the PR's tests cover the
   pieces but nothing drives the whole flow. Comment, docs, or other
   cosmetic changes are never findings. Neither is a change whose
   promise is genuinely unit-scoped and unit-tested. When in doubt,
   it is not a finding.

4. **File one proposal per distinct gap.** For each finding:

   - Generate a kebab-case id for the missing test, prefixed `e2e-`
     and named after the flow under test (e.g.
     `e2e-two-commit-trailer-flow`).
   - Check for an existing proposal first:
     `gh issue list --repo <org>/<repo> --label minion-proposal --state open --search "<id>" --limit 5`,
     and also scan the open proposals' titles for the same gap under
     a different id. If one already covers it, reuse it — record the
     existing issue in your verdict reasoning and file nothing. One
     repair applies: when the covering proposal's `<!-- program: ... -->`
     file does not exist on `origin/main`, write that file (step
     above) and push it (step 5) so the proposal stays buildable —
     still without filing a new issue.
   - Write a program file to `.minions/programs/<id>.md` with
     frontmatter (`id`, `target_repos`, `acceptance_criteria`,
     `pr_labels`) and a body that tells a fresh minion session
     exactly what end-to-end test to build: the flow to drive, where
     the test lives, and what it asserts. The acceptance criteria
     must be concrete enough to build from the issue alone.
   - Create the issue:
     `gh issue create --repo <org>/<repo> --label minion-proposal --title "e2e: <flow under test>" --body <body>`.
     The body starts with `Minion audit — ` followed by the PR
     reference that triggered it, then the gap and why unit tests
     cannot close it, the acceptance criteria, a link to the program
     file, and the marker
     `<!-- program: .minions/programs/<id>.md -->`.

5. **Commit and push the program files** to `main`, the way the
   propose program does — only if step 4 wrote any:

   ```bash
   git add .minions/programs/
   git commit -m "chore: add minion proposals"
   git push origin HEAD:main
   ```

   Your worktree sits on a session branch, so a bare `git push` will
   not reach `main` — push `HEAD:main` explicitly, and confirm it
   succeeded. If it is rejected because your base is behind, run
   `git fetch origin main && git rebase origin/main` and push once
   more. Do not leave the commit unpushed for the runtime to turn
   into a PR of its own: the proposal is unusable until its program
   file is on `main`.

6. **Write the verdict — your final act.** Create the directory and
   write `$MINION_AUDIT_DIR/verdict.json`:

   ```json
   {
     "status": "pass",
     "findings": [
       {
         "location": "internal/hooks/post_commit.go:88",
         "reasoning": "trailer stamping is only proven by unit tests of each hook; no test drives two rapid commits through both hooks — filed e2e-two-commit-trailer-flow"
       }
     ]
   }
   ```

   `status` is `pass` whenever you ran to completion — findings do
   not make it `fail`. List every gap (filed or deduplicated) with a
   non-empty `location` (path:line anchoring the flow in the head
   state) and `reasoning` naming why only an e2e test proves it plus
   the proposal id it maps to. No gaps means `"findings": []`. A
   missing or malformed file is treated as a crashed audit and fails
   the workflow, so write it even when there are no findings.

Do **not** comment on the PR, edit files in its branch, or run any
`git` write command other than the program-file commit in step 5.
Read-only `git`/`gh` is otherwise fine. The verdict file lives
outside your checkout on purpose — leave the working tree exactly as
you found it.
