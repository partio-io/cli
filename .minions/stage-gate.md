# Stage gate

<!-- partio:stage-gate:v1 -->

What a stage does with a premise verdict is described here, once. The research
and implement programs apply this description. They do not each invent a way to
stop, and they do not restate the procedure below.

Verification is a separate description: `.minions/premise-verifier.md` decides
whether each claim holds. This file starts from its verdict.

A premise is checked at every stage, against the tree in front of that stage.
A claim that was true when the issue was filed can be false months later, so no
stage trusts an earlier stage's verdict. Each one gathers the evidence again.

## Inputs

1. **A verdict for the block**, from `.minions/premise-verifier.md`: `holds`,
   `fails` or `unresolved`, plus every claim, the evidence it named, and the
   excerpt that produced its verdict.
2. **The issue** the premise belongs to, in `$MINION_ISSUE_NUMBER`.

## When the premise does not hold

`fails` and `unresolved` both stop the stage. A claim nothing settled has not
been checked, so it is not a pass.

1. **Stop before you produce anything.** No PRD, no slice plan, no branch, no
   pull request. The stage ends here.
2. **Add the `do-not-build` label** to the issue. The approval program already
   treats this label as a skip condition, so blocking needs no new label:

   ```
   gh issue edit "$MINION_ISSUE_NUMBER" --repo partio-io/cli --remove-label do-not-build
   gh issue edit "$MINION_ISSUE_NUMBER" --repo partio-io/cli --add-label do-not-build
   ```

   Remove it first, every time, including when the issue does not carry it — a
   label that is already present fires no fresh trigger, so adding it twice
   leaves the pipeline silent and the operator with no signal that this run
   also blocked. `--remove-label` on an issue without the label is not an error.
3. **Post one comment saying what you checked and what you found.** This
   comment is the operator's whole account of the decision. Its first line is
   exactly the marker:

   ```
   <!-- partio:premise-gate:v1 -->
   ```

   Then, for every claim you checked, whichever way it went — and whether it
   came from a block or was extracted from the issue prose:

   - the claim, word for word
   - the evidence the claim named
   - the verdict for that claim
   - the excerpt that produced it — the path and the lines you read, or the
     command and the output it printed

   Name the claims that failed and the evidence that contradicted them. A
   verdict with nothing behind it is the prose assertion the premise block
   replaced.
4. **Do not close the issue.** Do not add `minion-done`. The operator decides
   what happens to a blocked proposal; the stage only reports.

## How the operator overrules

The operator removes the `do-not-build` label. That is the whole mechanism.

A stage never reads the label as a decision it can trust. It re-verifies
against the current tree on every run and reaches its own verdict, so a wrong
verdict is corrected by removing the label and running again, and a correct one
is reached again from the same evidence.

## When the premise holds

1. **Refresh the premise block in the issue**, but only if the issue carries
   one, with the evidence you just gathered, so the next stage checks against
   today's facts rather than the ones the proposer recorded. Keep the
   `## Premise` section and its `<!-- partio:premise:v1 -->` marker; replace the
   evidence excerpts with what this run read.

   An issue with no block gets none. You extracted its claims from its prose,
   and that extraction is not written back: leave the body as the operator
   wrote it and record the evidence in step 2 instead.
2. **Record the verdict.** A run that passed carries its evidence too — the
   claims, the evidence, the `holds` verdict, and the excerpts behind it.
3. **Continue.** Only a block that `holds` lets the stage go on.
