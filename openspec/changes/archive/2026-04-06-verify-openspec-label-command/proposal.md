## Why

The `openspec-verify-label` workflow currently models `verify-openspec` as a filtered `pull_request:labeled` event and then asks the agent to remove that label through `remove-labels` safe output cleanup. GitHub Agentic Workflows already provide `label_command` for this pattern, so the workflow can use the built-in one-shot label trigger and drop redundant cleanup behavior from both frontmatter and prompt instructions. The compiled `.lock.yml` may still expand `label_command` into lower-level trigger wiring, so this change should describe the authored workflow contract without requiring a duplicate repository-authored label verification gate just because the compiler output is more explicit.

## What Changes

- Change the authored `openspec-verify-label` workflow trigger from filtered `pull_request` `labeled` events to `label_command` for the `verify-openspec` label.
- Restrict the new trigger to pull requests only so the workflow keeps its current PR-only activation behavior.
- Remove the explicit repository-authored trigger-label verification and cleanup contract that exists only to support the old `labeled` trigger path.
- Allow the compiled `.lock.yml` to normalize `label_command` into compiler-managed lower-level event wiring without reintroducing a custom `verify_label` gate.
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
- Requirement language that distinguishes authored `label_command` semantics from compiler-expanded lockfile wiring
