## Why

The `schema-coverage-rotation` workflow now relies on repository-local Go scripts and Node-based repository setup, but it still depends on the runner's default toolchain. Recent failures show that this is not sufficient: the default Go toolchain is too old to execute the memory helper commands reliably, so the workflow needs the same repository toolchain bootstrap pattern already used by `openspec-verify-label`.

## What Changes

- Add deterministic repository toolchain setup steps to the `schema-coverage-rotation` workflow before agent reasoning begins: install Go from `go.mod`, export `GOROOT`, `GOPATH`, and `GOMODCACHE`, install Node from `package.json`, and run `make setup`.
- Allow the schema-coverage rotation workflow's AWF network policy to use the repository's required ecosystems for this bootstrap path, including Go and Node alongside the default allowlist.
- Update the workflow requirements so schema-coverage rotation runs its repository-local commands against the provisioned repo toolchain instead of relying on runner-default versions.

## Capabilities

### New Capabilities
- `ci-schema-coverage-rotation-toolchain`: repository toolchain bootstrap and AWF network allowances for the schema-coverage rotation workflow

### Modified Capabilities
<!-- None. -->

## Impact

- `.github/workflows-src/schema-coverage-rotation/`
- `.github/workflows/schema-coverage-rotation.md`
- `.github/workflows/schema-coverage-rotation.lock.yml`
- OpenSpec requirements for schema-coverage rotation workflow bootstrap behavior
