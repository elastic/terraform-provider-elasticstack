## MODIFIED Requirements

### Requirement: Label trigger (REQ-002)
The workflow SHALL run on `pull_request_target` events of type `labeled`. A deterministic pre-activation step SHALL verify that `github.event.label.name` (or equivalent injected label data) is exactly `verify-openspec` and SHALL publish the result for downstream jobs. The workflow SHALL not perform verification or archive steps when that deterministic check indicates a different label.

#### Scenario: Correct label runs automation from base repository context
- **GIVEN** a pull request receives the label `verify-openspec`
- **WHEN** GitHub dispatches the `pull_request_target` `labeled` event
- **THEN** the deterministic pre-activation gate SHALL mark the run eligible for the agentic verification path

#### Scenario: Other labels do not start verification
- **GIVEN** a pull request receives a label other than `verify-openspec`
- **WHEN** the `pull_request_target` `labeled` event fires
- **THEN** the workflow SHALL not perform verification or archive steps for that event

### Requirement: Permissions for read, review, and push (REQ-003)
The workflow SHALL request permissions sufficient to read the repository, submit pull request reviews and review comments, push commits to the pull request branch via `push-to-pull-request-branch` for deterministically eligible same-repository pull requests, and remove the `verify-openspec` label from the triggering pull request via a deterministic script step. At minimum this SHALL include `contents: write`, `pull-requests: write`, and `issues: write` unless the agentic compiler emits a narrower equivalent that still allows those operations. The workflow SHALL request the label-mutation write scope in the deterministic path that performs label cleanup rather than depending on agent safe-output processing to hold that authority.

#### Scenario: Same-repository archive push and deterministic label cleanup are permitted
- **GIVEN** the agent archives the change and produces a commit on a deterministically eligible same-repository pull request branch
- **WHEN** `push-to-pull-request-branch` runs and the deterministic cleanup step removes `verify-openspec`
- **THEN** the workflow token SHALL have authority to push to the PR head branch and mutate the triggering pull request label set under normal repository settings

#### Scenario: Fork pull requests still permit review and cleanup
- **GIVEN** the triggering pull request head repository differs from the base repository
- **WHEN** the workflow submits a review and removes `verify-openspec`
- **THEN** the workflow permissions SHALL still allow review and label cleanup without requiring cross-repository push authority

### Requirement: Safe outputs for review and push (REQ-004)
The workflow SHALL declare safe outputs for:

- `create-pull-request-review-comment` with `max` large enough for verification and unassociated-file commentary.
- `submit-pull-request-review` with `max: 1` and `target` appropriate to the triggering pull request.
- `push-to-pull-request-branch` with `max: 1` (or documented policy) and `target: triggering`, plus any `checkout` `fetch` / `title-prefix` / `labels` required by repository policy and [GitHub Agentic Workflows - Push to PR branch](https://github.github.io/gh-aw/reference/safe-outputs-pull-requests/#push-to-pr-branch-push-to-pull-request-branch).

The workflow SHALL NOT declare a `remove-labels` safe output for `verify-openspec` cleanup because label cleanup is handled by a deterministic workflow step rather than by agent terminal outputs.

#### Scenario: One review decision per run
- **GIVEN** one workflow run completes verification
- **WHEN** reviews are submitted
- **THEN** at most one final submitted pull request review SHALL represent the approval decision before any archive push

#### Scenario: No redundant label-cleanup safe output is declared
- **GIVEN** maintainers inspect the workflow frontmatter
- **WHEN** they review the configured safe outputs
- **THEN** the workflow SHALL omit `remove-labels` for `verify-openspec` cleanup

### Requirement: Discover active change id from PR files (REQ-005)
The workflow SHALL use deterministic pre-activation steps to load the pull request changed files list, including each file entry's status (`added`, `modified`, `removed`, `renamed`, and so on), and to classify whether the triggering pull request is a same-repository pull request or a fork pull request. It SHALL consider only paths matching `openspec/changes/<id>/...` where `<id>` is a single path segment and `archive` is not the first segment (that is, exclude `openspec/changes/archive/**`). For each such path, it SHALL record the status of that file entry and SHALL publish pre-activation outputs that include the gate result, the selected active change id when selection succeeds, a deterministic review disposition that distinguishes approval-eligible modified-only changes from comment-only net-new change proposals, a deterministic verification mode (`workspace` for same-repository pull requests, `api-only` for fork pull requests), a deterministic archive/push eligibility result, and deterministic reason strings that explain those classifications.

#### Scenario: Selected change and execution mode are exposed to the agent
- **GIVEN** exactly one active change satisfies the change-selection rules
- **WHEN** the deterministic pre-activation steps complete
- **THEN** the workflow SHALL expose that change id, the deterministic review disposition, the deterministic verification mode, the deterministic archive/push eligibility result, and their reason strings as pre-activation outputs for the later agent job

#### Scenario: Fork pull request is classified before agent reasoning
- **GIVEN** the triggering pull request head repository differs from the base repository
- **WHEN** deterministic pre-activation classification runs
- **THEN** the workflow SHALL publish `api-only` verification mode and archive/push ineligibility before the agent starts

### Requirement: Verification using active OpenSpec tooling (REQ-007)
For the selected change id and deterministic execution outputs published by the pre-activation steps, the agent SHALL follow `.agents/skills/openspec-verify-change/SKILL.md` while respecting the deterministic verification mode. In `workspace` mode, the agent SHALL use standard OpenSpec commands, including where applicable `npx openspec status --change "<id>" --json` and `npx openspec instructions apply --change "<id>" --json`, and SHALL perform verification with context rooted at `openspec/changes/<id>/`. In `api-only` mode, the agent SHALL verify from pull request changed files, diffs, and deterministic workflow outputs and SHALL NOT require repository bootstrap or execution of fork-controlled workspace commands before review submission. The prompt SHALL consume the selected change id, review disposition, verification mode, archive/push eligibility, and their reasons from workflow outputs rather than requiring the agent to rediscover pull request trust or archive/push policy before verification.

#### Scenario: Same-repository pull request uses workspace verification
- **GIVEN** the triggering pull request head repository matches the base repository and deterministic setup selected a single active change
- **WHEN** verification runs
- **THEN** the workflow SHALL permit the agent to use local OpenSpec tooling rooted at `openspec/changes/<id>/`

#### Scenario: Fork pull request uses API-only verification
- **GIVEN** the triggering pull request head repository differs from the base repository and deterministic setup selected a single active change
- **WHEN** verification runs
- **THEN** the agent SHALL review the change from pull request metadata and diffs without depending on repository bootstrap in the trusted workflow context

### Requirement: Review event APPROVE vs COMMENT (REQ-012)
The agent SHALL submit a pull request review with `APPROVE` if and only if the deterministic review disposition is approval-eligible and there are zero CRITICAL issues and zero `unassociated` files; otherwise `COMMENT`. The agent SHALL submit `COMMENT` whenever the deterministic review disposition is comment-only, including cases where verification finds zero CRITICAL issues and zero `unassociated` files. It SHALL not use `REQUEST_CHANGES`. WARNING and SUGGESTION alone SHALL NOT block `APPROVE` for approval-eligible runs. Deterministic archive/push eligibility SHALL NOT by itself force the review event to `COMMENT`.

#### Scenario: Fork pull request may still receive APPROVE
- **GIVEN** deterministic gating marked the selected active change approval-eligible, verification found zero CRITICAL issues and zero `unassociated`, and archive/push eligibility is disallowed because the pull request comes from a fork
- **WHEN** the review is submitted
- **THEN** `event` SHALL be `APPROVE`

#### Scenario: Net-new change proposal remains comment-only
- **GIVEN** deterministic gating marked the selected active change comment-only because it includes added files
- **AND** verification found zero CRITICAL issues and zero `unassociated`
- **WHEN** the review is submitted
- **THEN** `event` SHALL be `COMMENT`

### Requirement: Archive after APPROVE only (REQ-013)
Only when the agent submits a pull request review with `APPROVE` for an approval-eligible run and deterministic archive/push eligibility is allowed, the workflow SHALL archive the selected change `<id>` using repository-standard automation (for example, `openspec archive <id>` and/or steps aligned with `openspec-archive-change`), updating `openspec/changes/archive/` and canonical specs per project policy.

#### Scenario: Fork approval does not archive
- **GIVEN** the review event is `APPROVE`
- **AND** deterministic archive/push eligibility is disallowed because the pull request comes from a fork
- **WHEN** the run completes
- **THEN** the workflow SHALL NOT move the change to `archive/` or mutate canonical specs for this id via this workflow

#### Scenario: Same-repository approval may archive
- **GIVEN** the review event is `APPROVE`
- **AND** deterministic archive/push eligibility is allowed
- **WHEN** the run completes
- **THEN** the workflow SHALL archive the selected change according to repository policy

### Requirement: Push archive result to PR branch (REQ-014)
After a successful archive step on a deterministically eligible same-repository pull request, the workflow SHALL commit the working tree changes and SHALL apply `push-to-pull-request-branch` so the pull request head branch contains the archive commit(s). When deterministic archive/push eligibility is disallowed, the workflow SHALL NOT attempt to push to the pull request head branch.

#### Scenario: Same-repository PR branch updated after approval
- **GIVEN** archive produced local commits on the checked-out PR branch for a deterministically eligible same-repository pull request
- **WHEN** push safe output succeeds
- **THEN** the open pull request SHALL show new commits from this workflow run

#### Scenario: Fork pull request does not attempt push
- **GIVEN** deterministic archive/push eligibility is disallowed because the pull request comes from a fork
- **WHEN** the review run completes
- **THEN** the workflow SHALL NOT call `push-to-pull-request-branch`

### Requirement: Remove trigger label with a deterministic workflow step (REQ-015)
For a run triggered by applying the `verify-openspec` label, the workflow SHALL remove that same label from the triggering pull request through a deterministic repository-authored script step after the label has been verified, rather than through agent safe outputs. The cleanup step SHALL remove only `verify-openspec`; it SHALL NOT remove unrelated pull request labels, and the workflow SHALL NOT rely on terminal agent safe outputs or a separate post-agent cleanup job for this behavior.

#### Scenario: Deterministic cleanup removes only the trigger label
- **GIVEN** a pull request receives the `verify-openspec` label
- **WHEN** deterministic trigger handling confirms that the run was activated by `verify-openspec`
- **THEN** the workflow step SHALL remove `verify-openspec` from the triggering pull request and SHALL NOT remove any other labels

#### Scenario: Cleanup does not depend on agent execution
- **GIVEN** a `verify-openspec`-triggered run is later skipped by deterministic gating or ends without archive/push behavior
- **WHEN** deterministic trigger handling has completed
- **THEN** trigger-label cleanup SHALL already be handled without waiting for agent safe outputs

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins only for deterministic `workspace` verification mode. At a minimum, that trusted workspace mode SHALL set up Node using `actions/setup-node` with `node-version-file: package.json`, SHALL configure Go in the runner environment through `actions/setup-go` with `go-version-file: go.mod`, SHALL export `GOROOT`, `GOPATH`, and `GOMODCACHE` after Go setup for AWF chroot mode, SHALL allow the Go ecosystem in the workflow's AWF network policy, and SHALL NOT use workflow frontmatter `runtimes.go` for Go provisioning. Fork pull requests in deterministic `api-only` mode SHALL NOT require this bootstrap before review submission.

#### Scenario: Same-repository verification prepares toolchains
- **GIVEN** deterministic verification mode is `workspace`
- **WHEN** the review environment is prepared
- **THEN** the workflow SHALL provision Node, Go, Terraform, and exported Go paths before agent reasoning begins

#### Scenario: Fork verification does not require trusted workspace bootstrap
- **GIVEN** deterministic verification mode is `api-only`
- **WHEN** the review run prepares the agent context
- **THEN** the workflow SHALL NOT require repository toolchain bootstrap as a prerequisite for submitting the review

### Requirement: Review environment installs repository dependencies before verification
Before the agent performs verification in deterministic `workspace` mode, the workflow SHALL run `make setup` in the agent workspace after runtime provisioning completes. This bootstrap SHALL make `npx openspec` available locally, SHALL prepare repository Go dependencies needed by agent-invoked Go commands through the repository's standard setup path, and SHALL preserve access to the prepared Go workspace and module cache for AWF agent commands during verification. Deterministic `api-only` mode SHALL NOT require `make setup` before review submission.

#### Scenario: Same-repository workspace runs repository setup
- **GIVEN** deterministic verification mode is `workspace`
- **WHEN** the workflow prepares the repository for agent verification
- **THEN** it SHALL run `make setup` in the review workspace before agent reasoning begins

#### Scenario: Fork review skips repository setup
- **GIVEN** deterministic verification mode is `api-only`
- **WHEN** the workflow prepares the review run
- **THEN** it SHALL be able to reach review submission without running `make setup`

### Requirement: Deterministic agent setup before verification
The workflow SHALL use deterministic custom workflow steps to prepare the repository workspace before agent reasoning begins only when deterministic verification mode is `workspace`. In that mode, after the review toolchains are provisioned, it SHALL run `make setup` at the repository root so `npx openspec` is available and repository Go dependencies are prepared per the review-environment bootstrap requirement, without the prompt having to rediscover those steps. In deterministic `api-only` mode, the workflow SHALL not require workspace bootstrap to start review reasoning.

#### Scenario: Same-repository workspace is ready before agent reasoning
- **GIVEN** deterministic verification mode is `workspace`
- **WHEN** the agent job starts for a verification run
- **THEN** deterministic custom steps SHALL complete `make setup` before the agent uses `npx openspec`

#### Scenario: Fork review starts without workspace bootstrap
- **GIVEN** deterministic verification mode is `api-only`
- **WHEN** the agent job starts for a verification run
- **THEN** the workflow SHALL provide review context without requiring `make setup`

### Requirement: Deterministic gates may skip agent execution
The workflow SHALL use deterministic pre-activation outputs to decide whether the expensive agent job runs. When label verification or change-selection gating determines that the pull request is not eligible for verification, the workflow SHALL skip the agent job rather than starting it only to emit a no-op result. Fork classification alone SHALL NOT skip the agent job; instead it SHALL select `api-only` verification mode and archive/push ineligibility while still allowing review execution.

#### Scenario: Ineligible label or change selection skips agent job
- **GIVEN** deterministic pre-activation gating concludes that the label is not `verify-openspec` or the pull request is not eligible for change verification
- **WHEN** downstream job conditions are evaluated
- **THEN** the workflow SHALL skip the agent job

#### Scenario: Fork pull request still runs review path
- **GIVEN** deterministic pre-activation gating selected a single active change and classified the pull request as a fork
- **WHEN** downstream job conditions are evaluated
- **THEN** the workflow SHALL continue to the agent job in `api-only` mode
