---
id: doc-update
target_repos:
  - docs
acceptance_criteria:
  - "Documentation accurately reflects the changes in the source PR"
  - "MDX frontmatter is valid"
  - "No broken links or references"
pr_labels:
  - minion
  - documentation
---

# Update documentation for a merged PR

A pull request merged on the cli repo. Update the documentation in the `docs`
repo to reflect any user-facing changes it introduced.

The source PR — its reference (`org/repo#number`), title, body, and full diff —
is provided in the **## Pull Request** section of this prompt. Everything you
need is already there; you do not need to fetch it.

## Steps

1. **Understand what changed.** From the PR diff, determine what *user-facing*
   behavior changed — new commands, changed flags, new config options, changed
   output. Ignore internal-only changes (refactors, tests, CI).

2. **If nothing user-facing changed, stop and make no edits.** A docs PR is only
   opened when you change files, so finishing with no edits is the correct
   "no update needed" outcome.

3. **Read the existing docs.** Your working directory is the `docs` repo root.
   Read its `CLAUDE.md` for structure guidance and find the pages affected.

4. **Update the docs.** Edit the relevant MDX files to reflect the change. Match
   the existing style and structure. Keep edits scoped to this PR.

5. **Verify.** Ensure MDX frontmatter is valid and there are no broken links or
   references.

Do **not** run `git`/`gh` or create a branch or PR yourself. The runtime commits
your edits and opens the docs PR (labels `minion`, `documentation`) for you.

## Context

- `docs/CLAUDE.md`
