---
id: e2e-model-name-checkpoint-flow
target_repos:
  - cli
acceptance_criteria:
  - "A test initialises a real git repo in a temp directory"
  - "It writes a Claude Code JSONL session file with an assistant entry whose `message` field is a JSON object containing a `model` key (e.g. `claude-sonnet-4-6`)"
  - "It writes a pre-commit state file (as runPreCommit would) and makes a real git commit via --no-verify"
  - "It calls runPostCommit directly against the temp repo"
  - "After the hook runs, it reads the checkpoint via checkpoint.Read and asserts Checkpoint.Model equals the model from the JSONL"
  - "It runs `partio explain <checkpoint-id>` (or calls the explain logic directly) and asserts the model name appears in the output"
  - "A second sub-test supplies a JSONL with no model field and asserts checkpoint.Model is empty and explain shows 'unknown'"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Build an end-to-end test: model name flows from JSONL through commit hook to explain output

## Background

PR partio-io/cli#622 adds model name tracking: the JSONL parser extracts the
model name from the `message.model` field on assistant entries, propagates it
through `SessionData.Model → Checkpoint.Model → Metadata.Model`, and the new
`partio explain` command displays it.

The PR adds no tests. The JSONL test fixture data does not include a `model`
field, so `ParseJSONL` is never exercised for model extraction. No test drives
the post-commit hook against a real JSONL file to verify the model reaches the
stored checkpoint. No test calls the `explain` logic at all.

Each component (JSONL parsing, struct propagation, checkpoint storage, explain
display) can behave correctly in isolation while the full pipeline fails — for
example if the `message` field in real Claude Code JSONL is a JSON-encoded
string rather than an inline object, the `json.Unmarshal` in `parse_jsonl.go`
would silently produce an empty model with no error, and unit tests of the
struct fields would still pass.

## What to build

Add `internal/hooks/e2e_model_checkpoint_test.go` (package `hooks`) that
drives the real post-commit hook with a synthetic JSONL session file and
verifies the model name appears in the stored checkpoint and in explain output.

### JSONL format to use

The model is embedded in the `message` field of assistant entries as a
JSON-encoded object (based on `internal/agent/claude/parse_jsonl.go`):

```jsonl
{"type":"human","role":"human","message":"Build a widget","timestamp":1700000000,"sessionId":"sess-model-test"}
{"type":"assistant","role":"assistant","message":{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"Done."}]},"timestamp":1700000010}
```

The `message` field is `json.RawMessage` in `jsonlEntry`, so this JSON object
form is what `assistantMessage` unmarshals from.

### Flow to exercise

1. Create a temporary directory, initialise a real git repo (`git init`, set
   `user.email`, `user.name`), make an initial empty commit so HEAD exists.
2. Write the JSONL session file to the path where `ParseJSONL` will be called.
   The session discovery path must be set so the hook finds it — either write
   the file to the conventional Claude Code session directory (`~/.claude/projects/<hash>/`)
   or stub the session discovery by writing the JSONL path directly into the
   pre-commit state JSON (whichever matches how `runPostCommit` resolves the
   session path).
3. Write the pre-commit state file to `.partio/state/pre-commit.json` with
   `agent_active: true`.
4. Stage a file and make a real commit with `--no-verify`.
5. Call `runPostCommit(dir, cfg)` directly.
6. Read the created checkpoint using `checkpoint.Read(checkpointID)` (discover
   the checkpoint ID from the commit trailer via `git log --format=%B -1`) and
   assert `cp.Model == "claude-sonnet-4-6"`.
7. Assert `cp.Metadata.Model == "claude-sonnet-4-6"` (via `cp.ToMetadata()`).
8. Call `runExplain` (the unexported function in `cmd/partio/explain.go`) or
   reproduce its output logic directly, and assert the output contains
   `Model:      claude-sonnet-4-6`.

### Sub-test: missing model degrades gracefully

Repeat the flow with a JSONL file that has no `model` key in the assistant
`message` object. Assert:
- `cp.Model == ""`
- The explain output contains `Model:      unknown`

### Where the test lives

`internal/hooks/e2e_model_checkpoint_test.go`, package `hooks` (same package
as `postcommit.go` so it can call `runPostCommit` directly).

For the explain assertion, either:
- Add an `internal/checkpoint/explain_test.go` that directly tests the display
  logic given a checkpoint struct with and without `Model` set, or
- Accept a captured stdout from a subprocess call to the built binary.

The simpler route is a unit test of the explain output logic (since `cmd/partio/explain.go`
only formats fields — no git I/O), paired with the e2e hook test for the
storage half.

### Notes

- The checkpoint ID appears in the commit message as a `Partio-Checkpoint:`
  trailer. Parse it with `git log --format=%B -1 | grep Partio-Checkpoint`.
- `checkpoint.Read` lives in `internal/checkpoint/` and reads the orphan
  branch `partio/checkpoints/v1`. The temp repo must have that branch for
  `runPostCommit` to write it — it is created by the checkpoint store on first
  write.
- If session discovery cannot be stubbed easily, write an integration helper
  that sets the `CLAUDE_SESSION_FILE` env var (if such a hook exists) or
  patches the config path — check `internal/agent/claude/find_session_dir.go`
  for the lookup logic.
- Do not add sleeps. The state file is written immediately before `runPostCommit`.
