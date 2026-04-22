# `ci-pr-auto-fix` — Label-gated failed-CI remediation for pull requests

Workflow implementation: authored source under `.github/workflows-src/`, compiled to `.github/workflows/`.

## Purpose

Define requirements for a GitHub Agentic Workflow that reacts to failed CI on opted-in pull requests, classifies supported failure profiles deterministically, and asks Copilot to remediate or summarize the failure directly on the pull request.

## ADDED Requirements

### Requirement: Workflow artifacts and compilation
The remediation workflow SHALL be authored as a GitHub Agentic Workflow markdown source and SHALL include a compiled `.lock.yml` generated from that source and committed with the change. Contributors SHALL NOT hand-edit the compiled lock artifact.

#### Scenario: Source and lock stay paired
- **WHEN** maintainers change the remediation workflow behavior
- **THEN** the authored source and compiled lock artifact SHALL match the compiler output committed in the repository

### Requirement: Trigger on failed CI workflow runs
The remediation workflow SHALL run from a `workflow_run` trigger for the repository's main pull-request CI workflow and SHALL continue only when the triggering workflow run concluded with `failure` for a `pull_request` event.

#### Scenario: Failed pull-request CI run is eligible for gating
- **WHEN** the configured CI workflow completes with conclusion `failure` for a pull-request event
- **THEN** the remediation workflow SHALL proceed to deterministic eligibility checks

#### Scenario: Successful CI run does not remediate
- **WHEN** the configured CI workflow completes with conclusion `success`
- **THEN** the remediation workflow SHALL NOT invoke remediation for that run

#### Scenario: Non-pull-request CI run does not remediate
- **WHEN** the configured CI workflow completes for a non-`pull_request` event such as `push` or `workflow_dispatch`
- **THEN** the remediation workflow SHALL NOT invoke remediation for that run

### Requirement: Same-repository PR resolution and `auto-fix` label gate
Before agent reasoning starts, deterministic repository-authored steps SHALL resolve the open pull request associated with the failed run, SHALL require exactly one matching same-repository pull request, and SHALL require that pull request to carry the `auto-fix` label. The workflow SHALL skip remediation for fork pull requests, unlabeled pull requests, and ambiguous pull-request matches.

#### Scenario: Same-repository labeled PR is eligible
- **WHEN** deterministic resolution finds exactly one open same-repository pull request for the failed run and that pull request has label `auto-fix`
- **THEN** the workflow SHALL mark the run eligible for remediation

#### Scenario: Unlabeled PR is skipped
- **WHEN** deterministic resolution finds the associated pull request but it does not have label `auto-fix`
- **THEN** the workflow SHALL skip remediation for that run

#### Scenario: Fork PR is skipped
- **WHEN** deterministic resolution finds that the failed run belongs to a pull request whose head repository differs from the base repository
- **THEN** the workflow SHALL skip remediation for that run

#### Scenario: Ambiguous branch match is skipped
- **WHEN** deterministic resolution finds zero or more than one matching open same-repository pull request for the failed run
- **THEN** the workflow SHALL skip remediation for that run

### Requirement: Deterministic failure classification and context capture
Before agent reasoning starts, deterministic repository-authored steps SHALL inspect the failed run's jobs and classify the run into supported remediation profiles. The workflow SHALL support at least `lint` failures and `acceptance` failures. It SHALL capture the source workflow run URL and the URL of each failed supported job, and it SHALL capture version-specific context for failed acceptance test jobs.

#### Scenario: Lint failure is classified
- **WHEN** the failed run includes a failed `Lint` job
- **THEN** the remediation context SHALL include the `lint` profile and the failed lint job URL

#### Scenario: Acceptance failures are classified with version context
- **WHEN** the failed run includes one or more failed matrix acceptance test jobs
- **THEN** the remediation context SHALL include the `acceptance` profile, the failed job URLs, and the version-specific identity of each failed matrix entry

#### Scenario: Unsupported-only failure is skipped
- **WHEN** the failed run contains no supported remediation profiles
- **THEN** the workflow SHALL skip agent remediation for that run

### Requirement: Security gating before checkout
Because the remediation workflow runs from `workflow_run`, it SHALL keep pull-request eligibility checks and failure classification deterministic and SHALL NOT checkout the pull-request head or execute repository code from that head branch until the run has been confirmed as an eligible same-repository `auto-fix` pull request.

#### Scenario: Ineligible run never checks out PR head
- **WHEN** deterministic gating concludes that the failed run is unsupported, fork-based, unlabeled, or otherwise ineligible
- **THEN** the workflow SHALL exit without checking out the pull-request head branch

#### Scenario: Eligible run may checkout PR head after gating
- **WHEN** deterministic gating concludes that the failed run is an eligible same-repository `auto-fix` pull request
- **THEN** the workflow MAY checkout the pull-request head branch for remediation

### Requirement: Lint remediation behavior
When the remediation context includes a `lint` failure, the agent SHALL receive the failed lint job link and SHALL attempt to produce a fix on the pull-request branch. If the agent cannot produce a safe fix, the workflow SHALL leave pull-request feedback summarizing the blocker instead of pretending remediation succeeded.

#### Scenario: Lint fix is pushed
- **WHEN** the remediation run includes a `lint` failure and the agent produces a concrete fix
- **THEN** the workflow SHALL update the pull-request branch with that fix

#### Scenario: Lint fix is not available
- **WHEN** the remediation run includes a `lint` failure but the agent cannot produce a safe fix
- **THEN** the workflow SHALL leave pull-request feedback summarizing why the fix was not applied

### Requirement: Acceptance remediation behavior
When the remediation context includes `acceptance` failures, the agent SHALL receive the failed acceptance job links and failed version context. The agent SHALL attempt a code fix only when there is a clear path to resolving the failure; otherwise the workflow SHALL leave pull-request feedback summarizing the issue.

#### Scenario: Acceptance failure has a clear fix
- **WHEN** the remediation run includes one or more acceptance failures and the agent determines there is a clear path to fixing them
- **THEN** the workflow SHALL update the pull-request branch with the resulting fix

#### Scenario: Acceptance failure is analysis-only
- **WHEN** the remediation run includes one or more acceptance failures and the agent does not determine a clear path to fixing them
- **THEN** the workflow SHALL leave pull-request feedback summarizing the failure instead of pushing speculative changes

### Requirement: Pull-request feedback is visible and idempotent
For skipped, unsupported, and analysis-only outcomes, the workflow SHALL create or update a marker-based pull-request comment so maintainers can inspect what happened without comment spam. Reruns for the same source workflow run SHALL update the existing marked comment instead of creating duplicates.

#### Scenario: First analysis-only run creates a comment
- **WHEN** a remediation run needs to report a skipped or analysis-only outcome and no marked comment exists for the source workflow run
- **THEN** the workflow SHALL create one pull-request comment with that outcome

#### Scenario: Rerun updates the existing comment
- **WHEN** a remediation rerun reports the same source workflow run outcome and a marked comment already exists
- **THEN** the workflow SHALL update that comment instead of creating another one

### Requirement: Branch updates stay on the triggering PR branch
When the agent produces a fix, the workflow SHALL push changes only to the triggering same-repository pull-request branch. The workflow SHALL NOT attempt to push branch updates for fork pull requests or other ineligible runs.

#### Scenario: Same-repository fix updates the PR branch
- **WHEN** an eligible remediation run produces a fix
- **THEN** the workflow SHALL push the resulting commit or commits to the triggering pull-request branch

#### Scenario: Fork run never pushes
- **WHEN** the failed run belongs to a fork pull request
- **THEN** the workflow SHALL NOT attempt to push branch updates

### Requirement: Follow-up CI reruns are supported by documented workflow configuration
The remediation workflow and its maintainer documentation SHALL define how agent-authored branch updates can trigger CI again when repository operators configure the required GH AW CI-trigger authentication.

#### Scenario: CI trigger support is documented
- **WHEN** maintainers inspect the remediation workflow documentation
- **THEN** they SHALL be able to identify the repository configuration needed for follow-up CI runs after agent-authored pushes
