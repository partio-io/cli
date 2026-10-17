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

## Bare AWS access key IDs not redacted without a paired secret

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2808 (Bare AWS access key IDs are not redacted without a paired secret)`
- reason: `irrelevant`
- note: Partio's redaction layer uses entropy-based heuristics and a small set of hardcoded patterns (`internal/redact/patterns.go`). It has no betterleaks integration and no component-based matching rules. The gap described — betterleaks' `aws-access-token` rule requiring a paired `aws-secret-access-key` component — is specific to Entire's betterleaks pipeline, which Partio does not replicate.

## OPF pre-push batching leaves CR intact, causing multiple JSON documents for CRLF input

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2806 (OPF pre-push batching leaves CR intact, causing multiple JSON documents for CRLF input)`
- reason: `irrelevant`
- note: Partio has no OpenAI Privacy Filter (OPF). Redaction runs entirely in-process before writing to the checkpoint branch; there is no external OPF subprocess or batching step.

## Remote helper: fall back to login when a credential helper returns a non-Entire JWT

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2748 (Remote helper: fall back to login when a credential helper returns a non-Entire JWT)`
- reason: `irrelevant`
- note: Partio has no git remote helper. Checkpoint pushes go through the standard `git push` command; there is no `git-remote-partio` binary or credential-aware transport layer.

## READMEs on entire.io show raw HTML and drop images / extra empty lines in code blocks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2746, #2745 (READMEs on entire.io show raw HTML; code blocks render with extra empty line)`
- reason: `irrelevant`
- note: These are rendering issues in the entire.io web application, not in the CLI. Partio has no web application component.

## Write pages for repo, auth, agent_help_cmd, search, lifecycle, checkpoint-signing, release notes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2740, #2739, #2738, #2737, #2736, #2735, #2734, #2733, #2732 (Write a page for repo/auth/agent_help_cmd/search/lifecycle; add agent and checkpoint changes to docs; publish v0.11.4 release notes)`
- reason: `irrelevant`
- note: These are documentation tasks for entireio/cli's own docs site. The Partio ingest prompt explicitly excludes documentation proposals — the doc minion handles those after code PRs merge.

## Redaction misses short sk- keys written as bare values (LiteLLM virtual keys)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2728 (Redaction misses short sk- keys written as bare values in a prompt)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #767. Not re-filed to avoid duplicates.

## Windows CLI: prepare-commit-msg hook stalls when finalizing checkpoint

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2726, #2725 (Windows CLI 0.11.4: prepare-commit-msg hook stalls; hook hangs indefinitely on Windows while Claude Code is active)`
- reason: `irrelevant`
- note: Partio has no `prepare-commit-msg` hook. Its installed hooks are `pre-commit`, `post-commit`, and `pre-push` only (`internal/git/hooks/hooks.go`). The stall is specific to Entire's prepare-commit-msg pipeline.

## --author and --branch filters for checkpoint list

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2718 (Feature request: --author and --branch filters for entire checkpoint list)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #761 (`Add --author filter to partio rewind --list`). Not re-filed to avoid duplicates.

## user-prompt-submit hits 30s timeouts: .git/info/exclude paths, stale-base diff

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2689 (user-prompt-submit hits 30s timeouts: .git/info/exclude paths counted as user changes, unbounded per-prompt state, stale-base diff after HEAD moves)`
- reason: `irrelevant`
- note: Partio has no `user-prompt-submit` hook. Its capture is triggered by standard git hooks (pre-commit, post-commit). The timeout issues described are specific to Entire's per-prompt state tracking mechanism.

## Control-plane HTTP client does not honor HTTP_PROXY/HTTPS_PROXY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2686 (Control-plane HTTP client does not honor HTTP_PROXY/HTTPS_PROXY)`
- reason: `irrelevant`
- note: Partio makes no outbound HTTP calls. Checkpoint storage uses local git plumbing operations; pushes use `git push` which honors proxy settings through git's own configuration. There is no Partio HTTP client to fix.

## Claude Code Workflow-spawned agents leave no task record

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2685 (Claude Code Workflow-spawned agents leave no task record: only the Agent tool is matched, and SubagentStop is skipped)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #760 (`Include Claude Code subagent session transcripts in checkpoint data`). Not re-filed to avoid duplicates.

## pre-push v1 fallback fetch is SIGKILLed at the push budget

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2664 (pre-push v1 fallback fetch is SIGKILLed at the push budget, leaves tmp_pack_* behind, and never converges)`
- reason: `irrelevant`
- note: Partio's pre-push hook calls `git push --no-verify origin partio/checkpoints/v1` directly. There is no v1 fallback fetch loop, no pack budget enforcement, and no tmp_pack cleanup path. The issue is specific to Entire's legacy git-branch backend.

## Shadow branches re-store the whole transcript every turn (168 GB)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2663 (Shadow branches re-store the whole transcript every turn: storage grows quadratically with session length)`
- reason: `irrelevant`
- note: Partio has never had shadow branches. Checkpoints are written once per commit to the orphan branch `partio/checkpoints/v1`. There is no per-turn intermediate storage that could accumulate quadratically.

## Background subagent's lines attributed to the human when parent commits

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2653 (Background subagent's lines are attributed to the human when the parent commits in the same turn)`
- reason: `irrelevant`
- note: Partio has no subagent tracking. Attribution is binary: 0% or 100% based on whether an agent was detected (`internal/attribution/calculate.go`). There are no per-subagent line counts to misattribute.

## Surface a refused agent relocation variable in entire status

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2627 (Surface a refused agent relocation variable in entire status)`
- reason: `irrelevant`
- note: Partio does not yet honor `CLAUDE_CONFIG_DIR` at all (issue #750 is open). There is no validation step that would refuse a relative path, so there is nothing to surface. This issue is a follow-up to a feature Partio has not yet implemented.

## pi: honor PI_CODING_AGENT_SESSION_DIR when resolving the session dir

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2626 (pi: honor PI_CODING_AGENT_SESSION_DIR when resolving the session dir)`
- reason: `irrelevant`
- note: Partio has no Pi agent integration. The registered agents are `claude-code` and `codex` only (`internal/agent/registry.go`).

## Native repos as primary host: outbound mirror to GitHub

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2623 (Native repos as primary host: outbound mirror to GitHub, webhooks, deploy path)`
- reason: `irrelevant`
- note: Partio has no repository hosting or mirroring infrastructure. Checkpoints are stored on a local orphan branch; there is no Partio-native git host.

## Codex hook trust check ignores single-quoted TOML keys

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2606 (Codex hook trust check ignores single-quoted TOML keys, so every hook shows as needing approval on Windows)`
- reason: `irrelevant`
- note: Partio does not write to Codex's `.codex/hooks.json` or any TOML configuration. Codex sessions are captured through the standard git hooks (`pre-commit`, `post-commit`), not through Codex's own hook system.

## OpenCode 2: V1 plugin API rejected and V1 export command prints help

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2588 ([OpenCode] Captures nothing on OpenCode 2: V1 plugin API rejected and V1 export command prints help)`
- reason: `irrelevant`
- note: Partio has no OpenCode agent integration.

## Claude Code: a subagent backgrounded by default is treated as foreground
<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2556 (Claude Code: a subagent backgrounded by default is treated as foreground, and its task record is dropped at launch)`
- reason: `premise-failed`
- claim: Claude Code stores subagent session transcripts in separate JSONL files outside the parent session's file, which Partio does not discover [evidence: `internal/agent/claude/find_latest_session.go`]
- verdict: `unresolved`
- found: `FindLatestSession` reads the single most-recently-modified JSONL in the session directory and returns it; the code makes no attempt to walk sibling JSONL files or follow parent/child session links. Whether Claude Code actually stores subagent transcripts as separate JSONL files (rather than embedding them in the parent JSONL) cannot be verified from the Partio source tree alone. The claim about Claude Code's on-disk format is not groundable here, so the premise is unresolved and the idea is dropped.

## Forged Partio-Checkpoint trailer in commit body could select unrelated checkpoint

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (Forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `premise-failed`
- claim: Partio reads `Partio-Checkpoint` trailers from commit messages to look up or mutate checkpoint storage [evidence: `internal/git/amend_trailers.go`, `internal/hooks/postcommit.go`]
- verdict: `fails`
- found: Searching `amend_trailers.go` and `postcommit.go` shows that Partio only _writes_ `Partio-Checkpoint: <id>` into commits as a human-readable reference; no code in this repository reads that field back from commit messages to select, reuse, or mutate checkpoints. Checkpoints are looked up exclusively by walking the sharded tree on `partio/checkpoints/v1`. The attack surface Entire's issue described does not exist in Partio's architecture.

## SubagentStop hook event correlation missing

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop is dropped: Entire correlates on tool_use_id, which the hook does not send)`
- reason: `irrelevant`
- note: Partio does not install Claude Code lifecycle hooks (UserPromptSubmit, SubagentStart, SubagentStop, etc.). It captures sessions via git hooks (pre-commit, post-commit, pre-push) by reading the JSONL file at commit time. The correlation problem described is specific to Entire's real-time hook pipeline and has no analogue in Partio.

## Issues #2066–#2089: Entire platform issues with no Partio analogue

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2066–#2089 (checkpoint explain scan limit false negative, hook latency, performance regression, empty tree commit, finalizeAllTurnCheckpoints deadline, test git isolation, token counting, persistent-ref lock files)`
- reason: `irrelevant`
- note: These issues concern Entire's server-side or advanced pipeline features: checkpoint explain's server-side scan limits (#2089), remaining hook latency work (#2091), agent performance regression in CI (#2098), an index emptied between staging and commit (#2111, a git behaviour issue), finalizeAllTurnCheckpoints total deadline (#2148, Entire's Claude hook pipeline), structural git isolation for tests (#2202), and persistent-ref lock file accumulation (#2197). Partio shares git hooks and JSONL parsing but lacks the async checkpoint pipeline, test-isolation infrastructure mismatch, and server-side scan paths these issues target.

## Issues #2249–#2318: Entire CLI and platform bugs with no Partio analogue

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249–#2318 (stale configure --agent hints, re-install discards newer backup hook, SubagentStop dropped, test isolation, lock files, git-refs write bookkeeping, stale-state cleanup, post-push cleanup on malformed state, install script error hiding)`
- reason: `irrelevant`
- note: `#2249` references `entire configure --agent`, which Partio does not have. `#2215`/`#2237`/`#2256`/`#2255`/`#2264`/`#2274` all have Partio analogues already filed (#717, #710, #763, #751). `#2318` is about Entire's install script; Partio has no install script. `#2350` and `#2378` concern Entire's shadow-branch and async cleanup pipeline, which Partio does not replicate.

## Issues #2393–#2416: push bookkeeping, case-insensitive FS refs, non-ASCII filenames, doctor prompts

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393–#2416 (git-refs write bookkeeping, ULID case-insensitive FS, checkpoint linking non-ASCII, doctor stuck-session prompt, entire enable --force without TTY)`
- reason: `irrelevant`
- note: `#2393` (push queue bookkeeping race) has no Partio analogue — Partio pushes synchronously and has no queue. `#2401` (ULID case collision on macOS/Windows) does not apply — Partio uses lowercase hex IDs which cannot collide on case-insensitive filesystems. `#2398` (non-ASCII filenames) is already filed as #713. `#2416` (doctor prompt blocks in CI) does not apply — `partio doctor` prints to stdout only and has no interactive prompts. `#2535` (enable --force exits 0 without TTY) is not applicable — `partio enable` has no --force flag or interactive flow.

## Issues #2551–#2612: session resume, OpenCode, hooks, commit linking, worktrees, shadow branches

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2551–#2612 (Unicode branch names, session resume ignored files, OpenCode 2 integration, git config-based hooks, SubagentStop, worktree disable, enable --force, commit linking, shadow branch storage, hook $0 break, user-prompt-submit timeout)`
- reason: `irrelevant`
- note: `#2551` (Unicode NFC/NFD) → filed as #718. `#2604` (resume overwrites ignored files) → filed as #730. `#2588` (OpenCode 2) → closed as not_planned (#98). `#2580` (git config hooks) → filed as #726. `#2556` (SubagentStop) → logged as dropped above. `#2274` (worktree disable) → filed as #751. `#2535` (enable --force) → not applicable. `#2612` (Codex worktree commit linking) → filed as #755. `#2663` (shadow branch storage) → Partio has no shadow branches. `#2673` ($0 hook chaining) → filed as #747. `#2689` (user-prompt-submit timeout) → Partio installs no user-prompt-submit hook.

## Issues #2623–#2746: native repos, amend drops trailers, subagent attribution, worktree settings, README rendering

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2623–#2746 (native repo primary host, amend -m drops trailers, subagent attribution, worktree local settings, git commit --amend $0, doctor interactive CI, checkpoint list filters, Windows hook stall, README rendering bugs)`
- reason: `irrelevant`
- note: `#2637` (amend drops trailers) → filed as #735. `#2653` (subagent attribution) → filed as #755. `#2634` (worktree local settings) → filed as #751. `#2718` (checkpoint list filters) → filed as #761. `#2726` (Windows prepare-commit-msg hang) → Partio installs no prepare-commit-msg hook. `#2745` and `#2746` (README rendering bugs) → entire.io web UI issues, entirely irrelevant. `#2623` (native repos as primary host) → Partio is a local capture tool with no hosting platform.

## changelog 0.10.3–0.11.5: Entire platform features with no Partio analogue

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.5`
- reason: `irrelevant`
- note: The changelog spans nine versions (0.10.3 through 0.11.5). Most changes concern Entire's hosted platform: trail management, native repo hosting, cluster routing, OPF redaction, mirror workflows, org invitations, device-code auth, Windows PE metadata, Antigravity integration, token breakdowns, shadow-branch removal, and line-attribution changes. These have no Partio analogue. Items with Partio relevance: transcript redaction parallelization (already filed as #712), checkpoint delete by ID (already filed as #461), git config-based hooks (already filed as #726), amend drops trailers (already filed as #735), pre-push checkpoint push timeout (filed this run as #776). Named pipes and confined filesystem operations were evaluated but could not be grounded from the Partio source tree without knowing Entire's specific failure scenario.

## PRs #2074–#2791: Entire platform PRs with limited Partio relevance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2791 (dependency bumps, trail features, native repo routing, OPF redaction, mirror async, Windows build, hook manager fixes, token counting, agent add/list/remove, JSONL transcript refactoring)`
- reason: `irrelevant`
- note: The vast majority are dependency updates (#2764, #2765, #2766, #2788, #2787), platform routing (#2778, #2756), trail/review features (#2749, #2759, #2760, #2776, #2780), mirror and cluster features (#2751, #2762, #2768, #2782, #2783, #2785, #2786), and Entire-specific bug fixes. PRs with Partio analogues already have proposals filed: #2790 (hook reinstall) → #710, #2713 (preserve checkpoint links across amend) → #735, #2634 (worktree local settings) → #751, #2643 (Husky v9 chaining) → #747, #2611 (ignore files during checkout) → #730, #2684 (sync_prompts) → #752, #2668 (session-commit mismatch) → #755.

## entireio/cli#2637: git commit --amend -m/-F drops the Entire-Checkpoint trailer

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2637`
- reason: `duplicate`
- note: Already filed as Partio issue #735 (post-commit: restore Partio-Checkpoint trailers on git commit --amend -m when agent is no longer running).

## entireio/cli#2612: Commit linking attaches an unrelated Codex session from another worktree

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2612`
- reason: `duplicate`
- note: Already filed as Partio issue #773 (Codex FindLatestSession picks up main-worktree session when committing from a linked worktree nested inside the main worktree).

## entireio/cli#2673: Wrapped hooks chained under renamed backup path break hooks that check $0

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2673`
- reason: `duplicate`
- note: Already filed as Partio issue #747 (Chaining to backup hook by renamed path breaks tools that dispatch on $0 (Trunk, Husky v9)).

## entireio/cli#2685: Claude Code Workflow-spawned agents leave no task record

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2685`
- reason: `duplicate`
- note: Already filed as Partio issue #760 (Include Claude Code subagent session transcripts in checkpoint data).

## entireio/cli#2626: Honor PI_CODING_AGENT_SESSION_DIR when resolving session dir

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2626`
- reason: `duplicate`
- note: Already filed as Partio issue #750 (Honor CLAUDE_CONFIG_DIR when locating Claude Code sessions and plans).

## Hook execution timeout

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.5 (each Entire Claude Code hook has a timeout)`
- reason: `duplicate`
- note: Already filed as Partio issue #212 (Add configurable hook execution timeout with graceful degradation).

## entireio/cli#2718: --author and --branch filters for entire checkpoint list

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2718`
- reason: `duplicate`
- note: Already filed as Partio issue #761 (Add --author filter to partio rewind --list).

## entireio/cli#2663: Shadow branches re-store the whole transcript every turn (quadratic storage growth)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2663`
- reason: `premise-failed`
- claim: Partio re-stores session content on every checkpoint write, causing the total size of the checkpoint branch to grow quadratically across sessions.
- verdict: `fails`
- found: `internal/checkpoint/store.go` uses `git hash-object -w --stdin` to write blobs. Git's object model deduplicates identical blobs by content hash; if the same JSONL content appears in two checkpoint commits, git stores only one blob object. The checkpoint branch grows linearly (one new tree + metadata per commit), not quadratically. The source's issue was specific to Entire's architecture of storing full transcripts on every turn; Partio captures once per commit and benefits from git's native deduplication.

## entireio/cli#2689: user-prompt-submit .git/info/exclude paths counted as user changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2689`
- reason: `irrelevant`
- note: Partio has no UserPromptSubmit hook. Its attribution logic (`internal/attribution/calculate.go`) uses `git diff --numstat` on the actual commit, not on all files including excluded ones. The issue's `unbounded per-prompt state` and `stale-base diff` are Entire-specific concepts not present in Partio.

## entireio/cli#2653: Background subagent's lines are attributed to the human when the parent commits in the same turn

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2653`
- reason: `irrelevant`
- note: Partio's attribution is binary (0% or 100%) based solely on `state.AgentActive` in `internal/attribution/calculate.go`. Partio has no subagent concept and no multi-agent attribution logic. The source's issue requires a subagent model that does not exist in Partio.

## entireio-cli-issues #2748, #2746, #2745, #2740–#2732, #2728, #2726, #2725, #2686, #2664, #2627, #2623, #2606: various cloud/UI/docs/Windows/auth issues

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2748, #2746, #2745, #2740–#2732, #2728, #2726, #2725, #2686, #2664, #2627, #2623, #2606`
- reason: `irrelevant`
- note: These issues concern: remote helper auth (#2748), README rendering (#2746, #2745), docs/page writing tasks (#2740–#2732), Windows CLI hook stalls (#2728, #2726, #2725), HTTP proxy support (#2686), pre-push budget kill / tmp_pack cleanup (#2664), surface agent relocation in status (#2627), native repos as primary host cloud feature (#2623), and TOML key parsing in a Codex hook trust check (#2606). None of these map to Partio's architecture: Partio is Unix-only (bash hooks), has no credential helper, no cloud backend, no HTTP client, no Codex hook trust check, and no TOML configuration.

## entireio-cli-pulls #2795, #2793, #2792, #2791, #2789, #2786, #2785, #2784, #2783, #2782, #2781, #2780, #2779, #2778, #2777, #2776, #2775, #2774, #2773, #2772, #2771: dependency bumps, cloud/trail/transcript/token-counting changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2795, #2793, #2792, #2791, #2789, #2786, #2785, #2784, #2783, #2782, #2781, #2780, #2779, #2778, #2777, #2776, #2775, #2774, #2773, #2772, #2771`
- reason: `irrelevant`
- note: These PRs cover: CDN fallback releases (#2795), trail server URL (#2793), search pagination (#2792), external agent discovery (#2791), JSONL conversion architecture (#2786), redo commit checkpoints (#2785), Brain checkpoint hint (#2784), Claude fork launch message counting (#2783), per-agent token sources (#2782), session state save timing (#2781), full-ref default branch resolution (#2780), turn step/token deduplication (#2779), structured tool result subagent counting (#2777), reviewer prompts on stdin (#2776), hook exec ETXTBSY retry (#2775), antigravity/pi summary prompts (#2773, #2772), trail-review reporting (#2771), and the 0.11.5 changelog (#2774). Partio has no trail, no cloud dispatch, no token counting, no JSONL conversion pipeline, no multi-agent turn tracking, and no stdin-based reviewer. The hooks ETXTBSY retry (#2775) could be relevant on Linux but is a test infrastructure concern, not a production bug in Partio's hook scripts.

## entireio/cli#2806: OPF pre-push batching leaves CR intact (CRLF multi-document failure)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2806 (OPF batch sanitizer leaves CR intact, causing multiple JSON documents for CRLF input)`
- reason: `irrelevant`
- note: Partio has no OpenAI Privacy Filter integration and no batch sanitization pipeline. The pre-push hook pushes the checkpoint branch with a plain `git push` and performs no transcript redaction at push time. The CRLF / multi-document failure is specific to Entire's OPF pipeline.

## entireio-cli-pulls #2796–#2805: redaction fixes, trail features, dependency bumps

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2796–#2805 (session sweep redaction loading fix, trail merge branch reporting, custom redaction rule validation, git-remote-entire TLS connection reuse, redaction prefix cache invalidation, dispatch polling on refused redirect, git-remote-entire error messages, golang.org/x/net bump, external agents transcript compaction, project trail URL agent-help)`
- reason: `irrelevant`
- note: These PRs are all Entire-specific: redaction pipeline fixes (#2796, #2800), trail merge output (#2797), custom rule validation (#2798), TLS reuse in git-remote-entire (#2799), dispatch polling (#2801), git-remote-entire error formatting (#2802), a dependency bump (#2803), external agent transcript compaction via plugin (#2804), and agent-help project trail URL guidance (#2805). Partio has no redaction pipeline, no trail management, no git remote helper, and no dispatch client. None of these map to Partio's domain.

## `entire enable --force` exits 0 without a TTY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2535 (entire enable --force exits 0 and refreshes nothing without a TTY)`
- reason: `irrelevant`
- note: `partio enable` has no `--force` flag (`cmd/partio/enable.go`). The command always reinstalls hooks unconditionally; the TTY-gated skip logic described does not exist in Partio.

## `entire doctor` stuck-session prompt bypasses interactive check, blocks in scripts

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416 (entire doctor stuck-session prompt bypasses CanPromptInteractively(), so it blocks in scripts and CI)`
- reason: `irrelevant`
- note: `partio doctor` has no interactive prompts. It prints diagnostics and exits; there is no `huh` form or bubbletea menu. The stuck-session prompt that blocks in scripts is specific to Entire's doctor implementation.

## `entire enable --force` writes hook wrapper through a symlink

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2410 (entire enable --force writes the hook wrapper through a symlink, clobbering the target file)`
- reason: `irrelevant`
- note: `partio enable` has no `--force` flag. The symlink-on-reinstall issue is already filed for Partio as issue #714 (`installHooks writes through a symlink at the hook path on re-enable`), which covers the same underlying behavior without the `--force` context.

## ULID checkpoint refs become permanently unreadable on case-insensitive filesystem

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2402, #2401 (ULID checkpoint refs become permanently unreadable when shard case-collides on macOS/Windows)`
- reason: `irrelevant`
- note: Partio uses 12-character lowercase hex strings for checkpoint IDs (`internal/checkpoint/checkpoint.go`). Shards are the first two hex characters (always lowercase 0–9, a–f). On a case-insensitive filesystem there is no uppercase variant to collide with. Partio also stores checkpoints in git trees on an orphan branch, not as per-checkpoint git refs, so the ref-collision mechanism described does not apply.

## Documentation requires tapping before trusting (Homebrew 6)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394 (Documentation tries to tap before trusting which is not allowed in Homebrew 6)`
- reason: `irrelevant`
- note: Partio is not distributed via Homebrew. Installation is via `go install` or direct binary download.

## git-refs writes can return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393 (git-refs writes can return success without durable push bookkeeping)`
- reason: `irrelevant`
- note: Partio has no asynchronous push queue and no `enqueueForPush` mechanism. The pre-push hook calls `git push` synchronously; there is no separate queue that could silently drop a ref.

## Cleanup can make uncondensed checkpoint commits unreachable after state expiry

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 (Cleanup can make uncondensed checkpoint commits unreachable after state expiry or a ref read failure)`
- reason: `irrelevant`
- note: Partio has no shadow branches and no mechanism that keeps uncondensed checkpoints reachable through a separate branch. Once the commit is on the main branch and the checkpoint is on `partio/checkpoints/v1`, there is no parallel shadow structure that cleanup could delete.

## applyBackfilledSessionTokenUsage overwrites the session-wide token total

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2368 (applyBackfilledSessionTokenUsage overwrites the session-wide token total with the checkpoint window delta for every non-Copilot agent)`
- reason: `irrelevant`
- note: Partio has no per-checkpoint token window tracking. Session token usage is a single cumulative field (`internal/agent/types.go`). There is no checkpoint condensation loop that could overwrite the session-wide total with a delta.

## partio status collapses multiple sessions per agent and mutates state while reporting

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2362 (entire status collapses multiple sessions per agent, and mutates session state while reporting it)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #765 (`partio status mutates session state as a side effect of reading`). Not re-filed to avoid duplicates.

## changelog 0.10.3–0.11.5: trail, repo, mirror, cluster, and auth management features

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.5 (trail merge/create/show/update/delete; repo mirror detach/delete/cascade; repo clone --nearest; cluster list; auth switch; org invite; project-scoped trails; native repo references; BFF fallback removal; cell routing; jurisdiction handling)`
- reason: `irrelevant`
- note: Partio has no trail, repo mirror, cluster, native repo, org, project, or auth management commands. These features are specific to Entire's cloud-hosted platform. Partio stores checkpoints locally on an orphan branch and has no control plane.

## changelog 0.10.3–0.11.5: external agents (Pi, Antigravity, OpenCode, Cursor, Copilot CLI)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.5 (Antigravity CLI agent integration; Pi review runner; OpenCode session export; Cursor transcript improvements; Copilot CLI subagent tracking; Factory AI Droid Windows hook wrappers)`
- reason: `irrelevant`
- note: Partio supports only `claude-code` and `codex`. None of these agent integrations exist in the codebase, and adding them requires new agent detectors and parsers not inspired by a changelog entry alone.

## changelog 0.10.3–0.11.5: betterleaks, OPF, and redaction pipeline improvements

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.5 (betterleaks 1.8.0/1.8.1/1.9.0 updates; OPF CAS ref updates; OPF entity span splits; custom redaction rules; bare password= coverage; redaction prefix caching; empty-regex guard; settings trust gate for OPF)`
- reason: `irrelevant`
- note: Partio uses entropy-based heuristics and a small set of hardcoded patterns for redaction. It has no betterleaks, no OPF, and no custom redaction rule system. Improvements to those external engines do not apply.

## changelog 0.10.3–0.11.5: search, dispatch, and API routing features

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.5 (search TUI continuous scroll; search result matching; dispatch --jurisdiction; entire api cell routing; cloud dispatch; search --json; search routed to one home placement)`
- reason: `irrelevant`
- note: Partio has no search, dispatch, or API routing functionality. These are specific to Entire's cloud platform.

## changelog 0.10.3–0.11.5: shadow branch removal and session state topology changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.5 (agent turns no longer write shadow branches; breaking change: shadow branch removal migration; shadow branch cleanup via doctor --force)`
- reason: `irrelevant`
- note: Partio never had shadow branches. This breaking change and its migration tooling do not apply.

## changelog 0.11.5: line attribution removal (breaking change)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.5 (line attribution removed: entire blame reports checkpoint linkage instead of AI/Human/Mixed labels)`
- reason: `irrelevant`
- note: Partio still carries binary attribution (0% or 100% agent) in checkpoint metadata. Entire's decision to remove line attribution from its product does not imply Partio should do the same; Partio's attribution model is simpler and serves a different purpose.

## PRs #2074–#2809: implementation PRs, dependency bumps, CI fixes, non-Partio agents

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2809 (implementation PRs for trails, repos, mirrors, search, dispatch, auth; Antigravity/Pi/OpenCode/Cursor/Copilot/Factory agent changes; dependency bumps; CI test fixes; internal refactors; token counting; transcript adapters)`
- reason: `irrelevant`
- note: The PRs in this range implement the changelog and issue changes above. Those already captured as relevant have their own records. The remainder — trail/repo/mirror/search/dispatch infrastructure, non-Partio agent integrations, dependency updates, CI configuration, and internal refactors — do not map to Partio's domain of git hook-based session capture and checkpoint storage on an orphan branch.