## Why

`Test Validation` currently reports success even when the `preflight` gate intentionally skips the entire acceptance workflow. That makes the check look like acceptance coverage passed when no validation job or matrix test actually ran, which is misleading for maintainers and branch-protection consumers.

## What Changes

- Update the CI requirements so `Test Validation` is skipped when `preflight` outputs `should_run=false`.
- Preserve the existing validation behavior for runs where `preflight` allows downstream CI: openspec-only changes may still skip `test` and pass validation, while provider changes still require a successful `test` job.
- Align the documented validation helper contract with the new workflow behavior so the preflight-skip path is no longer described as a passing validation result.
- Regenerate the compiled workflow after the authored template changes so the checked-in workflow artifact stays in sync.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-build-lint-test`: Change `Test Validation` so preflight-disabled workflow runs skip the validation job instead of reporting a successful test result.

## Impact

- **Specs**: `openspec/specs/ci-build-lint-test/spec.md`
- **Workflow sources**: `.github/workflows-src/test/workflow.yml.tmpl`, `.github/workflows-src/lib/validate-test-result.js`, `.github/workflows-src/lib/validate-test-result.test.mjs`
- **Generated workflow**: `.github/workflows/test.yml`
- **Verification**: workflow source generation/tests such as `make workflow-test`
