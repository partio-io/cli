# Migrate Minions to partio-io/cli

## Context

The minion system currently uses `partio-io/minions` as the principal repo — all proposal issues, workflows, and status tracking live there. This is awkward because the actual product being improved is `partio-io/cli`. We want cli to be the single source of truth: issues, workflows, and minion orchestration all live in cli.

## What to Migrate

### 1. Issues: All 93 open issues → partio-io/cli

For each of the 93 open issues on `partio-io/minions`:
- Create a new issue on `partio-io/cli` with the same title, body, and labels
- Close the original issue on `partio-io/minions` with a comment linking to the new issue

**Script approach** (bash with `gh`):
```bash
# For each open issue on partio-io/minions:
# 1. Read title, body, labels
# 2. Create on partio-io/cli with same content
# 3. Comment + close on partio-io/minions
```

Labels to ensure exist on `partio-io/cli` first:
- `minion-proposal`
- `minion-approved`
- `minion-failed`
- `minion-executing`
- `minion-done`
- `minion-planning`

### 2. Workflows → partio-io/cli

Move these workflows from `partio-io/minions` to `partio-io/cli`:
- `minion.yml` — main task execution
- `plan.yml` — plan generation
- `propose.yml` — changelog monitoring + proposal creation
- `approve.yml` — auto-approve proposals after review window
- `doc-minion.yml` — documentation PR generation

Changes needed in each workflow:
- The principal repo checkout now uses `actions/checkout@v4` directly (it IS the repo)
- `project.yaml` path adjusts (it's in the same repo now, not a separate checkout)
- References to `${{ github.repository }}` already resolve correctly

### 3. Project config update

Update `.minions/project.yaml` principal to point to `partio-io/cli`:
```yaml
principal:
  name: cli
  full_name: partio-io/cli
```

### 4. CLI code updates

Update hardcoded fallback defaults:
- `cmd/minions/propose.go` — fallback `issueRepo` changes from `partio-io/minions` to `partio-io/cli`
- `cmd/minions/approve.go` — fallback `--repo` default
- `internal/pipeline/pipeline.go` — fallback `principalRepo`
- `internal/pr/crosslink.go` — fallback in `fullNameFn`

These are the backward-compat fallbacks for when no `project.yaml` exists. Since cli is now the principal, they should default to cli.

## Execution Order

1. Create labels on `partio-io/cli`
2. Migrate all 93 issues (create on cli, close with link on minions)
3. Update `project.yaml` principal
4. Update fallback defaults in Go code
5. Copy workflows to `partio-io/cli/.github/workflows/`
6. Adjust workflow paths for cli-as-principal

## Verification

- `gh issue list --repo partio-io/cli --label minion-proposal` shows 93 issues
- `gh issue list --repo partio-io/minions --state open` shows 0
- Workflows in cli repo trigger correctly on label/comment events
- `minions propose --dry-run` targets cli repo
- `minions approve --dry-run` lists proposals from cli repo
