## Why

The `openspec-verify-label` workflow currently removes `verify-openspec` with a separate `completion_cleanup` job and an inline script. That duplicates behavior the agentic workflow can already express through safe outputs, and it keeps the label-cleanup contract outside the agent prompt and safe-output configuration.

## What Changes

- Update the `openspec-verify-label` workflow contract to remove `verify-openspec` through the `remove-labels` safe output instead of a dedicated cleanup job.
- Instruct the agent prompt to emit the safe output that removes only the triggering `verify-openspec` label when the run finishes handling the PR.
- Remove the workflow-specific cleanup script/job that exists only to mutate labels after agent execution.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: change trigger-label cleanup from a bespoke completion job to the built-in `remove-labels` safe output contract

## Impact

- `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`
- Compiled `openspec-verify-label` workflow outputs and permissions
- `ci-aw-openspec-verification` OpenSpec requirements and any prompt text that defines end-of-run cleanup behavior
