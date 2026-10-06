# Premise verifier

<!-- partio:premise-verifier:v1 -->

Verification is described here, once. The propose, research and implement
programs apply this description. They do not each carry their own idea of what
verification means, and they do not restate the procedure below.

Given a premise block and a checked-out tree, gather the evidence each claim
names and return a verdict plus what was found.

## Inputs

1. **A set of claims.** Normally a premise block: the `## Premise` section that
   carries the `<!-- partio:premise:v1 -->` marker, one claim per line, each
   naming the path, symbol or command that settles it. When the issue carries
   that marker, the block is everything you need, and you do not verify the
   prose around it. When it carries no marker, read `## When there is no
   block` below.
2. **A checked-out tree.** The repository the claims are about, already on
   disk in your working directory. Read it. Do not answer from the source
   material the idea came from, and do not answer from memory of how similar
   projects work.

## When there is no block

Hundreds of proposals were filed before the block format existed. They state
their facts in prose, and they are not exempt.

When the issue carries no `<!-- partio:premise:v1 -->` marker, extract the
factual claims from the issue prose and verify those. A factual claim is a
statement about this repository that the tree can settle. Take each one word
for word from the body, and name the path, symbol or command that settles it.
An assumed fact, below, is the one exception: quote its phrase word for word,
and report the claim you state from it next to that phrase.

Extract these:

- what the code does today, or does not do
- how it behaves, performs or is structured
- a fact that a requested behaviour, a design instruction or an acceptance
  criterion takes for granted about today's code or data

**An assumed fact.** A request can carry a fact about today's tree. Quote the
phrase that carries the fact, then state the fact again as a checkable claim,
and name the evidence that settles it. For example, #31 asked for "a merge
strategy appropriate for append-only checkpoint data". The phrase
"appropriate for append-only checkpoint data" carries the claim "nothing in
Partio deletes or rewrites checkpoint data". Verify that claim, not the
request around it.

Leave these, because no tree settles them:

- the behaviour the proposal asks for, but not the facts that behaviour rests
  on: extract those as assumed facts above
- what a sibling product does, and where the idea came from
- whether the idea is worth doing

Then verify the claims you extracted by the same procedure, the same verdicts
and the same output as a block's claims. Nothing below this section changes
because the claims came from prose. An older proposal meets the same bar as one
filed today.

Some proposals only describe what to build. If the prose makes no checkable
factual claim, there is nothing to verify: continue. This is not `unresolved`.
An unresolved claim is one you could not settle; here there was no claim to
settle, and treating the two alike would block most of the backlog at once.
Say so in your report: state that the prose makes no checkable claim, and name
what you read to decide that.

Check the proposal in front of you, and no other.

Extraction happens here, at check time. The verifier writes nothing itself: it
decides each claim and reports the claim, its evidence, its verdict and its
excerpt. What a caller writes with that result, on the issue or anywhere else,
belongs to the caller, and the caller states it in its own instructions.

## Procedure

For each claim, in order, whether it came from a block or from the prose:

1. **Gather the evidence the claim names.**
   - A path: read the file. If the path is a directory, list it and read the
     files the claim is about.
   - A symbol: find it, for example with `grep -rn '<symbol>' .`, then read
     where it is defined.
   - A command: run it and keep the output.
2. **Decide the claim against what you gathered, and against nothing else.**
3. **Record the verdict and the excerpt that produced it** — the path and the
   lines you read, or the command and the output it printed.

Gather first, decide second. A claim you gathered no evidence for is
`unresolved`. It is never `holds`.

**A sweeping claim.** A claim such as "append-only", "nothing deletes" or
"never changes" covers every path, so one path that agrees with it does not
settle it. Settle a sweeping claim with a command that lists every path that
changes the data, for example `grep -rn --include='*.go' 'update-ref\|"-D"' .`
for the checkpoint branch, and read each path on the list. One counterexample
fails it: for #31, `partio prune` removes checkpoints, and `partio reset`
deletes and recreates the checkpoint branch.
This applies to a claim from a block and to a claim from the prose.

## Verdicts

For one claim:

- `holds` — the gathered evidence confirms the claim.
- `fails` — the gathered evidence contradicts the claim.
- `unresolved` — the named evidence does not settle the claim. The path is
  missing, the symbol is absent, or the command printed nothing to read.

For the block: `fails` if any claim fails. Otherwise `unresolved` if any claim
is unresolved. Otherwise `holds`.

## Output

Report every claim, whichever way it went:

- the claim, word for word (for an assumed fact, the quoted phrase and the
  claim stated from it)
- the evidence the claim named
- the verdict
- the excerpt that produced the verdict

Report this for a block that holds as well as for one that does not. A stage
that passes carries its evidence too.

## What the caller does

Only a premise that `holds` lets the work continue. One that `fails` or is
`unresolved` stops the calling stage. Where its claims came from makes no
difference: an extracted claim that fails stops the stage exactly as a
block claim that fails does.

What "stop" means belongs to the caller: the proposer drops the idea before it
becomes an issue, and a later stage labels the issue and hands control back to
the operator. No caller treats `fails` or `unresolved` as a pass, and no caller
decides a claim it did not gather evidence for.
