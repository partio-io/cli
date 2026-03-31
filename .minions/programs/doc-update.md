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

When a code PR merges on the cli repo, update the documentation to reflect the changes.

## Steps

1. **Identify the source PR.** The PR reference is passed as context. Read the PR title, body, and diff:
   ```bash
   gh pr view <number> --repo <repo> --json title,body,files
   gh pr diff <number> --repo <repo>
   ```

2. **Understand what changed.** Analyze the diff to determine what user-facing behavior changed — new commands, changed flags, new config options, etc.

3. **Read existing docs.** Check `docs/` for relevant pages that need updating. Look at the CLAUDE.md in the docs repo for structure guidance.

4. **Update the docs.** Edit the relevant MDX files to reflect the changes. Keep the existing style and structure.

5. **Verify.** Ensure MDX frontmatter is valid and no broken references.

6. **Create a PR** on the docs repo:
   ```bash
   git checkout -b minion/doc-update-<source-pr-number>
   git add -A
   git commit -m "docs: update for <repo>#<number>"
   git push -u origin minion/doc-update-<source-pr-number>
   gh pr create --repo <docs-repo> --title "[docs] Update for <repo>#<number>" --label minion --label documentation
   ```

## Context

- `docs/CLAUDE.md`
