---
id: propose
target_repos:
  - cli
---

# Propose features from monitored sources

Scan monitored sources for new content and create feature proposals.

This program runs one agent, `proposer`, as a single one-shot Claude
session. The agent reads the monitored-source cursors, fetches what is
new since each cursor, applies the ingest prompt to decide what is
relevant to this project, and files one proposal per relevant idea as a
GitHub issue. The issue is the whole proposal: a build reads the issue
and nothing else. It then advances the cursors and commits the result.

The repository for all `gh` calls against this project is
`partio-io/cli`.

## Context

- `.minions/sources.yaml`
- `.minions/ingest-prompt.md`
- `.minions/premise-verifier.md`
- `.minions/project.yaml`

## Agents

### proposer

```capabilities
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - WebFetch
max_turns: 80
```

You are the proposer. Scan the monitored sources for new content and
create feature proposals. Work through the steps below in order.

1. **Read sources config** from `.minions/sources.yaml` in this repo. It lists changelog URLs, GitHub issue/PR repos, and the `last_version` cursor for each.

2. **Read the ingest prompt** from `.minions/ingest-prompt.md` — it describes the project and how to extract features. Read `.minions/premise-verifier.md` as well: it describes how to check a premise against the checked-out tree, and step 4 applies it before anything is filed.

3. **For each source**, fetch new content since `last_version`:
   - **changelog** sources: fetch the URL, find version headers newer than `last_version`
   - **issues** sources: run `gh issue list --repo <repo> --json number,title,body --limit 50`, filter items with number > last_version
   - **pulls** sources: run `gh pr list --repo <repo> --json number,title,body --limit 50`, filter items with number > last_version

4. **For each source with new content**, use the ingest prompt to analyze what's relevant to this project. For each relevant feature idea:
   - Generate a kebab-case ID
   - **Check for an earlier proposal of the same idea.** Run the duplicate search in your working directory, with the idea's source item and its title: `go run ./cmd/minion-review dupes --source "<owner/repo#N>" --title "<title>"`

     Pass the source item as `owner/repo#N` (for example `entireio/cli#2075`), and repeat `--source` when the idea comes from more than one item. A changelog entry has no item number: leave `--source` out, and the title carries the search. The command prints a JSON list of candidates: `minion-proposal` issues, open and closed, that cite the same source item (`"match": "source"`) or have a strongly overlapping title (`"match": "title"`). Each candidate carries its `number`, `state`, `state_reason` and `labels`. The list is the whole check: do not search the issues in another way.

     A shared source item or title does not make the same idea. Read each candidate with `gh issue view <n>`, and decide whether it asks for the same change. A closed candidate counts as much as an open one when the review closed it with a verdict. A candidate closed as `completed` counts only when this tree shows the work: a build can close an issue while its pull request never merges (#31 is such a case), so the closed state alone proves nothing.
   - **If a candidate is the same idea, file no issue for it.** Keep the idea and the number of the issue it matched for the summary. A duplicate is not a rejection: the matched issue is its record, so it gets no entry in `.minions/rejections.md`. Go on to the next idea.
   - Build the premise section from the idea's `premise` field. Use this format exactly:

     ```markdown
     ## Premise

     <!-- partio:premise:v1 -->

     - <one factual claim this proposal depends on> [evidence: `<path, symbol or command that settles it>`]
     ```

     One claim per line, and every claim names the path, symbol or command that settles it. A claim you cannot attach evidence to is not filed with an empty field: find the evidence, or drop the idea.
   - **Verify the premise before you file anything.** This repository is already checked out in your working directory. Apply `.minions/premise-verifier.md` to the premise section you just built, against that tree. Gather the evidence each claim names — read the file, find the symbol, run the command — and decide each claim from what you gathered. The ideas come from a sibling product, so a claim that is true of that product is not thereby true here. Verify it here.
   - **If the block does not hold, drop the idea.** Create no issue for it. A `fails` or `unresolved` verdict is not a pass. Keep the claim, the evidence and the verdict for the summary.
   - **Record every rejection in `.minions/rejections.md`.** Append one entry per dropped idea. Never rewrite an entry that is already there. Use this format exactly:

     ```markdown
     ## <what the idea was, in one line>

     <!-- partio:rejection:v1 -->

     - source: `<source name and the item within it>`
     - reason: `premise-failed`
     - claim: <the claim that failed> [evidence: `<the path, symbol or command it named>`]
     - verdict: `fails`
     - found: <what the evidence actually showed>
     ```

     The cursor moves past this idea in step 5, so the source item is never offered again. This entry is the only record that the idea was seen at all. Write it for a reader: someone who wants to know whether the bar is set right reads this file and nothing else.
   - **Record a skipped source item in the same log, under its own reason.** An item the ingest prompt found irrelevant was never an idea, and no claim was ever checked against this repository. Its entry carries no claim, no verdict and no finding — only why it was not about this project:

     ```markdown
     ## <what the source item proposed, in one line>

     <!-- partio:rejection:v1 -->

     - source: `<source name and the item within it>`
     - reason: `irrelevant`
     - note: <why this project has no use for it>
     ```

     Keep the two kinds apart. A dropped idea was checked and this repository contradicted it; a skipped item was never about this project. A reader who cannot tell them apart cannot tell a bar set too high from a source that has gone quiet, which is the one question this log answers.
   - If the block holds, create a GitHub issue: `gh issue create --repo <this-repo> --label minion-proposal --label minion-reviewed --title "<title>" --body "<description + what to build + acceptance criteria + premise section + gathered evidence + source line + proposal id line>"`

     The issue is the whole proposal. A build reads the issue and nothing else, so the body carries everything a fresh session needs to build from: what to build, in full, and the acceptance criteria as a `- [ ]` checklist under `## Acceptance Criteria`. Write no file for the proposal. `.minions/programs/` holds the programs the workflows run, and a proposal is not one of them.

     The `minion-reviewed` label tells the review sweep that this issue already passed today's bar: the duplicate check and the premise check above. The sweep does not review it again.

     The gathered evidence is the verifier's output: each claim, the evidence it named, the verdict, and the excerpt that produced it. A proposal that passed the check carries the evidence that passed it.

     Cite the source item on a line of its own: `Source: <owner/repo#N>`, one reference for each item the idea came from. For a changelog entry, name the repo and the version (`Source: <owner/repo> v<version>`). The duplicate check above finds an earlier proposal by this line, so an issue without it can be filed a second time on a later run.

     End the body with the id on a line of its own: `Proposal id: <id>`.

5. **Update `last_version`** in `.minions/sources.yaml` for each processed source (latest version string for changelogs, highest item number for issues/pulls).

6. **Commit and push** the updated sources.yaml and the rejection log, in one commit:
   ```bash
   git add .minions/sources.yaml .minions/rejections.md
   git commit -m "chore: advance minion sources and record rejections"
   git push
   ```

   The log and the cursor travel together. The cursor advances past a rejected idea whether or not the log survives, so a commit that carries one without the other loses the idea silently. If this run rejected nothing, the log is unchanged and `git add` stages nothing from it; commit the cursor as usual.

7. **Print summary** of what the run did. Report these four outcomes separately, with a count for each:
   - **filed** — ideas whose premise held. One issue each.
   - **dropped** — ideas the ingest prompt produced whose premise did not hold. Name the claim, the evidence it named, and the verdict for each one.
   - **skipped** — source items the ingest prompt found irrelevant to this project. These never became ideas.
   - **duplicate** — ideas an earlier proposal, open or closed, already holds. Name the idea and the issue it matched for each one, for example `#710`.

   A duplicate is not a dropped idea: nothing contradicted it, and the issue it matched is its record. A dropped idea and a skipped item are not the same event either. An idea was dropped because this repository contradicted it; an item was skipped because it was never about this project. Do not merge the two counts. The cursor advances past both, so this summary is the only place the difference survives the run.
