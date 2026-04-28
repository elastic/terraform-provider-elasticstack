## 1. Extract the shared changelog engine

- [x] 1.1 Inventory the existing JavaScript changelog-generation logic under `.github/workflows-src/changelog-generation/scripts/*.inline.js` and `.github/workflows-src/lib/*.js` (release-context resolution, merged-PR discovery, PR-body changelog parsing/validation, rendering, and `CHANGELOG.md` rewriting) that must move behind the shared engine interface
- [x] 1.2 Implement the shared engine in JavaScript/Node by composing the existing helpers (do NOT reimplement in Go), so it accepts explicit mode inputs (`unreleased` or `release`, with explicit target version for release mode), resolves merged PRs through the GitHub API using the workflow token, and emits structured outputs needed by workflows
- [x] 1.3 Preserve and extend the existing `.github/workflows-src/lib/*.test.mjs` coverage so the shared engine has automated tests covering deterministic assembly, release/unreleased mode selection, GitHub-backed merged-PR resolution, and changelog-section rewriting behavior

## 2. Update changelog-generation workflow behavior

- [x] 2.1 Refactor `.github/workflows-src/changelog-generation/workflow.yml.tmpl` and generated workflow output to remove `pull_request_target`, add explicit `workflow_dispatch` inputs for release mode, and invoke the shared engine in both scheduled unreleased mode and manual release mode
- [x] 2.2 Keep unreleased-mode branch/PR management in workflow orchestration, ensuring scheduled or manually dispatched unreleased runs still maintain the singleton `generated-changelog` PR
- [x] 2.3 Preserve or update workflow-source helper tests/fixtures so compiled workflow behavior and manual release-mode dispatch semantics are covered

## 3. Update release preparation workflow

- [x] 3.1 Refactor `.github/workflows/prep-release.yml` to invoke the shared changelog engine in release mode after applying the version bump and before PR creation/reuse
- [x] 3.2 Update release preparation to produce a single deterministic commit containing both the version bump and final release changelog update, while retaining deterministic branch naming, PR reuse, and `no-changelog` labeling
- [x] 3.3 Verify the updated workflow behavior and sync the canonical specs affected by release preparation and changelog generation
