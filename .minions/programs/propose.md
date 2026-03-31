---
id: propose
target_repos:
  - cli
---

# Propose features from monitored sources

Scan monitored sources for new content and create feature proposals.

## Steps

1. **Read sources config** from `.minions/sources.yaml` in this repo. It lists changelog URLs, GitHub issue/PR repos, and the `last_version` cursor for each.

2. **Read the ingest prompt** from `.minions/ingest-prompt.md` — it describes the project and how to extract features.

3. **For each source**, fetch new content since `last_version`:
   - **changelog** sources: fetch the URL, find version headers newer than `last_version`
   - **issues** sources: run `gh issue list --repo <repo> --json number,title,body --limit 50`, filter items with number > last_version
   - **pulls** sources: run `gh pr list --repo <repo> --json number,title,body --limit 50`, filter items with number > last_version

4. **For each source with new content**, use the ingest prompt to analyze what's relevant to this project. For each relevant feature idea:
   - Generate a kebab-case ID
   - Check if a proposal already exists: `gh issue list --repo <this-repo> --label minion-proposal --search "<feature-id>" --limit 1`
   - If not, write a program file to `.minions/programs/<id>.md` with frontmatter (id, target_repos, acceptance_criteria, pr_labels) and description
   - Create a GitHub issue: `gh issue create --repo <this-repo> --label minion-proposal --title "<title>" --body "<description + link to program file + <!-- program: .minions/programs/<id>.md --> marker>"`

5. **Update `last_version`** in `.minions/sources.yaml` for each processed source (latest version string for changelogs, highest item number for issues/pulls).

6. **Commit and push** all new program files and the updated sources.yaml:
   ```bash
   git add .minions/programs/ .minions/sources.yaml
   git commit -m "chore: add minion proposals"
   git push
   ```

7. **Print summary** of what was created.

## Context

- `.minions/sources.yaml`
- `.minions/ingest-prompt.md`
- `.minions/project.yaml`
