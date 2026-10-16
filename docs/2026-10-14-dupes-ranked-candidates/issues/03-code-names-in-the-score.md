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

- [ ] The score of a candidate is its title overlap plus its code-name overlap, with one weight table for each, built once for each search from the listed proposals. A strong title match still needs a title overlap of 0.6 or more.
- [ ] For a ranked candidate, `shared` holds the shared title words and the shared code names, in a stable order.
- [ ] A query or a proposal with no code names is ranked by its title overlap alone.
- [ ] A rare shared code name ranks a candidate above a candidate with a common shared code name.
- [ ] A span in escaped backticks still counts, and a name inside a fenced block does not count.
- [ ] A path counts by its base name, and a `()` at the end is cut.
- [ ] Spans under 3 bytes, spans of digits only and the common names do not count. A span with spaces stays whole.
- [ ] Each code-name case above shows in `shared` of a ranked candidate, through `FindDupes` and the fake server.
- [ ] The body of an `--issue` query gives its code names.
- [ ] `--body-file <path>` and `--body-file -` give the text of the query, and the command test covers both. A file that does not exist exits 2 with its path. `--issue` with `--body-file` exits 2.
- [ ] Live check, read-only: `GH_TOKEN=$(gh auth token) GITHUB_REPOSITORY=partio-io/cli go run ./cmd/minion-review dupes --issue 771` prints #750 as the first ranked candidate, with `claude_config_dir` in its `shared`.
- [ ] `make test` and `make lint` pass.

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
