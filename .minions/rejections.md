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

## changelog 0.10.5: doctor session lock announcement and mirror async

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5`
- reason: `irrelevant`
- note: The notable items (doctor announcing waiting for session lock; async mirror creation; session full.jsonl released after condensation) are all specific to Entire's server-side session-lock architecture and async mirror infrastructure. Partio has no session lock, no mirror creation, and no condensation pipeline.

## changelog 0.10.6: named pipes at config paths hang processes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (named pipes at config paths no longer hang processes)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #707 (Validate .partio/ directory and settings file types before opening). Not re-filed.

## changelog 0.10.6: session and checkpoint file locks revalidate inode after acquiring

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (session and checkpoint file locks revalidate inode after acquiring)`
- reason: `irrelevant`
- note: Partio does not use persistent file locks for sessions or checkpoints. Session state is written to `.partio/sessions/` as JSON files with no inode-validated lock objects. The race condition described is specific to Entire's lock implementation.

## changelog 0.10.6: Windows PowerShell installer and install-detection improvements

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (Windows PowerShell installer, install detection per-OS with brew/mise/Scoop)`
- reason: `irrelevant`
- note: Partio has no install scripts. The Windows binary artifact work is already filed as Partio issue #695. The installer UX improvements are Entire-specific.

## changelog 0.10.6: concurrent checkpoint write discards OPF rewrite

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (concurrent checkpoint write no longer silently discards OPF rewrite)`
- reason: `irrelevant`
- note: Partio's checkpoint storage uses git plumbing commands (hash-object, mktree, commit-tree, update-ref). There is no OPF (Object Processing Framework) layer. Concurrent checkpoint writes do not occur in normal Partio usage since git serializes hook execution per-commit.

## changelog 0.11.0: control-plane, cloud infrastructure, and authentication changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (auth switch, repo view, repo mirror add, cluster list, repo protection, repo remote url, GitLab checkpoint remote, native /et/ repo refs, async mirror, device-code Linux login, control-plane E2E)`
- reason: `irrelevant`
- note: Partio has no control plane, no cloud mirror infrastructure, no authentication, no cluster concept, no native repo references, and no checkpoint remote that can be configured as GitLab. Checkpoints are stored on a local orphan branch and pushed to whatever `origin` the repo already has.

## changelog 0.11.0: Copilot CLI and OpenCode subagent tracking

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (subagent tracking for Codex and Copilot CLI; OpenCode summaries and tailoring)`
- reason: `irrelevant`
- note: Partio has no subagent tracking and no OpenCode support. Partio captures the entire JSONL session file at commit time; there is no per-turn or per-subagent recording pipeline.

## changelog 0.11.0: status --json one entry per session

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (entire status --json reports one entry per session with session_id, worktree_path, branch)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #220, #337, #482, and #597 (Add --json flag to partio status). Not re-filed.

## changelog 0.11.0: already-committed files no longer re-checkpointed at turn-end

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (already-committed files no longer re-checkpointed; comparisons respect autocrlf and gitattributes)`
- reason: `irrelevant`
- note: Partio's post-commit hook uses `git diff` for all file comparisons (internal/git/diff.go). Git diff already respects gitattributes and autocrlf. Partio has no direct byte-by-byte file comparison that could be confused by line-ending differences.

## changelog 0.11.0: entire session current uses agent session IDs and process ancestry

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (entire session current uses agent session IDs and process ancestry for accurate identification)`
- reason: `irrelevant`
- note: Process-identity-based session linking is already filed for Partio as issue #673 (Link commits to sessions by process identity). Not re-filed.

## changelog 0.11.0: git hook installation uses confined filesystem operations

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (git hook installation uses confined filesystem operations; doctor reports unsafe directories)`
- reason: `irrelevant`
- note: The hook symlink write is already filed as Partio issue #714. The FIFO/non-regular-file check for config paths is already filed as issue #707. Not re-filed.

## changelog 0.11.0: JSONL redaction verifies replacements with fallback

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (JSONL redaction verifies replacements with fallback to structural rewriting or failing closed)`
- reason: `irrelevant`
- note: Partio's `internal/redact/redact.go` applies pattern-based and entropy-based redaction via `Text()` but does not verify that replacements were actually made. However, Partio's redaction does not operate on JSONL structure — it treats session content as plain text. A structural-rewrite fallback would require JSONL parsing in the redaction layer, which is absent. The Entire fix addresses a different implementation class (structured JSONL redaction with CAS verification).

## changelog 0.11.0: entire status no longer deletes stale records while reading

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0 (entire status no longer deletes stale records while reading)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #709 (partio status silently transitions sessions to ENDED: move CleanupStale out of the status command). Not re-filed.

## changelog 0.11.2: control-plane protocol migration (RFD-026), native repo .git suffix, session ID validation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.2`
- reason: `irrelevant`
- note: Changelog 0.11.2 covers RFD-026 snake_case/cursor-pagination migration (cloud-specific), trailing .git preservation in native repo names (Partio has no native repo concept), ENTIRE_CHECKPOINT_TOKEN scoping (Partio has no checkpoint token), and session ID character validation (Partio generates IDs as 12-char lowercase hex via crypto/rand, which is always URL-safe).

## entireio-cli-issues #2556: Claude Code subagent backgrounded by default dropped at launch

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2556 (Claude Code subagent backgrounded by default is treated as foreground, task record dropped at launch)`
- reason: `irrelevant`
- note: Partio has no subagent tracking. Session capture reads the entire JSONL file at commit time; there is no per-launch or per-subagent task record in Partio's architecture.

## entireio-cli-issues #2551: session resume Unicode NFC/NFD branch name

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2551 (session resume reports branch not found for Unicode-equivalent NFC/NFD branch name)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #718 (FindByBranch uses byte-exact comparison: resume --branch-name silently finds no checkpoints for Unicode-equivalent branch names). Not re-filed.

## entireio-cli-issues #2535: enable --force exits 0 without TTY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2535 (entire enable --force exits 0 and refreshes nothing without a TTY)`
- reason: `irrelevant`
- note: Partio's `partio enable` has no `--force` flag and no TTY-dependent behavior. The command is fully non-interactive and idempotent by design.

## entireio-cli-issues #2518: repo grant list picks unreachable placement

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2518 (repo grant list picks placement without checking active credentials can reach it)`
- reason: `irrelevant`
- note: Partio has no access grants, placements, or credential-scoped repo listings.

## entireio-cli-issues #2416: doctor stuck-session prompt blocks in CI

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416 (entire doctor stuck-session prompt bypasses interactive check, blocks in scripts and CI)`
- reason: `irrelevant`
- note: Partio's `partio doctor` (cmd/partio/doctor.go) has no interactive prompts. It performs read-only checks (directory, hook files, branch existence) and prints results without waiting for user input.

## entireio-cli-issues #2410: enable --force writes hook through symlink

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2410 (entire enable --force writes hook wrapper through symlink, clobbering target)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #714 (installHooks writes through a symlink at the hook path on re-enable). Not re-filed.

## entireio-cli-issues #2402, #2401: ULID checkpoint refs unreadable on case-insensitive filesystems

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2402, #2401 (ULID checkpoint refs permanently unreadable when shard case-collides on macOS/Windows)`
- reason: `irrelevant`
- note: Partio's checkpoint IDs are 12-character lowercase hex strings generated by `hex.EncodeToString(rand.Read)` (internal/checkpoint/checkpoint.go:NewID). Shard keys are the first two lowercase hex characters. Case collisions cannot occur with all-lowercase IDs on case-insensitive filesystems.

## entireio-cli-issues #2398: checkpoint linking fails for non-ASCII filenames

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2398 (checkpoint linking fails for non-ASCII filenames)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #713 (DiffNameOnly returns C-escaped filenames for non-ASCII paths, missing -z flag). Not re-filed.

## entireio-cli-issues #2394: documentation Homebrew tap trust issue

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394 (documentation tries to tap before trusting, not allowed in Homebrew 6)`
- reason: `irrelevant`
- note: Partio has no Homebrew tap or formula.

## entireio-cli-issues #2393: git-refs writes return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393 (git-refs writes can return success without durable push bookkeeping)`
- reason: `irrelevant`
- note: Partio has no git-refs push queue. The pre-push hook calls `git push origin partio/checkpoints/v1` synchronously; push success is determined by git's return code. No separate bookkeeping layer exists.

## entireio-cli-issues #2378: cleanup makes uncondensed checkpoint commits unreachable

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 (cleanup can make uncondensed checkpoint commits unreachable after state expiry or ref read failure)`
- reason: `irrelevant`
- note: Partio's checkpoints are commits on a single linear orphan branch (partio/checkpoints/v1). There is no shard-ref cleanup, no shadow ref, and no separate condensation pipeline that could orphan commits.

## entireio-cli-issues #2368: applyBackfilledSessionTokenUsage overwrites session total

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2368 (applyBackfilledSessionTokenUsage overwrites session-wide token total with checkpoint window delta)`
- reason: `irrelevant`
- note: Partio stores total_tokens in checkpoint metadata but has no `applyBackfilledSessionTokenUsage` function or checkpoint-window token delta. Token accounting uses a single field written once from the session JSONL parse.

## entireio-cli-issues #2367: session resume writes to ~/.claude instead of $CLAUDE_CONFIG_DIR

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2367 (session resume writes to ~/.claude instead of $CLAUDE_CONFIG_DIR)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #715 (ReadPlanFile hardcodes ~/.claude/plans/: plan content silently missed when CLAUDE_CONFIG_DIR is set). Not re-filed.

## entireio-cli-issues #2362: entire status collapses multiple sessions and mutates state while reporting

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2362 (entire status collapses multiple sessions per agent, mutates session state while reporting)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #709 (partio status silently transitions sessions to ENDED: move CleanupStale out of the status command). Not re-filed.

## entireio-cli-issues #2360: checkpoint_push_remote election accepts pushurl-only remote

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2360 (checkpoint_push_remote election accepts pushurl-only remote that status then reports as working)`
- reason: `irrelevant`
- note: Partio has no `checkpoint_push_remote` configuration option. The pre-push hook always targets `origin`. The election/validation path described does not exist in Partio.

## entireio-cli-issues #2350: post-push cleanup deletes uncondensed shadow ref

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350 (post-push cleanup deletes uncondensed shadow ref when session state is malformed)`
- reason: `irrelevant`
- note: Partio has no shadow refs, no condensation pipeline, and no post-push cleanup step. The pre-push hook pushes the checkpoint branch and returns.

## entireio-cli-issues #2328: Add Devin agent support

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2328 (Add Devin agent support)`
- reason: `irrelevant`
- note: Adding a Devin agent to Partio requires Devin's session transcript format and process-discovery mechanism. The issue provides no technical specification of Devin's session format that could be grounded in Partio's agent detection interface (`internal/agent/`). No claim about Devin's session format can be verified against the checked-out tree, so no premise can be built and the idea cannot be filed.

## entireio-cli-issues #2318, #2276: install script error handling and Windows URL

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318 (install script hides GitHub API auth errors), #2276 (Windows install script URL gives login page)`
- reason: `irrelevant`
- note: Partio has no install scripts (confirmed: no files at scripts/install*). Windows binary distribution is already filed as issue #695; when implemented, install script robustness will be relevant. Not applicable now.

## entireio-cli-issues #2274: disable does not reach linked worktrees

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274 (entire disable does not reach linked worktrees: settings.local.json is per worktree)`
- reason: `irrelevant`
- note: Partio's `partio disable` removes hooks from `git rev-parse --git-common-dir`, which is shared across all worktrees. Disabling from any worktree removes hooks everywhere. Entire's issue is caused by per-worktree settings.local.json being recreated on new worktrees; Partio's settings live in the main repo's `.partio/` directory and are not per-worktree.

## entireio-cli-issues #2271: docs security page missing data-residency information

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271 (docs security page has no data-residency information)`
- reason: `irrelevant`
- note: Partio stores all data locally (orphan branch). There is no cloud storage and no data-residency disclosure needed.

## entireio-cli-issues #2264: entire status reports sync when hooks not installed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2264 (entire status reports checkpoints sync to origin when git hooks are not installed)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #704 (partio status reports "Hooks: installed" for foreign hook files). Not re-filed.

## entireio-cli-issues #2263: Lefthook classified as non-overwriting, weaker warning given

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2263 (Lefthook is classified as not overwriting Entire's hooks, so users get weaker warning and no permanent fix)`
- reason: `irrelevant`
- note: Partio's `DetectHookManagers` (internal/git/hooks/detect_hook_managers.go) has no overwriting/non-overwriting classification for detected hook managers. It provides generic integration instructions. Doctor integration warnings are already filed as issue #716. The overwriting-classification distinction is specific to Entire's hook-manager severity model, which Partio does not implement.

## entireio-cli-issues #2260: Read deny rule breaks auto/unattended permission modes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260 (Read deny rule for .entire/metadata/** breaks auto/unattended permission modes)`
- reason: `irrelevant`
- note: Partio does not write a Read deny rule to `.claude/settings.json` or any agent settings file. Partio adds gitignore entries for `.partio/sessions/`, `.partio/state/`, and `.partio/settings.local.json`, but not agent permission rules.

## entireio-cli-issues #2257: external agent settings scope leak

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2257 (adding external agent writes merged settings back to one scope, leaking values between settings.json and settings.local.json)`
- reason: `irrelevant`
- note: Partio has no `partio agent add` command. Agents are selected via the `PARTIO_AGENT` environment variable or the `agent` config field. There is no merge-and-write-back path for agent configuration.

## entireio-cli-issues #2256: AppendCheckpointTrailer emits trailer git rejects

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256 (AppendCheckpointTrailer can emit a trailer the final-block parser and git reject)`
- reason: `irrelevant`
- note: Partio's `AmendTrailers` (internal/git/amend_trailers.go) appends `Partio-Checkpoint: <12-char-hex>` and `Partio-Attribution: <N>% agent`. Both values are always ASCII-safe. The existing non-deterministic key ordering is already filed as issue #717. No trailer content Partio produces would be rejected by git's final-block parser.

## entireio-cli-issues #2255: forged Entire-Checkpoint lines select unrelated checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `irrelevant`
- note: Partio reads checkpoints from the orphan branch (partio/checkpoints/v1) by traversing the git object tree, not by parsing commit trailers of user commits. `FindByBranch` (internal/checkpoint/find_by_branch.go) uses `git ls-tree` on the checkpoint branch. A forged `Partio-Checkpoint` trailer in a user's commit body does not affect checkpoint storage or retrieval.

## entireio-cli-issues #2249: stale configure --agent hints in error messages

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249 (stale entire configure --agent hints in investigate and review error messages)`
- reason: `irrelevant`
- note: Partio has no `partio configure --agent` command. Agent selection uses `PARTIO_AGENT` or the `agent` config field. There are no investigate or review commands in Partio.

## entireio-cli-issues #2237: re-install discards newer hook and chains to stale backup

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2237 (re-install discards another tool's newer hook and chains to stale backup)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #710 (installHooks silently overwrites the original hook backup when a foreign hook appears on reinstall). Not re-filed.

## entireio-cli-issues #2218: Windows entire.exe version 0.0.0.0 in PE metadata

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2218 (Windows entire.exe reports version 0.0.0.0 in PE metadata)`
- reason: `irrelevant`
- note: Partio does not produce Windows PE binaries in its current release process. Windows binary support is already filed as issue #695. When implemented, PE version embedding will be relevant.

## entireio-cli-issues #2215: Claude Code SubagentStop dropped (tool_use_id correlation)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop is dropped: Entire correlates on tool_use_id which the hook does not send)`
- reason: `irrelevant`
- note: Partio has no SubagentStop hook and no per-event correlation. Session data is read from the JSONL file in full at commit time. There is no event-level streaming or per-tool-call tracking.

## entireio-cli-issues #2204: OPF writers still CAS via go-git

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2204 (OPF writers still CAS checkpoint refs via go-git, racing native update-ref)`
- reason: `irrelevant`
- note: Partio uses only native git plumbing commands (hash-object, mktree, commit-tree, update-ref). There is no go-git dependency and no OPF layer.

## entireio-cli-issues #2203, #2202, #2201: infrastructure and test isolation

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2203 (install scripts never shellchecked), #2202 (make test git isolation structural), #2201 (auth-go lock dir isolation)`
- reason: `irrelevant`
- note: Partio has no install scripts (#2203). Test git isolation is already filed as Partio issue #696 (#2202). Auth-go is an Entire-specific dependency Partio does not use (#2201).

## entireio-cli-issues #2197, #2196: lock file accumulation and migration performance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2197 (persistent-ref lock files accumulate per checkpoint), #2196 (MigrateBranchToRefs re-runs take a flock per already-imported checkpoint)`
- reason: `irrelevant`
- note: Partio does not use persistent lock files for checkpoint refs. There is no branch-to-refs migration in Partio's architecture; checkpoints have always been stored on the orphan branch via git plumbing.

## entireio-cli-issues #2160: Goose and Qwen Code agent integrations

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2160 (feature request: Goose and Qwen Code agent integrations)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #681 (Add Goose coding agent integration) and #684 (Add Qwen Code coding agent integration). Not re-filed.

## entireio-cli-issues #2157: redaction corrupts Grok reasoning blocks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2157 (redaction corrupts Grok reasoning blocks: encrypted_content not covered by signature skip rule)`
- reason: `irrelevant`
- note: Partio has no Grok agent support and no reasoning-block or signature-skip rule in its redaction layer (`internal/redact/`). The redaction operates on plain text tokens; there is no structured block handling.

## entireio-cli-issues #2148: finalizeAllTurnCheckpoints has no total deadline

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2148 (finalizeAllTurnCheckpoints has no total deadline: slow remote outlasts agent hook timeout)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #700 (Post-commit hook has no total deadline: a hung operation blocks git commit forever). Not re-filed.

## entireio-cli-issues #2137, #2126: OpenCode Desktop and agent picker UI

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2137 (doesn't work with OpenCode Desktop), #2126 (agent picker hides last options in entire enable)`
- reason: `irrelevant`
- note: Partio has no OpenCode support (#2137). Partio's `partio enable` has no interactive agent picker; the agent is set via config field or environment variable (#2126).

## entireio-cli-issues #2121: semantic search 401 for EU jurisdiction

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2121 (semantic search returns 401 for EU-jurisdiction account)`
- reason: `irrelevant`
- note: Partio has no semantic search feature. Checkpoints are stored locally on an orphan branch with no cloud search backend.

## entireio-cli-issues #2115: flat 30s hook timeout is wrong in both directions

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2115 (hook timeouts: a flat 30s is wrong in both directions, cap is doing telemetry's job)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #700 (Post-commit hook has no total deadline). Not re-filed.

## entireio-cli-issues #2111: index emptied between staging and commit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2111 (index emptied between staging and commit resulting in git commit recording an empty tree)`
- reason: `irrelevant`
- note: The index-consumption issue specific to Partio's `git commit --amend` is already filed as issue #693 (AmendTrailers consumes the git index, silently including staged files in the amended commit). The specific emptying scenario described in #2111 is Entire-specific and involves a different code path.

## entireio-cli-issues #2098, #2091, #2089, #2087: performance, latency, and OPF

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2098 (CLI performance regression), #2091 (hook latency remaining work), #2089 (checkpoint explain scan limit false negative), #2087 (OPF 9th layer never runs on git-refs)`
- reason: `irrelevant`
- note: #2098 is Entire-specific (full-repository activation overhead). #2091 covers hook latency improvements beyond those already filed as Partio issue #700. #2089 is about `checkpoint explain` which Partio does not implement. #2087 is OPF-specific; Partio has no OPF layer.

## PRs #2074–#2569: dependency bumps, CI, cloud infrastructure, Entire-specific features

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2569`
- reason: `irrelevant`
- note: The bulk of PRs in this range are dependency version bumps (#2520, #2530, #2540), CI fixes (#2465, #2526, #2527, #2537, #2542, #2555, #2565, #2567, #2569), control-plane command unification (#2502, #2547), cloud infrastructure (trail RFD-026 migration #2507, repo protection #2516–#2517, GitLab #2528, cluster list #2480, mirror management #2529), auth consolidation (#2471, #2494–#2495, #2509–#2512, #2539), agent changes (IsPreview removal #2554, worktree hook wrapper #2487, Codex transcript #2536, #2501), checkpoint delivery (#2519, #2521–#2522, #2533, #2550), and Entire-specific fixes (#2463–#2475, #2476–#2499). None of these map to Partio's domain of local git hook-based session capture and checkpoint storage on an orphan branch.
