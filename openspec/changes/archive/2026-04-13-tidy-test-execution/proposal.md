## Why

The repository's unit-style test execution is split across `make test`, `make check-lint`, and CI-specific workflow steps, which makes it easy for contributors and automation to exercise different test coverage. Consolidating these checks makes `make test` the single entry point for unit-level verification and keeps CI aligned with that contract.

## What Changes

- Move workflow generation tests from the lint-oriented path into the `test` aggregate so `make test` covers all unit-style test suites.
- Add hook test execution for `.agents/hooks/*.test.mjs` to the repository test workflow.
- Update CI build job requirements so workflow tests and hook tests run alongside the existing build-oriented checks.
- Clarify the Makefile and CI specs so they describe the new test placement and CI execution behavior.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `makefile-workflows`: Change the `test` target requirements to include workflow tests and hook JavaScript tests, and remove workflow-test from the lint/check-lint contract.
- `ci-build-lint-test`: Change the `build` job requirements to run workflow tests and hook tests in CI.

## Impact

Affected areas include the root `Makefile`, JavaScript hook tests under `.agents/hooks/`, and `.github/workflows/test.yml`. The change does not add external services or new runtime dependencies, but it does change which commands contributors and CI use for unit-level verification.
