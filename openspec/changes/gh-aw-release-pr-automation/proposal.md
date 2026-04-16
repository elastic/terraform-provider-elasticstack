## Why

Preparing a provider release PR is currently a manual process that depends on hand-maintained changelog entries and repeated maintainer judgment over which merged changes are actually user-facing. The repository needs to split that problem into ongoing changelog maintenance and release preparation: a changelog generator should keep `CHANGELOG.md` current from merged PR history, while release preparation should become a small deterministic workflow that only handles versioning and release PR creation.

## What Changes

- Add a GitHub Agentic Workflow changelog generator with two operating modes:
  - a scheduled and manually dispatched mode that regenerates the full `## [Unreleased]` section from merged PR history and opens or updates a singleton PR from branch `generated-changelog`
  - a `pull_request` mode for `prep-release-*` branches that regenerates the concrete `## [x.y.z] - <date>` release section inside the triggering release PR branch
- Require the changelog generator to build proof-carrying PR evidence in workflow memory, generate changelog text strictly from PR-level summaries, and return structured provenance for every emitted changelog bullet.
- Add a deterministic release-preparation workflow that creates the `prep-release-x.y.z` branch and PR with the simple version bump changes, then relies on the changelog generator to populate the release section for that PR.
- Add a thin `make` target that dispatches the release-preparation workflow through `gh workflow run`, while keeping release logic in the workflows and their helpers.
- Extend pull-request auto-approval so the generated changelog PR can auto-merge once green, with tightly scoped rules:
  - branch name is exactly `generated-changelog`
  - every commit is authored by `github-actions[bot]`
  - only `CHANGELOG.md` is modified
- Update the existing PR workflow wiring so changelog-only generated PRs can still reach the auto-approve path without accidentally running the full provider CI suite.

## Capabilities

### New Capabilities
- `ci-changelog-generation`: GH AW changelog maintenance for `Unreleased` and release-section generation using PR evidence and provenance-backed validation
- `ci-release-pr-preparation`: deterministic workflow-dispatched release PR preparation that bumps versions and creates `prep-release-x.y.z` pull requests

### Modified Capabilities
- `ci-pr-auto-approve`: add a narrowly scoped generated-changelog approval category for `generated-changelog` PRs authored only by `github-actions[bot]` and touching only `CHANGELOG.md`
- `ci-build-lint-test`: adjust pull-request workflow behavior so generated changelog PRs can reach the auto-approve path without running the full build, lint, and acceptance matrix unnecessarily
- `makefile-workflows`: add a maintainer-facing target that dispatches the release-preparation workflow with a validated bump mode input

## Impact

- New authored GH AW changelog workflow source under `.github/workflows-src/` and compiled workflow artifacts under `.github/workflows/`
- New deterministic helpers for PR evidence gathering, workflow-memory manifest persistence, provenance validation, and changelog section regeneration
- A smaller deterministic release-preparation workflow that limits release PR changes to version bump plumbing while changelog content is maintained separately
- Changes to `.github/workflows/test.yml` and `scripts/auto-approve/` so generated changelog PRs can be auto-approved and auto-merged safely
- Changes to `CHANGELOG.md` maintenance expectations, maintainer release docs, and the Makefile dispatch ergonomics for release preparation
