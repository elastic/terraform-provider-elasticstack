## MODIFIED Requirements

### Requirement: Workflow artifacts and operating modes
The changelog generator SHALL be authored as a deterministic workflow template under `.github/workflows-src/` and SHALL generate a checked-in workflow YAML artifact under `.github/workflows/` via `scripts/compile-workflow-sources/main.go`. The workflow SHALL NOT require a GH AW markdown source or a compiled `.lock.yml`. The workflow SHALL support:

- a scheduled mode
- a manual `workflow_dispatch` mode
- explicit workflow-dispatch inputs that allow maintainers to select `unreleased` or `release` execution mode

#### Scenario: Source and compiled workflow stay paired
- **WHEN** maintainers change changelog generator behavior
- **THEN** the `.github/workflows-src/` template and generated `.github/workflows/changelog-generation.yml` SHALL match the current repository compiler output

#### Scenario: Manual dispatch can enter release mode explicitly
- **WHEN** a maintainer dispatches the changelog generator with inputs selecting `release` mode and a target version
- **THEN** the workflow SHALL enter release-section generation mode for the checked out target branch without relying on pull-request event metadata

### Requirement: Full-section regeneration uses authoritative ranges
The changelog generator SHALL regenerate the full target section on each run from an authoritative range rather than appending only “since last run” changes.

- In scheduled/manual unreleased mode, it SHALL regenerate the full `## [Unreleased]` section from the previous semver release tag to `main`.
- In explicit release mode, it SHALL regenerate the full concrete `## [x.y.z] - <date>` section for the checked out release branch from the previous semver release tag to that branch head.
- In both modes, release-note content SHALL be assembled from merged pull requests in that authoritative range by parsing their PR-body changelog contract and labels.

#### Scenario: Scheduled run rebuilds full Unreleased
- **WHEN** the changelog generator runs in scheduled mode
- **THEN** it SHALL regenerate the entire `## [Unreleased]` section from the authoritative release range instead of incrementally appending entries

#### Scenario: Manual release run rebuilds full release section
- **WHEN** the changelog generator runs in explicit release mode for a checked out `prep-release-*` branch
- **THEN** it SHALL regenerate the full `## [x.y.z] - <date>` section for that branch from the authoritative release range instead of incrementally appending entries

### Requirement: Scheduled/manual mode updates the singleton generated changelog PR
In scheduled or manually dispatched unreleased mode, after deterministic validation succeeds, the workflow SHALL rewrite only the `## [Unreleased]` section of `CHANGELOG.md` and SHALL push the result to the singleton branch named `generated-changelog`. The workflow SHALL use repository-authored GitHub Actions logic to look up an existing pull request from `generated-changelog` to `main`, create that pull request when none exists, and update the existing PR body when one already exists.

#### Scenario: Generated changelog branch PR is reused
- **GIVEN** the singleton branch `generated-changelog` already has an open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator runs again
- **THEN** it SHALL update that same branch and refresh the existing pull request instead of opening a duplicate

#### Scenario: Missing generated changelog PR is created
- **GIVEN** the singleton branch `generated-changelog` has no open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator produces an updated `CHANGELOG.md`
- **THEN** the workflow SHALL create the pull request after pushing the branch update

### Requirement: Explicit release mode updates only the targeted release section
In explicit release mode, after deterministic validation succeeds, repository-authored helper logic SHALL update only the concrete `## [x.y.z] - <date>` section for the checked out release branch and SHALL push that change only to the targeted release branch. Manual release-mode execution MAY refresh release PR metadata when the corresponding pull request is known, but release-mode changelog generation SHALL NOT depend on `pull_request_target` event metadata or automatic pull-request triggers.

#### Scenario: Release mode updates only the targeted branch
- **WHEN** the changelog generator runs in explicit release mode for a release-preparation branch
- **THEN** it SHALL push changelog updates only to that targeted release branch

#### Scenario: Release mode does not rewrite Unreleased on the release branch
- **WHEN** the changelog generator runs in explicit release mode
- **THEN** it SHALL regenerate the concrete release section needed for that branch without treating the branch as the singleton `Unreleased` maintenance branch

### Requirement: Shared changelog engine reuses existing JavaScript helpers
The shared changelog engine SHALL be implemented in JavaScript and SHALL run on the Node.js runtime already used by `actions/github-script` and the workflow-source helpers. It SHALL be built by extracting and composing the existing JavaScript modules under `.github/workflows-src/changelog-generation/scripts/` and `.github/workflows-src/lib/` rather than by reimplementing changelog parsing, rendering, or PR resolution in another language. The engine SHALL NOT be authored in Go and SHALL NOT introduce a parallel Go implementation of changelog generation.

#### Scenario: Engine is implemented in JavaScript
- **WHEN** the shared changelog engine is invoked from any workflow
- **THEN** it SHALL execute as JavaScript on Node.js and SHALL reuse the existing repository-authored JS helpers for release-context resolution, merged-PR resolution, PR-body changelog parsing, and changelog rendering

#### Scenario: No Go implementation of the changelog engine
- **WHEN** maintainers add or modify shared changelog engine logic
- **THEN** that logic SHALL live in the existing JavaScript helper tree and SHALL NOT be added under `scripts/` as a Go program

### Requirement: Merged PR changelog metadata is gathered for deterministic assembly
Before changelog rendering starts, a shared repository-authored changelog engine SHALL gather the merged pull requests in the authoritative release range and capture the metadata needed for rendering from those PRs by using the GitHub API with the workflow's repository token. For each merged PR, the engine SHALL capture at least the pull request number, URL, merge commit SHA, labels, and pull request body.

#### Scenario: Merged PR body is available to the renderer
- **WHEN** the shared changelog engine resolves merged pull requests in the authoritative range
- **THEN** each merged PR considered for changelog assembly SHALL include its body and labels in the renderer input

#### Scenario: Engine uses workflow token for GitHub API lookups
- **WHEN** changelog assembly needs to resolve merged PRs in the authoritative range
- **THEN** the shared changelog engine SHALL authenticate GitHub API requests with the workflow-provided repository token rather than requiring a separate credential source
