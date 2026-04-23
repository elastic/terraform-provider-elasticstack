## 1. Extract the shared changelog engine

- [ ] 1.1 Identify the current changelog-generation logic that must move behind a shared repository-authored script interface, including release-context resolution, merged-PR discovery, changelog parsing/validation, rendering, and `CHANGELOG.md` rewriting
- [ ] 1.2 Implement the shared engine so it accepts explicit mode inputs (`unreleased` or `release`, with explicit target version for release mode), resolves merged PRs through the GitHub API using the workflow token, and emits structured outputs needed by workflows
- [ ] 1.3 Preserve or add automated tests covering deterministic assembly, release/unreleased mode selection, GitHub-backed merged-PR resolution, and changelog-section rewriting behavior

## 2. Update changelog-generation workflow behavior

- [ ] 2.1 Refactor `.github/workflows-src/changelog-generation/workflow.yml.tmpl` and generated workflow output to remove `pull_request_target`, add explicit `workflow_dispatch` inputs for release mode, and invoke the shared engine in both scheduled unreleased mode and manual release mode
- [ ] 2.2 Keep unreleased-mode branch/PR management in workflow orchestration, ensuring scheduled or manually dispatched unreleased runs still maintain the singleton `generated-changelog` PR
- [ ] 2.3 Preserve or update workflow-source helper tests/fixtures so compiled workflow behavior and manual release-mode dispatch semantics are covered

## 3. Update release preparation workflow

- [ ] 3.1 Refactor `.github/workflows/prep-release.yml` to invoke the shared changelog engine in release mode after applying the version bump and before PR creation/reuse
- [ ] 3.2 Update release preparation to produce a single deterministic commit containing both the version bump and final release changelog update, while retaining deterministic branch naming, PR reuse, and `no-changelog` labeling
- [ ] 3.3 Verify the updated workflow behavior and sync the canonical specs affected by release preparation and changelog generation
