## Why

The `kibana-spec-impact` workflow's `pre_activation` job checks out the repo-memory branch (`memory/kibana-spec-impact`) to `/tmp/gh-aw/repo-memory/kibana-spec-impact`. GitHub Actions `actions/checkout@v4` requires the `path` input to be a **relative path under `$GITHUB_WORKSPACE`**. An absolute path to `/tmp` causes the pre-activation step to fail, which prevents the deterministic impact check from running and gates the entire agent workflow.

This workflow is currently non-functional. Fixing it restores automated detection of Kibana OpenAPI/spec changes that may require Terraform provider updates.

## What Changes

- Change the pre-activation `actions/checkout` step for the repo-memory branch to use a **workspace-relative path** (e.g. `gh-aw-repo-memory/kibana-spec-impact` instead of `/tmp/gh-aw/repo-memory/kibana-spec-impact`)
- Update the `--memory` flag path in the `go run ./scripts/kibana-spec-impact pre-activation` command to match the new checkout location
- Re-generate the compiled lockfile (`.github/workflows/kibana-spec-impact.lock.yml`) via `scripts/compile-workflow-sources`

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `ci-kibana-spec-impact-issues`: adds a requirement that the pre-activation repo-memory checkout path must be workspace-relative to comply with `actions/checkout@v4` validation

## Impact

- **`.github/workflows-src/kibana-spec-impact/workflow.md.tmpl`** — source template for the pre-activation checkout path and `go run` command
- **`.github/workflows/kibana-spec-impact.lock.yml`** — compiled/lockfile produced by the workflow compiler (must be regenerated)
- **Scripts** — none; the `kibana-spec-impact` Go tool accepts any `--memory` path, no code changes needed
