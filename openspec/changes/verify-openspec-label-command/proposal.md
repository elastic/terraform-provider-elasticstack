## Why

The `openspec-verify-label` workflow currently models `verify-openspec` as a filtered `pull_request:labeled` event and then asks the agent to remove that label through `remove-labels` safe output cleanup. GitHub Agentic Workflows already provide `label_command` for this pattern, so the workflow can use the built-in one-shot label trigger and drop redundant cleanup behavior from both frontmatter and prompt instructions.

## What Changes

- Change the `openspec-verify-label` workflow trigger from filtered `pull_request` `labeled` events to `label_command` for the `verify-openspec` label.
- Restrict the new trigger to pull requests only so the workflow keeps its current PR-only activation behavior.
- Remove the explicit trigger-label verification and cleanup contract that exists only to support the old `labeled` trigger path.
- Remove the `remove-labels` safe output and any agent instructions that tell the workflow to clean up `verify-openspec` at the end of a run.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: change the verify workflow from a filtered `labeled` trigger plus agent-managed cleanup to a PR-only `label_command` trigger with built-in label removal

## Impact

- `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`
- Generated `openspec-verify-label` workflow artifacts under `.github/workflows/`
- `openspec/specs/ci-aw-openspec-verification/spec.md`
- Any workflow prompt text or deterministic setup that currently depends on explicit trigger-label verification or `remove-labels`
