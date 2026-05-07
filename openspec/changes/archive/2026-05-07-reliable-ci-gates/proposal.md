## Why

The `Build/Lint/Test` workflow has three active bugs — broken required checks when a PR moves from draft to ready, a race condition where a push event can overwrite a failing acceptance-test result with a spurious success, and changelog-only PRs that either run 48+ unnecessary acceptance-test jobs or get their `build`/`lint` required checks stuck in a skipped state. All three bugs share the same root cause: a monolithic 100-line `preflight` job with six conditional branches trying to handle every event shape in one place.

## What Changes

- **BREAKING** — Replace `.github/workflows/test.yml` (`Build/Lint/Test`) with three focused workflows: `provider.yml`, `openspec.yml`, and `workflows.yml`, each backed by its own source template under `.github/workflows-src/`.
- Remove the `preflight` job and all its branching logic.
- Remove the `test-validation` job; replace with a `gate` job in each workflow.
- Change workflow triggers: keep `push` only on `main` (post-merge CI); use `pull_request: [opened, synchronize, reopened]` (drop `ready_for_review`); add `workflow_dispatch`.
- Extend change classification to skip provider CI for CHANGELOG-only, `openspec/`, `.agents/`, and non-test `.github/` changes (currently only `openspec/` skips).
- Remove `build` and `lint` as standalone required checks; replace with three gate required checks (`provider / gate`, `openspec / gate`, `workflows / gate`).
- Remove the `auto-approve` changelog special-case (generated-changelog branch detection, auto-merge step).
- Move workflow tests and hook tests from the `build` job into the dedicated `workflows.yml`.
- Update `scripts/auto-approve`: remove the `generated-changelog` category.
- Update `.github/workflows-src/manifest.json` and the compile script to reflect the new three-workflow layout; delete the `test/` source directory.

## Capabilities

### New Capabilities

- `ci-openspec-workflow`: New `openspec.yml` workflow that runs `make check-openspec` and always publishes a `gate` result as the required check.
- `ci-workflows-workflow`: New `workflows.yml` workflow that runs `make workflow-test` and `make hook-test` for `.github/` changes and always publishes a `gate` result as the required check.

### Modified Capabilities

- `ci-build-lint-test`: Restructured into `provider.yml`; new triggers (push main-only, PR without `ready_for_review`), no `preflight`, extended classify logic, new `gate` job, `auto-approve` job updated to drop changelog path.
- `ci-pr-auto-approve`: Remove the `generated-changelog` category selector, commit-author gate, and file-allowlist gate.

## Impact

- **`.github/workflows/`**: `test.yml` deleted; `provider.yml`, `openspec.yml`, `workflows.yml` added.
- **`.github/workflows-src/`**: `test/` directory replaced by `provider/`, `openspec/`, `workflows/` directories; `manifest.json` updated.
- **`scripts/auto-approve/`**: `generated-changelog` category and its unit tests removed.
- **`.github/workflows-src/lib/classify-changes.js`**: skip-path set extended.
- **GitHub branch protection**: required checks updated from `Build/Lint/Test / Build`, `Build/Lint/Test / Lint`, `Build/Lint/Test / Test Validation` to `provider / gate`, `openspec / gate`, `workflows / gate` (manual step post-deploy).
