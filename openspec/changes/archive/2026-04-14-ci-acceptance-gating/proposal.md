## Why

The main CI workflow exposes required acceptance checks as per-version matrix jobs. That makes branch protection awkward for non-provider pull requests, because skipping the workflow with path filters leaves required checks pending while requiring every matrix leg blocks `openspec/**`-only changes that do not need provider acceptance coverage.

## What Changes

- Define the proposed CI behavior in OpenSpec change artifacts first; the workflow YAML implementation remains a follow-up task before the change can be considered implemented.
- Update the `ci-build-lint-test` workflow requirements so matrix acceptance tests run only when the current change set includes provider-impacting files; in the first iteration, changes limited to `openspec/**` SHALL be treated as non-provider changes.
- Add a dedicated workflow validation job that always reports a final test status: it passes when acceptance tests are intentionally skipped for non-provider changes, and it fails when provider changes require acceptance coverage but the matrix test job does not succeed.
- Change the auto-approve gate to depend on the validation job result instead of raw matrix test success so spec-only pull requests can still satisfy the existing automation contract.
- Document the GitHub branch-protection follow-up to require the stable validation check instead of the per-version matrix acceptance checks.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-build-lint-test`: Add diff-aware acceptance-test gating and a stable validation job for required-check integration.

## Impact

- **Specs**: `openspec/specs/ci-build-lint-test/spec.md`
- **Workflow logic**: `.github/workflows/test.yml` is unchanged in this proposal-only change set and still needs a follow-up implementation change for the change-classification step, conditional matrix-test execution, `Test Validation`, and updated auto-approve gating.
- **Repository operations**: GitHub branch protection or rulesets must be updated to require the stable validation check instead of individual matrix acceptance checks.
