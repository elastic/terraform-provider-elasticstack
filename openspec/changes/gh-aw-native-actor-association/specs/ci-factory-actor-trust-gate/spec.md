# `ci-factory-actor-trust-gate` — Native gh-aw actor association gating for factory workflows

Workflow implementation: authored sources under `.github/workflows/` (markdown-frontmatter GH AW format); compiled lock files alongside each source.

## Purpose

Define requirements for replacing the runtime JS-based `check_actor_trust` pre-activation step in
all four factory workflows (research, code, change, reproducer) with the declarative native
`on.skip-author-associations` field provided by gh-aw v0.74.4+.

## MODIFIED Requirements

### Requirement: Actor trust gating uses native `skip-author-associations`
All four factory workflow sources (research, code, change, reproducer) SHALL gate actor trust
declaratively via the `on.skip-author-associations` frontmatter field rather than through a
runtime `getCollaboratorPermissionLevel` API call. The skip list SHALL include `none`,
`first_timer`, `first_time_contributor`, and `contributor` associations, allowing `OWNER`,
`MEMBER`, and `COLLABORATOR` associations to proceed.

#### Scenario: Trusted actor activates the workflow
- **GIVEN** the factory workflow receives a qualifying event
- **WHEN** the triggering actor has a GitHub `author_association` of `OWNER`, `MEMBER`, or `COLLABORATOR`
- **THEN** the workflow SHALL activate and proceed to downstream pre-activation steps

#### Scenario: Untrusted actor is skipped at the job level
- **GIVEN** the factory workflow receives a qualifying event
- **WHEN** the triggering actor has a GitHub `author_association` of `none`, `first_timer`, `first_time_contributor`, or `contributor`
- **THEN** the workflow SHALL NOT activate; the `skip-author-associations` gate SHALL prevent the job from running

#### Scenario: change-factory covers both issues and issue_comment triggers
- **GIVEN** the change-factory workflow has both an `issues` trigger and an `issue_comment`/`slash_command` trigger
- **WHEN** configuring `skip-author-associations`
- **THEN** the field SHALL include both `issues` and `issue_comment` event keys with the same association skip list

### Requirement: `check_actor_trust` step is removed from all four workflows
The `check_actor_trust` pre-activation step SHALL be removed from all four factory workflow
sources. No downstream step SHALL gate on `steps.check_actor_trust.outputs.actor_trusted`.

#### Scenario: No runtime permission-level API call
- **GIVEN** a qualifying event from a trusted actor
- **WHEN** the pre-activation job runs
- **THEN** it SHALL NOT call `github.rest.repos.getCollaboratorPermissionLevel` for actor trust purposes

### Requirement: `normalize_context` SHALL emit `actor_trusted=true` unconditionally for issue-event paths
After the `check_actor_trust` step is removed, each factory workflow's `normalize_context` step SHALL emit `actor_trusted=true` for the issue-event path without reading a step output. The native gate guarantees trust for any event that reaches this step.

#### Scenario: normalize_context actor_trusted is always true for issue-event path
- **GIVEN** the native gate has passed (workflow is running)
- **WHEN** `normalize_context` processes the issue-event branch
- **THEN** it SHALL set `actor_trusted=true` unconditionally

### Requirement: `check-actor-trust.js` and its library functions are deleted
The file `.github/scripts/workflows/lib/factory-runners/check-actor-trust.js` SHALL be deleted.
The functions `factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` SHALL be removed
from `.github/scripts/workflows/lib/factory-issue-shared.js` along with their exports.

#### Scenario: check-actor-trust.js does not exist in the repository
- **GIVEN** the change is implemented
- **WHEN** the repository is inspected
- **THEN** `.github/scripts/workflows/lib/factory-runners/check-actor-trust.js` SHALL NOT exist

#### Scenario: actor-trust functions are not exported from factory-issue-shared.js
- **GIVEN** the change is implemented
- **WHEN** the repository is inspected
- **THEN** `factory-issue-shared.js` SHALL NOT export `factoryCheckActorTrust` or `factoryActorTrustWhenSenderMissing`

### Requirement: Actor-trust unit tests are removed
Test cases that test `factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` SHALL be
removed from `factory-issue-shared.test.mjs`, `code-factory-issue.test.mjs`, and
`change-factory-issue.test.mjs`. All other test coverage in those files SHALL remain intact.

#### Scenario: Actor-trust test cases are absent
- **GIVEN** the change is implemented
- **WHEN** the test files are inspected
- **THEN** no test case SHALL reference `factoryCheckActorTrust` or `factoryActorTrustWhenSenderMissing`

### Requirement: Lock files are regenerated
All four factory workflow lock files SHALL be regenerated via `gh aw compile` after source
changes and committed together with the source edits.

#### Scenario: Lock files match compiled source
- **GIVEN** the workflow sources have been updated
- **WHEN** `gh aw compile` is run for all four workflows
- **THEN** the resulting lock files SHALL be committed and SHALL match the compiler output
