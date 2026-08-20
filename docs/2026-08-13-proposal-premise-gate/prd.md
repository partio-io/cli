# Proposal Premise Gate

## Problem Statement

The minion pipeline files feature proposals twice a day, and it writes
them about code it has never read.

**A proposal can assert something false about this repository and
nothing will catch it.** Issue #30 said Partio's post-commit hook
"performs a full tree walk to determine which files were changed in a
commit… O(N) in the number of tracked files". The hook never did that.
It ran a two-commit `git diff`, which is already a tree-to-tree
comparison and never touches the working tree. The claim was imported
from a sibling product's changelog, where the problem was real. Nobody
checked whether it was real here.

The issue was filed on 2026-03-26 and built on 2026-08-10. It took
three builds to produce a pull request, and that pull request improves
nothing. Measured directly, the replacement is level with the original
on a thousand-file repository and roughly thirty to forty percent
slower on a ten-thousand-file one; process startup dominates both. The
one genuine defect it fixes — first commits recording zero attributed
lines, because the fallback pointed at a git object that does not exist
in any repository — was found by accident and is unrelated to the
stated goal.

**The proposer improvises, because its instructions are thrown away.**
The program format keeps the title prose, the context section, the
planner section and the agents section. Everything else is discarded
without warning. The propose program carries its seven steps under a
`## Steps` heading and defines no agents, so the executor synthesizes
an implicit agent whose entire instruction set is the one-sentence
description under the title. The steps that tell it to apply the ingest
prompt, to check for an existing proposal, and to advance the source
cursor never reach the model. It reads three files named as context
hints and invents the rest of the workflow. Five programs are written
this way; every program that behaves reliably uses an agents section.

**The result is a queue nobody can read.** There are 534 open
proposals. Fifty-six already carry the approved label and will build
when touched. Ten have been closed. Nine minion pull requests have ever
merged — under two percent of what the proposer has filed. A bad
proposal does not stand out in a queue that long, and the operator is
the only gate between a false premise and a build.

## Solution

Make a proposal state what it believes about this repository, and make
every stage check that belief before spending anything on it.

- **A proposal carries its premise.** Each one records the factual
  claims it rests on, one per line, each paired with the path or
  command that proves it. A premise written as prose cannot be
  rechecked; a premise written as claims can.
- **The proposer reads the code before it writes the claim.** The
  source tree is already checked out when the proposer runs. It stops
  guessing from a sibling product's changelog and grounds every claim
  in what is actually here. Only ideas whose premise survives get
  filed, so volume falls at the source rather than being pruned later.
- **Every stage rechecks, because time passes.** Propose, research and
  implement each verify the premise against the tree in front of them.
  A claim that was true in March can be false by August, and issue #30
  sat for four and a half months.
- **A failed check stops the stage.** It adds the blocking label,
  comments what it checked and what it found, and ends. It does not
  close the issue and it does not build anyway. The operator decides
  what happens next.
- **Rejections are recorded.** Ideas the proposer drops are appended to
  a log it commits, with the reason. Without that record a bar that is
  set too high is indistinguishable from a source that has gone quiet.
- **Instructions reach the model.** The propose program and the ingest
  prompt move their instructions into the agents section, where the
  format actually carries them. A test then fails any program that
  hides instructions in a section the parser drops.

## User Stories

1. As the operator, I want a proposal to state the facts it assumes about my code, so that I can judge it without reading the codebase myself.
2. As the operator, I want each assumed fact paired with a command that proves it, so that I can verify a claim in seconds rather than minutes.
3. As the operator, I want the proposer to read my repository before it describes my repository, so that proposals stop describing a product I do not have.
4. As the operator, I want a proposal whose premise is already false to never be filed, so that my queue holds only work that could still matter.
5. As the operator, I want the premise rechecked when research runs, so that a proposal that aged badly is caught before it is planned.
6. As the operator, I want the premise rechecked when the build starts, so that a proposal approved months ago cannot build against a tree that has moved on.
7. As the operator, I want a failed check to stop the stage, so that no compute is spent on work that is already known to be pointless.
8. As the operator, I want a failed check to add a blocking label, so that the issue cannot be picked up again by accident.
9. As the operator, I want the failing stage to comment what it checked and what it found, so that I can confirm or overrule the machine from the issue alone.
10. As the operator, I want the machine never to close an issue on its own, so that I stay the one who decides what is abandoned.
11. As the operator, I want to overrule a failed check and build anyway, so that a verifier mistake does not permanently block real work.
12. As the operator, I want rejected ideas written to a file in the repository, so that I can audit whether the bar is dropping things I wanted.
13. As the operator, I want each rejection to record its reason, so that I can tell a false premise from an irrelevant source item.
14. As the operator, I want the rejection log committed by the same run that advances the source cursor, so that a dropped idea is never silently unrecoverable.
15. As the operator, I want proposal volume to fall because fewer bad ideas are filed, so that I do not have to bulk-close a backlog I never read.
16. As the operator, I want my existing 534 proposals left alone, so that nothing I might still want disappears in a cleanup.
17. As the operator, I want an older proposal that has no premise block checked against its prose claims instead, so that the backlog is not exempt from the gate.
18. As the operator, I want the fifty-six already-approved proposals checked when a stage next touches them, so that no special sweep is needed.
19. As the operator, I want the propose program's instructions to actually reach the model, so that it follows the workflow I wrote instead of inventing one.
20. As the operator, I want a test that fails when a program hides instructions in a discarded section, so that this class of bug cannot return quietly.
21. As the operator, I want that test to name the offending program, so that I can fix it without bisecting the format rules.
22. As the operator, I want the ingest prompt to demand claims grounded in the checked-out tree, so that the model cannot satisfy it with a plausible guess.
23. As the operator, I want the proposal format to keep working for sources that are genuinely relevant, so that raising the bar does not silence the pipeline.
24. As the operator, I want the change to ship as program text in this repository, so that no engine release, tag and version pin is needed to adopt it.
25. As the operator, I want the gate to reuse the blocking label the pipeline already understands, so that no new state is introduced.
26. As the researcher stage, I want to read a machine-checkable premise, so that I can validate before planning rather than planning something void.
27. As the researcher stage, I want to refresh the premise block when I run, so that later stages check against the newest evidence.
28. As the build stage, I want to refuse a proposal whose premise has failed, so that I never open a pull request that solves nothing.
29. As the operator, I want a stage to record what it verified even when the premise holds, so that a green build carries its evidence too.
30. As the operator, I want the gate scoped to the CLI repository, so that the change lands where every source and every target already points.
31. As a Partio user, I want first commits to record their attributed lines, so that my earliest work is not silently counted as zero.
32. As a Partio user, I want hook changes justified by measurement, so that my post-commit path does not get slower in the name of speed.
33. As the operator, I want a benchmark that compares against the old behaviour, so that "faster" is demonstrated rather than asserted.
34. As the operator, I want the pipeline's hit rate to become legible, so that I can tell whether the proposer is worth running twice a day.

## Implementation Decisions

**The premise block is the contract between stages.** Every proposal
carries a section listing the factual claims it depends on. Each claim
is one line and names the evidence that settles it — a path, a symbol,
or a command whose output decides the matter. The block is the only
thing a later stage needs in order to recheck; it does not depend on
the prose that surrounds it.

**One verifier, three callers.** Verification is a single described
behaviour: given a premise block and a checked-out tree, produce a
verdict plus the evidence gathered. The propose, research and implement
programs call it. They do not each carry their own idea of what
verification means.

**Verification runs at every stage, not once.** The operator's decision
was explicit: code moves between filing, approval, research and build,
so a single check at any one point is worthless by the time the next
stage runs. Research additionally refreshes the block so that implement
checks the newest evidence.

**A failed verdict stops the stage and hands control back.** The stage
adds the existing `do-not-build` label, comments the claims it checked
and the evidence that contradicted them, and exits without further
work. It does not close the issue. This follows the standing rule that
the machine does not change state the operator has not seen. The label
is reused rather than invented because the approval program already
treats it as a skip condition.

**Proposals predating the block are checked against their prose.** The
534 existing proposals carry no block. A stage that meets one extracts
the factual claims from the body and verifies those, with the same stop
behaviour on failure. The backlog is not swept ahead of time and is not
pruned; each proposal is checked when a stage next touches it.

**Volume is reduced at the source.** The proposer files only what
survives its own premise check. No cap on proposals per run and no bulk
close of the existing queue — the operator chose to raise the bar
rather than bound the output or discard the backlog.

**Rejections are durable.** Ideas dropped by the proposer append to a
log file that the run commits alongside the source cursor it already
updates. Reason is recorded with each entry. This exists because the
cursor advances past a rejected item, so without the log the idea is
gone with no trace.

**Instructions move into the agents section.** This is a prerequisite,
not a nicety. The program format retains the title prose, the context
section, the planner section and the agents section, and silently
discards the rest; a program with no agents runs with only its
one-sentence description as instruction. Premise instructions added to
a steps section would vanish exactly as the current seven steps do. The
propose program is converted. The ingest prompt is a template read as
context rather than a parsed program, so it needs new content but no
conversion. The approval, ingest, documentation-update and
readme-update programs have the same defect and are named here as a
known gap rather than fixed in this work; of those, only the
documentation-update program is reachable from a workflow.

**The ingest prompt gains premise fields and a grounding rule.** Its
output schema carries the claims and their evidence, and its
instructions require every claim to be grounded in the checked-out
tree. Adaptation from a source product remains the goal; asserting the
source product's problems as this product's problems does not.

**Nothing changes in the engine.** All edits are program text in this
repository, so adoption needs no engine release, tag, or version pin
bump. The single exception is the shape test, which is ordinary Go.

**Language.** The shape test is written in Go, because this repository
is Go and the existing language policy makes the repository's language
the default. Everything else in this work is Markdown program text and
implies no service or script.

**Scope is this repository.** It is the only entry in the proposer's
target list, and all three configured sources point at the same sibling
repository.

## Testing Decisions

**What a good test looks like here.** A test asserts behaviour an
outside caller can observe, not the shape of the implementation behind
it. For program text, the observable behaviour is what actually reaches
the model; for the shape test, it is whether a given program's
instructions survive parsing. Tests do not assert on prompt wording,
because wording is expected to change and asserting on it produces a
suite that fails on every edit without ever catching a defect.

**Every module is tested, and the untestable parts are named.**

- `program-shape-test` is the machine-testable core and carries real
  coverage. It reads every program in this repository and fails any one
  whose instructions sit in a section the format discards, naming the
  offending program. It covers the positive case (a program with an
  agents section passes), the negative case (a program with steps and
  no agents fails), and the boundary (a program whose whole instruction
  set lives in the title prose passes, since that text does reach the
  model).
- `premise-block` is tested through the shape of what stages produce
  and consume: a block with claims and evidence is recognised, a block
  whose claims carry no evidence is rejected, and a proposal with no
  block at all is handled by the prose path rather than crashing.
- `premise-verifier`, `stage-gate`, `rejection-log`, `propose-program`,
  `ingest-prompt`, `research-program` and `implement-program` are
  prompt-driven behaviours. Their instruction text is verified
  mechanically only to the extent that the shape test proves it reaches
  the model. **Their runtime behaviour is not unit-testable and is
  covered by dry-run inspection instead** — a dry run renders the
  prompt that would be sent, and the rendered prompt is read to confirm
  the verification, gate and logging instructions are present. This is
  a deliberate and visible gap: no automated test proves that a stage
  actually stops on a false premise. Closing it would require an
  end-to-end harness that runs a real model against a fixture
  repository, which this work does not build.

**Prior art.** The program parser in the engine repository has
table-driven parser tests covering section handling, and they are the
model for the shape test's structure. Within this repository, the house
pattern is table-driven tests on the standard library only, using
temporary directories for filesystem isolation, with no external test
framework. The shape test follows both.

**A regression fixture for the triggering case.** The premise of issue
#30 is kept as a fixture for the shape and block tests, since it is a
known-false claim with known evidence and makes a clean negative case.

## Out of Scope

- **Pruning or capping the existing queue.** The 534 open proposals
  stay. No bulk close, and no limit on proposals per run.
- **A sweep of the fifty-six approved proposals.** They are checked when
  a stage next touches them, not ahead of time.
- **The approval, ingest, documentation-update and readme-update
  programs.** They have the same dropped-instruction defect. Only the
  documentation-update program is reachable from a workflow; the other
  three are unreferenced. They are named in this document so the gap is
  visible, and left for separate work.
- **The approval program's automation.** It describes auto-approval
  after a review window, but no workflow runs it, so approval is
  effectively manual today. Whether to wire it up is a separate
  decision and is not settled here.
- **Any change to the engine.** No release, tag, or version pin bump.
- **Ranking or scoring proposals.** The gate decides whether a premise
  holds, not whether an idea is worth doing.
- **Other target repositories.** App, site, extension and docs are
  untouched.
- **Issue #30 and its pull request.** The operator's decision is to
  salvage the first-commit attribution fix as its own change and close
  both. That is separate work and is not part of this document.
- **An end-to-end harness that runs a real model against a fixture
  repository.** Named in Testing Decisions as the reason a gap remains.

## Further Notes

The measurements behind the problem statement were taken directly
rather than inherited. On a thousand-file repository the original and
the replacement are level; on a ten-thousand-file repository the
replacement is consistently slower across interleaved runs. Git process
startup is roughly two thirds of the total in both cases, which is the
more useful finding: the diff strategy was never the bottleneck in this
path.

The first-commit defect is worth restating because it was found by
accident and is the only real bug in the episode. The removed fallback
compared against the well-known empty-tree object, and that object is
not present in a fresh repository — nor in this one. Both paths
therefore failed on a first commit, and attribution recorded zero
lines. Any work that touches this area should keep that fix.

Two facts about the pipeline are worth watching once the gate is in.
The first is the hit rate: nine merged pull requests against 544 filed
proposals. If the gate works, the ratio should move, and if it does not
move the problem is upstream of the premise. The second is the
rejection log: if it fills with ideas that look good, the bar is
catching the wrong thing, and the log is the only place that would show
it.

The sibling product remains a legitimate source of inspiration. The
failure was never that a changelog was read; it was that a changelog
was believed.
