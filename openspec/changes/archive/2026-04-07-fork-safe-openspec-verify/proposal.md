## Why

The `openspec-verify-label` workflow needs a repository-authored `pull_request_target` trigger, deterministic trigger-label cleanup, and deterministic archive/push gating for fork pull requests. The current implementation does not maintain a separate `api-only` verification mode, so the change artifacts should describe the simpler model the repository actually implements.

## What Changes

- Change the authored `openspec-verify-label` workflow trigger from `label_command` back to an explicitly filtered labeled pull request workflow using `pull_request_target`.
- Restore deterministic trigger-label verification and add a deterministic script step that removes only `verify-openspec`, instead of relying on `label_command` activation behavior or agent safe outputs.
- Add deterministic pre-activation outputs for active-change selection, review disposition, and archive/push eligibility, and expose those results and reason strings to the agent prompt.
- Prevent archive or push behavior for fork pull requests before the agent starts, while preserving the normal review decision rules for runs that otherwise pass verification.
- Declare the needed deterministic pre-activation token scopes through workflow frontmatter `on.permissions`.
- Preserve the existing review/bootstrap flow for verification rather than introducing a separate `api-only` verification mode.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: change trigger and cleanup behavior to use `pull_request_target`, deterministic label cleanup, explicit pre-activation permissions, and deterministic archive/push eligibility gating

## Impact

- `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`
- `.github/workflows-src/openspec-verify-label/scripts/verify_label.inline.js`
- `.github/workflows-src/openspec-verify-label/scripts/remove_trigger_label.inline.js`
- `.github/workflows-src/lib/verify-label.js`
- `.github/workflows-src/lib/openspec-verify-label.test.mjs`
- Generated `openspec-verify-label` workflow artifacts under `.github/workflows/`
- `openspec/specs/ci-aw-openspec-verification/spec.md`
- Workflow prompt text and deterministic pre-activation outputs related to trigger verification, cleanup, review disposition, and archive/push eligibility
