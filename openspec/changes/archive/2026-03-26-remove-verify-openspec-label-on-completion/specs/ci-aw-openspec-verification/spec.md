## MODIFIED Requirements

### Requirement: Permissions for read, review, and push (REQ-003)

The workflow SHALL request permissions sufficient to read the repository, submit pull request reviews and review comments, push commits to the pull request branch via `push-to-pull-request-branch`, and remove the `verify-openspec` label from the triggering pull request when the run completes. At minimum this SHALL include `contents: write`, `pull-requests: write`, and `issues: write` unless the agentic compiler emits a narrower equivalent that still allows those operations.

#### Scenario: Push safe output and label cleanup are permitted

- GIVEN the agent archives the change and produces a commit on the PR branch
- WHEN `push-to-pull-request-branch` runs and the workflow removes `verify-openspec` during completion cleanup
- THEN the token SHALL have authority to push to the PR head branch and mutate the pull request labels under normal repository settings

## ADDED Requirements

### Requirement: Remove trigger label after workflow completion (REQ-015)

For a run triggered by applying the `verify-openspec` label, the workflow SHALL remove that same label from the triggering pull request before the run fully completes, regardless of whether the verification outcome is `APPROVE`, `COMMENT`, `noop`, or failure after the workflow has started. The workflow SHALL remove only `verify-openspec`; it SHALL NOT remove unrelated pull request labels as part of this cleanup behavior.

#### Scenario: Approved run clears trigger label

- GIVEN a `verify-openspec`-triggered run submits an `APPROVE` review and completes archive or push steps as applicable
- WHEN the workflow enters its final completion handling
- THEN the workflow SHALL remove the `verify-openspec` label before the run completes

#### Scenario: Non-approval run clears trigger label

- GIVEN a `verify-openspec`-triggered run ends with `COMMENT`, `noop`, or failure
- WHEN the workflow enters its final completion handling
- THEN the workflow SHALL remove the `verify-openspec` label before the run completes
