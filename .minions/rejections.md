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

## Windows: prepare-commit-msg hook hangs while Claude Code session is active

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2725, #2726 (Windows prepare-commit-msg hook stalls indefinitely; Windows CLI 0.11.4 prepare-commit-msg hook stalls when finalizing temporary checkpoint)`
- reason: `irrelevant`
- note: Partio does not use a prepare-commit-msg hook. Its git hooks are pre-commit, post-commit, and pre-push only. The stall is specific to Entire's per-prompt checkpoint finalization in that hook.

## Redaction misses short sk- keys (LiteLLM virtual keys)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2728 (Redaction misses short sk- keys written as bare values in a prompt)`
- reason: `irrelevant`
- note: Partio has no sk- pattern in `internal/redact/patterns.go`. However, Partio's entropy layer (threshold 4.5 bits/char, minLen 20) catches high-entropy tokens at or above that length regardless of prefix. A typical LiteLLM virtual key (`sk-xxxx…` ≥ 20 chars with random suffix) scores above the threshold and is already redacted by the entropy pass. Short ambiguous sk- values below 20 chars are excluded deliberately; they produce too many false positives on ordinary identifiers.

## Docs pages, website rendering, and release notes tasks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2732–#2740 (write lifecycle/search/agent_help_cmd/repo/auth doc pages, add checkpoint and agent changes to guides, publish v0.11.4 release notes); #2745–#2746 (raw HTML in READMEs; extra blank line in code blocks on entire.io)`
- reason: `irrelevant`
- note: These are documentation and website tasks specific to entire.io. Partio has no hosted doc site and no analogous doc pages.

## Remote helper credential fallback and service accounts

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2748 (Remote helper: fall back to login when credential helper returns non-Entire JWT); #2894 (Documented way to create service accounts / API keys for non-interactive CLI auth)`
- reason: `irrelevant`
- note: Partio has no authentication layer, remote helper, or credential system. Access to the checkpoint orphan branch uses the repo's existing git remote configuration.

## OPF pre-push CRLF batching leaves CR in JSON documents

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2806 (OPF pre-push batching leaves CR intact, causing multiple JSON documents for CRLF input)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes the checkpoint branch with a plain `git push` call. There is no OPF (Object Pack Format) batching layer.

## Bare AWS access key IDs not redacted without paired secret

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2808 (Bare AWS access key IDs are not redacted without a paired secret)`
- reason: `irrelevant`
- note: Partio's `aws-access-key-id` pattern (`internal/redact/patterns.go`) is `(?i)(AKIA|ASIA|AROA|AIDA|ANPA|ANVA|APKA)[A-Z0-9]{16}`, which matches a bare access key ID with no requirement for a paired secret. The premise does not hold for Partio.

## Support OpenCode 2.0 and Goose agents

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2810 (Support OpenCode 2.0); #2812 (Add goose support)`
- reason: `irrelevant`
- note: Partio has no OpenCode or Goose agent detector/parser. Adding new agents is a self-contained effort with no basis in the existing codebase structure for those agents; these are new integrations, not refinements of existing behaviour.

## Codex Windows hooks: emit native commandWindows

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2884 (Codex Windows hooks: emit native commandWindows and preserve it across managed rewrites)`
- reason: `irrelevant`
- note: Partio's Codex support uses standard JSONL parsing only; there is no Windows-native hook format or `commandWindows` field in Partio's Codex parser.

## Post-commit background fetch leaves large tmp_pack files

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2895 (Post-commit background fetch leaves a 115 MB tmp_pack_* file after every commit)`
- reason: `irrelevant`
- note: Partio's post-commit hook does not perform a background fetch. Checkpoints are written to a local orphan branch with no network activity in the hook itself.

## Generated hooks fail silently when binary cannot run (mise shim)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2901 (Generated hooks fail silently when `entire` is on PATH but can't run, e.g. a mise shim with no version set); PR #2922`
- reason: `irrelevant`
- note: Partio hooks call `partio _hook <type>` and exit non-zero if the call fails. A failed hook exits with a non-zero code, which git surfaces as a hook error. Silent failure would require the hook script to suppress exit codes, which Partio's generated scripts do not do. The scenario is specific to Entire's hook design.

## checkpoint explain 120s foreground fetch timeout

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2902 (checkpoint explain: 120s foreground fetch timeout turns the sidecar metadata fetch into infinitely repeated discarded work)`
- reason: `irrelevant`
- note: Partio has no `checkpoint explain` command and no foreground sidecar fetch mechanism.

## Subagent task record skipped when subagent commits its own work

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2903 (Subagent task record skipped when the subagent commits its own work — "no file changes detected")`
- reason: `irrelevant`
- note: Partio writes one checkpoint per git commit and has no concept of subagent task records or subagent-initiated commits. Each commit triggers the hook independently.

## git commit -q anchored to branch tip rather than the committing turn

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2908 (import claude-code: turns that commit with "git commit -q" are anchored to the branch tip)`
- reason: `irrelevant`
- note: Partio has no `partio import` command. Session-to-commit linking is handled by the post-commit hook reading the live session at commit time.

## No CLI command returns why a gates check failed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2924 (No CLI command returns why an Entire Gates check failed)`
- reason: `irrelevant`
- note: Partio has no Gates feature or CI gate integration.

## Windows: background hooks flash black windows and steal focus

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2926 (Windows: background Entire hooks repeatedly flash black windows and steal focus)`
- reason: `irrelevant`
- note: Partio hooks are POSIX shell scripts. Windows-specific console window management is an Entire-specific concern; Partio has no Windows hook implementation.

## OpenCode v2 plugin fails (no default export)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2928 (OpenCode v2 fails to load the Entire plugin: v1 plugin format with no default export)`
- reason: `irrelevant`
- note: Partio has no OpenCode plugin or plugin format.

## Pi extension ENTIRE_PI_NESTED survives session resets

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2930 (Pi extension: ENTIRE_PI_NESTED survives /reload, /new and /resume; unhandled stdin error can crash Pi)`
- reason: `irrelevant`
- note: Partio has no Pi extension.

## Configure Trail gates as code (per-repo and per-org)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2932 (Configure Trail gates as code per repository and per organization)`
- reason: `irrelevant`
- note: Partio has no Trail CI integration or gates configuration.

## Windows: large Codex transcript saturates CPU during post-commit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2933 (Windows: large Codex transcript saturates CPU during post-commit 12m43s, interrupted capture leaves checkpoint trailer without remote data)`
- reason: `irrelevant`
- note: Partio's post-commit hook reads the Codex JSONL file synchronously. There is no background post-processing loop that could saturate CPU; the session file is read once and written to the orphan branch.

## Show transcript images in session timeline

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2942 (Show transcript images/screenshots in the entire.com session timeline)`
- reason: `irrelevant`
- note: Partio has no web UI or session timeline. JSONL transcript content is stored verbatim on the orphan branch but not rendered.

## Entire doctor conflicts with Brain setup

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2949 (Entite doctor conflict with brain setup)`
- reason: `irrelevant`
- note: Partio has no Brain integration. The conflict is specific to Entire's Brain product and its interaction with the Entire doctor command.

## Safer defaults for public remotes: checkpoint push on by default, redaction bypasses

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2951 (Safer defaults for public remotes: checkpoint push is on by default; Bearer headers and binary transcripts slip past redaction)`
- reason: `irrelevant`
- note: Partio's redaction is configurable via `redact.enabled` and is applied before writing to the orphan branch. Bearer token redaction is already in `internal/redact/patterns.go`. Binary transcript handling does not apply since Partio stores text JSONL. The push-on-by-default concern is Entire-specific; Partio's `push_sessions` is off by default. No new Partio proposal emerges from this compound issue.

## entire disable --uninstall leaves Entire's pre-push hook behind Brain's wrapper

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2957 (entire disable --uninstall leaves Entire's pre-push hook running behind entire-brain's wrapper)`
- reason: `irrelevant`
- note: Partio has no Brain product. The hook layering issue is specific to Entire's multi-product architecture.

## Re-wrapping a pre-push hook rotates the user's own hook out of the chain

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2960 (Re-wrapping a pre-push hook that chains through Entire's moved copy rotates the user's own hook out of the chain)`
- reason: `irrelevant`
- note: Already captured in Partio issue #747 (hook chaining breaks $0-based dispatchers). Not re-filed.

## PII phone redaction misses non-North American numbers

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2976 (PII phone redaction misses every non-North American number)`
- reason: `irrelevant`
- note: Partio has no phone number redaction pattern in `internal/redact/patterns.go`. Partio's redaction covers API keys, tokens, and high-entropy strings; it does not include PII patterns such as phone numbers or addresses.

## PII address redaction misses European street addresses

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2977 (PII address redaction misses European street addresses)`
- reason: `irrelevant`
- note: Partio has no address redaction pattern. Same reason as #2976.

## Sibling-worktree commit linking picks stale idle session

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2979 (Sibling-worktree commit linking picks a session idle for days; every commit from another worktree condenses its multi-GB transcript)`
- reason: `irrelevant`
- note: Partio's `FindLatestSession` searches up from the repo root and takes the most recently modified session directory. The multi-worktree session mismatch scenario for Partio is already captured in issues #564 and #787. Not re-filed.

## Post-push session attach publishes linked_commits but checkpoint remains absent

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2980 (Post-push session attach publishes linked_commits but checkpoint remains absent from activity and Dispatch)`
- reason: `irrelevant`
- note: Partio has no post-push session attach command or Dispatch integration.

## Promotional credit not granted after first preset agent run

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2981 (First preset Agent run completed but no $100 promotional credit was granted)`
- reason: `irrelevant`
- note: Billing issue; no software behaviour to improve.

## git-remote-entire push memory and unbounded replica cache

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2985 (git-remote-entire: push uses ~4× the pack size in memory; replica cache grows without bound)`
- reason: `irrelevant`
- note: Partio has no git remote helper implementation. Checkpoints are pushed via a plain `git push` shell call.

## Windows: batch checkpoint ref push fails beyond ~340 refs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2986 (Windows: batch checkpoint ref push fails once queue passes ~340 refs, per-ref fallback drains only ~30 per push)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes a single branch reference (`partio/checkpoints/v1`) unconditionally. There is no batched ref queue.

## changelog 0.11.5–0.12.0: agent diagnostics, checkpoint search, session attach GA

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.5–0.12.0 (agent diagnostics, checkpoint semantic search, session attach GA, trail runners, Cursor project directory naming, manual push policy introduction, prompt-out-of-commits, Brain pre-push wrapper)`
- reason: `irrelevant`
- note: The new features in these versions (semantic search, session attach command, manual push policy UI, trail runners, agent diagnostics) are all Entire-specific products or UX layers with no direct Partio analog. The Brain pre-push wrapper and Cursor naming are platform-specific. No new Partio proposal emerges from these entries that is not already covered by existing issues.

## changelog 0.12.2: entireio enable lists config file sources

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.12.2 (entire enable now lists the config files it will write before prompting)`
- reason: `irrelevant`
- note: Already captured in Partio issue #793 (partio enable should list config files before writing). Not re-filed.

## PRs #2914–#2987: auth, CI, trail commands, Codex/Pi fixes, Brain wrapper, transcript changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2914–#2987 (credential root fix, mirror test, race-condition sharding, Windows canary, subagent store fix, checkpoint fetch restart fix, manual push policy, hook shim warning, commit linking, manual push hold, Claude turn baseline, installer path fix, dial latency fix, consistent flag resolve, push policy prompt, config JSON fail-early, squash strategy trailers, review snippet limit, default-branch naming, checkpoint trailer parsing, transcript images, trail runner replacement, remote checkpoint filter, file runner removal, Pi lifecycle fix, Cursor naming, Codex streaming adapter, trail 401 hint, trail Core resolve, trail URL selector, prompt-out-of-commits, doctor Brain wrapper count, finalize failure reporting, git-dir cache, Codex image URI redaction, service-account keys, core.hooksPath in-place edit, trail commands batch, Windows headless workers, Pi transcript decode, import q anchor, commit signature preservation, Claude tool result blocks, trail no-base branch, trail findings cards, git-remote pack hold)`
- reason: `irrelevant`
- note: The bulk of PRs in this range are CI fixes, trail command additions, Codex/Pi/Cursor-specific session handling, Brain integration, auth improvements, and Entire-specific feature work. The two PRs implementing ideas already captured as Partio proposals (#2966 for core.hooksPath → Partio #797; #2978 for commit signature preservation → Partio #801) are recorded as duplicates, not rejections.

## PRs #2074–#2913: dependency bumps, trail features, cloud infrastructure, platform-specific

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2913 (dependency updates, trail gate management, cloud native repos, shadow branch GC, session cache memory diagnostics, Windows-specific hook wrappers, Pi/OpenCode/Cursor plugin changes, ULID shard handling, git-remote-entire memory cache, auth improvements, CI fixes)`
- reason: `irrelevant`
- note: PRs in this range are cloud infrastructure (native repos, dispatch routing, EU mirrors), trail gate features, session cache and memory diagnostics, Windows hook wrappers, Pi/OpenCode/Cursor/Codex plugin changes, shadow branch GC and ULID sharding, git-remote-entire memory cache, and a large volume of dependency bumps and CI fixes. None are applicable to Partio's git-hook + orphan-branch checkpoint architecture. Captured as a batch alongside the #2914–#2987 batch above.

## `--author` / `--branch` filters for checkpoint list

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2718 (add --author and --branch flags to checkpoint list output)`
- reason: `irrelevant`
- note: Partio has no checkpoint list command. Checkpoint data is stored on the orphan branch and accessible only via raw git commands or `partio rewind`.

## UserPromptSubmit hook timeouts under per-session flock contention

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2689 (UserPromptSubmit hook blocks ~30s and times out due to per-session flock contention)`
- reason: `irrelevant`
- note: Partio has no UserPromptSubmit hook. Its hooks are standard git lifecycle hooks without per-session file locking.

## Shadow branches quadratic storage growth

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2663 (shadow branches accumulate quadratic storage as commit history grows)`
- reason: `irrelevant`
- note: Partio has no shadow branches. Checkpoints are stored on a single orphan branch (`partio/checkpoints/v1`) without per-commit branch proliferation.

## Background subagent commits attributed to wrong session

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2653 (background subagent commits attributed to the wrong session / wrong agent percentage)`
- reason: `irrelevant`
- note: Partio has no background subagent concept. Attribution is calculated per-commit by `attribution.Calculate()` based on the single active session detected at commit time.

## Native repos as primary host

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2623 (native repos as primary host — move git storage from git-remote-entire to native git server)`
- reason: `irrelevant`
- note: Partio stores data on a local orphan branch using git plumbing. There is no cloud host, git-remote-entire backend, or remote storage migration path.

## Session resume overwrites untracked files

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2604 (session resume overwrites user's untracked files when restoring to a checkpoint)`
- reason: `irrelevant`
- note: Partio's `partio rewind` creates a new branch from the checkpoint commit rather than switching the working tree in place. Untracked files in the current worktree are not touched.

## Session resume Unicode NFC/NFD branch name mismatch

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2551 (session resume fails to find the branch when the name contains Unicode composed/decomposed differently on macOS vs Linux)`
- reason: `irrelevant`
- note: Partio's rewind command looks up checkpoints by ID or by branch name from `git rev-parse --abbrev-ref HEAD`; there is no user-supplied Unicode branch name comparison.

## `entire enable --force` exits 0 without a TTY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2535 (entire enable --force silently exits 0 when no TTY is available, skipping agent selection prompt)`
- reason: `irrelevant`
- note: Partio's `partio enable` has no `--force` flag and no interactive agent selection prompt. The agent is set via config or environment variable.

## ULID checkpoint refs unreadable due to shard case-collision

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2401 (ULID-based checkpoint refs become unreadable on case-insensitive filesystems due to shard path collision)`
- reason: `irrelevant`
- note: Partio uses a single orphan branch (`partio/checkpoints/v1`) rather than per-checkpoint git refs with ULID sharding. The case-collision problem does not exist in Partio's storage model.

## git-refs writes return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393 (git-refs backend acknowledges writes before confirming durable push, risking silent data loss)`
- reason: `irrelevant`
- note: Partio uses local git plumbing (`hash-object`, `mktree`, `commit-tree`, `update-ref`) with no git-refs backend or async push acknowledgement. The durability model is standard git object storage.

## `partio clean` makes uncondensed checkpoints unreachable

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 (partio clean removes checkpoint refs before sessions are condensed, making them unreachable)`
- reason: `irrelevant`
- note: Partio's `clean` command (`cmd/partio/clean.go`) is a stub that lists entries and tells the user to run `partio reset`. It performs no actual cleanup, so the described data-loss scenario cannot occur yet.

## Post-push cleanup deletes uncondensed shadow ref

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350 (post-push cleanup hook deletes the shadow ref for a session that hasn't been condensed yet)`
- reason: `irrelevant`
- note: Partio has no shadow refs and no post-push cleanup hook. The pre-push hook only pushes the checkpoint branch; it does not delete any refs.

## Owner access required for auto-created organisations

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2989 (auto-created orgs require owner access for admin/billing setup)`
- reason: `irrelevant`
- note: Partio has no organisation creation, billing, or cloud account management. Checkpoints are stored on a local orphan branch; there is no Entire cloud backend.
