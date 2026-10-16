# 03 — Code names in the score

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — The review checks by issue number](./02-review-checks-by-issue-number.md)

## What to build

The score of a candidate adds the overlap of the code names of the
bodies to the title overlap. A query gets its body from `--issue`
(slice 02) or from the new flag `--body-file`.

**Code names** (new, in the `review` package, a file of its own). A
function takes a body and returns the set of its code names: the
inline code spans of the body. The rules come from the analysis of
2026-10-13 (`codeNames` in
`/home/arvos/Code/workspaces/workspace-minion-reboot/review-sweep-tools/dupes-analysis/dupesig/main.go`):

1. An escaped backtick (`` \` ``) counts as a backtick.
2. Fenced blocks are left out.
3. An inline span on one line counts.
4. Each span is trimmed and put in lower case, and a `()` at the end
   is cut.
5. A span with a `/` and no space becomes its base name.
6. Spans under 3 bytes, spans of digits only, and the common names
   `partio`, `true`, `false`, `nil`, `git` and `go` are dropped.
7. A span with spaces stays whole.

**Weights.** One weight table for title words (as today) and one for
code names. Each is built once for each search, from the listed
proposals, with the current formula: the weight of an item is
ln((n+1)/(df+1)) + 1, and the overlap is the weight of the shared
items divided by the weight of the smaller set. The `wordWeights` type
already does this for any set of strings, and slice 01 made its sums
run in sorted order.

**The score.** The score of a candidate is its title overlap plus its
code-name overlap. A query or a proposal with no code names gets its
title overlap alone. A strong title match stays a title overlap of 0.6
or more, on the title alone. For a ranked candidate, `shared` holds the
shared title words and the shared code names.

**The query body.** `DupesQuery` gets a body text. The body of an
`--issue` query is the body of that issue. The text of `--body-file`
gives only code names: the source items still come from `--source`.

**The command.** `--body-file <path>` reads the text of the query from
a file, and `--body-file -` reads standard input, so that the proposer
can pass its idea in a quoted heredoc with no shell escapes. `--issue`
takes no `--body-file`: add it to the rule of slice 02. A file that the
command cannot read exits 2, with the path and the cause. The command
function takes standard input as a parameter, so that the test can give
it a reader.

On the data of 2026-10-13, the analysis measured title + code, top 10,
at 0.67 of the 42 hand-found pairs and 0.90 of the 78 tool-found pairs,
against 0.62 and 0.86 for the title alone. A model of this search on
the live backlog of 2026-10-14 put #750 first for `--issue 771` (score
1.08), with `claude_config_dir` among the shared code names.

## User stories covered

5, 6, 7, 8, 13 (the flag), 14 (the flag), 25 (for `--body-file`), 29.

## Acceptance criteria

- [x] The score of a candidate is its title overlap plus its code-name overlap, with one weight table for each, built once for each search from the listed proposals. A strong title match still needs a title overlap of 0.6 or more.
- [x] For a ranked candidate, `shared` holds the shared title words and the shared code names, in a stable order.
- [x] A query or a proposal with no code names is ranked by its title overlap alone.
- [x] A rare shared code name ranks a candidate above a candidate with a common shared code name.
- [x] A span in escaped backticks still counts, and a name inside a fenced block does not count.
- [x] A path counts by its base name, and a `()` at the end is cut.
- [x] Spans under 3 bytes, spans of digits only and the common names do not count. A span with spaces stays whole.
- [x] Each code-name case above shows in `shared` of a ranked candidate, through `FindDupes` and the fake server.
- [x] The body of an `--issue` query gives its code names.
- [x] `--body-file <path>` and `--body-file -` give the text of the query, and the command test covers both. A file that does not exist exits 2 with its path. `--issue` with `--body-file` exits 2.
- [x] Live check, read-only: `GH_TOKEN=$(gh auth token) GITHUB_REPOSITORY=partio-io/cli go run ./cmd/minion-review dupes --issue 771` prints #750 as the first ranked candidate, with `claude_config_dir` in its `shared`.
- [x] `make test` and `make lint` pass.

## Modules touched

- Code names (a new file in `internal/review/`)
- Weights (`internal/review/title_words.go`, or a file of its own when the type now serves both tables)
- Duplicate search (`internal/review/dupes.go`: the query body and the score)
- The `dupes` command (`cmd/minion-review/main.go`: `--body-file` and standard input)

## Test prior art

- `internal/review/dupes_test.go`: the fake backlog. Give each case a body with the span under test, and assert `shared` of the ranked candidate (PRD, Testing Decisions: the code names are tested through the duplicates operation).
- `internal/review/source_refs_test.go`: a table-driven test of a text rule, for the layout of the cases.
- `cmd/minion-review/main_test.go`: the command test with `t.Setenv`.

## Out of scope

- The proposer program: [04 — The proposer checks with the text of its idea](./04-proposer-checks-with-its-idea-text.md).
- The recall test: [06 — Recall test and the count after the merge](./06-recall-test-and-count-after-merge.md).
- Body words that are not code names: the score uses only titles and code names (PRD).
- A match by equal proposal id (PRD, Out of Scope).

## Evidence

- 1 red: TestFindDupesAddsTheCodeNameOverlapToTheScore — code_names_test.go:38: code-name part of the score = 0, want 1
- 1 green: TestFindDupesAddsTheCodeNameOverlapToTheScore — --- PASS: TestFindDupesAddsTheCodeNameOverlapToTheScore (0.00s)
- 2 red: TestFindDupesAddsTheCodeNameOverlapToTheScore — code_names_test.go:41: shared = [claude], want [claude claude_config_dir]
- 2 green: TestFindDupesAddsTheCodeNameOverlapToTheScore — --- PASS: TestFindDupesAddsTheCodeNameOverlapToTheScore (0.00s)
- 3 red: TestFindDupesRanksByTitleAloneWithNoCodeNames — code_names_test.go:78: #760 1, #761 1 [faster upload], want #761 one lower with [faster upload]
- 4 red: TestFindDupesRanksARareCodeNameAboveACommonOne — code_names_test.go:101: rare score 1, common score 1, want rare higher
- 5 red: TestFindDupesSharesEachCodeNameCase/escaped_backticks_count — code_names_test.go:134: shared = [tweak], want [tweak escaped_name]
- 6 red: TestFindDupesSharesEachCodeNameCase/path_counts_by_its_base_name — code_names_test.go:134: shared = [tweak], want [tweak post_commit.go]
- 7 red: TestFindDupesSharesEachCodeNameCase/a_span_with_spaces_stays_whole — code_names_test.go:134: shared = [tweak], want [tweak git commit --amend]
- 8 red: TestFindDupesSharesEachCodeNameCase — code_names_test.go:134: shared = [tweak], want [tweak kept_name]
- 3 green: TestFindDupesRanksByTitleAloneWithNoCodeNames — PASS TestFindDupesRanksByTitleAloneWithNoCodeNames (0.00s)
- 4 green: TestFindDupesRanksARareCodeNameAboveACommonOne — PASS TestFindDupesRanksARareCodeNameAboveACommonOne (0.00s)
- 5 green: TestFindDupesSharesEachCodeNameCase — PASS TestFindDupesSharesEachCodeNameCase (0.00s)
- 6 green: TestFindDupesSharesEachCodeNameCase — PASS TestFindDupesSharesEachCodeNameCase (0.00s)
- 7 green: TestFindDupesSharesEachCodeNameCase — PASS TestFindDupesSharesEachCodeNameCase (0.00s)
- 8 green: TestFindDupesSharesEachCodeNameCase — PASS TestFindDupesSharesEachCodeNameCase (0.00s)
- 9 green: TestFindDupesByIssueTakesTheCodeNamesOfItsBody — PASS TestFindDupesByIssueTakesTheCodeNamesOfItsBody (0.00s)
- 9 red: TestFindDupesByIssueTakesTheCodeNamesOfItsBody — code_names_test.go:149: #750 not in []
- 10 red: TestDupesTakesTheQueryTextFromABodyFile — cmd/minion-review/main_test.go:38:28: too many arguments in call to dupes
- 10 green: TestDupesTakesTheQueryTextFromABodyFile — --- PASS: TestDupesTakesTheQueryTextFromABodyFile (0.00s)
- 11 smoke: dupes --issue 771 (live, read-only) — first ranked #750 score 1.08, shared has claude_config_dir
- 12 smoke: make test && make lint — make test exit 0; golangci-lint 0 issues

## Report

**What changed**
- New `internal/review/code_names.go`: `codeNames(body)`, with the seven rules of the analysis of 2026-10-13.
- `internal/review/dupes.go`: one weight table for code names next to the table for titles. The score is the title overlap plus the code-name overlap. A strong match still needs a title overlap of 0.6 or more. `shared` holds the title words, then the code names, each sorted. The query body comes from `--issue` or from `Body`.
- `cmd/minion-review/main.go`: `--body-file <path>` and `--body-file -`. `dupes` takes standard input as a parameter. A file that the command cannot read exits 2 with the path and the cause. `--issue` refuses `--body-file`.
- Tests: new `internal/review/code_names_test.go` (5 tests, 11 code-name cases). `cmd/minion-review/main_test.go`: 2 new tests, and `--body-file` added to the `--issue` rule test.
- `title_words.go` did not change: `wordWeights` serves both tables as it is.

**Suite.** `make test` and `make lint`: green at step 1 and at step 5 (lint: 0 issues). No repair outside "Modules touched".

**Live check** (read-only, run again after the review fixes): `dupes --issue 771` gives #750 first, score 1.08, with `claude_config_dir` in `shared`.

**Evidence records:** 10 red, 10 green, 2 smoke, 0 prefactor, 0 screenshot. I wrote `codeNames` in full for the tracer test. So the red runs of criteria 3 to 9 come from the new tests against the `dupes.go` of `HEAD`, which I restored for that run.

**Review findings** (`code-review medium --fix`)
1. Escapes in names: other backslash escapes stayed in a name. The review fixed it, and undid every Markdown escape. Removed: rule 1 asks only for escaped backticks, and the spec check flagged it. Ask item.
2. Body scans: the code names of the proposals are now built only when the query has code names. Kept: the result does not change, and the table is still built once for each search.
3. Evidence lines: the masker had changed three `--- PASS:` lines. The review restored them. Kept.
4. Score formula: one common shared code name, on a side with one code name only, gives the full code-name score of 1. Skipped by the review: criterion 1 sets the formula. Ask item.
5. Fence pairing: an unclosed fence or a `~~~` fence lets code count. Skipped. Ask item.
6. Slash rule: `owner/repo#12` gives `repo#12`, and two `main.go` files share a name. Skipped. Ask item.
7. `--body-file -` waits on a terminal when nothing is piped in. Skipped: most CLIs do the same. Ask item.

**Spec findings**
- (a) none. (c) none.
- (b) The review fix that undid every backslash escape. Removed, see review finding 1.

**Marked with `git add -N`:** `internal/review/code_names.go`, `internal/review/code_names_test.go`.

**Check time:** about 2 minutes for the one parallel wait (08:09:35 to 08:11:53 CEST, which includes the removal of the spec finding).

**Against the PRD**
- A name can show twice in `shared`: once as a title word and once as a code name. For example, the live #750 has `claude` twice. This shows which part matched, but a reader can think it is a fault.
- The live `shared` lists are long, with up to 13 items. The proposer of slice 04 reads them.
