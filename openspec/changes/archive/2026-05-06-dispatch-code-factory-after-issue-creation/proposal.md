## Why

Three repository-authored analysis workflows currently create follow-up issues intended for `code-factory` automation by attaching a `code-factory` label at issue creation time. Because those issues are created through GitHub Actions `GITHUB_TOKEN`-backed safe-output processing, the downstream issue-triggered `code-factory` workflow is not activated, so the issues are opened but never handed off to the implementation worker.

## What Changes

- Remove `code-factory` from the issue labels created by the semantic function refactor, schema coverage rotation, and flaky test catcher workflows.
- Add a deterministic post-safe-outputs dispatch phase to each of those producer workflows that reads the safe-output temporary issue ID map and dispatches the `code-factory` workflow once per created issue.
- Extend the `code-factory` issue intake workflow to support a `workflow_dispatch` entrypoint for internally dispatched single-issue runs, alongside the existing manual issue-event entrypoint.
- Refactor `code-factory` intake so implementation context is derived from normalized intake data rather than directly from `github.event.issue.*`, and so dispatch-triggered runs fetch the live issue body/title by issue number before proceeding.
- Preserve the existing manual `code-factory` label-based intake path for maintainers while clearly separating it from producer-driven automation.

## Capabilities

### New Capabilities
- `ci-code-factory-dispatch-fanout`: Deterministic fan-out from issue-producing analysis workflows into one `workflow_dispatch` run of `code-factory` per created issue using the safe-output temporary ID map.

### Modified Capabilities
- `ci-code-factory-issue-intake`: Add a `workflow_dispatch` intake mode, normalize issue context resolution across entrypoints, and keep duplicate-PR suppression and single-issue PR semantics for dispatched runs.
- `ci-semantic-refactor-workflow`: Change follow-up issue behavior from label-trigger handoff to explicit post-safe-outputs dispatch and remove `code-factory` from created issue labels.
- `flaky-test-catcher`: Change follow-up issue behavior from label-trigger handoff to explicit post-safe-outputs dispatch and remove `code-factory` from created issue labels.
- `ci-schema-coverage-rotation-issue-slots`: Extend the schema coverage workflow contract so created follow-up issues are dispatched to `code-factory` explicitly rather than relying on `code-factory` labels at creation time.

## Impact

- Affected workflow sources under `.github/workflows-src/` for `code-factory-issue`, `semantic-function-refactor`, `schema-coverage-rotation`, and `flaky-test-catcher`.
- Affected generated workflow artifacts under `.github/workflows/` and compiled `.lock.yml` files.
- Shared workflow helper logic and tests under `.github/workflows-src/lib/` for intake normalization and dispatch metadata handling.
- OpenSpec capability specs for `code-factory` intake and the three producer workflows.
