# `ci-changelog-generation` — GH AW changelog generator

Workflow implementation: authored GH AW source under `.github/workflows/`, backed by repository-authored template/source under `.github/workflows-src/`, with compiled `.lock.yml` committed from `gh aw compile`.

## Purpose
Define a GitHub Agentic Workflow that maintains `CHANGELOG.md` from merged pull-request history. The workflow has one agentic changelog engine and two operating modes: maintaining the full `## [Unreleased]` section on branch `generated-changelog`, and regenerating a concrete release section for release-preparation pull requests.
## Requirements
### Requirement: Workflow artifacts and operating modes
The changelog generator SHALL be authored as a deterministic workflow template under `.github/workflows-src/` and SHALL generate a checked-in workflow YAML artifact under `.github/workflows/` via `scripts/compile-workflow-sources/main.go`. The workflow SHALL NOT require a GH AW markdown source or a compiled `.lock.yml`. The workflow SHALL support:

- a scheduled mode
- a manual `workflow_dispatch` mode
- a `pull_request_target` mode for same-repository release-preparation branches matching the repository's `prep-release-*` naming contract

#### Scenario: Source and compiled workflow stay paired
- **WHEN** maintainers change changelog generator behavior
- **THEN** the `.github/workflows-src/` template and generated `.github/workflows/changelog-generation.yml` SHALL match the current repository compiler output

#### Scenario: Release-preparation pull request activates release mode
- **GIVEN** a same-repository pull request whose head branch matches the configured `prep-release-*` pattern
- **WHEN** the changelog generator runs for that pull request through `pull_request_target`
- **THEN** it SHALL enter release-section generation mode for the triggering pull request branch

### Requirement: Full-section regeneration uses authoritative ranges
The changelog generator SHALL regenerate the full target section on each run from an authoritative range rather than appending only “since last run” changes.

- In scheduled/manual mode, it SHALL regenerate the full `## [Unreleased]` section from the previous semver release tag to `main`.
- In release-pull-request mode, it SHALL regenerate the full concrete `## [x.y.z] - <date>` section for the triggering release branch from the previous semver release tag to that branch head.
- In both modes, release-note content SHALL be assembled from merged pull requests in that authoritative range by parsing their PR-body changelog contract and labels.

#### Scenario: Scheduled run rebuilds full Unreleased
- **WHEN** the changelog generator runs in scheduled or manual mode
- **THEN** it SHALL regenerate the entire `## [Unreleased]` section from the authoritative release range instead of incrementally appending entries

#### Scenario: Release PR run rebuilds full release section
- **WHEN** the changelog generator runs for a `prep-release-*` pull request
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
In scheduled/manual mode, after deterministic validation succeeds, the workflow SHALL rewrite only the `## [Unreleased]` section of `CHANGELOG.md` and SHALL push the result to the singleton branch named `generated-changelog`. The workflow SHALL use repository-authored GitHub Actions logic to look up an existing pull request from `generated-changelog` to `main`, create that pull request when none exists, and update the existing PR body when one already exists.

#### Scenario: Generated changelog branch PR is reused
- **GIVEN** the singleton branch `generated-changelog` already has an open pull request to `main`
- **WHEN** the scheduled/manual changelog generator runs again
- **THEN** it SHALL update that same branch and refresh the existing pull request instead of opening a duplicate

#### Scenario: Missing generated changelog PR is created
- **GIVEN** the singleton branch `generated-changelog` has no open pull request to `main`
- **WHEN** the scheduled/manual changelog generator produces an updated `CHANGELOG.md`
- **THEN** the workflow SHALL create the pull request after pushing the branch update

### Requirement: Release-pull-request mode updates only the triggering release section
In release-pull-request mode, after deterministic validation succeeds, repository-authored helper logic SHALL update only the concrete `## [x.y.z] - <date>` section for the triggering release branch and SHALL push that change only to the triggering release pull request branch. When the workflow refreshes release PR metadata, it SHALL use the pull request number from event metadata rather than performing branch-based PR discovery.

#### Scenario: Release mode updates only the triggering branch
- **WHEN** the changelog generator runs for a `prep-release-*` pull request
- **THEN** it SHALL push changelog updates only to the triggering release pull request branch

#### Scenario: Release mode does not rewrite Unreleased on the release branch
- **WHEN** the changelog generator runs for a release-preparation pull request
- **THEN** it SHALL regenerate the concrete release section needed for that branch without treating the branch as the singleton `Unreleased` maintenance branch

### Requirement: Merged PR changelog metadata is gathered for deterministic assembly
Before changelog rendering starts, deterministic repository-authored steps SHALL gather the merged pull requests in the authoritative release range and capture the metadata needed for rendering from those PRs. For each merged PR, the workflow SHALL capture at least the pull request number, URL, merge commit SHA, labels, and pull request body.

#### Scenario: Merged PR body is available to the renderer
- **WHEN** deterministic pre-render steps resolve merged pull requests in the authoritative range
- **THEN** each merged PR considered for changelog assembly SHALL include its body and labels in the renderer input

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

