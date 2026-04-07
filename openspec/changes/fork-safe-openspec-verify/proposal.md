## Why

The `openspec-verify-label` workflow should be runnable for pull requests from forks, but the current `label_command` trigger and archive/push flow assume same-repository pull request mechanics. Moving the workflow to a trusted pull request trigger requires the workflow contract to distinguish between pull requests that may be reviewed only and pull requests that may also archive and push changes.

## What Changes

- Change the authored `openspec-verify-label` workflow trigger from `label_command` back to an explicitly filtered labeled pull request workflow, using `pull_request_target` so maintainers can trigger verification for fork pull requests from the base repository context.
- Restore deterministic trigger-label verification and add a deterministic script step that removes only `verify-openspec`, instead of relying on `label_command` activation behavior or agent safe outputs.
- Add a deterministic pre-activation capability check that classifies whether the triggering pull request is allowed to archive and push, and expose that result and its reason to the agent prompt.
- Prevent archive or push behavior for fork pull requests before the agent starts, while preserving the normal review decision rules for runs that otherwise pass verification.
- Preserve the existing active-change selection rules, relevance review, and verification structure for pull requests that remain eligible to archive and push.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: change trigger and cleanup behavior to support fork pull requests under `pull_request_target`, and add deterministic archive/push eligibility gating

## Impact

- `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`
- `.github/workflows-src/openspec-verify-label/scripts/verify_label.inline.js`
- `.github/workflows-src/openspec-verify-label/scripts/remove_trigger_label.inline.js`
- `.github/workflows-src/lib/verify-label.js`
- `.github/workflows-src/lib/openspec-verify-label.test.mjs`
- Generated `openspec-verify-label` workflow artifacts under `.github/workflows/`
- `openspec/specs/ci-aw-openspec-verification/spec.md`
- Workflow prompt text and deterministic pre-activation outputs related to trigger verification, cleanup, and archive/push eligibility
