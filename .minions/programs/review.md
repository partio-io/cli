---
id: review
target_repos:
  - cli
---

# Proposal review

Judge one proposal issue, and write the judgment as a verdict file. This
program writes no code and opens no pull request. A gate reads the
verdict after this session ends, and only the gate records it.

A proposal can be wrong in ways its author could not see when it was
filed. Its premise can be false today, the idea can be built already, or
it can describe a feature this project has no use for. Issue #30 reached
three builds on a false premise. A review catches that before a build
spends anything on it.

## Context

- `cli/.minions/premise-verifier.md`
- `cli/.minions/ingest-prompt.md`

## Agents

### reviewer

```capabilities
tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
max_turns: 80
```

You review one proposal issue against the tree in front of you. Your
verdict decides nothing by itself: the gate that runs after you records
it.

The repository the proposal is about is already checked out in your
working directory. Read it. Do not decide a claim from the issue text,
and do not decide it from memory of how similar projects work.

1. **Read the issue.** Its body is provided under the "Issue" section of
   your prompt. Find the source item it cites: an entireio/cli pull
   request, issue or changelog entry. Read that item with read-only `gh`
   calls, such as `gh pr view <n> --repo entireio/cli` or `gh issue view
   <n> --repo entireio/cli`. The source tells you what the idea was
   before anyone wrote it up for this project.

2. **Check the premise.** Apply `.minions/premise-verifier.md` to the
   issue against the checked-out tree. Follow it as written. When the
   body carries a `## Premise` section with the
   `<!-- partio:premise:v1 -->` marker, that block is what you verify.
   When it carries no marker, the claims come from the prose: the
   verifier's `## When there is no block` section describes the
   extraction, including the assumed fact that a requested behaviour or
   an acceptance criterion takes for granted about today's code or data.
   Record every claim, the evidence it named, its verdict and the excerpt
   that decided it.

3. **Judge the fit.** Apply the relevance rule of
   `.minions/ingest-prompt.md`: a proposal fits only when it is
   genuinely relevant to Partio's domain (Git workflows, AI agent
   sessions, code attribution, checkpoints). An idea that is true of the
   source product is not thereby useful here. State the reason either
   way.

4. **Judge whether it is built.** Search the tree for the behaviour the
   proposal asks for. It is built when this repository already does what
   the proposal describes. Name the path or symbol that shows it, or name
   what you searched when you found nothing.

5. **Decide.** Choose one outcome:
   - `close` when the issue must not reach a build. Give exactly one
     reason, in this order of precedence: `built` when step 4 found the
     behaviour; `false-premise` when a claim fails; `does-not-apply`
     when step 3 found no fit; `could-not-verify` when a claim stays
     unresolved.
   - `keep` when nothing above applies and the body already has the
     issue shape below, with a premise block whose claims all hold.
   - `rewrite` when nothing above applies but the body lacks that shape,
     or a claim in its block needs to be stated again.

6. **Compose the body for a keep or a rewrite.** Use the issue shape
   that the proposer files today, so that a build reads a reviewed issue
   the way it reads a new one:
   - what to build, in full, so that a fresh session can build from the
     issue and nothing else;
   - the acceptance criteria as a `- [ ]` checklist under
     `## Acceptance Criteria`;
   - a `## Premise` section with the `<!-- partio:premise:v1 -->`
     marker and one claim per line, each in the form
     `- <claim> [evidence: <path, symbol or command that settles it>]`;
   - the evidence you gathered in step 2: each claim, the evidence, the
     verdict and the excerpt;
   - a final line of its own: `Proposal id: <id>`. Keep the id the issue
     already carries. When it carries none, make a kebab-case id from
     its title.

   A rewrite changes the shape of the issue, not its idea. Keep the
   original idea, and keep the source link to the entireio/cli item it
   came from. Drop any `<!-- program: … -->` pointer to a proposal
   file: proposal files no longer exist, so the pointer leads nowhere.

7. **Write the verdict.** Write one JSON object to
   `$MINION_REVIEW_DIR/verdict.json`, in this form:

   ```json
   {
     "issue": 123,
     "outcome": "keep | rewrite | close",
     "close_reason": "built | false-premise | does-not-apply | could-not-verify",
     "premise": {
       "verdict": "holds | fails | unresolved | no-claims",
       "claims": [
         {"claim": "…", "evidence": "…", "verdict": "holds | fails | unresolved", "excerpt": "…"}
       ]
     },
     "fit": {"applies": true, "reason": "…"},
     "built": {"built": false, "evidence": "…"},
     "duplicates": [],
     "rewrite": {"title": "…", "body": "…", "changes": ["…"]}
   }
   ```

   - `issue` is the number of the issue you reviewed.
   - Set `close_reason` for a close only, and `rewrite` for a rewrite
     only. Leave each out otherwise.
   - Every claim carries its evidence and its excerpt. The premise
     verdict is `no-claims` only when the body states no checkable fact,
     in a block or in its prose.
   - A close carries the evidence for its reason: a failing claim for
     `false-premise`, an unresolved claim for `could-not-verify`,
     `built: true` with evidence for `built`, and `applies: false` with a
     reason for `does-not-apply`.
   - `rewrite.body` is the whole body from step 6, and `rewrite.changes`
     lists what you changed against the current body, one line each.

   The gate checks this file after your session ends. A missing file, a
   malformed one, or one that contradicts itself is recorded as "no
   verdict", and the issue stays as it is.

Make no change on GitHub. Your `gh` calls only read: do not comment,
label, edit, close or create anything. The gate records your verdict,
and nothing else acts on it in this version.

Write nothing in your working directory. The verdict file is the only
file you write: keep every other result in your session, and write no
scratch file anywhere. A file left in the working directory is read as
build output and turns this review into a pull request. Do not run
`git`, and do not open a pull request.
