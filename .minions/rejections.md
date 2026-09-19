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

## No new pulls > #2073

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–present`
- reason: `irrelevant`
- note: `gh pr list --repo entireio/cli` returned no pull requests with number > 2073 in this run. Nothing to ingest.

## changelog 0.10.3: subagent survival into checkpoints, process-ancestry session linking, zombie self-heal, redaction config, doctor trace, telemetry, search TUI, git status lock fix, condensation/OPF fixes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3`
- reason: `irrelevant`
- note: The bulk of 0.10.3 covers Entire-specific cloud infrastructure. Subagent checkpoint survival and process-ancestry session linking are analogous to Partio proposals already filed as issues #673 and #677. Zombie self-heal was already filed as #679. Redaction engine config, doctor trace, telemetry signals, and search TUI are Entire-specific. The `git status --no-optional-locks` fix is for Entire's background status calls; Partio does not call `git status` in any hook. The condensation/OPF pipeline fixes have no equivalent in Partio's orphan-branch checkpoint model.

## changelog 0.10.4: Entire-native repo cloning, jurisdiction dispatch, trail JSON output, os.Root filesystem anchors, OAuth and plugin resolver hardening

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4`
- reason: `irrelevant`
- note: All 0.10.4 items are Entire-specific: Entire-native `/et/` path cloning, `--jurisdiction` dispatch flag, trail `original_branch` JSON field, reduced runner scaffolding, `os.Root` filesystem confinement for Entire's plugin/server paths, OAuth cross-host redirect restriction, and plugin resolver `$PATH` hardening. Partio has no equivalent paths for any of these.

## changelog 0.10.5: async mirror creation, full.jsonl release after condensation, cross-federation git clone, cluster login domain restriction, checkpoint explain verification

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5`
- reason: `irrelevant`
- note: All 0.10.5 items are Entire-specific: async mirror requests, session full.jsonl release on condensation (Partio has no condensation), cross-federation clone auth, cluster login domain scoping, and checkpoint explain server-response verification. None of these concepts exist in Partio's orphan-branch model.

## changelog 0.10.6: Windows PowerShell installer, runner setup, trail pre-push hook, symlink validation in enable, concurrent checkpoint/OPF fixes, UTF-8 truncation, settings trust security

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6`
- reason: `irrelevant`
- note: All 0.10.6 items are Entire-specific: Windows PowerShell installer (Partio has no Windows installer), enhanced runner setup (no runners), trail pre-push hook (no trails), symlink validation in Entire's enable flow (different from Partio's hook installation — Partio's install.go renames symlinks via os.Rename before writing, so the Entire-specific `--force` clobbering scenario does not arise), concurrent OPF checkpoint writes (no OPF), UTF-8 truncation in model prompts (Partio does not send prompts to models in its core hooks), and settings trust gate hardening for external_agents (no external agent commands in Partio).

## issues #2518, #2402, #2401, #2393, #2378, #2368, #2362, #2350, #2360, #2204, #2197, #2196, #2148, #2091, #2087: Entire cloud/OPF/condensation infrastructure bugs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2518, #2402, #2401, #2393, #2378, #2368, #2362, #2350, #2360, #2204, #2197, #2196, #2148, #2091, #2087`
- reason: `irrelevant`
- note: These issues describe bugs in Entire's cloud and storage infrastructure: mirror credential routing (#2518), ULID checkpoint ref case-collision on case-insensitive filesystems (#2402/#2401), git-refs async push bookkeeping (#2393), uncondensed checkpoint cleanup races (#2378), backfilled session token accounting (#2368), multi-session status collapsing (#2362), shadow-ref cleanup on malformed session state (#2350), pushurl-only remote election (#2360), go-git OPF CAS race (#2204), persistent lock file accumulation (#2197), MigrateBranchToRefs per-checkpoint lock overhead (#2196), finalizeAllTurnCheckpoints deadline (#2148), hook latency remaining work (#2091), and OPF 9th-layer bypass (#2087). None of these concepts (ULID shards, OPF layers, condensation, async push queues, cloud mirrors) exist in Partio's orphan-branch model.

## issues #2394, #2318, #2276, #2218, #2203, #2098: platform and installer bugs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394, #2318, #2276, #2218, #2203, #2098`
- reason: `irrelevant`
- note: These cover Entire-specific platform and installer issues: Homebrew tap-before-trust docs error (#2394), install.sh hiding API auth errors (#2318), Windows installer URL serving a login page (#2276), Windows PE metadata reporting version 0.0.0.0 (#2218), install.sh missing shellcheck (#2203), and a CLI performance regression after enabling full-repository features (#2098). Partio has no install.sh, no Windows installer, no PE metadata, and no full-repository indexing feature.

## issues #2328, #2160: Devin and Goose/Qwen Code agent integrations

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2328 (Add Devin agent support), #2160 (Feature request: Goose and Qwen Code agent integrations)`
- reason: `irrelevant`
- note: Partio already has an open proposal for Goose integration (#681). Devin and Qwen Code have no published session JSONL format or hook interface documented in this repository; adding them would require a new detector with no grounding in the checked-out tree. These are new integrations, not fixes to existing behavior.

## issues #2271, #2121, #2126, #2137, #2089: docs, semantic search, agent picker UI, OpenCode Desktop, explain scan limit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271, #2121, #2126, #2137, #2089`
- reason: `irrelevant`
- note: Docs security page data-residency (#2271) is marketing documentation. Semantic search 401 for EU accounts (#2121) and checkpoint explain scan-limit false negative (#2089) require Entire's cloud search backend. Agent picker height bug (#2126) is in Entire's interactive TUI. OpenCode Desktop hooks (#2137) requires OpenCode support. None of these exist in Partio.

## issues #2201, #2202, #2203: Entire test and CI infrastructure

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2201, #2202, #2203`
- reason: `irrelevant`
- note: auth-go lock directory isolation (#2201), structural git isolation in tests (#2202), and install.sh shellcheck (#2203) are Entire's internal test/CI infrastructure. Partio has its own test conventions (table-driven, t.TempDir, no external frameworks) that don't involve these components.

## issues #2115, #2098, #2091: hook timeouts, performance regression, hook latency

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2115, #2098, #2091`
- reason: `irrelevant`
- note: Entire's flat 30s hook timeout (#2115) and hook latency work (#2091) are specific to Entire's multi-agent stop/start hook pipeline and telemetry. Partio's pre-commit and post-commit hooks do not have configurable timeouts. The CLI performance regression (#2098) is tied to Entire's full-repository indexing feature which Partio does not have.

## issues #2260, #2257, #2249: metadata read rules, external agent settings leak, stale configure hints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260, #2257, #2249`
- reason: `irrelevant`
- note: The `.entire/metadata/` read-deny rule (#2260) is specific to Entire's Claude Code permission config. The external-agent settings scope leak (#2257) requires Entire's `agent add` command. Stale `entire configure --agent` hints (#2249) reference Entire's configure subcommand. Partio has none of these: no metadata permission rules, no `partio agent` subcommand, no `partio configure` command.

## issues #2255, #2256: forged checkpoint trailer lines, AppendCheckpointTrailer grammar mismatch

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (forged Entire-Checkpoint lines), #2256 (AppendCheckpointTrailer emits invalid trailer)`
- reason: `irrelevant`
- note: Partio writes `Partio-Checkpoint: <id>` trailers via `git.AmendTrailers` (`internal/git/amend_trailers.go`) but does not parse checkpoint IDs back from commit bodies — the checkpoint branch is the source of truth, not commit trailers. The trailer forging and grammar-mismatch bugs in Entire depend on a round-trip where the trailer is parsed to select or mutate checkpoint storage; Partio has no such round-trip.

## issues #2416, #2410: doctor stuck-session interactive prompt, enable --force symlink clobber

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416 (doctor stuck-session prompt bypasses interactive check), #2410 (enable --force through symlink)`
- reason: `irrelevant`
- note: Partio's `doctor` command (`cmd/partio/doctor.go`) has no interactive prompts — it reports status and exits, so #2416 has nothing to bypass. Partio's `enable` command has no `--force` flag; the specific clobber scenario in #2410 requires `--force` to skip the backup-existence check that would otherwise prevent it.

## issues #2274, #2263: disable doesn't reach linked worktrees, lefthook overwrite classification

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274 (disable doesn't reach linked worktrees), #2263 (lefthook classified as not overwriting hooks)`
- reason: `irrelevant`
- note: Partio's `disable` removes git hooks (installed to `git rev-parse --git-common-dir`, shared across worktrees) and optionally removes `.partio/`; it does not write `enabled: false` to `settings.local.json`, so the per-worktree re-enable scenario in #2274 does not apply. Partio's hook-manager detection (`internal/git/hooks/detect_hook_managers.go`) emits the same warning for all managers regardless of whether they overwrite hooks; there is no `OverwritesHooks` classification to be wrong about, so #2263 has no equivalent bug to port.

## issues #2215, #2111: Claude Code SubagentStop dropped, git index emptied between staging and commit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop dropped), #2111 (index emptied between staging and commit → empty tree)`
- reason: `irrelevant`
- note: Partio has no subagent session tracking; it captures one JSONL transcript per commit via git hooks rather than correlating `SubagentStop` events with hook invocations (#2215). The empty-tree-via-index-lock race (#2111) requires concurrent `git status` writes from Entire's background stop hooks; Partio does not call `git status` in any hook path.

## issues #2367, #2237, #2264: session dir ignores CLAUDE_CONFIG_DIR, hook backup overwritten on reinstall, status reports installed for foreign hooks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2367, #2237, #2264`
- reason: `irrelevant`
- note: All three ideas extracted from these issues were verified against the checked-out tree and the premises held, but proposals for identical findings had already been filed in a prior run: #708 (FindSessionDir ignores CLAUDE_CONFIG_DIR), #710 (installHooks silently overwrites existing backup), #704 (status reports "Hooks: installed" for foreign hook files). Not re-filed to avoid duplicates.
