## MODIFIED Requirements

### Requirement: Label trigger (REQ-002)
The workflow SHALL use `label_command` for the `verify-openspec` label and SHALL restrict that command to `pull_request` events only. Applying `verify-openspec` to a pull request SHALL activate the workflow's verification path for that pull request, and labels other than `verify-openspec` SHALL NOT activate this workflow. The workflow SHALL NOT rely on a separate deterministic label-verification step to confirm the trigger label after activation.

#### Scenario: Correct label runs automation on pull requests
- **GIVEN** a pull request receives the label `verify-openspec`
- **WHEN** the `label_command` trigger activates
- **THEN** the workflow SHALL mark the run eligible for the agentic verification path subject to the remaining deterministic change-selection gate

#### Scenario: Other labels do not start verification
- **GIVEN** a pull request receives a label other than `verify-openspec`
- **WHEN** pull request labeling activity occurs
- **THEN** this workflow SHALL NOT activate for that label

#### Scenario: Non-pull-request items are out of scope
- **GIVEN** an issue or discussion receives the label `verify-openspec`
- **WHEN** labeling activity occurs
- **THEN** this workflow SHALL NOT activate because the command is configured for pull requests only

### Requirement: Permissions for read, review, and push (REQ-003)

The workflow SHALL request permissions sufficient to read the repository, submit pull request reviews and review comments, push commits to the pull request branch via `push-to-pull-request-branch`, and allow `label_command` activation to remove the `verify-openspec` label from the triggering pull request automatically. At minimum this SHALL include `contents: write`, `pull-requests: write`, and any additional compiler-required scope for automatic trigger-label removal on pull requests.

#### Scenario: Push safe output and trigger-label cleanup are permitted

- GIVEN the agent archives the change and produces a commit on the PR branch
- WHEN `push-to-pull-request-branch` runs and `label_command` activation removes the trigger label
- THEN the workflow token SHALL have authority to push to the PR head branch and mutate the triggering pull request label set under normal repository settings

### Requirement: Safe outputs for review and push (REQ-004)

The workflow SHALL declare safe outputs for:

- `create-pull-request-review-comment` with `max` large enough for verification and unassociated-file commentary.
- `submit-pull-request-review` with `max: 1` and `target` appropriate to the triggering pull request.
- `push-to-pull-request-branch` with `max: 1` (or documented policy) and `target: triggering`, plus any `checkout` `fetch` / `title-prefix` / `labels` required by repository policy and [GitHub Agentic Workflows - Push to PR branch](https://github.github.io/gh-aw/reference/safe-outputs-pull-requests/#push-to-pr-branch-push-to-pull-request-branch).

The workflow SHALL NOT declare a `remove-labels` safe output for `verify-openspec` cleanup because trigger-label removal is handled by `label_command` activation.

#### Scenario: One review decision per run

- GIVEN one workflow run completes verification
- WHEN reviews are submitted
- THEN at most one final submitted pull request review SHALL represent the approval decision before any archive push

#### Scenario: No redundant label-cleanup safe output is declared

- GIVEN maintainers inspect the workflow frontmatter
- WHEN they review the configured safe outputs
- THEN the workflow SHALL omit `remove-labels` for `verify-openspec` cleanup

### Requirement: Remove trigger label after workflow completion (REQ-015)

For a run triggered by applying the `verify-openspec` label to a pull request, the workflow SHALL rely on `label_command` automatic removal of that same label from the triggering pull request as part of activation rather than instructing the agent to request `remove-labels` before it concludes handling. The cleanup SHALL remove only `verify-openspec`; it SHALL NOT remove unrelated pull request labels, and the workflow SHALL NOT rely on terminal agent safe outputs or a separate post-agent cleanup job for this behavior.

#### Scenario: Activation removes the trigger label automatically

- GIVEN a pull request receives the `verify-openspec` label
- WHEN the `label_command` workflow activates
- THEN the triggering pull request SHALL lose the `verify-openspec` label without requiring an agent-emitted `remove-labels` output

#### Scenario: Comment or skipped runs do not require agent cleanup

- GIVEN a `verify-openspec`-triggered run later submits `COMMENT` or is skipped by deterministic gating before agent execution
- WHEN workflow activation has completed
- THEN trigger-label cleanup SHALL already be handled without waiting for terminal agent safe outputs

### Requirement: Deterministic gates may skip agent execution
The workflow SHALL use deterministic pre-activation outputs to decide whether the expensive agent job runs. When change-selection gating determines that the pull request is not eligible for verification, the workflow SHALL skip the agent job rather than starting it only to emit a no-op result. Trigger-label cleanup for `verify-openspec` SHALL already be handled by `label_command` activation and SHALL NOT depend on whether the agent job runs.

#### Scenario: Ineligible run skips agent job
- **GIVEN** deterministic gating concludes the pull request is not eligible for verification
- **WHEN** downstream job conditions are evaluated
- **THEN** the workflow SHALL skip the agent job
