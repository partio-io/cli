# Rejections

Ideas the proposer did not file, and why. Each run appends; nothing here
is ever rewritten. The cursor has already moved past every item below, so
this file is the only record that the idea was seen at all.

## Protect against duplicate checkpoints when post-commit is killed mid-execution

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2035 (Checkpoint duplicated when a session-end hook is killed between the v1 write and the state save)`
- reason: `premise-failed`
- claim: Partio's post-commit hook can write a duplicate checkpoint for the same commit if the hook is killed between writing the checkpoint and saving the commit cache (`internal/hooks/postcommit.go:201–216`) [evidence: `internal/hooks/postcommit.go`]
- verdict: `fails`
- found: The pre-commit state file is deleted at line 38 of `internal/hooks/postcommit.go` — before any checkpoint is written (line 201). If the hook is killed after writing the checkpoint, the state file is already gone. Any subsequent re-invocation (e.g. from `git commit --amend`) finds no state file and returns nil at line 35 — no duplicate. The amend triggers pre-commit anew, which creates a new state file and a new commit hash; post-commit then writes a checkpoint for that new hash, which is a different commit, not a duplicate. Partio's design already prevents the duplicate-checkpoint scenario that Entire's issue described.

## Cursor: Cursor transcript <timestamp> wrapper leaking into prompt fields

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2065 (Cursor transcript-derived prompts leak <timestamp> wrapper tags)`
- reason: `irrelevant`
- note: Partio has no Cursor agent support. The only agents registered are claude-code and codex; Cursor integration does not exist in this codebase.

## Codex last_assistant_message discarded on Stop/SubagentStop

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2064 (Codex last_assistant_message is parsed then discarded on Stop/SubagentStop)`
- reason: `irrelevant`
- note: Partio's Codex session parser (`internal/agent/codex/parse_jsonl.go`) has no `stopRaw` or `subagentStopRaw` event types and does not track turn-end summaries. The bug described is specific to Entire's Codex event pipeline, which Partio does not replicate.

## Subagent transcripts skip image externalization before size cap

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2063 (Subagent transcripts skip image externalization before size cap / redaction)`
- reason: `irrelevant`
- note: Partio has no subagent transcript pipeline and no image externalization. The session capture writes a single JSONL file per commit; there is no multi-agent or image-embedding path.

## Claude Code parallel Task subagents misattribute TodoWrite incremental checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2062 (Claude Code parallel Task subagents misattribute TodoWrite incremental checkpoints)`
- reason: `irrelevant`
- note: Partio writes one checkpoint per git commit. It has no incremental TodoWrite checkpoints and no parallel subagent tracking. This class of attribution bug does not exist in Partio's architecture.

## Cursor subagentStop has no subagent_id; correlation key goes empty

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2061 (Cursor subagentStop has no subagent_id; Entire correlation key goes empty)`
- reason: `irrelevant`
- note: Partio has no Cursor support and no subagent correlation logic.

## Codex input_image base64 silently destroyed by JSONL redaction

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2060 (Codex input_image base64 silently destroyed by JSONL redaction)`
- reason: `irrelevant`
- note: Partio's redaction (`internal/redact/`) operates on text tokens and entropy-based heuristics; it has no special handling for Codex image payloads and no image types in its Codex parser. The redaction path described in the issue is Entire-specific.

## Subagent data does not survive condensation

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2058 (Subagent data does not survive condensation — durable storage for subagent transcripts)`
- reason: `irrelevant`
- note: Partio has no subagent tracking, no condensation pipeline, and no durable subagent transcript storage. The feature request is entirely specific to Entire's architecture.

## --checkpoint-remote supports only the github provider

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2033 (--checkpoint-remote supports only the github provider — no self-hosted/generic git remote option)`
- reason: `irrelevant`
- note: Partio stores checkpoints on a local orphan branch (`partio/checkpoints/v1`) using git plumbing commands. There is no configurable checkpoint remote provider; the remote is whatever `origin` the repo already has.

## Codex sessions not captured in linked worktrees (hooks written to worktree-local .codex/hooks.json)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2020 (Codex sessions are not captured in linked worktrees because hooks are installed in an ignored worktree-local location)`
- reason: `irrelevant`
- note: Partio uses git hooks (pre-commit, post-commit, pre-push) to capture Codex sessions, not Codex's own `.codex/hooks.json` mechanism. Git hooks are installed to `git rev-parse --git-common-dir`, which is shared across worktrees. Partio does not write to `.codex/hooks.json` at all.

## OpenCode Desktop app: hooks never fire (plugin uses Bun globals, desktop runs Node)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2014 (OpenCode hooks never fire in the OpenCode Desktop app)`
- reason: `irrelevant`
- note: Partio has no OpenCode support. There is no OpenCode agent, plugin, or hook configuration in this codebase.

## git-refs push queue can drop a newer same-ref update during an in-flight push

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2008 (git-refs push queue can drop a newer same-ref update during an in-flight push)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes the checkpoint branch synchronously via a `git push` call. There is no asynchronous push queue; the race described does not apply.

## Replace obsolete `entire configure --agent` guidance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2007 (Replace obsolete entire configure --agent guidance left by #1062)`
- reason: `irrelevant`
- note: Partio uses the `PARTIO_AGENT` environment variable and `agent` config field, not a dedicated `agent add/configure` command. The stale docs issue is specific to Entire's CLI command structure.

## OpenCode: preserve user prompts when message events arrive in sequence

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2001 (OpenCode: Preserve user prompts when message events arrive in sequence)`
- reason: `irrelevant`
- note: Partio has no OpenCode support.

## Discover untracked OpenCode sessions during session attach

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1992 (discover untracked OpenCode sessions during session attach)`
- reason: `irrelevant`
- note: Partio has no OpenCode support and no `session attach` command.

## Warn when a coding agent commits in an enabled repo but no session is being recorded

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1965 (warn when a coding agent commits in an enabled repo but no session is being recorded)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #650 and #664. Not re-filed to avoid duplicates.

## Spam link

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1962 (re-create-gamegambling.app)`
- reason: `irrelevant`
- note: Not a feature or bug report; link to a gambling site.

## `entire agent add/list` does not discover external agents

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1928 (entire agent add/list does not discover external agents)`
- reason: `irrelevant`
- note: Partio has no `partio agent` subcommand. Agent selection is via `PARTIO_AGENT` env var or config field. The issue is specific to Entire's agent management UX.

## agent remove claude-code deletes unrelated .claude/settings.json keys

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1924 (agent remove claude-code deletes unrelated .claude/settings.json keys)`
- reason: `irrelevant`
- note: Partio has no `partio agent remove` command. Hook installation and removal is handled by `partio enable` and `partio disable`.

## Document or support fallback when semantic search is unavailable in a repository region

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1923 (Document or support fallback when semantic search is unavailable in a repository region)`
- reason: `irrelevant`
- note: Partio has no search feature. Checkpoints are stored on a local orphan branch with no cloud search backend.

## Concurrent git-refs checkpoint writes can silently lose completed data

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1917 (Concurrent git-refs checkpoint writes can silently lose completed data)`
- reason: `irrelevant`
- note: Partio writes checkpoints from git hooks, which git serialises per-commit. Concurrent writes to the orphan branch do not occur in normal usage; the race condition described is specific to Entire's async checkpoint pipeline.

## Integrate Muse Code as a supported agent

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1912 (Integrate Muse Code as a supported agent)`
- reason: `irrelevant`
- note: Partio supports claude-code and codex. Adding Muse Code would require a new agent detector and session parser. This is a new integration, not a feature inspired by the source material; no Muse Code session format is documented in this codebase to base a premise on.

## Docs: phone PII redaction is NANP-only but documented without locale qualifier

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1909 (Docs: phone PII redaction is NANP-only in practice but documented with no locale qualifier)`
- reason: `irrelevant`
- note: Partio's redaction (`internal/redact/`) uses entropy-based heuristics, not named PII patterns. There is no phone number pattern in Partio's redaction configuration.

## Repo relocation silently loses commit trailers (findSessionsForWorktree has no fallback)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1890 (Repo relocation silently loses commit trailers on 0.9.0)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #606 and #615. Not re-filed to avoid duplicates.

## dispatch returns 404 for ready EU mirror while regional recap sees checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1889 (dispatch returns 404 for ready EU mirror while regional recap sees checkpoints)`
- reason: `irrelevant`
- note: Partio has no dispatch command or cloud mirror infrastructure. Checkpoints are local.

## UserPromptSubmit hook blocks ~30s and times out on most prompts (per-session flock contention)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1887 (UserPromptSubmit hook blocks ~30s and times out on most prompts)`
- reason: `irrelevant`
- note: Partio has no UserPromptSubmit hook. Its hooks (pre-commit, post-commit, pre-push) are standard git hooks without per-session file locking.

## Checkpoint metadata.json exceeds GitHub's 100 MB limit (files_touched never deduplicated)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1927 (Checkpoint metadata.json exceeds GitHub's 100 MB limit: files_touched is never deduplicated, and nested git worktrees are walked)`
- reason: `irrelevant`
- note: Partio's `checkpoint.Metadata` type (`internal/checkpoint/checkpoint.go`) has no `files_touched` field. The checkpoint branch stores session content (JSONL, diff, context) but not a deduplicated file-path list. The size issue described does not exist in Partio's checkpoint schema.

## changelog 0.10.2: subagent hooks for Codex and Cursor (SubagentStart/SubagentStop)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.2 (better subagent handling for Codex and Cursor)`
- reason: `irrelevant`
- note: Partio has no subagent tracking for any agent. Adding SubagentStart/SubagentStop support would require a new hook mechanism that does not exist in Partio's architecture.

## changelog 0.10.2: --object-format flag for repo create (sha1/sha256)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.2 (entire repo create --object-format)`
- reason: `irrelevant`
- note: Partio has no repo create command. Checkpoints are written to an orphan branch in the existing repo using git plumbing.

## changelog 0.10.1: checkpoint restore error handling for out-of-boundary restores

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.1 (entire checkpoint restore now returns an error when restoring outside a session boundary)`
- reason: `irrelevant`
- note: Partio has no checkpoint restore command. Rewinding is done via `partio rewind` which applies git operations, not Entire's server-side restore API.

## changelog 0.10.1: self-heal zombie sessions via detached sweep from session-start

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.1 (feat: self-heal zombie sessions via detached sweep from session-start) / PR #2029`
- reason: `irrelevant`
- note: Filed as Partio issue #679. Not logged as an idea rejection — listed here for traceability only, as the analogous Partio proposal was filed.

## changelog 0.10.0: Rust-based git library for pull/checkout tracking

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (Entire now includes a full Rust-based git library)`
- reason: `irrelevant`
- note: Partio is a pure Go project with no Rust components. Pull/checkout tracking is not part of Partio's scope; it captures commits, not individual git operations.

## changelog 0.10.0: session cache-stats command

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (entire session cache-stats lists session caches with estimated memory size)`
- reason: `irrelevant`
- note: Partio has no concept of session caches or memory limits. Sessions are JSONL files on disk; there is no in-process cache to inspect.

## changelog 0.10.0: session list --where filter by model name

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (entire session list --where can now filter by model name)`
- reason: `irrelevant`
- note: Partio has no session list command. Model information is stored in checkpoints but not exposed through a queryable interface.

## changelog 0.10.0: SSH key signing for commits in Claude Code/Codex/Cursor

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (embeds reference implementation for signing commits with SSH keys)`
- reason: `irrelevant`
- note: Partio uses `git commit --amend` to add trailers; it defers commit signing to git's own configuration. Partio does not manage signing keys.

## changelog 0.10.0: link mid-task background-subagent commits to their session (PR #2034)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 / PR #2034 (feat: link mid-task background-subagent commits to their session)`
- reason: `irrelevant`
- note: Partio has no background subagent concept. Commit-to-session linking via process identity is already filed as Partio issue #673.

## changelog 0.9.5: session pause/resume flags

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.5 (entire session update --pause and --resume flags)`
- reason: `irrelevant`
- note: Partio's session state machine (ACTIVE, IDLE, ENDED) does not include a paused state and there is no interactive session management UI. Partio is a background capture tool, not a session controller.

## changelog 0.9.5: global --timestamp flag across all commands

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.5 (--timestamp flag moved to global; supports RFC3339, Unix, natural language)`
- reason: `irrelevant`
- note: Partio's commands (enable, disable, status, doctor, rewind, etc.) operate on a single repo's state and do not filter output by time range.

## changelog 0.9.3: entire repo copy command

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.3 (new entire repo copy command for copying git repositories with full history)`
- reason: `irrelevant`
- note: Partio has no repo management commands. It hooks into git operations but does not copy repositories.

## changelog 0.9.3: session diagnose command with memory/CPU usage

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.3 (new diagnostic command entire session diagnose)`
- reason: `irrelevant`
- note: Partio has `partio doctor` for health checks. `partio doctor` checks file existence and hook installation; it does not monitor process memory/CPU. Adding detailed session diagnostics would require a running server, which Partio does not have. A deeper doctor mode is already filed as issue #661.

## changelog 0.9.2: search --format=json and --context-lines flag

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.2 (entire search --format=json, --context-lines flag)`
- reason: `irrelevant`
- note: Partio has no search command. Checkpoint content is readable via `partio rewind` and raw git commands, not a search interface.

## changelog 0.9.1: checkpoint explain streaming and partial output on cancel

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.1 (entire checkpoint explain partial explanations, streaming)`
- reason: `irrelevant`
- note: Partio has no checkpoint explain command. The `partio explain` command (`cmd/partio/explain.go`) calls an external LLM API; it is not part of Partio's checkpoint storage architecture.

## PRs #1882–#2073: dependency bumps, CI fixes, platform infrastructure

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #1882–#2073 (dependency updates, CI hang fix, trail backend retirement, dispatch routing, telemetry, release channel, OpenCode plugin, Cursor/Codex subagent fixes, search TUI changes, changelog PRs)`
- reason: `irrelevant`
- note: The bulk of PRs in this range are dependency version bumps, CI configuration fixes (e.g. #2072 apt hang), platform routing changes (e.g. #2046–#2051 cell targets, dispatch), telemetry (#2023–#2024), trail backend retirement (#2021, #2037), OpenCode plugin changes (#2018, #2027, #2053), Cursor/Codex subagent bug fixes (#2066–#2071), search TUI updates (#2022, #2044), and changelog/credit PRs (#2039, #2050, #2073). None of these map to Partio's domain of git hook-based session capture and checkpoint storage.

## `git diff --name-only` output used for path comparison would fail on non-ASCII filenames

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2398 (Checkpoint linking fails for non-ASCII filenames)`
- reason: `premise-failed`
- claim: `DiffNameOnly` output (from `git diff --name-only`, without `-z`) is used to compare file paths in the session-linking or checkpoint-writing flow [evidence: `internal/git/diff_name_only.go`, `internal/hooks/postcommit.go`]
- verdict: `fails`
- found: `DiffNameOnly` is called at `postcommit.go:95` inside a `slog.LevelDebug` guard (`if slog.Default().Enabled(context.Background(), slog.LevelDebug)`). Its result is only logged for diagnostics; it is not compared against session file paths or used in any decision. The path-mismatch bug Entire fixed does not exist in Partio.

## Control-plane HTTP proxy not honored

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2686 (Control-plane HTTP client does not honor HTTP_PROXY/HTTPS_PROXY)`
- reason: `irrelevant`
- note: Partio has no control-plane HTTP client. All operations are local git plumbing commands or GitHub API calls made via the `gh` CLI. There is no network-dialing code in Partio's CLI.

## Workflow-spawned agents leave no task record

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2685 (Claude Code Workflow-spawned agents leave no task record)`
- reason: `irrelevant`
- note: Partio has no subagent or task tracking. It captures the main Claude Code session JSONL once per commit. There is no task-record, transcript-per-subagent, or token-accounting path.

## Shadow branches grow quadratically (whole transcript re-stored every turn)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2663 (Shadow branches re-store the whole transcript every turn)`
- reason: `irrelevant`
- note: Partio has no shadow branches and no per-turn checkpoint commits. Each `git commit` by the developer produces exactly one checkpoint commit on the orphan branch. The per-turn accumulation issue is specific to Entire's shadow-branch strategy.

## Background subagent lines attributed to human when parent commits in same turn

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2653 (Background subagent's lines are attributed to the human when the parent commits in the same turn)`
- reason: `irrelevant`
- note: Partio has no subagent tracking. Attribution is binary: 0% or 100% based on whether any agent was detected running during pre-commit. There is no per-subagent line attribution.

## Native repos as primary host with outbound mirror

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2623 (Native repos as primary host: outbound mirror to GitHub, webhooks, deploy path)`
- reason: `irrelevant`
- note: Partio has no hosted Git service and no mirroring capability. It is a CLI hook tool that stores checkpoints on a local orphan branch.

## session resume can silently overwrite ignored untracked files during branch checkout

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2604 (session resume can silently overwrite ignored untracked files during branch checkout)`
- reason: `irrelevant`
- note: Partio has no `session resume` command. The `partio resume` command (`cmd/partio/resume.go`) continues a Partio session within the current branch; it does not perform branch checkouts.

## OpenCode 2 integration (V1 plugin API rejected, V1 export command fails)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2588 (Captures nothing on OpenCode 2: V1 plugin API rejected and V1 export command prints help)`
- reason: `irrelevant`
- note: Partio has no OpenCode support. There is no OpenCode agent, plugin, or hook configuration in this codebase.

## git-refs writes return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393 (git-refs writes can return success without durable push bookkeeping)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes the checkpoint branch synchronously via a direct `git push` call. There is no asynchronous push queue or separate push bookkeeping file. The failure mode described is specific to Entire's push-queue architecture.

## Cleanup can make checkpoint commits unreachable after state expiry

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 (Cleanup can make uncondensed checkpoint commits unreachable after state expiry or a ref read failure)`
- reason: `irrelevant`
- note: Partio's cleanup (`cmd/partio/clean.go`) removes stale session state files from `.partio/`. It does not maintain shadow refs or decide reachability by session-state inventory. The cleanup race described is specific to Entire's shadow-branch/session-state coupling.

## Post-push cleanup deletes uncondensed shadow ref when session state is malformed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350 (Post-push cleanup deletes an uncondensed shadow ref when its session state is malformed)`
- reason: `irrelevant`
- note: Partio has no shadow branches and no post-push cleanup that consults session state to decide whether to delete refs. The race between malformed state and shadow-ref deletion is specific to Entire's architecture.

## Install script hides GitHub API authentication and rate-limit errors

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318 (Install script hides GitHub API authentication and rate-limit errors)`
- reason: `irrelevant`
- note: Partio has no install script. It is distributed as a Go binary built with `make build` or installed via `make install`.

## `disable` does not reach linked worktrees

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274 (entire disable does not reach linked worktrees: settings.local.json is per worktree)`
- reason: `irrelevant`
- note: Partio's `partio disable` removes git hooks from `git rev-parse --git-common-dir`, which is shared across all worktrees. Disabling Partio in one worktree disables it for all worktrees in the same repo. The per-worktree re-enable problem does not apply.

## issues #2066–#2115, #2148: subagent tracking, hook timeouts, and index-lock hazard

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2066–#2115, #2148 (Cursor/Codex subagent fixes, hook timeouts, index emptied between staging and commit, finalizeAllTurnCheckpoints deadline)`
- reason: `irrelevant`
- note: These issues cover subagent lifecycle tracking (Partio has no subagent support), hook timeout budgets (Partio's git hooks have no configurable timeout mechanism), and index-lock hazards from concurrent git-status invocations (Partio does not call git status in its hooks). None apply to Partio's architecture.

## issues #2197–#2215: persistent-ref lock files, test isolation, auth, and SubagentStop

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2197–#2215 (persistent-ref lock files accumulate, test git isolation, auth-go lock dir, SubagentStop dropped)`
- reason: `irrelevant`
- note: Partio stores checkpoints on a single orphan branch via git plumbing; it has no per-ref lock files. The test isolation and auth issues are Entire-specific infrastructure. SubagentStop requires subagent tracking that Partio lacks entirely.

## issues #2249, #2255, #2256: stale configure --agent hints and trailer grammar bugs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249 (stale configure --agent hints), #2255 (forged Entire-Checkpoint lines), #2256 (AppendCheckpointTrailer grammar)`
- reason: `irrelevant`
- note: Partio uses PARTIO_AGENT env var and a `partio agent` config field, not Entire's `agent add/configure` commands. Partio's trailer writing is a simple string append (`git/amend_trailers.go`) with no separate strict parser that could reject its own output. Partio also has no code path that reads `Partio-Checkpoint` from commit messages to mutate checkpoint storage.

## issue #2260: Read deny rule in .claude/settings.json breaks auto mode

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260 (Read(./.entire/metadata/**) deny rule makes ordinary commands ask for approval)`
- reason: `irrelevant`
- note: Partio does not write any `permissions.deny` rules to `.claude/settings.json` or any other settings file. It writes only its own `.partio/` directory. No deny rules are installed.

## Forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (Forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `irrelevant`
- note: Partio only writes `Partio-Checkpoint:` trailers via `AmendTrailers`; it never reads checkpoint trailer lines back from commit messages to resolve or mutate checkpoint storage. A forged line would be cosmetically appended alongside the real trailer but would not affect any checkpoint operation.

## PRs #2074–#2091: dependency bumps, OpenCode fixes, performance profiling, trail changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2091 (dependency bumps, OpenCode plugin fixes, hook latency profiling, checkpoint explain fix, index-emptied fix, performance regression, trail changes, git-config cleanup)`
- reason: `irrelevant`
- note: This range consists of: dependency version bumps, OpenCode-specific fixes (#2076, #2078, #2089), hook latency profiling for Entire's per-prompt hooks (#2091), a checkpoint explain scan-limit fix (#2089), an index-emptied race condition in Entire's staging pipeline (#2111), CLI performance regressions specific to Entire's per-prompt attribution model (#2098), and infrastructure/trail changes. None map to Partio's git-hook checkpoint architecture.

## PRs #2550–#2570: trail/project features, mirror/remote changes, auth and cell routing

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2550–#2570 (trail merge, AGENTS.md swap, auth cells, trail/project scope, resume, dispatch, repo delete, repo clone hint, lint, OPF, activity date DST, transcript reader)`
- reason: `irrelevant`
- note: This range includes trail and project management PRs, hosted cell and auth routing, repo clone/delete commands, and Entire-specific lint/infrastructure. None apply to Partio's architecture.

## PRs #2605–#2645: hook manager compatibility, import/attach, background scan, interactive wizards

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2605–#2645 (session import, remote checkpoint listing, hook chaining fixes already filed, Husky v9, trail models, background OPF scan, wizard for repo/project create, strategy fixes, secure enclave tokens, worktree local settings, auth fail-closed, capture growing sessions)`
- reason: `irrelevant`
- note: The hook-chaining and worktree-settings fixes (#2643, #2634) in this range are already covered by open Partio proposals (#747 for hook chaining; worktree disable is not applicable per the #2274 analysis above). The remaining PRs deal with Entire-specific features: session import/attach, remote checkpoint listing, trail models, background privacy filter, interactive repo/project wizards, secure-enclave tokens, and auth infrastructure.

## PRs #2692–#2699: shadow branch removal, activity display, summary isolation, transcript reader, checkpoint delete (filed), Pi reviewer, subagent token count, Claude stream event

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2692–#2699`
- reason: `irrelevant`
- note: PR #2692 removes Entire's shadow branches and line attribution (Partio has neither). #2693 fixes activity display for external agent names (Partio has no activity command). #2694 isolates summary generation from external plugins (Partio has no summary generation). #2695 resolves commit SHAs in cross-repo checkpoint explain (Partio has no cross-repo feature). #2696 adds `entire checkpoint delete` — the analogous `partio checkpoint delete` was filed as Partio issue #757. #2697 shows why a Pi reviewer failed (Partio has no Pi reviewer). #2698 counts subagent API calls written after stop (Partio has no subagent tracking). #2699 fixes Claude Code 2.1.x stream-json system events causing false review failures (Partio has no review runner).

## changelog 0.10.6: Windows installer, runner setup, trail/native repo features, named-pipe hang

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6`
- reason: `irrelevant`
- note: The additions in 0.10.6 (Windows PowerShell installer, `entire runner setup` rework, `entire trail create` pre-push hooks, trail commands against native repos, named-pipe config-file hang fix, concurrent OPF checkpoint-write fix) are either platform infrastructure Partio has no equivalent of (installer, runner, trails, native repos) or internal fixes for Entire's cloud checkpoint pipeline. Partio stores checkpoints on a local orphan branch with no installer script, no trail concept, and no OPF pipeline.

## changelog 0.11.0: cloud platform additions and non-Claude agent support

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0`
- reason: `irrelevant`
- note: The bulk of 0.11.0 changes are cloud-only (native repository references, `entire cluster list`, `entire repo protection list|add|remove`, `entire repo remote url`, `entire logout` revocation, `entire enable` no longer creates GitHub remotes) or support agents Partio does not implement (Codex and Copilot CLI subagent tracking, OpenCode tailored summary runners, factoryai-droid Windows fix). The fixes for Windows console handles and PE version metadata are Windows-specific build concerns. The `investigate` plugin extraction and control-plane rename changes have no equivalent in Partio's command set.

## changelog 0.11.2: trail RFD and credential transport security fix

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.2`
- reason: `irrelevant`
- note: 0.11.2 changes a trail request wire format (RFD-026) and patches a security bug where `ENTIRE_CHECKPOINT_TOKEN` was attached to non-git transports. Partio has no trail concept and no `PARTIO_CHECKPOINT_TOKEN`; checkpoint pushes are plain `git push` calls with no bearer token.

## changelog 0.11.3: new agent, org/repo management, squash trailers, session lifecycle

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.3`
- reason: `irrelevant`
- note: Antigravity CLI (`agy`) is a new agent Partio does not support. `entire org invite`, `entire repo clone --nearest`, and interactive multi-select for grants are cloud management features with no Partio equivalent. The squash-trailer fix preserves existing `Entire-Checkpoint` trailers across `git merge --squash`; Partio's post-commit hook creates fresh trailers on every new commit, so squash commits get a new checkpoint rather than carrying old ones — the scenario Entire fixed does not arise. The "hooks no longer delete live idle sessions" and pushurl-only remote-election fixes are Entire-internal session lifecycle and cloud-push concerns.

## changelog 0.11.4: checkpoint explain subagents, trail API, review security, agent relocation at attach

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.4`
- reason: `irrelevant`
- note: `entire checkpoint explain` subagent-task and token-breakdown support requires a subagent pipeline Partio does not have. `entire trail create --repo` and account display names are cloud-management additions. The `entire review` agent-config-isolation fix is specific to Entire's review command. `StopFailure` turn-end handling is specific to Entire's hook-based session lifecycle; Partio reads session JSONL at commit time and does not use Claude Code's stop hooks. The session resume/attach agent-home relocation improvements are already captured as filed proposals (#708, #715).

## Issues and PRs not individually logged above: remaining items from this run

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2066–#2689 and entireio-cli-pulls #2074–#2706 (remaining items not logged individually)`
- reason: `irrelevant`
- note: Items from this run not individually logged: issues #2689 (UserPromptSubmit timeout — no such hook in Partio), #2664 (pre-push SIGKILL temp packs — Partio's pre-push is a simple git push), #2653 already covered above, #2612 (Codex daemon worktree session sharing — daemon model is Codex-specific), #2606 (Codex TOML hook trust — no Codex hook trust mechanism in Partio), #2588 already covered, #2556 (backgrounded subagent treated as foreground — no subagent tracking), #2535 (enable --force TTY dialog — Partio's enable has no interactive agent selection), #2518 (repo grant list — cloud only), #2416 (doctor interactive prompt blocks CI — Partio's doctor has no interactive prompts), #2402/#2401 (ULID shard case collision — Partio uses only lowercase hex IDs), #2394 (Homebrew pre-trust docs — no Homebrew formula), #2368 (token overwrite — no incremental token accounting), #2362 (status collapses sessions — Partio has no multi-session status view), #2360 (checkpoint_push_remote pushurl-only — no configurable checkpoint push remote), #2328 (Devin agent — already filed as #585, duplicate), #2276 (Windows install URL — no install script), #2271 (docs security page — not a code issue), #2264 (status reports sync when hooks missing — Partio's status already checks hooks at cmd/partio/status.go:65–83), #2249 (stale agent hints — no such hints in Partio), #2218 (Windows PE metadata — not applicable), #2215 (SubagentStop correlation — no subagent tracking), #2204 (OPF writers race — no OPF pipeline), #2203 (install.sh shellcheck — no install script), #2410 (enable writes through symlink — Partio's os.Rename renames the symlink itself before writing, so target is never touched; backup-overwrite aspect already filed as #710), #2257 (AddExternalAgent settings scope — no external-agent management command), #2256 (AppendCheckpointTrailer grammar — Partio's AmendTrailers always appends without grammar detection), #2689–#2392 issues otherwise not covered, and PRs #2700–#2706 (#2700 per-commit build stamp — not user-facing; #2701 GIT_HTTP_USER_AGENT — no HTTP; #2702 repo delete mirrors — cloud only; #2703 agent home registry — cloud only; #2704 HTTP credential security — no HTTP; #2705/#2706 dependency bumps).

## PRs #2707–#2716: dead code removal, dispatch fixes, reviewer config, amend trailer (filed), checkpoint push bound, adopt transcript homes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2707–#2716 (remove dead code, fix repository discovery in dispatch/search, review profile agent config, preserve checkpoint links across amends, Codex CI installer, bound git-refs pre-push push, accept transcripts under recorded agent homes)`
- reason: `irrelevant`
- note: #2707 removes dead and test-only code from Entire's codebase — not a feature idea. #2708 fixes repository discovery in Entire's dispatch wizard and search completion; Partio has no dispatch command. #2711 lets review profiles carry per-reviewer agent config; Partio has no review command. #2713 preserves checkpoint links across `git commit --amend` message changes — already filed as Partio issue #735. #2714 updates Entire's CI to install Codex via its official installer; Partio's CI setup is unrelated. #2715 bounds Entire's git-refs pre-push checkpoint push to prevent unbounded retries; Partio uses a single synchronous `git push` with no retry queue (the analogous improvement — making the push non-blocking — is filed as Partio issue #759). #2716 lets `entire session adopt` accept transcripts stored under any of an agent's recorded homes; Partio's session discovery uses a hardcoded home path and the general CLAUDE_CONFIG_DIR fix is already filed as Partio issue #708.

## Issues #2690–#2718 and PRs #2717–#2721: remaining items from this proposer run

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2690–#2718 and entireio-cli-pulls #2717–#2721 (items beyond the previous run's cursor)`
- reason: `irrelevant`
- note: Issues #2690–#2718: #2690 (activity date DST fix — no activity view in Partio), #2691 (repo clone name hint — no hosted repos), #2692 (shadow branch removal — Partio has none), #2693–#2699 already covered in the batch above, #2627 (refused agent relocation variable in status — no equivalent mechanism in Partio), #2626 (Pi session dir — no Pi support), #2612 (Codex daemon session cross-worktree — Codex-specific daemon sharing), #2551 (Unicode branch name in session resume — Partio has no session resume branch checkout), #2367 (session resume writes to ~/.claude — Partio has no session resume with transcript restore), #2318 (install script error hiding — Partio has no install script), #2410 (enable --force writes through symlink — Partio has no --force flag on enable). Issue #2718 was filed as new proposal #761; issue #2685 was filed as new proposal #760. PRs #2717–#2721: #2717 (drop BFF fallback for activity/recap — cloud only), #2718 cloud-dispatch change, #2719 (dependency bump), #2720 (link shell-command file changes — Partio tracks sessions not individual shell operations), #2721 (cloud dispatch cell routing — no dispatch in Partio).
