# 05 — Escaped backticks in the whole body

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

After the proposer files an issue, it counts the escaped backticks in
the whole body with one command, and it expects 0.

Today the read-back of `.minions/programs/propose.md` (step 4, after
`gh issue create`) says: read the body back with `gh issue view
<number> --repo <this-repo> --json body --jq .body`, and write it again
if a backslash stands in front of a backtick. In the run that filed
#771, the agent sent that output through `head -30`, so it checked 30
lines only.

The new read-back text:

- The count command, in a fenced block:

  ```bash
  gh issue view <number> --repo <this-repo> --json body --jq '[.body | scan("\\\\`")] | length'
  ```

- The command prints the count of escaped backticks in the whole body.
  The agent expects `0`.
- On any other count, the agent writes the body again, without the
  backslashes, with `gh issue edit <number> --repo <this-repo>
  --body-file -` and the same quoted heredoc. Then it runs the count
  again.

The command was checked on 2026-10-14, in zsh and in bash. It printed
`54` for #758 (closed, with escapes) and `0` for #771.

The text stays in the `## Agents` prose of the `### proposer` agent,
with no `####` heading. A static test cannot prove that the text
reaches the agent; only a real dry run can.

## User stories covered

18, 19, 30 (for the propose program).

## Acceptance criteria

- [ ] The read-back step shows the count command in a fenced block, as above. `<number>` and `<this-repo>` are its only parts to fill.
- [ ] The step says that the count must be 0. On any other count, the agent writes the body again with `gh issue edit <number>` and the same heredoc, and counts again.
- [ ] `internal/premise/repo_propose_test.go` asserts the command, the expected 0, and the rule to write again and count again. The current assertions of `TestProposeProgramFilesTheBodyAsWritten` pass.
- [ ] Live check, read-only: the command, copied from the program file with the parts filled, prints `0` for #771 and a count above 0 for #758.
- [ ] A real dry run of the propose program prints the count command in the agent instructions, and the text of steps 5, 6 and 7 after it. See "Dry run" below.
- [ ] `make test` and `make lint` pass.

## Dry run

```bash
GOBIN=<scratchpad>/bin go install github.com/partio-io/minions/cmd/minions@v0.0.14
cd /home/arvos/Code/workspaces/workspace-minion-reboot/cli
GH_TOKEN=$(gh auth token) WORKSPACE_ROOT=/home/arvos/Code/workspaces/workspace-minion-reboot \
  <scratchpad>/bin/minions run .minions/programs/propose.md --dry-run --project .minions/project.yaml
```

`<scratchpad>` is the scratchpad of the session. In the output, check
the agent instructions: the fenced count command, then the steps after
it.

## Modules touched

- The proposer program (`.minions/programs/propose.md`)
- Its repo tests (`internal/premise/repo_propose_test.go`)

## Test prior art

- `internal/premise/repo_propose_test.go`: `TestProposeProgramFilesTheBodyAsWritten` and the helper `fencedBlockContaining`.

## Out of scope

- The old bodies that a review "keep" leaves unchanged, with escapes (32 of the 640 proposals of 2026-10-13 had them): a separate PR later (PRD, Out of Scope).
- The duplicate step: [04 — The proposer checks with the text of its idea](./04-proposer-checks-with-its-idea-text.md).
