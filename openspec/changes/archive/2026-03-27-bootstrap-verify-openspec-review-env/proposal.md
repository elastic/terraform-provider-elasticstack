## Why

The `openspec-verify-label` review workflow currently tells the agent to install only Node dependencies before verification, even though repository checks and agent-invoked commands can rely on Go and Terraform tooling as well. Because the repository now declares `go 1.26.1`, review runs that fall back to the runner's default Go installation can fail before verification work even starts.

## What Changes

- Update the OpenSpec verification workflow contract so the review environment bootstraps the same core toolchain layers as the `lint` job before the agent begins reasoning.
- Require the review path to provision Node using the version range declared in `package.json` engines, Go from `go.mod`, and Terraform CLI in a way that does not depend on runner-default language versions.
- Require the review workflow to run `make setup` in the review workspace after those runtimes are installed so `npx openspec` and agent-invoked Go commands can run against repo-standard dependencies.
- Regenerate the compiled `.github/workflows/openspec-verify-label.lock.yml` after the source workflow is updated.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: bootstrap the review environment with repo-standard Node, Go, Terraform, and dependency setup before agent verification starts

## Impact

- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- Review-environment bootstrap behavior for `verify-openspec` runs
- Alignment with the `lint` job's initial setup in `.github/workflows/test.yml`
- Repository bootstrap commands referenced from `Makefile`
