# `ci-factory-actor-trust-gate` — gh-aw role-based actor trust gating for factory workflows

Workflow implementation: authored sources under `.github/workflows/` (markdown-frontmatter GH AW format); compiled lock files alongside each source.

## Purpose

Define requirements for replacing the runtime JS-based `check_actor_trust` pre-activation step in
all four factory workflows (research, code, change, reproducer) with gh-aw's built-in role-based
membership checking and explicit bot allow-listing.

## MODIFIED Requirements

### Requirement: Actor trust gating uses gh-aw `check_membership` via implicit `roles` default
All four factory workflow sources (research, code, change, reproducer) SHALL rely on gh-aw's
built-in `check_membership` pre-activation step, which verifies the triggering actor holds one of
`admin`, `maintainer`, or `write` repository roles. This check is injected by the gh-aw compiler
by default for `issues` triggers with `roles` defaulting to `[admin, maintainer, write]`.

#### Scenario: Trusted actor activates the workflow
- **GIVEN** the factory workflow receives a qualifying event
- **WHEN** the triggering actor has a repository permission of `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL activate and proceed to downstream pre-activation steps

#### Scenario: Untrusted actor is blocked at the pre-activation level
- **GIVEN** the factory workflow receives a qualifying event
- **WHEN** the triggering actor has a repository permission below `write` (or has no repository permission)
- **THEN** the workflow SHALL NOT activate; the role check SHALL prevent the job from running

### Requirement: `code-factory` explicitly allows `github-actions[bot]` via `on.bots`
The `code-factory` workflow SHALL declare `on.bots: [github-actions[bot]]` so that
`workflow_dispatch` triggers originating from other workflows (e.g. `semantic-function-refactor`)
are allowed alongside the default role-based check.

#### Scenario: Workflow dispatch from `semantic-function-refactor` is allowed
- **GIVEN** `semantic-function-refactor` dispatches a `code-factory` run via `workflow_dispatch`
- **WHEN** the dispatch is sent by `github-actions[bot]`
- **THEN** the `code-factory` workflow SHALL activate because the bot is in the explicit allow list

### Requirement: `skip-author-associations` is NOT used
No factory workflow SHALL declare `skip-author-associations`. The role-based `check_membership`
check provides equivalent permission-level gating without relying on the potentially-unreliable
`author_association` webhook payload field.

### Requirement: `check_actor_trust` step is removed from all four workflows
The `check_actor_trust` pre-activation step SHALL be removed from all four factory workflow
sources. No downstream step SHALL gate on `steps.check_actor_trust.outputs.actor_trusted`.

#### Scenario: No runtime permission-level API call in JS
- **GIVEN** a qualifying event from a trusted actor
- **WHEN** the pre-activation job runs
- **THEN** its JavaScript steps SHALL NOT call `github.rest.repos.getCollaboratorPermissionLevel`

### Requirement: `normalize_context` SHALL emit `actor_trusted=true` unconditionally for issue-event paths
Each factory workflow's `normalize_context` step SHALL emit `actor_trusted=true` for the
issue-event path because gh-aw's `check_membership` has already filtered untrusted actors before
the job runs. The dispatch-intake path already hardcodes `actor_trusted=true`.

#### Scenario: normalize_context actor_trusted is always true for issue-event path
- **GIVEN** the role gate has passed (workflow is running)
- **WHEN** `normalize_context` processes the issue-event branch
- **THEN** it SHALL set `actor_trusted=true` unconditionally

### Requirement: `check-actor-trust.js` and its library functions are deleted
The file `.github/scripts/workflows/lib/factory-runners/check-actor-trust.js` SHALL be deleted.
The functions `factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` SHALL be removed
from `.github/scripts/workflows/lib/factory-issue-shared.js` along with their exports.

### Requirement: Actor-trust unit tests are removed
Test cases that test `factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` SHALL be
removed from `factory-issue-shared.test.mjs`, `code-factory-issue.test.mjs`, and
`change-factory-issue.test.mjs`. All other test coverage in those files SHALL remain intact.

### Requirement: Lock files are regenerated
All four factory workflow lock files SHALL be regenerated via `gh aw compile` after source
changes and committed together with the source edits.
