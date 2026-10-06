# Proposal Review

## Problem Statement

The operator cannot trust the text of a proposal, and the pipeline builds
on that text anyway.

**Issue #31 shows the failure from end to end.** The proposer filed it on
2026-03-26 from a sibling product's changelog (entireio/cli#533). Its
description asks for "a merge strategy appropriate for append-only
checkpoint data". The checkpoint data is not append-only: `partio prune`
removes checkpoints, and `partio reset` deletes and recreates the
checkpoint branch. The build-time premise gate passed the issue twice, on
2026-09-24 and on 2026-09-25. The first run names two claims; the second
shows two before its log is cut. No claim is about how checkpoints change.
The premise checker inside research extracted no claims at all, because
the issue had no premise block. Research then designed a union merge on
the false premise, and the build delivered it as PR #727, with every
check green.

Reproduced against #727:

- The next agent commit after `partio prune` brings the pruned
  checkpoints back.
- Every agent commit between two pushes adds a merge commit to the
  checkpoint branch. Five writes after one push gave four merge commits.
- Every agent commit makes two SSH round trips to the remote. Each took
  2.1 to 3.9 seconds from the operator's workstation.
- `partio prune --dry-run` moves the branch.
- A normal `partio enable` on a second device prints a false warning and
  tells the user to retry.
- After a force-push to the remote, reconciliation fails on every later
  call.

**Two rules let the false premise through.**

- The verifier extracts only sentences that describe today's code. It
  skips what a proposal wants to build, and it skips the acceptance
  criteria. "Appropriate for append-only checkpoint data" sits inside a
  request, so the checker skipped it. But the phrase asserts a fact about
  today's data, and one search settles that fact.
- Research skips any issue that has no premise block. The August work
  removed this escape from the build gate and left it in research.

**The backlog has the same origin, and it is large.** On 2026-10-05 there
were 604 open proposals:

| Filed | Open |
|---|---|
| 2026-03 | 91 |
| 2026-04 | 125 |
| 2026-05 | 124 |
| 2026-06 | 48 |
| 2026-07 | 107 |
| 2026-08 | 76 |
| 2026-09 | 29 |
| 2026-10 | 4 |

Almost all of them derive from entireio/cli: 361 cite a pull request, 181
an issue, and 106 a changelog entry. Most were filed before the proposer
checked anything against this repository.

- 50 carry a verified premise block.
- 170 carry acceptance criteria as a checklist. 172 carry no criteria
  section at all.
- 500 point at a proposal file that does not exist on main. None of the
  486 distinct paths exists.
- 55 carry the approved label, but automation applied it at filing time,
  not the operator. On #31 it arrived one second after the proposal
  label, together with `minion-failed`. 44 carry both approved and
  failed, and 10 still carry `minion-executing`.

**A fix to one issue at a time does not scale.** The operator can correct
#31 by hand. The same proposer wrote the other 603, mostly under weaker
rules, and each one can steer research in the same way. The August
decision checks a proposal only when a stage touches it. So the text stays
wrong until a build starts, and at that point research plans from it.

**The proposer refills what a cleanup removes.** Its duplicate check
searches open issues only, by an id that the model invents on each run.
Confirmed pairs: #710 and #736, #714 and #720, #681 and #688. The
proposer rejected entireio/cli#2035 as premise-failed and still filed it
as #678. The proposer writes its rejection log, but nothing reads it.

**A passing check leaves no trace.** The runtime logs the first 500 bytes
of each agent report, and a passing gate writes nothing to the issue.
Nobody can see from #31 what the gate checked.

## Solution

Review every proposal once, against today's code, in a fresh session.
Keep only what today's proposer would file. Review again before each
build.

- **One bar, the proposer's bar.** An issue stays open only if today's
  proposer would file it: its premise holds against today's tree, it
  applies to Partio, it is not built yet, and it is not a duplicate. New
  proposals already pass this bar.
- **A failed issue is closed, with its evidence.** The review closes it
  with a comment that names each claim, its evidence and its verdict. The
  operator reopens an issue to overrule the review, and the review never
  closes a reopened issue again.
- **A kept issue is rewritten in today's format.** The review rewrites it
  in place, in the shape that today's proposer files: what to build, the
  acceptance criteria as a checklist, a verified premise block with
  evidence, and a proposal id. GitHub's edit history keeps the old text.
- **The sweep runs at night, after a dry run.** First, 20 issues from
  different months run as a dry run: the verdicts go to a report, and
  nothing changes on GitHub. The operator reads the report and then
  switches the sweep on. The sweep runs between 00:00 and 07:00 local
  time, so it does not compete with the operator's daytime Claude use.
- **Every build reviews its issue first.** A build stops when the review
  reaches no verdict. A build of a rewritten issue gets fresh research,
  because the old plan describes the old text.
- **The operator's own issues get a facts-only check.** The review checks
  an issue that the operator wrote, or reopened after a review closed it,
  for a false premise. It never closes or rewrites that issue.
- **The checks catch what #31 hid.** The verifier and the ingest prompt
  also extract the facts that a request or an acceptance criterion takes
  for granted. Research loses its escape for an issue without a premise
  block.
- **The proposer stops a second filing.** Its duplicate check matches the
  source item and similar titles across open and closed issues, so an
  idea that the review closed does not come back.
- **Every verdict leaves evidence.** Each reviewed issue gets a comment
  with its evidence, and one tracking issue lists every verdict.

**This reverses four August decisions, on purpose.** The premise-gate PRD
said that the backlog is not swept, that the machine never closes an issue
on its own, that an old proposal is never rewritten, and that the existing
proposals stay untouched. #31 is the evidence that a check at touch time
comes too late. The operator made the new decisions on 2026-10-05.

## User Stories

1. As the operator, I want every open proposal reviewed against today's code, so that the backlog shows only work that can still matter.
2. As the operator, I want each proposal reviewed in its own fresh session, so that the context of one issue never leaks into the verdict on another.
3. As the operator, I want the review to apply the same bar that the proposer applies to new ideas, so that old and new proposals meet one standard.
4. As the operator, I want a proposal with a false premise closed, so that research never plans on it again.
5. As the operator, I want a proposal that is already built closed, so that the backlog does not ask for work that exists.
6. As the operator, I want a proposal that does not apply to Partio closed, so that the internals of a sibling product stop filling the queue.
7. As the operator, I want a duplicate closed with a link to the issue that stays, so that one idea has one issue.
8. As the operator, I want a proposal that the review cannot verify closed as "could not verify", so that an unchecked idea never passes as a checked one.
9. As the operator, I want each close to carry a comment with every claim, its evidence and its verdict, so that I can audit a close from the issue alone.
10. As the operator, I want each close to use GitHub's state reason, completed for a built idea and not planned for the rest, so that GitHub's own filters separate the cases.
11. As the operator, I want to overrule a close by reopening the issue, so that a wrong verdict costs me one click.
12. As the operator, I want the review never to close a reopened issue again, so that my overrule stays in force.
13. As the operator, I want a proposal that still has value rewritten in today's proposal format, so that research and the build read a clear and current specification.
14. As the operator, I want the rewrite to keep the original idea and its source link, so that the review corrects the issue instead of replacing it with a different one.
15. As the operator, I want the rewrite to carry a verified premise block with evidence, so that every later stage checks the same claims again.
16. As the operator, I want the rewrite to carry the acceptance criteria as a checklist, so that the build has testable targets.
17. As the operator, I want the rewrite to remove the pointer to a proposal file that no longer exists, so that the issue carries the whole proposal.
18. As the operator, I want a comment that says what each rewrite changed, so that I can judge it without a diff of the edit history.
19. As the operator, I want GitHub's edit history to keep the old text, so that no rewrite destroys information.
20. As the operator, I want the review to remove the automatic March approval from a kept proposal, so that "approved" means that I chose the issue.
21. As the operator, I want the review to remove stale build labels from a kept proposal, so that the labels describe the current issue and not a run from March.
22. As the operator, I want a dry run on 20 issues from different months before anything changes, so that I can judge the verdicts before 600 of them act.
23. As the operator, I want the dry-run report to show each proposed rewrite in full, so that I can judge the rewrites as well as the closes.
24. As the operator, I want to add known cases such as #30 and #31 to the dry run, so that I can see whether the review catches premises that I already know are false.
25. As the operator, I want the sweep to act only after I switch it on, so that the dry run is a real gate.
26. As the operator, I want the sweep to run only at night, so that my daytime Claude usage stays free.
27. As the operator, I want the sweep to take approved proposals first and then the oldest, so that the issues nearest to a build and the issues most likely obsolete go first.
28. As the operator, I want the sweep to stop taking new issues at the end of the night window, so that it never runs into my day.
29. As the operator, I want one tracking issue that lists every verdict with links, so that I can follow the sweep and undo a bad night.
30. As the operator, I want an issue whose review failed twice listed for me and skipped, so that one broken issue does not use every night.
31. As the operator, I want to stop the sweep with one switch, so that I can halt it without a code change.
32. As the operator, I want each build to review its issue first, so that an issue filed months ago is current when research reads it.
33. As the operator, I want a build to stop when the review reaches no verdict, so that a crashed check never lets a build through.
34. As the operator, I want a build of a rewritten issue to run research again, so that no build follows a plan written for the old text.
35. As the operator, I want a build whose review closes the issue to stop without a failure mark, so that a closed proposal does not look like a broken run.
36. As the operator, I want my own issues checked for false premises only, so that the review never closes or rewrites what I wrote.
37. As the operator, I want a false premise in my own issue to stop the build with the blocking label and a comment, so that I can correct the text or overrule.
38. As the operator, I want the verifier to extract the facts that a request takes for granted, so that "appropriate for append-only checkpoint data" gets the same check as any other claim.
39. As the operator, I want the verifier to extract the facts that an acceptance criterion takes for granted, so that a criterion cannot hide a false premise.
40. As the operator, I want a sweeping claim such as "append-only" settled by a list of every path that changes the data, so that one counterexample is enough to fail it.
41. As the operator, I want the ingest prompt to put those assumed facts into the premise block of a new proposal, so that new issues carry them from the start.
42. As the operator, I want research to verify an issue that has no premise block, so that no stage exempts old proposals again.
43. As the operator, I want the duplicate check of the proposer to search closed issues too, so that an idea that the review closed is not filed again.
44. As the operator, I want the duplicate check of the proposer to match the source item, so that one upstream pull request never produces two issues.
45. As the operator, I want the proposer to mark what it files as reviewed, so that the sweep does not review a new proposal a second time.
46. As the operator, I want every review verdict to leave a comment, a pass included, so that a green verdict carries its evidence.
47. As the operator, I want an open pull request from an older build named in the review comment and left alone, so that I decide what happens to it.
48. As the operator, I want every GitHub change that the review makes to go through tested code, so that closes and rewrites follow rules and not model behaviour.
49. As the operator, I want the review program unable to change GitHub itself, so that a confused session can at worst write a bad verdict, which the gate rejects.
50. As the operator, I want a malformed rewrite treated as no verdict, so that a broken body never replaces a working one.
51. As the build stage, I want one verdict that says proceed or stop, and says whether the issue changed, so that the workflow needs no logic of its own.
52. As the research stage, I want to plan from an issue that passed the review, so that my design rests on checked facts.
53. As the proposer, I want one command that lists candidate duplicates, so that I check the backlog in the same way as the review.
54. As a Partio user, I want features built from correct assumptions about Partio, so that a new feature does not undo my prunes or slow my commits.
55. As the operator, I want the close counts per reason in the tracking issue, so that I can see whether the source, the verifier or the bar needs work.
56. As the operator, I want the whole change to ship in this repository, so that it needs no engine release and no version pin change.

## Implementation Decisions

**The bar is the proposer's bar.** An issue passes when four conditions
hold. Its premise holds against the checked-out tree. It applies to
Partio under the relevance rule of the ingest prompt (Git workflows, AI
agent sessions, code attribution, checkpoints). It is not built yet. It
is not a duplicate. The review does not judge priority or worth. The
operator chose this bar so that the backlog and new filings meet one
standard.

**Verdicts.** There are three outcomes: keep, rewrite and close. A close
carries one reason: built, false premise, does not apply, duplicate, or
could not verify. "Could not verify" closes the issue, because today's
proposer drops an idea that it cannot verify. A premise with no checkable
claim is not "could not verify": the current rule of the verifier
applies, and the other three conditions of the bar decide.

**One program judges, and one Go gate acts.** The review program runs one
fresh session per issue. It reads the issue, the source item and the
tree. It applies the verifier, and it judges fit, built and duplicate.
Then it writes one verdict file. It makes no change on GitHub, and it
writes nothing into its working directory, because a file there turns
the run into a pull request. The `minion-review` command reads the
verdict and makes every GitHub change. This follows the audit-gate
pattern that this repository already uses.

**The verdict file is the contract.** It names the issue and the
outcome. It gives the close reason and, for a duplicate, the issue that
stays. It lists every claim with the evidence that the claim named, its
verdict and the excerpt that decided it, and it gives the block verdict.
It records the fit decision with one line of reason and the built
decision with its evidence. It lists each duplicate candidate that the
program considered and why the candidate is or is not the same idea. For
a rewrite, it carries the new title, the new body and a summary of the
changes. The program writes the file to an absolute path that the
workflow supplies, outside the worktree of the session.

**The gate fails closed.** These cases are "no verdict": a missing file,
malformed JSON, an unknown outcome or reason, a close without evidence, a
duplicate without a target, and a rewrite whose body fails validation.
No verdict never counts as a pass.

**Rewrite validation.** A rewritten body must parse as the issue shape
of the proposer. It must have a premise block that the premise package
accepts (the marker is present, and every claim has evidence), an
acceptance-criteria section with at least one checklist item, a final
proposal id line, and the original source reference when the old body
had one (20 open proposals state no source, and 3 come from the e2e
audit). The title must not be empty.

**Go decides the mode, not the model.** The gate checks facts only when
the issue has no `minion-proposal` label, because then the operator wrote
it, or when someone reopened it after a review closed it. A `reopened`
event after the close comment of the review counts as the overrule of
the operator. The runner posts as the operator, so the gate cannot tell
the two actors apart, but the review itself never reopens an issue. In
facts-only mode the gate ignores keep, rewrite and close, and it applies
only the premise result. The program has one mode, and the guarantee
lives in tested code.

**What the gate does in each mode.**

- **Sweep, full check.** Keep and rewrite add `minion-reviewed`; remove
  `minion-approved`, `minion-failed`, `minion-executing` and
  `do-not-build`; and post the evidence comment. A rewrite also replaces
  the title and the body, and its comment lists the changes. A close
  posts the evidence comment, adds `minion-reviewed`, and closes the
  issue with state reason completed for a built idea and not planned for
  the other reasons. A duplicate close links the issue that stays.
- **Build, full check.** Keep proceeds. Rewrite applies the changes and
  proceeds with "changed". Close applies the close and stops the build as
  blocked, not as failed.
- **Build, facts only.** A premise that holds, or that has no checkable
  claim, proceeds with an evidence comment. A premise that fails or is
  unresolved gets `do-not-build`, removed first and then added as the
  stage gate does so that the label fires again. The gate posts the
  marked comment and stops the build. It never closes or edits an issue
  of the operator.
- **Dry run.** No issue changes. The gate adds one row per issue to the
  tracking issue, with the full proposed rewrite in a collapsed section.

**Labels.** `minion-reviewed` is a new label. The gate adds it to every
proposal that it keeps, rewrites or closes. The sweep selects open
proposals without it, so the sweep never selects a reopened issue again.
The proposer adds the label when it files, because it applied the same
bar.

**Duplicates.** The first issue reviewed stays.

- A candidate that is open and already reviewed is the issue that stays,
  and this issue closes as its duplicate.
- A candidate that is open and not yet reviewed lets this issue stay. The
  other one closes when its turn comes.
- A candidate that the review already closed makes this issue its
  duplicate, because the idea already has a verdict.
- A candidate closed as built makes this issue built.

**Code does the duplicate search.** `minion-review dupes` takes a source
reference and a title. It searches every `minion-proposal` issue, open
and closed. It returns each candidate with its state, state reason,
labels, review verdict, and the reason for the match: the same source
item (entireio/cli#N in any form, or an issue or pull request URL), or a
strong overlap of the normalized titles. The program decides whether a
candidate is the same idea. The review program and the proposer both use
the command.

**The sweep workflow.** A scheduled run starts at 23:00 UTC and takes no
new issue after 05:00 UTC. That stays inside 00:00–07:00 local time in
summer and in winter. The run acts only when the repository setting
`PROPOSAL_REVIEW_SWEEP` is `on`. The operator switches it on after the
dry run, and off to stop the sweep. For each issue, one job runs the
review program and then the gate, one issue after the other, so the
24-hour queue limit of GitHub never applies. A manual dispatch runs the
dry run, with a sample size or an explicit list of issues.
`minion-review next` picks the batch: approved proposals first, then the
oldest. It skips an issue that has two "no verdict" rows in the tracking
issue and lists it there under a "needs you" heading. The sample spreads
across the filing months.

**The tracking issue.** There is one tracking issue. The gate finds it by
a marker and creates it on first use. Each night gets one comment, with
one row per issue: number, title, verdict, reason and link. The issue
body keeps the totals per verdict and the "needs you" list. The runner
posts as the operator, so GitHub sends no notification. The operator
reads the issue.

**The build workflow.** The review replaces the premise gate as the first
step after the install. The output of the gate tells the workflow whether
to proceed and whether the issue changed. When the issue changed, the
workflow skips the slice-plan lookup and runs research. Research
publishes a new PRD and a new plan, and the newest plan wins, as the
runtime already does. No verdict fails the job, and the current failure
step marks the issue. The premise-gate program is removed.

**The verifier.** Its extraction rule gets a third kind of claim: a fact
that a requested behaviour, a design instruction or an acceptance
criterion takes for granted about today's code or data. The checker
quotes the phrase that carries the fact and states it again as a
checkable claim. A command that lists every path that changes the data
settles a sweeping claim, for example "append-only" or "nothing
deletes". The rule that skips requests now skips only the behaviour that
the request asks for, not the facts that the request rests on. The lines
that forbid a sweep and forbid a rewrite leave the verifier. Verification
stays a pure check, and each caller decides what it writes.

**The ingest prompt** gets the same rule, so a new proposal puts its
assumed facts into its premise block.

**The research program.** Its premise checker extracts claims from the
prose when an issue has no premise block, as the build gate already did,
instead of writing a pass. Research keeps its own stop behaviour from
the stage gate. This work does not move that behaviour into Go.

**The propose program.** Its duplicate check uses `minion-review dupes`
instead of a search of open issues by the generated id.

**Open pull requests.** When a reviewed issue has an open pull request
from an older build, the comment names it. The gate never changes a pull
request.

**Language and shape.** The `review` package and the `minion-review`
command are in Go, because this repository is in Go. One package with
three operations (gate, next and duplicates) and one command with three
subcommands keep the number of seams small. The review program keeps all
its instructions in its agents section, and the shape test enforces it.

**No engine change.** Everything ships in this repository. The current
minions version on the runner runs the new program as it is, so no
release, tag or pin change is necessary. Sessions use the default model
of the runner, as the premise gate does today.

**Throughput.** One runner runs one Claude session at a time. A premise
check took 59 to 96 seconds on recent builds, and a review does more. The
604 issues need several nights. The runner signs in with the
subscription of the operator, so the cost is usage limits and runner
time, not a bill.

## Testing Decisions

**What makes a good test.** A test asserts behaviour that a caller can
see. For the gate, that is the GitHub requests it makes and the decision
it returns, for a given verdict file and issue state, not how it builds
them. For program and workflow text, that is what reaches the model,
which workflow runs what and in which order, and the contract markers
and instructions whose removal is a regression, as the current repo
tests do. No unit test judges model behaviour.

**Seams.**

- **The `review` package, through a fake GitHub server.** This is a new
  seam, at the highest point: the gate, next and duplicates operations,
  each called against a fake GitHub API, as the tests of the audit gate
  do. Every GitHub write and the build decision sit behind it, so one
  seam covers them. The command is a thin wrapper and, like the other
  commands in this repository, has no tests of its own.
- **The repo tests for programs and workflows.** This is a current seam.
  The premise and program-shape packages already read the program and
  workflow files and assert on them. The new and changed text gets its
  tests there.
- **The premise package** is not a new seam. The gate uses its parser to
  validate rewrites, and the tests of the gate cover that use.

**Every module is tested.**

- **`review` gate.** It fails closed in each malformed case: a missing
  file, bad JSON, an unknown outcome or reason, a close without evidence,
  a duplicate without a target, and an invalid rewrite. Keep, rewrite and
  close each make exactly the expected requests. Each close reason gets
  its state reason. Labels are added and removed as decided. Facts-only
  mode never closes or edits, for an issue of the operator and for an
  issue reopened after a review close. A facts-only failure removes and
  adds `do-not-build` and posts the marked comment. Build mode returns
  proceed, stop and changed correctly. A dry run changes no issue and
  writes one tracking row with the proposed body.
- **`review` next.** Approved proposals come first, then the oldest.
  Issues with `minion-reviewed` are excluded. Two "no verdict" rows
  exclude an issue and list it. The sample spreads across months and is
  deterministic.
- **`review` duplicates.** It matches the same source item in each form
  (shorthand, issue URL, pull request URL). It finds closed issues. It
  orders candidates by the reason for the match. It returns nothing for
  an unrelated idea. It excludes the issue under review.
- **The review program.** It exists, keeps its instructions in the
  agents section, writes only the verdict file, makes no GitHub change,
  uses the duplicate command, and applies the verifier.
- **The sweep workflow.** It runs the program and then the gate for each
  issue. It acts only when the setting is on. It stops taking issues at
  the end of the window. Its dispatch offers the dry run, the sample size
  and the issue list.
- **The build workflow.** The review runs before research and the build.
  No verdict fails the job. "Changed" forces research. A close stops the
  build without the failure step. The premise-gate program is gone.
- **The verifier.** It carries the rule for assumed facts and the rule
  for sweeping claims. It no longer forbids a sweep or a rewrite.
- **The ingest prompt.** It carries the rule for assumed facts.
- **The research program.** It has no escape for an issue without a
  premise block, and it extracts claims from the prose. This adds the
  test that the August work added for the gate only.
- **The propose program.** It uses the duplicate command, searches closed
  issues, and adds `minion-reviewed` when it files.

**August tests that change meaning.** Some tests assert that the backlog
is not swept, that an old proposal is not rewritten, that a blocked build
never closes the issue, and that the build runs the premise gate. They
encode decisions that this PRD reverses, so they get a rewrite. "Never
closes" and "never rewrites" now hold for the issues of the operator and
for reopened issues. The build tests find the review step. Removal of the
label stays the overrule for a facts-only block.

**The visible gap.** No automated test proves that the review program
judges well. The dry run covers this gap: 20 issues across months plus
known cases (#30 and #31, whose false premises are on record), which the
operator reads before the sweep acts. To close the gap, a harness must
run a real model on fixture issues. This work does not build one.

**Prior art.** The tests of the audit gate (a fake GitHub server, the
fail-closed cases, one upserted comment), the checks-verdict tests, and
the repo tests in the premise and program-shape packages. The tests are
table-driven, use the standard library only and use temporary
directories, as the house style requires.

## Out of Scope

- **Closed proposals.** The sweep reads open issues only. Thirteen
  proposals are closed, nine with `minion-done`, and #31 is one of them.
  A closed issue gets its review when the operator builds it again.
- **Issue #31 and PR #727.** They are separate work: close #727, delete
  its branch, reopen #31, and build it again.
- **The source cursor of the proposer.** It reaches main only when the
  propose pull request merges, and #728 has been open since 2026-09-26.
- **A reader for the rejection log.** The duplicate checks now read the
  closed issues, where the verdicts of the review live.
- **The close of an issue when its pull request opens.** The build still
  marks an issue done at that point.
- **The stop behaviour of research in Go.** Research keeps the
  prompt-driven procedure of the stage gate.
- **Priority or ranking.** The operator chose the bar of the proposer,
  which judges facts and fit, not worth.
- **Automatic approval.** No workflow approves issues, and this work adds
  none.
- **Engine changes, model choice, and other target repositories.**
- **The four programs that keep instructions in dropped sections**
  (approval, ingest, readme update and documentation update). They stay
  a known gap.
- **The defects of PR #727** themselves.

## Further Notes

Known answers for the dry run: the false claim of #30 ("full tree walk",
refuted in the August PRD) and the "append-only" assumption of #31. A good
review rewrites #31 with a corrected premise and does not close it. The
reconciliation of diverged branches still has value; only the merge rule
must respect deletions.

Watch the close reasons. Many "could not verify" closes point at the
evidence gathering. Many "does not apply" closes point at the source: most
proposals copy entireio/cli pull requests, which often describe the
internals of that product. The August PRD recorded nine merged minion
pull requests against 544 filed proposals. After the sweep, that ratio
becomes measurable against a clean backlog.

The night window is in UTC because GitHub schedules run in UTC. 23:00 to
05:00 UTC is 00:00 to 06:00 local time in winter and 01:00 to 07:00 in
summer.

The 500-byte log limit stays. The verdict file and the evidence comment
are now the full record of every review.

## Decision Log

**2026-10-06, after the first dry run.** A failing claim closes an issue
only when the idea rests on it. When the problem that the idea solves
survives the correction, the review rewrites the issue on the correct
fact. Each failing claim in the verdict then carries a `correction`, and
the gate accepts a rewrite with a failing claim only when every failing
claim has one. A keep never carries a failing claim. #31 is this case:
its branches can still diverge, and only its merge rule took append-only
data for granted. This makes the known answer in Further Notes agree
with the verdict rules, which had closed every issue with a failing
claim.
