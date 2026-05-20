## Why

The dead-code removal rotation workflow removes symbols and occasionally their companion tests, then opens a PR immediately after verifying the build and unit tests. It does not run `make fmt` before committing, so the PR often fails the CI lint job due to simple `gofmt`-style formatting differences. Maintainers then have to manually reformat or close the PR, which defeats the purpose of the automation.

## What Changes

- Insert a `make fmt` step in the dead-code removal agent task, immediately after verification (build + unit tests) passes and before the `create-pull-request` safe output is called.
- If `make fmt` reports a non-zero exit code, record the attempt with reason `fmt_failed` and call `noop` rather than opening a PR with known formatting issues.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `deadcode-removal-rotation`: add a mandatory `make fmt` step between verification and PR creation to ensure all committed changes are correctly formatted.

## Impact

- `.github/workflows/ci-deadcode-removal-rotation.md` — update the agent task section to include `make fmt` after the verification step and before opening the PR.
- No changes to pre-activation logic, scripts, or Go source.
