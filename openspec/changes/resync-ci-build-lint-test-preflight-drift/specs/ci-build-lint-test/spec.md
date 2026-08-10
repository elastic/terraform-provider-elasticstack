## MODIFIED Requirements

### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Provider CI`. The workflow SHALL run on `push` to branch `main`. The workflow SHALL run on `pull_request` events of type `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Push to main

- GIVEN a `push` to `main`
- WHEN the change-classification job reports `provider_changes=true`
- THEN build, lint, and test jobs MAY run per other requirements

### Requirement: Job permissions (REQ-028–REQ-029)

The change-classification job SHALL request the minimum permissions required to inspect pull requests (`contents: read`, `pull-requests: read`). The acceptance test job SHALL request `contents: read`, `issues: write`, and `pull-requests: write` permissions.

#### Scenario: Change-classification permissions

- GIVEN the change-classification job definition
- WHEN permissions are evaluated
- THEN they SHALL match the minimum set for listing PRs

### Requirement: Change classification gate (REQ-032–REQ-033)

The workflow SHALL evaluate whether the `build`, `lint`, `golangci-lint`, and matrix acceptance `test` jobs are required for the current change set via a dedicated change-classification job (`classify`) that runs unconditionally on every trigger. For `pull_request` events, the classifier SHALL set `provider_changes=false` only when every changed file is non-impacting: exactly `CHANGELOG.md`, or any path under `openspec/`, or any path under `.agents/`, or any path under `.github/` other than `.github/workflows/provider.yml` itself. Any change set containing at least one path outside that non-impacting set, or an empty changed-file list, SHALL set `provider_changes=true`. For non-`pull_request` events (including `push` and `workflow_dispatch`), the classifier SHALL skip file inspection entirely and unconditionally set `provider_changes=true`.

When the change-classification job runs, it SHALL expose its result as a workflow output that downstream jobs can consume when deciding whether those jobs are required.

#### Scenario: OpenSpec-only change set

- **GIVEN** a `pull_request` workflow run whose changed files are all under `openspec/`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Provider-impacting change set

- **GIVEN** a `pull_request` workflow run whose changed files include at least one path outside the non-impacting set
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=true`

#### Scenario: Non-pull_request event always classifies as provider-impacting

- **GIVEN** a non-`pull_request` event triggering the workflow (including `push` or `workflow_dispatch`)
- **WHEN** the change-classification job runs
- **THEN** it SHALL report `provider_changes=true` without inspecting the changed-file list

### Requirement: Provider gate job (REQ-034–REQ-036)

The workflow SHALL publish a `gate` job ("Provider Gate") that always reports a final required-check result for the workflow run, evaluating the change-classification result together with the `build`, `lint`, `golangci-lint`, and matrix acceptance `test` job results.

The `gate` job SHALL succeed when either of the following is true:

* The change-classification job reports `provider_changes=false` and `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job are all intentionally skipped
* `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job all complete successfully (regardless of the classify result)

The `gate` job SHALL fail when any of the following is true:

* Any of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports `failure` or `cancelled`
* The change-classification job reports `provider_changes=true` and at least one of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports an unexpected `skipped` result

The `gate` job SHALL provide a stable required-check target that can be used by GitHub branch protection or rulesets instead of the per-version matrix acceptance checks or the individual `build`/`lint`/`golangci-lint` checks.

#### Scenario: OpenSpec-only pull request

- **GIVEN** a pull request whose changed files are all under `openspec/`
- **WHEN** the workflow reaches the `gate` job
- **THEN** `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job SHALL be treated as intentionally skipped
- **AND** the `gate` job SHALL succeed

#### Scenario: Provider change with failing acceptance coverage

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the matrix acceptance `test` job does not complete successfully
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

#### Scenario: Provider change with an unexpected skip

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** at least one of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports `skipped`
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

### Requirement: Auto-approve job (REQ-018–REQ-021)

The `auto-approve` job SHALL depend on the `gate` job and SHALL only run on `pull_request` events. `auto-approve` SHALL require the `gate` job to succeed before it runs, unconditionally — there is no event-type carve-out. The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after satisfied validation

- **GIVEN** a pull request workflow and a successful `gate` job result
- **WHEN** auto-approve runs
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

#### Scenario: Auto-approve does not run when the gate fails

- **GIVEN** a pull request workflow and a `gate` job result other than `success`
- **WHEN** the workflow evaluates whether to run `auto-approve`
- **THEN** the `auto-approve` job SHALL NOT run

## REMOVED Requirements

### Requirement: Preflight gate (REQ-023–REQ-027)

The workflow SHALL evaluate whether to execute CI jobs via a dedicated preflight gate job that emits a `should_run` output.

For `push` events, the preflight gate SHALL set `should_run=true` when either:

* No open pull request exists for the pushed branch in the same repository
* All commits in the push event were authored by an allowed bot user: Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)

For `push` events where **neither** of the above holds, the preflight gate SHALL set `should_run=false`.

For non-`push` events (`pull_request` and `workflow_dispatch`), the preflight gate SHALL set `should_run=true`, except for `pull_request` events of type `ready_for_review` where it SHALL set `should_run=false`.

The `build`, `lint`, and matrix acceptance `test` jobs SHALL only execute when the preflight gate outputs `should_run=true`.

#### Scenario: Push without open PR

- GIVEN a push to a branch with no open PR in the same repository
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and all commits by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND every commit in the push event was authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and a commit not by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND at least one commit in the push event was not authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `false` and downstream jobs SHALL be skipped

**Reason**: `.github/workflows/provider.yml` has no `preflight` job and no `should_run` output. The real
gate structure is: `classify` (Change Classification) runs unconditionally and exposes
`provider_changes`; `build`/`lint`/`golangci-lint`/`test` gate on
`needs.classify.outputs.provider_changes == 'true'`; and the terminal `gate` job (Provider Gate)
evaluates all of those results via `gateProvider()`. There is no separate job that decides whether CI
runs at all based on push-author or open-PR checks.

**Migration**: Readers looking for "should CI run at all" behavior should refer to the "Change
classification gate" requirement (the `classify` job) and the "Provider gate job" requirement (the
`gate` job) in this same capability.

### Requirement: Ready-for-review behavior (REQ-030)

On `ready_for_review` `pull_request` events, the workflow SHALL keep the preflight gate behavior that prevents the `build`, `lint`, change-classification, and matrix acceptance `test` jobs from running. The `Test Validation` job SHALL succeed based on the intentional preflight skip, and `auto-approve` SHALL remain eligible to run.

#### Scenario: Ready for review event

- **GIVEN** a `pull_request` with action `ready_for_review`
- **WHEN** the workflow runs
- **THEN** `build`, `lint`, change-classification, and matrix acceptance `test` SHALL be skipped by the preflight gate
- **AND** `Test Validation` SHALL succeed
- **AND** auto-approve SHALL be eligible to run

**Reason**: `provider.yml`'s `pull_request` trigger types are `[opened, synchronize, reopened]`;
`ready_for_review` is not a configured trigger at all, and `gateProvider()` has no event-type
branching. This requirement describes a skip path for an event the workflow does not even listen for.

**Migration**: None. There is no `ready_for_review`-specific behavior to preserve or relocate; standard
`pull_request` triggers (`opened`, `synchronize`, `reopened`) are already covered by "Workflow identity
and triggers" and the `gate`/`classify` requirements in this capability.

### Requirement: Generated changelog pull requests can reach auto-approve without full CI
The `Build/Lint/Test` workflow SHALL allow same-repository pull requests from branch `generated-changelog` that are authored by `github-actions[bot]` and modify only `CHANGELOG.md` to reach the `auto-approve` job without requiring the full build, lint, change-classification, or matrix acceptance test path to run. The skip condition MUST verify all three criteria — branch name, PR author, and file list — in the preflight gate before setting `should_run=false`.

#### Scenario: Generated changelog PR reaches auto-approve path
- **GIVEN** a same-repository pull request from branch `generated-changelog`
- **AND** the PR author is `github-actions[bot]`
- **AND** the pull request changes only `CHANGELOG.md`
- **WHEN** the workflow evaluates its execution path
- **THEN** the workflow SHALL produce a successful path that leaves `auto-approve` eligible to run
- **AND** it SHALL NOT require the full build, lint, and matrix acceptance test jobs for that PR

#### Scenario: Auto-merge is gated on the approval outcome
- **GIVEN** a `generated-changelog` PR
- **WHEN** the `auto-approve` job runs
- **THEN** auto-merge SHALL only be enabled if the auto-approve script determined `ShouldApprove` or `AlreadyApproved` is true (reported via a `GITHUB_OUTPUT` step output)
- **AND** auto-merge SHALL NOT be enabled if the auto-approve gates reject the PR

**Reason**: This requirement describes a preflight-gate CI bypass for `generated-changelog` PRs that
does not exist in `provider.yml`. Generated-changelog auto-approve policy is already owned and
documented by the `ci-pr-auto-approve` capability; keeping a duplicate (and inaccurate) CI-skip claim
here is harmful.

**Migration**: None for this capability. Refer to `openspec/specs/ci-pr-auto-approve/spec.md` for the
generated-changelog selector, commit-author, and file-allowlist requirements. Script implementation
parity with those specs is out of scope for this change.

### Requirement: Changelog-only bypass remains narrowly scoped
The `Build/Lint/Test` workflow SHALL keep the changelog-only bypass narrowly scoped to the generated changelog automation shape. Other changelog-only pull requests SHALL NOT gain the same bypass unless they satisfy all three repository-authored generated-changelog conditions: branch name `generated-changelog`, PR author `github-actions[bot]`, and files limited to `CHANGELOG.md`.

#### Scenario: Manual changelog-only PR does not inherit generated-changelog bypass
- **GIVEN** a pull request changes only `CHANGELOG.md`
- **AND** its head branch name is not `generated-changelog`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL NOT treat that pull request as the generated-changelog special case

#### Scenario: Wrong author does not inherit generated-changelog bypass
- **GIVEN** a pull request from branch `generated-changelog` changes only `CHANGELOG.md`
- **AND** the PR author is not `github-actions[bot]`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL run full CI rather than skipping to the auto-approve path

**Reason**: `provider.yml` has no preflight-gate-based changelog CI bypass to scope narrowly.
Generated-changelog auto-approve policy remains owned by `ci-pr-auto-approve`, which already
documents the selector requirements; this duplicate scoping claim is removed from
`ci-build-lint-test`.

**Migration**: None for this capability. Refer to `openspec/specs/ci-pr-auto-approve/spec.md` for
generated-changelog policy. Script implementation parity with those specs is out of scope for this
change.
