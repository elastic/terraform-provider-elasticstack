# `ci-code-factory-issue-intake` — Delta

## ADDED Requirements

### Requirement: Workflow status comments on the issue include the run link
The workflow SHALL set `status-comment: true` in the top-level `on:` configuration (see GitHub Agentic Workflows [status comments](https://github.github.com/gh-aw/reference/triggers/#status-comments-status-comment)) so the activation job posts a status comment on the triggering issue when the run starts and updates it when the run completes, including a link to the workflow run as provided by the framework.

#### Scenario: Status comment enabled
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow `on:` frontmatter
- **THEN** it SHALL include `status-comment: true` (or an object form that enables status comments for issues)

#### Scenario: No custom comment step for run linkage
- **WHEN** the workflow is authored for `code-factory` issue intake
- **THEN** the repository SHALL NOT rely on a custom implementation-job step solely to post the workflow run URL to the issue; run visibility SHALL be covered by `status-comment` as above

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `code-factory` label from the triggering issue **using the same mechanism as** OpenSpec verify (label): `actions/github-script@v9` with `x-script-include` to an inline script that delegates to the shared `.github/workflows-src/lib/remove-trigger-label.js` helper (generalized to accept the factory label name and issue number). The step SHALL run only when the workflow would proceed to the implementation agent (eligible qualifying issue event, trusted actor, and no open linked `code-factory` pull request per existing duplicate suppression). The workflow SHALL grant `issues: write` to pre-activation where required for label removal.

#### Scenario: Remove step mirrors verify workflow pattern
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow `on.steps`
- **THEN** it SHALL include a remove-label step structurally equivalent to OpenSpec verify (label), including step name `Remove trigger label`, `uses: actions/github-script@v9`, and an `x-script-include` reference for the inline script
- **AND** the included script SHALL reuse the generalized `remove-trigger-label` library (not a forked copy of the GitHub API logic)

#### Scenario: Label removed only when agent gate passes
- **WHEN** pre-activation determines the implementation agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `code-factory` from that issue

#### Scenario: Label retained when agent does not run
- **WHEN** pre-activation suppresses the agent (ineligible event, untrusted actor, or duplicate linked PR)
- **THEN** the workflow SHALL NOT remove `code-factory` solely as a side effect of this intake run
