---
id: ingest
target_repos:
  - cli
---

# Ingest features from a single source

Analyze content from a URL and generate feature proposals.

## Steps

1. **Fetch the content.** Use `curl` to download the URL provided as context.

2. **Read the ingest prompt** from `.minions/ingest-prompt.md`.

3. **Analyze the content** using the ingest prompt guidelines. Extract feature ideas that are relevant to this project.

4. **For each relevant feature:**
   - Generate a kebab-case ID
   - Check for duplicates: `gh issue list --repo <this-repo> --label minion-proposal --search "<id>" --limit 1`
   - Write a program file to `.minions/programs/<id>.md`
   - Create a GitHub issue linking to it with `<!-- program: .minions/programs/<id>.md -->`

5. **Commit and push** the new program files:
   ```bash
   git add .minions/programs/
   git commit -m "chore: add minion proposals from ingest"
   git push
   ```

6. **Print summary.**

## Context

- `.minions/ingest-prompt.md`
- `.minions/project.yaml`
