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

## changelog 0.10.3–0.11.3: cloud platform, cluster, and server-side features

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.11.3`
- reason: `irrelevant`
- note: The bulk of changelog items in this range address cloud platform concerns that Partio has no equivalent for: cluster selection and latency measurement, org invite management, outbound mirror workflows, native repository hosting, cross-site auth guards, data-API cell contract migrations, Secure Enclave token storage, Windows PE version metadata, shell completion error surfacing, and trail/review/thread request schema changes. Partio stores checkpoints on a local orphan branch with no cloud backend.

## entireio-cli #2623: Native repos as primary host (outbound mirror to GitHub, webhooks, deploy path)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2623`
- reason: `irrelevant`
- note: Partio has no concept of native repository hosting, outbound mirroring, or webhook infrastructure. Checkpoints are stored on a local orphan branch; the remote is whatever git remote the user already has.

## entireio-cli #2588: OpenCode 2 V1 plugin API rejected and V1 export command prints help

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2588`
- reason: `irrelevant`
- note: Partio has no OpenCode agent support.

## entireio-cli #2535: `entire enable --force` exits 0 and refreshes nothing without a TTY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2535`
- reason: `irrelevant`
- note: Partio's `enable` command has no `--force` flag and no interactive agent-management prompts. It is idempotent by design and does not branch on TTY availability.

## entireio-cli #2416: `entire doctor` stuck-session prompt bypasses interactive check, blocks in CI

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416`
- reason: `irrelevant`
- note: Partio's `doctor` command (`cmd/partio/doctor.go`) has no interactive prompts. It runs a series of read-only health checks and prints results; it does not ask the user for input at any point.

## entireio-cli #2393: git-refs writes return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393`
- reason: `irrelevant`
- note: Partio has no git-refs transport or asynchronous push queue. The checkpoint branch is pushed synchronously via `git push --no-verify` in the pre-push hook.

## entireio-cli #2378 and #2350: cleanup can make uncondensed checkpoints unreachable; post-push cleanup deletes uncondensed shadow ref

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 and #2350`
- reason: `irrelevant`
- note: Partio has no shadow-ref concept, no async cleanup pipeline, and no condensation pass that could make checkpoint commits unreachable. Checkpoints are written atomically via `git commit-tree` + `git update-ref` and are never garbage-collected by hook code.

## entireio-cli #2318: Install script hides GitHub API authentication and rate-limit errors

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318`
- reason: `irrelevant`
- note: Partio's install path is `go install` or `make install`. It has no shell install script that queries the GitHub API.

## entireio-cli #2274: `entire disable` does not reach linked worktrees

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274`
- reason: `irrelevant`
- note: Partio's hooks are installed to `git rev-parse --git-common-dir` (confirmed: `internal/git/repo.go:HooksDir`), which is the shared git directory across all worktrees. `partio disable` uninstalls from that shared location, so one `disable` call removes hooks for all worktrees. The per-worktree settings scenario that Entire's issue describes does not apply.

## entireio-cli #2271: Docs — security page has no data-residency information

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271`
- reason: `irrelevant`
- note: Documentation updates are handled by the doc minion after code changes merge.

## entireio-cli #2263: Lefthook is classified as not overwriting Entire's hooks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2263`
- reason: `irrelevant`
- note: Partio detects Lefthook and warns users to add calls to their Lefthook config (`internal/git/hooks/detect_hook_managers.go`). A differentiated warning for managers that overwrite hooks at install time is already covered by proposal #433 (Auto-recover hooks after external hook manager overwrites) and #716 (doctor: surface hook-manager integration instructions when a foreign hook is found).

## entireio-cli #2260: `Read(./.entire/metadata/**)` deny rule makes ordinary commands ask for approval

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260`
- reason: `irrelevant`
- note: Partio does not write to `.claude/settings.json` and does not install any `Read(...)` deny rules. The issue is specific to Entire's Claude Code permission injection.

## entireio-cli #2256: `AppendCheckpointTrailer` can emit a trailer the final-block parser rejects

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256`
- reason: `irrelevant`
- note: Partio's `AmendTrailers` (`internal/git/amend_trailers.go`) always separates the trailer block from the commit body with `"\n\n"` and only appends two well-formed lines (`Partio-Checkpoint: <12-char hex>` and `Partio-Attribution: <N>% agent`). Neither value can contain whitespace or special characters. The writer/reader grammar mismatch described in the source issue does not arise here.

## entireio-cli #2255: Forged `Entire-Checkpoint` lines in commit bodies can select and mutate unrelated checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255`
- reason: `irrelevant`
- note: Partio does not resolve checkpoint storage based on trailer lines in commit bodies. The post-commit hook mints a fresh checkpoint ID and writes it unconditionally; it does not read an existing `Partio-Checkpoint` trailer to decide which checkpoint to reuse or mutate. A forged trailer line would be ignored.

## entireio-cli #2249: Stale `entire configure --agent` hints in investigate and review error messages

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249`
- reason: `irrelevant`
- note: Partio has no `investigate` or `review` commands and no `configure --agent` command. Agent selection uses `PARTIO_AGENT` or the `agent` config field.

## entireio-cli #2215: Claude Code `SubagentStop` is dropped — Entire correlates on `tool_use_id`, which the hook does not send

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215`
- reason: `irrelevant`
- note: Partio has no subagent correlation logic. The JSONL parser (`internal/agent/claude/parse_jsonl.go`) processes entries by role/type to extract message text; a `SubagentStop` entry with no text content is silently skipped, which is the correct behavior for Partio's purpose of capturing the session transcript.

## entireio-cli #2202 and #2201: Test infrastructure — git isolation and auth-go lock dir

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2202 and #2201`
- reason: `irrelevant`
- note: These are internal test-infrastructure improvements to Entire's test suite. Partio's test patterns are documented in CLAUDE.md and are already table-driven with `t.TempDir()` isolation.

## entireio-cli #2197: Persistent-ref lock files accumulate one per checkpoint

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2197`
- reason: `irrelevant`
- note: Partio's checkpoint store uses direct git plumbing (`hash-object`, `mktree`, `commit-tree`, `update-ref`) without acquiring per-checkpoint lock files. There is no lock file accumulation concern.

## entireio-cli #2098: CLI performance regression after enabling full repository

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2098`
- reason: `irrelevant`
- note: The performance regression described is specific to Entire's full-repository analysis and agent permission injection. Partio's hooks do not perform full-repository scans. Hook latency and deadline concerns are already covered by proposal #700 (post-commit has no total deadline).

## entireio-cli #2091 and #2089: Hook latency and checkpoint explain --session false negative

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2091 and #2089`
- reason: `irrelevant`
- note: Hook latency remaining work is already filed as proposal #700. `entire checkpoint explain --session` false negatives are specific to Entire's cloud search index and scan-limit mechanism; Partio has no equivalent search path.

## entireio-cli PRs #2074–#2642: cloud platform, agent integrations, and already-filed improvements

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2642`
- reason: `irrelevant`
- note: The bulk of PRs in this range address cloud platform concerns (cluster routing, auth jurisdiction, trail/review APIs, Secure Enclave, native repo management), agent integrations with no Partio equivalent (OpenCode, Antigravity, Cursor subagents), internal infrastructure (dependency bumps, CI, Windows installer), and improvements to Entire features Partio does not have (session attach, checkpoint restore, OPF scanner, history import). The handful of PRs relevant to Partio's domain (#2641 hook backup rotation, #2611 ignored-file protection, #2586 non-ASCII filenames, #2576 background subagent records, #2552 Unicode normalization, #2534 symlink hook test, #2531 worktree session tracking, #2532 user settings tier) all map to proposals already on file (#710, #730, #713, #641–642, #718, #714, #665, #634 respectively).
