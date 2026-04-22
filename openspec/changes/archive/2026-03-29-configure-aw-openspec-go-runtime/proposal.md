## Why

The `ci-aw-openspec-verification` workflow currently provisions Go with `actions/setup-go` and `go-version-file: go.mod`, which configures Go in the runner workspace used for dependency installation and repository bootstrap. The `gh-aw` agent environment has its own runtime model, and the published `runtimes:` reference supports an explicit `go.version` string but does not document any equivalent file-based or external version source, so the workflow needs an explicit frontmatter Go version in addition to the existing runner setup.

## What Changes

- Update the `ci-aw-openspec-verification` workflow contract to require both frontmatter `runtimes.go.version` for the agent environment and an explicit `actions/setup-go` step for the runner environment used during repository setup.
- Require the frontmatter Go runtime version to be set to the repository's supported Go release and kept in sync with future `go.mod` version changes.
- Preserve the existing `actions/setup-go` behavior that reads `go.mod` for the runner workspace so dependency installation and bootstrap commands continue to use repository-declared Go configuration.
- Add Makefile targets to check for drift between `go.mod` and `runtimes.go.version`, and to sync the workflow frontmatter from `go.mod` on demand.
- Extend `make check-lint` to run the drift check so mismatches fail local validation and CI.
- Document the operational trade-off that frontmatter Go version alignment becomes an explicit maintenance responsibility because `runtimes:` cannot read the version directly from `go.mod`.
- Leave Renovate integration out of scope for this change; the sync target should exist for manual or future automation use, but `make renovate-post-upgrade` should remain unchanged for now.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: require separate Go configuration for runner bootstrap and agent review environments by keeping `actions/setup-go` for the runner while adding explicit `runtimes.go.version` for the agent
- `makefile-workflows`: add explicit targets for checking and syncing the workflow Go runtime, and require `check-lint` to fail when the workflow frontmatter drifts from `go.mod`

## Impact

- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- `Makefile`
- `openspec/specs/ci-aw-openspec-verification/spec.md`
- `openspec/specs/makefile-workflows/spec.md`
- Workflow maintenance guidance for keeping the frontmatter Go version aligned with `go.mod`
