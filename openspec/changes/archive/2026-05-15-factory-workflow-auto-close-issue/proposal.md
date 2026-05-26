## Why

The `change-factory` and `reproducer-factory` workflows create pull requests that are intended to remain linked to the source issue without closing it. Today, both workflows instruct agents to use `Related to #N` and explicitly avoid closing keywords, but gh-aw's default `create-pull-request` behavior still auto-inserts `Fixes #N` unless `auto-close-issue: false` is set. This creates a mismatch between workflow intent and runtime behavior and can cause GitHub to auto-close issues when these PRs are merged.

## What Changes

- Configure `safe-outputs.create-pull-request.auto-close-issue: false` in the `change-factory` workflow.
- Configure `safe-outputs.create-pull-request.auto-close-issue: false` in the `reproducer-factory` workflow.
- Clarify the workflow requirements so linked PRs created by these workflows must not gain auto-closing issue references from safe-output defaults.
- Regenerate compiled workflow artifacts after updating the repository-authored workflow sources.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `ci-change-factory-issue-intake`: enforce that the linked proposal PR remains non-closing by disabling gh-aw's automatic issue-closing text injection.
- `ci-reproducer-factory-issue-intake`: enforce that the linked reproducer PR remains non-closing by disabling gh-aw's automatic issue-closing text injection.

## Impact

- Affected authored workflow sources under `.github/workflows-src/` and generated workflow artifacts under `.github/workflows/`.
- Affected behavior of gh-aw `create-pull-request` safe outputs for `change-factory` and `reproducer-factory`.
- No provider Go code, Terraform schema, generated clients, or acceptance-test behavior changes.
