---
id: readme-update
target_repos:
  - cli
acceptance_criteria:
  - "README accurately reflects current CLI capabilities"
  - "No broken links"
pr_labels:
  - minion
---

# Update README for a merged PR

When a PR merges, update the repo's README to reflect any user-facing changes.

## Steps

1. **Read the source PR** diff and description:
   ```bash
   gh pr view <number> --repo <repo> --json title,body,files
   gh pr diff <number> --repo <repo>
   ```

2. **Read the current README.md** in the repo.

3. **Determine if the README needs updating.** Only update if:
   - New commands or flags were added
   - Existing commands changed behavior
   - Install instructions changed
   - Architecture changed significantly

   If no README update is needed, create a file `.no-update-needed` with a brief explanation and stop.

4. **Update README.md** with the relevant changes. Keep existing style.

5. **Create a PR**:
   ```bash
   git checkout -b minion/readme-update-<source-pr-number>
   git add README.md
   git commit -m "docs: update README for <repo>#<number>"
   git push -u origin minion/readme-update-<source-pr-number>
   gh pr create --repo <repo> --title "[readme] Update for #<number>" --label minion
   ```

## Context

- `README.md`
