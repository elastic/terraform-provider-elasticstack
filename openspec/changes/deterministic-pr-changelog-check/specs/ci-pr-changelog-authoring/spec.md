## REMOVED Requirements

### Requirement: Workflow artifacts and compilation
The PR changelog authoring workflow SHALL be authored from a repository template under `.github/workflows-src/` that generates a GitHub Agentic Workflow markdown file under `.github/workflows/` via `scripts/compile-workflow-sources/main.go`. The repository SHALL commit the generated `.md` workflow and the compiled `.lock.yml` produced by `gh aw compile`. Contributors SHALL NOT hand-edit the generated `.md` or `.lock.yml` artifacts.

#### Scenario: Source and compiled artifacts stay paired
- **WHEN** maintainers change the PR changelog authoring workflow behavior
- **THEN** the `.github/workflows-src/` template, generated `.md` workflow, and compiled `.lock.yml` SHALL match the committed compiler output

**Reason**: The workflow is replaced by a plain GitHub Actions `.yml` file. There is no longer a source template, compiled `.md`, or `.lock.yml` to maintain.
**Migration**: Delete `.github/workflows/pr-changelog-authoring.md`, `.github/workflows/pr-changelog-authoring.lock.yml`, `.github/workflows-src/pr-changelog-authoring/`, and remove the `pr-changelog-authoring` entry from `.github/workflows-src/manifest.json`. Create `.github/workflows/pr-changelog-check.yml` as a plain workflow.

### Requirement: Trigger on `Build/Lint/Test` workflow completion
The workflow SHALL run from a `workflow_run` trigger for the repository workflow named `Build/Lint/Test` and SHALL continue only when the source workflow run completed for a pull-request event.

#### Scenario: Pull-request CI completion is eligible for gating
- **WHEN** the `Build/Lint/Test` workflow completes for a `pull_request` event
- **THEN** the PR changelog authoring workflow SHALL continue to deterministic pull-request resolution and gating

#### Scenario: Non-pull-request workflow run is skipped
- **WHEN** the `Build/Lint/Test` workflow completes for a non-`pull_request` event such as `push` or `workflow_dispatch`
- **THEN** the PR changelog authoring workflow SHALL NOT invoke changelog validation or authoring for that run

**Reason**: The `workflow_run` trigger introduced latency (feedback only after CI completes), could not comment on fork PRs, and required a complex PR resolution step. Direct `pull_request_target` triggering provides immediate feedback and full write access for all PRs.
**Migration**: The new workflow triggers on `pull_request_target` with types `[opened, synchronize, labeled]`. No changes to `test.yml` are required.

### Requirement: Deterministic pull-request resolution and opt-out gate
Before agent reasoning starts, deterministic repository-authored steps SHALL resolve the pull request associated with the triggering `workflow_run`. The workflow SHALL skip agent authoring when the resolved pull request carries the `no-changelog` label.

#### Scenario: `no-changelog` label suppresses authoring
- **WHEN** deterministic resolution finds the triggering pull request and that pull request carries the `no-changelog` label
- **THEN** the workflow SHALL treat the pull request as explicitly exempt from changelog authoring

#### Scenario: Missing pull request fails gating
- **WHEN** deterministic resolution cannot identify exactly one pull request for the triggering workflow run
- **THEN** the workflow SHALL fail gating without invoking the agent, with the deterministic resolution step exiting non-zero and emitting an error message prefixed with `PR_CHANGELOG_GATING:`

**Reason**: Under `pull_request_target`, `context.payload.pull_request` is populated directly in the event payload. API-based PR resolution (head_sha + branch search) is no longer needed.
**Migration**: The `no-changelog` opt-out remains. The workflow reads `context.payload.pull_request.labels` directly without an API call.

### Requirement: Missing changelog sections are drafted from PR metadata
When the resolved pull request lacks a `## Changelog` section and is not exempt via `no-changelog`, the agent SHALL draft the missing section from the pull request title and description and SHALL update the pull request body with that drafted section.

#### Scenario: Missing changelog section is added
- **WHEN** deterministic gating concludes the pull request requires a changelog section and none is present
- **THEN** the workflow SHALL invoke the agent to draft the `## Changelog` section and update the pull request body with the result

**Reason**: LLM-based drafting introduces non-determinism and cost. Authors are expected to supply their own `## Changelog` section; the workflow fails with actionable feedback when it is absent.
**Migration**: Authors must add a `## Changelog` section to their PR body manually. The failure comment explains the required format.

## MODIFIED Requirements

### Requirement: Existing changelog section is validated deterministically
The workflow SHALL parse and validate the `## Changelog` section from the pull request body using `parseChangelogSectionFull` and `validateChangelogSectionFull`. The validator SHALL require `Customer impact` to be exactly one of `none`, `fix`, `enhancement`, or `breaking` (case-sensitive). The validator SHALL require a `Summary` line when `Customer impact` is not `none`. The validator SHALL reject a `### Breaking changes` subsection that is present but empty, and SHALL require that subsection when `Customer impact` is `breaking`.

When validation fails, the workflow SHALL post or update a PR comment identifying the failure reason. When validation passes, the workflow SHALL update any existing failure comment to indicate the check passed.

#### Scenario: Valid changelog section passes the check
- **WHEN** the pull request body contains a `## Changelog` section that satisfies all validation rules
- **THEN** the workflow SHALL succeed, and if a prior failure comment exists it SHALL be updated to a "check passed" message

#### Scenario: Malformed changelog section fails with comment
- **WHEN** the pull request body contains a `## Changelog` section that does not satisfy the validation rules
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment listing each validation error

#### Scenario: Missing changelog section fails with comment
- **WHEN** the pull request body contains no `## Changelog` section and the PR does not carry the `no-changelog` label
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment stating that no `## Changelog` section was found

### Requirement: Breaking changes subsection may be free-form markdown
Within the `## Changelog` contract, the optional `### Breaking changes` subsection SHALL allow free-form markdown content, including prose, bullet lists, and fenced code blocks. Validation SHALL treat that subsection as a delimited markdown block rather than a structured schema.

#### Scenario: Breaking changes block contains fenced code
- **WHEN** the pull request body includes `### Breaking changes` with fenced code blocks or migration prose
- **THEN** the workflow SHALL accept that subsection as valid when the block is non-empty

### Requirement: Minimal permissions for validation and PR comments
The workflow SHALL request only the permissions needed to read pull request metadata and post or update PR comments. At minimum the workflow SHALL declare `pull-requests: write` and `issues: write`. The `issues: write` scope is required because PR comments are created and updated via the `issues` REST API endpoints (`github.rest.issues.listComments`, `github.rest.issues.createComment`, `github.rest.issues.updateComment`).

#### Scenario: Workflow can comment on fork PRs
- **WHEN** the triggering pull request originates from a fork repository
- **THEN** the workflow SHALL have sufficient permissions to post or update a comment on that pull request

## ADDED Requirements

### Requirement: Trigger on pull request open, update, or label change
The workflow SHALL trigger on `pull_request_target` events with types `opened`, `synchronize`, `edited`, `labeled`, and `unlabeled`. It SHALL evaluate the changelog contract immediately on each trigger without waiting for any other workflow to complete.

#### Scenario: Check runs on PR open
- **WHEN** a pull request is opened against the base repository
- **THEN** the workflow SHALL evaluate the changelog section within the same CI round as other immediate checks

#### Scenario: Check re-runs on new push
- **WHEN** new commits are pushed to an open pull request
- **THEN** the workflow SHALL re-evaluate the changelog section and update any existing comment accordingly

#### Scenario: Check re-runs when PR body is edited
- **WHEN** the pull request body is edited (e.g., the author adds or corrects the `## Changelog` section)
- **THEN** the workflow SHALL re-evaluate the changelog section and update any existing comment accordingly

#### Scenario: Check re-runs when label is applied
- **WHEN** a label is applied to an open pull request
- **THEN** the workflow SHALL re-evaluate the changelog section, allowing a freshly applied `no-changelog` label to immediately pass the check

#### Scenario: Check re-runs when label is removed
- **WHEN** a label is removed from an open pull request (e.g., `no-changelog` is removed)
- **THEN** the workflow SHALL immediately re-evaluate the changelog section, requiring the PR body to contain a valid `## Changelog` section if no `no-changelog` label remains

### Requirement: `no-changelog` label suppresses the check
The workflow SHALL pass immediately when the pull request carries the `no-changelog` label, without parsing or validating the PR body.

#### Scenario: `no-changelog` label causes immediate pass
- **WHEN** the pull request labels include `no-changelog`
- **THEN** the workflow SHALL succeed without inspecting the PR body

### Requirement: Comment upsert uses a stable hidden marker
The workflow SHALL identify its own PR comments by the hidden HTML marker `<!-- pr-changelog-check -->` embedded in the comment body. It SHALL update an existing marked comment rather than creating a new one, preventing comment accumulation on repeated pushes.

#### Scenario: Repeated failures update rather than accumulate
- **WHEN** the changelog check fails on multiple successive pushes to the same pull request
- **THEN** only one failure comment from the workflow SHALL be present; each failure SHALL update the existing comment rather than posting a new one

#### Scenario: Pass after failure updates the failure comment
- **WHEN** a pull request that previously had a workflow failure comment is updated to include a valid `## Changelog` section
- **THEN** the workflow SHALL update the existing failure comment to indicate the check passed

### Requirement: Workflow does not check out pull request code
The workflow SHALL NOT check out or execute any code from the pull request branch. It SHALL operate exclusively on pull request metadata available in the event payload.

#### Scenario: Fork PR is evaluated without code execution
- **WHEN** the triggering pull request originates from a fork repository
- **THEN** the workflow SHALL read only `context.payload.pull_request` metadata and post a comment via the REST API, without checking out the fork's code
