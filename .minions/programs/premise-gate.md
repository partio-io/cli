---
id: premise-gate
target_repos:
  - cli
---

# Premise gate

Decide whether a build may start. This program writes no code and opens
no pull request. It runs before the implement program, against the tree
that build is about to modify.

The build is the most expensive stage to waste. Issue #30 reached three
builds and produced a pull request that improves nothing: measured
directly, the change it delivers is level with the original on a
thousand-file repository and roughly thirty to forty percent slower on a
ten-thousand-file one. Every minute of that was spent after the premise
had already been false for months.

Research checking the premise earlier does not settle it. A claim that
was true in March can be false in August, and a build can be triggered
by hand long after any research ran. This stage gathers the evidence
again, against the tree in front of it.

## Context

- `cli/.minions/premise-verifier.md`
- `cli/.minions/stage-gate.md`

## Agents

### premise-checker

```capabilities
tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
max_turns: 60
```

You check the parent issue's premise against the tree in front of you,
before this build writes anything. Only your verdict decides whether the
build proceeds.

The repository the claims are about is already checked out in your
working directory. Read it. Do not decide a claim from the issue text,
and do not decide it from memory of how similar projects work.

1. Read the parent issue body, which is provided under the "Issue"
   section of your prompt, and find the claims you must verify.

   When the body carries a `## Premise` section with the
   `<!-- partio:premise:v1 -->` marker, that block is what you verify,
   and you do not verify the prose around it.

   When the body carries no such marker, the claims come from the prose
   instead. Every proposal filed before the block format existed is in
   this state, and none of them is exempt. The verifier's `## When there
   is no block` section describes the extraction. Follow it as written.

2. Apply `.minions/premise-verifier.md` to the claims against the
   checked-out tree. It describes verification once, for every stage
   that needs it. Follow it as written — gather the evidence each claim
   names, then decide, then record the excerpt that decided it. Do not
   restate its procedure and do not invent your own. A claim from the
   prose meets the same bar as a claim from a block.

3. Apply `.minions/stage-gate.md` to the verdict it produced. That file
   describes what a stage does when a premise does not hold, and what it
   does when it holds. Follow it as written.

4. When the premise does not hold, the gate has already labelled the issue
   and posted the comment naming what you checked and what you found.
   Stop there. Write no code, produce no further artifact, and do not
   close the issue. Which label the gate applies, and how the operator
   overrules it, are the gate's to state — not this program's, and not
   yours to decide from the issue's current labels. Verify against the
   tree on every run, whatever labels the issue already carries.

5. When the premise holds, record every claim, the evidence it named,
   its verdict, and the excerpt that produced it — for a premise that
   holds as well as for one that does not. A build that proceeds carries
   the evidence it rests on.

   Refresh the issue body only when the issue carries a premise block.
   An issue whose claims you extracted from its prose gets no block: the
   extraction lives in your report and nowhere else. Leave that body as
   the operator wrote it. Do not backfill a block into it.

   To refresh a block, read the current body, replace the evidence
   excerpts inside the `## Premise` section with what you read on this
   run, and keep the section's marker and every claim. Leave the rest of
   the body untouched — you are refreshing evidence, not rewriting the
   proposal. Then write the whole body to the shared path and put it
   back:

   ```
   REFRESHED_BODY="/tmp/minion-build-body-${MINION_ISSUE_NUMBER:-0}.md"
   gh issue edit "$MINION_ISSUE_NUMBER" --repo partio-io/cli --body-file "$REFRESHED_BODY"
   ```

Some proposals only describe what to build. When the body states no
checkable fact about this repository — in a block or in its prose — there
is nothing to verify, and the build continues. That is not `unresolved`.
Say so in your report: state that the body makes no checkable claim, and
name what you read to decide that.

Do not create or modify any file in the working directory. Every file
you write goes under `/tmp`. A file left in the working directory is
read as build output and turns this check into a pull request, which is
the outcome this program exists to prevent. Do not run `git`, and do not
open a PR.
