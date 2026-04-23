## Why

The final release changelog update currently depends on `ci-changelog-generation` being triggered indirectly from release-preparation pull-request activity. That coupling makes release preparation less reliable than it should be, because the release PR can exist before the final changelog section has been regenerated. Release preparation should deterministically produce a ready-to-review PR, including the final concrete changelog section, and still preserve a manual recovery path.

## What Changes

- Refactor changelog-generation logic into a shared repository-authored script/engine that resolves merged PRs via the GitHub API using the workflow token, parses PR-body changelog contracts, and rewrites `CHANGELOG.md` deterministically.
- Update `prep-release.yml` to invoke the shared changelog engine directly in release mode during release preparation, so release preparation fails immediately when changelog assembly fails.
- Simplify `changelog-generation.yml` so its automatic triggers only cover scheduled unreleased maintenance, while `workflow_dispatch` also supports an explicit release mode via inputs.
- Remove the `pull_request_target` release trigger path from the changelog-generation workflow and replace event-inferred release mode with explicit workflow inputs.
- Keep workflow responsibilities limited to checkout, commit/push, PR create/reuse/labeling, and PR metadata updates around the shared engine.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `ci-changelog-generation`: Change workflow triggering and operating-mode selection so scheduled automation maintains the unreleased changelog, while release-mode execution is invoked explicitly and uses a shared repository-authored engine.
- `ci-release-pr-preparation`: Change release preparation so the workflow deterministically generates the final release changelog section during preparation and creates a single ready-to-review release-preparation commit/PR.

## Impact

- `.github/workflows/prep-release.yml`
- `.github/workflows/changelog-generation.yml`
- `.github/workflows-src/changelog-generation/workflow.yml.tmpl`
- changelog-generation helper code under `.github/workflows-src/lib/` and/or `scripts/changelog-generation/`
- OpenSpec specs for `ci-changelog-generation` and `ci-release-pr-preparation`
