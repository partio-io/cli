# Duplicate Search: Ranked Candidates

## Problem Statement

The proposer files one `minion-proposal` issue for each new idea that it
finds upstream. Before it files, it runs the duplicate check,
`minion-review dupes`. The review runs the same check before a build. The
check returns a candidate only for a strong match: an issue that cites the
same source item, or an issue with a title overlap of 0.6 or more. The
proposer program says: "The list is the whole check: do not search the
issues in another way." So each miss of the check is a miss of the
pipeline.

The check misses too many copies:

- Up to 2026-10-13, 120 proposals were closed as duplicates. The check
  found 78 of them, in the review sweep. The other 42 were found by hand,
  because the check returned no candidate for them.
- The misses continue. Since #750, the proposer filed 19 proposals, and 4
  of them were copies that the check missed: #757, #758, #768 and #771.
  Each one was closed by hand. Three of them are among the 42.
- #771 is a third copy of the `CLAUDE_CONFIG_DIR` idea (#708, #750). Its
  source item is different from theirs, and its title overlap with #750
  is 0.599, just under the cutoff. So the check returned nothing, and the
  proposer filed it.

No later step catches these copies. The proposer labels its issues
`minion-reviewed` when it files them, so the review sweep never picks
them. The review sees them only before a build.

The cutoff causes the misses, not the score. On the corpus of 2026-10-13
(640 proposals), all other proposals were ranked for each closed issue by
title overlap plus code-name overlap. The kept issue is in the top 10 for
28 of the 42 hand-found pairs, and for 70 of the 78 tool-found pairs. For
the 4 recent misses, the kept issue is at rank 1, 2, 2 and 3.

## Solution

The duplicate check returns the best candidates, not only the candidates
over a cutoff.

- The strong matches stay as they are.
- A ranked part adds the 10 other proposals with the best score. The score
  is the title overlap plus the overlap of the code names in the bodies.
  Code names are the inline code spans of a body, for example
  `FindSessionDir` or `postcommit.go`. Rare words and names weigh more
  than common ones.
- Each candidate shows its score, the words and code names that it shares
  with the query, and the start of its description. The agent judges most
  candidates from this output. It opens an issue only when the candidate
  can be the same idea.
- The proposer gives the check the text of the idea that it has at that
  step. The review gives the check only the issue number.
- Both programs judge a ranked candidate as they judge a strong match
  today. They also get rules for two cases that the text does not cover
  today: a candidate that was closed as a copy, and a candidate that the
  agent is not sure about.
- After the proposer files an issue, it counts the escaped backticks in
  the whole body, not only in the first lines.
- A recall test on the 120 labeled pairs keeps the found copies found. A
  later change cannot lower the recall in silence.

The full list of all proposal titles, for a judgment of meaning, is not
part of this change. It comes later, only if copies still get through.

## User Stories

Actors: the operator (jcleira), the proposer (the agent of the propose
program), the review (the agent of the review program), and a maintainer
(who changes the duplicate check later).

1. As the operator, I want the duplicate check to return the best candidates also when none reaches the cutoff, so that a copy with a weak title overlap is not filed.
2. As the operator, I want a copy with a different source item and a title overlap just under 0.6 (like #771) to be found, so that one idea has one issue.
3. As the operator, I want the strong matches to stay as they are, so that no copy that the check finds today gets lost.
4. As the operator, I want the ranked part to hold 10 candidates at most, so that each check stays small inside the turn limit of the proposer.
5. As the operator, I want the score to use the code names of the bodies as well as the title words, so that two proposals about the same function or file rank high also with different titles.
6. As the operator, I want rare words and names to weigh more than common ones, so that a shared `postcommit.go` or "hook" alone does not make two ideas look the same.
7. As the operator, I want a proposal with no code names to be ranked by its title alone, so that the check still works for it.
8. As the operator, I want a proposal with no source item (a changelog entry) to be ranked by title and code names, so that it is not a blind spot.
9. As the operator, I want the same output from two runs over the same issues, so that a second run does not change the candidates.
10. As the operator, I want equal scores to put the older issue first, so that the issue that most likely stays comes first.
11. As the proposer, I want each candidate to show its title, its state, its score, what it shares with my idea, and the start of its description, so that I can judge most candidates without one more tool call for each.
12. As the proposer, I want to open only the candidates that can be the same idea, so that I stay inside my turn limit.
13. As the proposer, I want to give the check the text of my idea (its description, acceptance criteria and premise from the ingest step), so that the code names of the idea count in the score.
14. As the proposer, I want to pass that text on standard input with a heredoc, so that backticks and quotes in it need no shell escapes.
15. As the proposer, I want to file no issue when a ranked candidate is the same idea, as for a strong match, so that the kind of match does not change the decision.
16. As the proposer, I want a candidate that was closed as a copy to stand for the issue that it names, so that I apply the rules to the issue that stays.
17. As the proposer, when I am not sure that a candidate is the same idea, I want to file the issue and name the candidate in a line of the body, so that no idea gets lost and a reader sees the possible copy.
18. As the proposer, I want one command that counts the escaped backticks in the whole body after I file, so that an escape after line 30 is not missed.
19. As the proposer, I want the count command to expect 0, and on any other count to write the body again and count again, so that the filed body is clean.
20. As the review, I want to give the check only the issue number, so that the tool reads the title, the body and the source items itself and leaves the issue out of the candidates.
21. As the review, I want the check to read an issue that the proposal listing does not hold, so that a build of an issue without the `minion-proposal` label still gets a check.
22. As the review, I want to judge a ranked candidate with the current rules (the first issue reviewed stays), so that the verdict keeps its form.
23. As the review, when I am not sure, I want to keep the issue and give the reason in the evidence, so that no issue closes on a weak guess.
24. As the operator, I want every candidate in the review verdict with a short reason, so that I can audit each decision.
25. As the proposer, I want a clear error for a bad mix of flags (for example `--issue` with `--title`), so that I correct the call and do not act on a wrong list.
26. As a maintainer, I want the JSON output to keep its current fields and only add new ones, so that a reader of the old form still works.
27. As a maintainer, I want a recall test on the 120 labeled pairs, so that a change to the tokenizer, the weights or the code names cannot lower the recall in silence.
28. As a maintainer, I want the recall test to print the count of found pairs and the floor when it fails, so that I see how far a change moved the recall.
29. As a maintainer, I want one test case for each code-name rule (escaped backticks, fenced blocks, paths, call parentheses, short and common names, spans with spaces), so that each rule is pinned.
30. As the operator, I want a real dry run of both programs with the new text, so that I know that the minions parser gives the text to the agent.
31. As the operator, I want the change to be live at the next proposer run after the merge, with no minions release and no pin change.
32. As the operator, I want the PR body to give a command that counts the proposals that were closed by hand as copies after the merge, so that I can decide about the full title list with data.

## Implementation Decisions

**Language.** Go, because the repository is in Go. No new dependency.

**Delivery.** One PR. Table-driven tests. All program text goes in the
`## Agents` prose, because the minions parser drops every other H2 section
and every `####` heading.

**Modules.**

- **Code names** (new, in the `review` package). It takes a body and
  returns the set of its code names. The rules come from the 2026-10-13
  analysis. An escaped backtick counts as a backtick. Fenced blocks are
  left out. Inline spans on one line count. Each span is trimmed and put
  in lower case, and a `()` at the end is cut. A span with a `/` and no
  space becomes its base name. Spans under 3 bytes, spans of digits only,
  and a short list of common names (partio, true, false, nil, git, go) are
  dropped. A span with spaces stays whole.
- **Weights** (changed from the title-word weights). One weight table for
  title words and one for code names. Each is built once from the listed
  proposals with the current formula: the weight of an item is
  ln((n+1)/(df+1)) + 1, and the overlap is the weight of the shared items
  divided by the weight of the smaller set. The sums run in sorted order,
  so that the same input gives the same score on every run.
- **Duplicate search** (the changed `FindDupes`). The query gets two new
  optional parts: a body text and an issue number. With an issue number,
  the search takes the title, the body and the source items of that issue
  from the listing, or reads the issue when the listing does not hold it,
  and it leaves the issue out. The search returns the strong matches
  first, as today. Then it returns up to 10 ranked candidates: the
  proposals with the best title + code score above zero that are not
  strong matches. Equal scores sort by issue number, older first. Each
  candidate also gets its score and an excerpt: the start of its
  description as plain text, about 300 characters, cut at a word.
- **The `dupes` command** (changed). New flags: `--body-file <path>`, where
  `-` reads standard input, for the text of the query; and `--issue <n>`.
  `--issue` takes no `--title`, `--source`, `--body-file` or `--exclude`.
  Without `--issue`, `--title` stays required. The JSON output keeps its
  fields and adds `score` (a number) and `excerpt` (a string). `match` gets
  a third value, `ranked`. For a ranked candidate, `shared` holds the
  shared title words and code names.
- **The proposer program** (changed text). The duplicate step passes the
  idea text with `--body-file -` and a heredoc. It names the three kinds of
  match. It tells the agent to read the title and the excerpt of each
  candidate, and to open only the candidates that can be the same idea.
  "The list is the whole check" stays. The read-back step gets the count
  command in a fenced block.
- **The review program** (changed text). The duplicate step calls the
  command with `--issue <n>`. Every candidate stays in the verdict with
  "same" and "why".
- **Recall fixture** (new test data). The 2026-10-13 corpus (640
  proposals: number, title, state, state reason, body), gzipped; the
  pre-merge bodies of the kept issues that the analysis used; and the two
  pair lists (42 hand-found pairs, 78 tool-found pairs).

**K = 10.** The top 10 by title + code finds 28 of the 42 hand-found pairs
and 70 of the 78 tool-found pairs. The top 20 finds about 32 of the 42.
10 keeps the output small for the proposer, which is at its turn limit:
one of its last 8 runs stopped at the limit.

**The proposer checks once, at the same step as now.** The check stays
before the premise, so a copy costs no premise work. At that step the
agent has the idea from the ingest step (title, source, description,
acceptance criteria, premise), not the full body. So it has fewer code
names than the analysis used, and the live recall can be a little lower
than the measured recall. The title score alone puts 26 of the 42
hand-found pairs in the top 10, so the loss is small. No second check
after the premise: the proposer has no spare turns.

**Rules for the programs.**

- A ranked candidate that is the same idea counts as a strong match does:
  the proposer files no issue, and the review applies its current rules.
- A candidate that was closed as a copy stands for the issue that it
  names. Both programs apply their rules to that issue. The review has
  this rule today; the proposer gets it.
- Not sure: the proposer files the issue and adds the line `Possible
  duplicate of #<n>: <reason>` above the source line. The review keeps
  the issue, with `"same": false` and the reason in "why".

**No match by proposal id.** The id is a slug that the model makes on each
run. Only 2 pairs share an id (#757 and #461, #771 and #708), and the
ranked part already puts both in its top 10.

**No gate change.** The verdict keeps its form. The gate does not compare
the verdict with the candidate list.

**No change to the e2e-audit program.** It does not use this check.

**Live at the next run.** The workflows read the programs from `main` and
run the command with `go run`, so a merge is live at the next proposer
run. No minions release and no pin change.

## Testing Decisions

**What makes a good test.** A test asserts what a caller sees: the
candidates and their fields for a given set of issues and a given query;
the exit code and the output of the command for given flags; and the text
that reaches the model. A test does not assert how the score is computed
inside. No test judges model behaviour.

**Seams.** This change uses three current seams and adds none.

- **The `review` duplicates operation, through the fake GitHub server.**
  The code names, the weights, the ranked part, the excerpt and the
  issue-number query are all tested here, through the candidates that the
  operation returns. The recall test uses this seam too: the fake server
  serves the fixture, and each pair runs as an issue-number query for the
  closed issue.
- **The `dupes` command test.** It covers the new flags: the body text
  from a file and from standard input, `--issue`, the bad mixes of flags,
  and the new JSON fields.
- **The repo tests for programs.** They read the program files and assert
  the contract text.

**Every module is tested.**

- **Code names**, through the duplicates operation. An escaped backtick
  still makes a span. A name inside a fenced block does not count. A path
  counts by its base name. A `()` at the end is cut. Short names, names of
  digits only and common names do not count. A span with spaces stays
  whole. Each case shows in `shared` of a ranked candidate.
- **Weights**, through the duplicates operation. A rare shared name ranks
  a candidate above a common shared name. Two runs over the same issues
  give the same scores in the same order.
- **Duplicate search.** Strong matches come first and do not come again in
  the ranked part. The ranked part holds 10 candidates at most. A
  candidate with a score of zero is not returned. Equal scores put the
  older issue first. The issue-number query takes the title, the body and
  the source items from the issue and leaves the issue out. An issue that
  the listing does not hold is read once. Each candidate has a score and
  an excerpt. The excerpt has no heading marks and no escapes, and it
  stops at a word.
- **Recall**, through the duplicates operation on the fixture. The output
  (strong or ranked) holds the kept issue for at least as many pairs as
  the first run of the test measures. The floors are written as counts,
  one for the hand-found pairs and one for the tool-found pairs. The
  expected values are near 28 of 42 and 70 of 78; they can be higher,
  because the output also holds the strong matches. On a failure, the
  test prints the counts and the floors.
- **The `dupes` command**, as in the seam above.
- **The proposer program.** The duplicate step passes the idea text on
  standard input. It names the ranked kind of match. It keeps "The list is
  the whole check". It has the rule for a candidate closed as a copy and
  the rule for "not sure". The read-back step has the count command and
  expects 0. The current assertions (the order of the steps, the same-idea
  bullet, the heredoc on create, "with no escapes") still pass. They change
  only where the text changes on purpose.
- **The review program.** The duplicate step uses `--issue`. Every
  candidate stays in the verdict. The rule for "not sure" says keep.

**Left untested.** The judgment of the agent on a candidate is model
behaviour, so no unit test covers it. The dry run proves only that the
text reaches the agent. The count of copies closed by hand after the merge
measures the judgment.

**Prior art.** The current duplicates tests, with the fake server and real
bodies from test data. The command test, which sets the environment and
calls the subcommand. The program tests in the premise package for the
proposer and for the review duplicate step.

## Out of Scope

- The full list of all proposal titles, for a judgment of meaning. It
  comes later, only if copies still get through.
- A second duplicate check in the proposer, after the premise.
- A match by equal proposal id.
- The old bodies that a review "keep" leaves unchanged (escapes, old
  headings, a dead program line as the last line). A separate PR later.
- The e2e-audit program.
- The label `minion-reviewed` that the proposer adds at creation, and the
  schedule of the review sweep.
- Facets of a dropped copy, for example its new source item, on the kept
  issue.
- The turn limit of the proposer.
- The gate and the minions runtime.

## Further Notes

**Data and tools.** The analysis is outside this repository, in
`review-sweep-tools/dupes-analysis/` of the workspace: `corpus.json` (the
640 proposals of 2026-10-13), `pairs-github-dup.tsv` (the 42 hand-found
pairs, from the `marked_as_duplicate` events), `pairs-tool.txt` (the 78
pairs that the review found), and `dupesig/` (the Go analysis). The
pre-merge bodies are in `review-sweep-tools/cache/`. The fixture comes
from these files. The gzipped corpus is about 0.5 MB.

**Measured numbers** (2026-10-13; rank of the kept issue among all other
proposals, for the closed issue):

| Pairs | Score | top 5 | top 10 | top 20 | median rank |
|---|---|---|---|---|---|
| 42 hand-found | title | 0.57 | 0.62 | 0.69 | 5 |
| 42 hand-found | code names | 0.26 | 0.48 | 0.64 | 13 |
| 42 hand-found | title + code | 0.57 | 0.67 | 0.76 | 4 |
| 78 tool-found | title | 0.77 | 0.86 | 0.91 | 2 |
| 78 tool-found | title + code | 0.81 | 0.90 | 0.92 | 2 |

In the analysis, the values moved by about 0.01 from run to run, because
the sums followed the map order. The tool sums in sorted order.

**The 14 hand-found pairs that the ranked part does not find.** They share
only common words or files (hook, checkpoint, `postcommit.go`). All are
old: the closed issues are #347 to #721. An example is #700 "Post-commit
hook has no total deadline" against #212 "Add configurable hook execution
timeout", at rank 46. Only a judgment of meaning finds these. That is the
job of the full title list, if it comes later.

**The count after the merge.** The PR body gives a `gh` command that lists
the proposals filed after the merge and closed by hand as copies. A count
by state reason alone can miss a copy: #758 was closed as completed, not
as a duplicate. With this count, the operator decides about the full
title list.

**The dry run.** Install minions v0.0.14 into a scratch directory, then run
`minions run <program> --dry-run --project .minions/project.yaml` from the
repository, with `WORKSPACE_ROOT` set to the directory that holds the
repository and `GH_TOKEN` set. Do this for the propose and the review
programs. The new text must be in the agent instructions of the output,
and the text after it must still be there.

**Turn budget.** The proposer often logs more turns than its limit of 80,
and one of its last 8 runs stopped at the limit. The ranked part adds
output, not calls: one command gives one result. The agent opens only the
likely candidates.
