## MODIFIED Requirements

### Requirement: Permissions for read, review, and push (REQ-003)

The workflow SHALL request permissions sufficient to read the repository, submit pull request reviews and review comments, push commits to the pull request branch via `push-to-pull-request-branch`, and remove the `verify-openspec` label from the triggering pull request via the `remove-labels` safe output. At minimum this SHALL include `contents: write`, `pull-requests: write`, and `issues: write` unless the agentic compiler emits a narrower equivalent that still allows those operations.

#### Scenario: Push safe output and label cleanup are permitted

- GIVEN the agent archives the change and produces a commit on the PR branch
- WHEN `push-to-pull-request-branch` and `remove-labels` safe outputs run
- THEN the token SHALL have authority to push to the PR head branch and mutate the triggering pull request label set under normal repository settings

### Requirement: Safe outputs for review and push (REQ-004)

The workflow SHALL declare safe outputs for:

- `create-pull-request-review-comment` with `max` large enough for verification and unassociated-file commentary.
- `submit-pull-request-review` with `max: 1` and `target` appropriate to the triggering pull request.
- `push-to-pull-request-branch` with `max: 1` (or documented policy) and `target: triggering`, plus any `checkout` `fetch` / `title-prefix` / `labels` required by repository policy and [GitHub Agentic Workflows - Push to PR branch](https://github.github.io/gh-aw/reference/safe-outputs-pull-requests/#push-to-pr-branch-push-to-pull-request-branch).
- `remove-labels` with `target: triggering`, `allowed` constrained to `verify-openspec`, and a `max` that permits the single trigger-label cleanup action.

#### Scenario: One review decision per run

- GIVEN one workflow run completes verification
- WHEN reviews are submitted
- THEN at most one final submitted pull request review SHALL represent the approval decision before any archive push

#### Scenario: Cleanup output is limited to the trigger label

- GIVEN the workflow requests label cleanup through safe outputs
- WHEN `remove-labels` is evaluated
- THEN the workflow configuration SHALL allow removal of `verify-openspec` and SHALL NOT require broader label-removal authority

### Requirement: Remove trigger label after workflow completion (REQ-015)

For a run triggered by applying the `verify-openspec` label, the workflow SHALL instruct the agent to request removal of that same label from the triggering pull request through the `remove-labels` safe output before the agent concludes its handling of the pull request. The cleanup request SHALL remove only `verify-openspec`; it SHALL NOT remove unrelated pull request labels, and the workflow SHALL NOT rely on a separate post-agent cleanup script or job for this behavior.

#### Scenario: Approved run requests trigger label cleanup

- GIVEN a `verify-openspec`-triggered run submits an `APPROVE` review and completes archive or push handling as applicable
- WHEN the agent emits its final safe outputs for the run
- THEN those outputs SHALL include removal of the `verify-openspec` label from the triggering pull request

#### Scenario: Non-approval run requests trigger label cleanup

- GIVEN a `verify-openspec`-triggered run ends with `COMMENT` or `noop` after agent handling begins
- WHEN the agent emits its final safe outputs for the run
- THEN those outputs SHALL include removal of the `verify-openspec` label from the triggering pull request
