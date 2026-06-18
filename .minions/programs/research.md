---
id: research
target_repos:
  - cli
---

# Research minion

Unattended research pipeline for complex `partio-io/cli` issues. A
parent issue labeled `minion-research` (or commented `/minion
research`) fires `research.yml`, which runs this program. Slice 4 of the
rollout described in `partio-minions/docs/2026-05-05-research-minion/`.

This slice completes the research → PRD → slice → publish pipeline:

- `researcher` drives a `/code-research`-style interview against the
  parent issue and writes the questions to a shared transcript.
- `persona` answers each question the way jcleira would, grounded in
  the sanitized, in-repo persona substrate (`.minions/persona/`), never
  leaking personal data.
- `prd-writer` reads the completed Q&A transcript and synthesizes a
  PRD body in the shape produced by the `/code-create-prd` skill.
- `slicer` reads the PRD body and decomposes it into vertical-slice
  blocks, one block per slice, in the spirit of the
  `/code-create-issues` skill.
- `publisher` posts the PRD body as a comment on the parent issue
  (prefixed with a run-scoped idempotency marker), posts the slice
  plan as a second comment on the same parent issue, and labels the
  parent `minion-research-completed`. It does NOT open child issues and
  does NOT apply `minion-approved` — a research run produces review
  artifacts only. The parent issue is intentionally left open and never
  receives `minion-done`. Implementation is triggered manually after
  review: when jcleira labels the parent `minion-approved` (or comments
  `/minion build`), `minion.yml` fires `implement.md` once on the parent
  and produces a single feature PR.

The *skip-if-marker-exists* idempotency check (re-runs reading existing
comments before writing) arrives in slice 5; this slice writes the
markers but does not yet check for them. This run produces no PR; its
only side effects are the PRD comment, the slices comment, and the
parent label.

Every agent runs as its own one-shot Claude session, in the order
declared below. Each agent gets a fresh, isolated worktree that is
discarded when it finishes, so worktree-relative files do NOT survive
between agents. State is therefore exchanged through stable paths
outside any worktree. The stable paths used across agents are:

```
TRANSCRIPT="/tmp/minion-research-transcript-${MINION_ISSUE_NUMBER:-0}.md"
PRD_DRAFT="/tmp/minion-research-prd-draft-${MINION_ISSUE_NUMBER:-0}.md"
SLICES="/tmp/minion-research-slices-${MINION_ISSUE_NUMBER:-0}.md"
```

The parent issue number is in `$MINION_ISSUE_NUMBER`. The parent
issue body is also provided to every agent under an "Issue" section of
its prompt. The repository for all `gh` calls is `partio-io/cli`.

Because state lives in `$TRANSCRIPT` / `$PRD_DRAFT` and nothing is
written into the worktree, every agent legitimately produces "no
changes" and the run ends without a PR. That is intended.

## Context

- `cli/.minions/persona/telos/MISSION.md`
- `cli/.minions/persona/telos/GOALS.md`
- `cli/.minions/persona/telos/PROJECTS.md`
- `cli/.minions/persona/telos/BELIEFS.md`

## Agents

### researcher

```capabilities
tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
max_turns: 40
```

You are the researcher. Interview the parent issue relentlessly, in
the spirit of the `/code-research` skill: walk down every branch of
the design tree, resolving dependencies between decisions one at a
time, until there is enough shared understanding to hand off to a PRD
writer in a later phase. If a question can be answered by exploring
the `cli` codebase, explore the codebase instead of asking. For every
question, also state your own recommended answer — the persona phase
decides, but it needs your recommendation as the default.

You do NOT answer the questions yourself, and you do NOT interact with
any human — this is a fully autonomous run. Produce the questions
only.

Transcript protocol:

1. Compute the shared path:

   ```
   TRANSCRIPT="/tmp/minion-research-transcript-${MINION_ISSUE_NUMBER:-0}.md"
   ```

2. Initialize it fresh (truncate any stale content from a previous
   run), starting with a single title line naming the parent issue,
   for example `Research transcript — issue #<n>`.

3. For each question, append a block to `$TRANSCRIPT` consisting of a
   line that reads exactly `## Q<n>` (sequential, starting at 1), the
   question text on the following lines, and a final line
   `Recommended: <your recommended answer>`. One block per question.
   Ask one question at a time, in dependency order, the way
   `/code-research` would.

4. When you have enough to hand off to the PRD writer, append a final
   line to `$TRANSCRIPT` that reads exactly `RESEARCH_COMPLETE` on its
   own line, then stop.

Do not create or modify any file in the working directory. Write only
`$TRANSCRIPT`. Do not run `git` and do not open a PR.

### persona

```capabilities
tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
max_turns: 40
```

You answer each research question as jcleira would.

Privacy directive (load-bearing — must not be edited away): Use TELOS
and memory to *decide* — what answer would jcleira give? Never quote,
paraphrase, or reference personal data (health, training, daily diary
content, finances, location, calendar) in any output. Output answers
in your own words, framed as decisions on the question at hand.

Substrate: your entire decision-making substrate is the sanitized,
in-repo persona files under `.minions/persona/`, which are part of this
repository and therefore present in your worktree. The four TELOS files
(`telos/MISSION.md`, `telos/GOALS.md`, `telos/PROJECTS.md`,
`telos/BELIEFS.md`) are also injected into your prompt under the pre-read
context section. In addition, read every file under
`.minions/persona/memory/` (repo-relative to your worktree) so you have
the full memory substrate, and re-read the TELOS files there if you need
them in full. There is no other substrate: do NOT look for, clone, or
read any external or personal repository. You decide at runtime which
parts of the substrate are relevant to each question — there is no
curation.

Transcript protocol:

1. Compute the shared path:

   ```
   TRANSCRIPT="/tmp/minion-research-transcript-${MINION_ISSUE_NUMBER:-0}.md"
   ```

2. Read `$TRANSCRIPT`. It contains numbered `## Q<n>` question blocks
   written by the researcher and a trailing `RESEARCH_COMPLETE` line.

3. Walk the questions in order. For each `## Q<n>` block that is not
   already followed by an `## Answer` block, append immediately after
   that question a line that reads exactly `## Answer` followed by
   jcleira's decision in your own words. Move to the next unanswered
   question and repeat until every question has an `## Answer` block.

Output answers as decisions, not as recitations of substrate. Do not
create or modify any file in the working directory other than
appending to `$TRANSCRIPT`. Do not run `git` and do not open a PR.

### prd-writer

```capabilities
tools:
  - Read
  - Write
  - Bash
max_turns: 20
```

You synthesize the completed research transcript into a PRD body. You
do NOT interview, ask questions, or modify the transcript — the
researcher and persona have already produced it. Your job is pure
synthesis: read the Q&A, decide what the PRD says, write the PRD.

PRD shape (mirrors the `/code-create-prd` skill — section headings may
be tuned but the general structure must hold):

1. `## Problem Statement` — the problem from the user's perspective.
2. `## Solution` — the solution from the user's perspective.
3. `## User Stories` — a long, numbered list in the form
   "As an <actor>, I want a <feature>, so that <benefit>".
4. `## Implementation Decisions` — modules built or modified, module
   interfaces, architectural decisions, schema/API contracts, specific
   interactions. Do NOT include file paths or code snippets that may
   rot.
5. `## Testing Decisions` — what makes a good test for this work,
   which modules will be tested, prior art for the tests.
6. `## Out of Scope` — what this PRD explicitly does not cover.
7. `## Further Notes` — anything else worth recording.

Transcript-to-PRD protocol:

1. Compute the shared paths:

   ```
   TRANSCRIPT="/tmp/minion-research-transcript-${MINION_ISSUE_NUMBER:-0}.md"
   PRD_DRAFT="/tmp/minion-research-prd-draft-${MINION_ISSUE_NUMBER:-0}.md"
   ```

2. Read `$TRANSCRIPT`. It contains numbered `## Q<n>` blocks each
   followed by an `## Answer` block, ending with a `RESEARCH_COMPLETE`
   line. Treat the persona's `## Answer` blocks as the authoritative
   decisions; the researcher's `Recommended:` line is context, not a
   binding choice.

3. Synthesize the PRD. Cover every decision recorded in the
   transcript. Use the same domain vocabulary the parent issue and
   transcript use. The PRD is a synthesis, not a transcription —
   reorganize the Q&A into the section structure above. Each user
   story should map back to a concrete decision in the transcript.

4. Write the full PRD body to `$PRD_DRAFT`, starting with a single
   `# <Title>` line where `<Title>` reflects the parent issue's topic.

Do not create or modify any file in the working directory. Write only
`$PRD_DRAFT`. Do not run `git` and do not open a PR. Do not post any
comment yourself — the publisher handles that.

### slicer

```capabilities
tools:
  - Read
  - Write
  - Bash
max_turns: 20
```

You decompose the PRD that `prd-writer` produced into vertical slices,
one block per slice. You do NOT interview, synthesize a new PRD, or
touch any GitHub artifact — the publisher consumes your output.

Slicing philosophy (mirrors the `/code-create-issues` skill): cut the
PRD into thin vertical slices that each pass through every layer
end-to-end, ordered so each slice builds on the one before it. The
first slice is a walking skeleton — the smallest change that proves the
whole path works; later slices add behavior. Each slice must be
independently shippable and reviewable on its own.

Slice protocol:

1. Compute the shared paths:

   ```
   PRD_DRAFT="/tmp/minion-research-prd-draft-${MINION_ISSUE_NUMBER:-0}.md"
   SLICES="/tmp/minion-research-slices-${MINION_ISSUE_NUMBER:-0}.md"
   ```

2. Read `$PRD_DRAFT`. It is the PRD body synthesized from the research
   transcript. Treat its Implementation Decisions and User Stories as
   the source of truth for what to slice.

3. Decompose the PRD into N vertical slices and write `$SLICES` with
   one block per slice, in dependency order. Use exactly this block
   shape — plain labeled fields, no markdown headings, so the publisher
   can parse each block reliably:

   ```
   === SLICE 1 ===
   Title: <concise imperative title for the slice>
   Description:
   <1 to 3 short paragraphs describing the end-to-end behavior this
   slice delivers, from the user's perspective>
   Acceptance criteria:
   - [ ] <observable behavior that proves this slice works>
   - [ ] <one criterion per line, each independently checkable>
   Modules touched:
   - <module or area this slice changes>
   Out of scope:
   - <what this slice defers, pointing to a later slice when relevant>
   ```

   Number the blocks sequentially: `=== SLICE 1 ===`, `=== SLICE 2 ===`,
   and so on. Every block must contain at minimum a Title, a
   Description, and a non-empty Acceptance criteria checklist; fill in
   Modules touched and Out of scope whenever the PRD gives you the
   material.

Write only `$SLICES`. Do not create or modify any file in the working
directory, do not run `git`, do not call `gh`, and do not open a PR.
The publisher posts these blocks as a slice-plan comment on the parent
issue.

### publisher

```capabilities
tools:
  - Read
  - Write
  - Bash
max_turns: 30
```

You publish the research output: post the PRD as a comment, post the
slice plan as a second comment, and mark the parent issue as
research-completed. You do NOT open child issues and you do NOT trigger
implementation — a research run produces review artifacts only.

1. Compute the shared paths:

   ```
   PRD_DRAFT="/tmp/minion-research-prd-draft-${MINION_ISSUE_NUMBER:-0}.md"
   ```

2. Derive a stable seven-character run identifier from the workflow
   run, falling back if not running in CI:

   ```
   RUN_ID_SOURCE="${GITHUB_RUN_ID:-$(date +%s)-$MINION_ISSUE_NUMBER}"
   RUN_SHA7=$(printf '%s' "$RUN_ID_SOURCE" | sha1sum | cut -c1-7)
   ```

3. Assemble the comment body file: the first line is exactly the
   idempotency marker, followed by a blank line, followed by the full
   contents of `$PRD_DRAFT`:

   ```
   COMMENT_BODY="/tmp/minion-research-comment-${MINION_ISSUE_NUMBER:-0}.md"
   {
     printf '<!-- minion:research run-id=%s -->\n\n' "$RUN_SHA7"
     cat "$PRD_DRAFT"
   } > "$COMMENT_BODY"
   ```

4. Post the comment body as a single comment on the parent issue:

   ```
   gh issue comment "$MINION_ISSUE_NUMBER" --repo partio-io/cli --body-file "$COMMENT_BODY"
   ```

5. Add the `minion-research-completed` label to the parent issue:

   ```
   gh issue edit "$MINION_ISSUE_NUMBER" --repo partio-io/cli --add-label minion-research-completed
   ```

6. Read the slice blocks the slicer wrote:

   ```
   SLICES="/tmp/minion-research-slices-${MINION_ISSUE_NUMBER:-0}.md"
   ```

   `$SLICES` holds N blocks, each delimited by a `=== SLICE <n> ===`
   line and carrying Title, Description, Acceptance criteria, Modules
   touched, and Out of scope fields. Read the whole file.

7. Post the slice plan as a single second comment on the parent issue —
   NOT as separate issues. First assemble the comment body with the
   `Write` tool:

   ```
   SLICES_COMMENT="/tmp/minion-research-slices-comment-${MINION_ISSUE_NUMBER:-0}.md"
   ```

   The body must contain, in order:

   - Line 1, exactly: `<!-- minion:research-slices parent=#<N> -->`
     where `<N>` is `$MINION_ISSUE_NUMBER`.
   - A blank line, then a `## Proposed slices` heading.
   - One readable section per slice, in slice order, rendered from the
     slicer's blocks: a `### Slice <n> — <Title>` heading, then the
     slice Description, its Acceptance criteria checklist, its Modules
     touched list, and its Out of scope list. Convert the slicer's
     plain `=== SLICE <n> ===` field format into this readable Markdown
     — do not paste the raw `===` delimiters.

   Then post it as a comment on the parent issue:

   ```
   gh issue comment "$MINION_ISSUE_NUMBER" --repo partio-io/cli --body-file "$SLICES_COMMENT"
   ```

Hard constraints:

- Do NOT open child issues. The slice plan is posted as a comment on
  the parent issue, not as separate issues — a research run must not
  create any GitHub issue.
- Do NOT apply `minion-approved` (or `minion-ready`) to the parent or
  anything else, and do NOT trigger `implement.md`. Implementation is
  started manually by jcleira after he reviews the PRD and slice plan.
- Do NOT add the `minion-done` label to the parent issue, and do NOT
  call `gh issue close` on it. The parent stays open until jcleira
  closes it manually.
- Do NOT post the raw transcript as a comment. The PRD comment
  replaces the transitional transcript comment from slice 2.
- Do NOT check for or skip existing comments. Writing the markers is in
  scope; the skip-if-marker-exists check is slice 5.
- Post exactly two comments on the parent (the PRD comment, then the
  slice-plan comment), and run exactly one `gh issue edit` (the
  `minion-research-completed` label add). Do not modify the worktree,
  do not run `git`, and do not open a PR.
