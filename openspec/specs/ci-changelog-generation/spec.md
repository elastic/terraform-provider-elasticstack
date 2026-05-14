# `ci-changelog-generation` — Changelog generator

Workflow implementation: repository-authored template under `.github/workflows-src/changelog-generation/workflow.yml.tmpl`, generating checked-in `.github/workflows/changelog-generation.yml` via `scripts/compile-workflow-sources/main.go`.

## Purpose

Define a deterministic workflow that maintains `CHANGELOG.md` from merged pull-request history using a shared JavaScript changelog engine. The workflow supports maintaining the full `## [Unreleased]` section on branch `generated-changelog`, and regenerating a concrete release section for release-preparation branches in explicit release mode.
## Requirements
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

### Requirement: Deterministic validation gates changelog mutation

Before updating `CHANGELOG.md`, deterministic repository-authored validation SHALL verify that parsed changelog content from merged PRs is structurally valid and matches the repository changelog format for the target section. The validator SHALL reject output when referenced PR metadata cannot be resolved from the authoritative merged range, when the rendered changelog shape is invalid for the target section, or when extracted breaking-change content cannot be preserved as markdown inside the top-level breaking-changes section.

#### Scenario: Unsupported pull request metadata is rejected

- **WHEN** deterministic assembly encounters rendered changelog content that references merged PR metadata outside the authoritative range
- **THEN** changelog mutation SHALL fail before `CHANGELOG.md` is updated

#### Scenario: Invalid breaking-changes block is rejected

- **WHEN** deterministic assembly encounters an extracted `### Breaking changes` subsection that cannot be preserved as markdown content in the target changelog section
- **THEN** changelog mutation SHALL fail before `CHANGELOG.md` is updated

### Requirement: Scheduled/manual mode updates the singleton generated changelog PR
In scheduled or manually dispatched unreleased mode, after deterministic validation succeeds, the workflow SHALL rewrite only the `## [Unreleased]` section of `CHANGELOG.md` and SHALL push the result to the singleton branch named `generated-changelog`. The workflow SHALL use repository-authored GitHub Actions logic to look up an existing pull request from `generated-changelog` to `main`, create that pull request when none exists, and update the existing PR body when one already exists. After pushing the changelog commit to `generated-changelog`, the workflow SHALL push an empty commit re-authenticated with the CI trigger token to trigger downstream CI.

#### Scenario: Generated changelog branch PR is reused
- **GIVEN** the singleton branch `generated-changelog` already has an open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator runs again
- **THEN** it SHALL update that same branch and refresh the existing pull request instead of opening a duplicate
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

#### Scenario: Missing generated changelog PR is created
- **GIVEN** the singleton branch `generated-changelog` has no open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator produces an updated `CHANGELOG.md`
- **THEN** the workflow SHALL create the pull request after pushing the branch update
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

### Requirement: Explicit release mode updates the targeted release section and removes Unreleased
In explicit release mode, after deterministic validation succeeds, repository-authored helper logic SHALL update only the concrete `## [x.y.z] - <date>` section for the checked out release branch and SHALL push that change only to the targeted release branch. After pushing the changelog commit to the release branch, the workflow SHALL push an empty commit re-authenticated with the CI trigger token to trigger downstream CI. Manual release-mode execution MAY refresh release PR metadata when the corresponding pull request is known, but release-mode changelog generation SHALL NOT depend on `pull_request_target` event metadata or automatic pull-request triggers.

In release mode, when the rewriter mutates `CHANGELOG.md` to emit the new `## [x.y.z] - <date>` section, it SHALL also remove any existing `## [Unreleased]` section (header and body) from the file. This SHALL hold both on the first run against a release-preparation branch (when no `## [x.y.z]` heading exists yet) and on any re-run (when the `## [x.y.z]` heading is already present alongside a stale `## [Unreleased]` section). Release-mode mutation SHALL NOT preserve, duplicate, or insert content alongside the Unreleased section; the resulting `CHANGELOG.md` SHALL contain exactly one block representing the work shipped in the release, headed by `## [x.y.z] - <date>`.

#### Scenario: Release mode updates only the targeted branch
- **WHEN** the changelog generator runs in explicit release mode for a release-preparation branch
- **THEN** it SHALL push changelog updates only to that targeted release branch
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

#### Scenario: Release mode does not regenerate Unreleased on the release branch
- **WHEN** the changelog generator runs in explicit release mode
- **THEN** it SHALL regenerate the concrete release section needed for that branch and SHALL NOT preserve or regenerate any `## [Unreleased]` section, without treating the branch as the singleton `Unreleased` maintenance branch

#### Scenario: Release mode replaces the Unreleased section with the new versioned section
- **GIVEN** `CHANGELOG.md` on a `prep-release-x.y.z` branch contains a `## [Unreleased]` section with body content and no `## [x.y.z]` heading
- **WHEN** the changelog generator runs in explicit release mode for that branch
- **THEN** the resulting `CHANGELOG.md` SHALL contain a single `## [x.y.z] - <date>` section in place of the previous `## [Unreleased]` section, with no `## [Unreleased]` heading remaining in the file

#### Scenario: Release mode re-run collapses lingering Unreleased section
- **GIVEN** `CHANGELOG.md` on a `prep-release-x.y.z` branch already contains a `## [x.y.z] - <date>` section and also contains a `## [Unreleased]` section
- **WHEN** the changelog generator runs in explicit release mode again for that branch
- **THEN** the resulting `CHANGELOG.md` SHALL contain a single regenerated `## [x.y.z] - <date>` section and SHALL NOT contain any `## [Unreleased]` heading

#### Scenario: Release mode with no prior Unreleased section prepends the new section
- **GIVEN** `CHANGELOG.md` contains no `## [Unreleased]` heading and no `## [x.y.z]` heading
- **WHEN** the changelog generator runs in explicit release mode for version `x.y.z`
- **THEN** the resulting `CHANGELOG.md` SHALL begin with the new `## [x.y.z] - <date>` section followed by the prior changelog content

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

### Requirement: Parsed PR-body changelog sections drive release-note assembly

The changelog generator SHALL parse the `## Changelog` section from each merged PR body and SHALL treat that contract as the authoritative source for release-note content. PRs labeled `no-changelog` and PR changelog entries with `Customer impact: none` SHALL be excluded from rendered changelog bullets. The optional `### Breaking changes` subsection, when present, SHALL be preserved as markdown content and rendered under the target version's top-level `### Breaking changes` section. A merged PR in the authoritative range that lacks both the `no-changelog` label and a parseable `## Changelog` section SHALL cause deterministic assembly to fail rather than being silently omitted or summarized from fallback text.

#### Scenario: User-facing summary becomes a changelog bullet

- **WHEN** a merged PR body contains a valid `## Changelog` section with `Customer impact` other than `none`
- **THEN** the rendered target section SHALL include a changelog bullet derived from that PR's `Summary` and citation

#### Scenario: `Customer impact: none` is excluded

- **WHEN** a merged PR body contains `Customer impact: none`
- **THEN** the rendered target section SHALL omit a normal changelog bullet for that PR

#### Scenario: Missing changelog contract fails assembly

- **WHEN** a merged PR in the authoritative range lacks both a parseable `## Changelog` section and the `no-changelog` label
- **THEN** deterministic assembly SHALL fail before mutating `CHANGELOG.md`, with a repository-authored validation error that identifies the offending pull request

#### Scenario: Breaking changes block is preserved

- **WHEN** a merged PR body contains a non-empty `### Breaking changes` subsection
- **THEN** the rendered target version SHALL include that markdown block under the top-level `### Breaking changes` section

### Requirement: Output normalization stays minimal

The changelog generator SHALL apply only simple normalization needed to keep `CHANGELOG.md` consistent. It MAY normalize bullet prefixes, pull-request citation shape, whitespace, and placement of preserved breaking-change blocks, but it SHALL NOT semantically rewrite author-provided summaries or breaking-change prose during scheduled/release assembly.

#### Scenario: Simple formatting is normalized without semantic rewrite

- **WHEN** deterministic assembly renders PR-body changelog content into `CHANGELOG.md`
- **THEN** it SHALL preserve the author-provided meaning while applying only the minimal formatting normalization required by repository changelog conventions

