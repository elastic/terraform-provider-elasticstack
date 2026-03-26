## Why

The `verify-openspec` label currently acts as a trigger but remains on the pull request after the workflow finishes. That leaves the PR in a misleading state and makes the label harder to use as an explicit "run verification now" signal for later updates.

## What Changes

- Update the `ci-aw-openspec-verification` requirements so the workflow removes the triggering `verify-openspec` label before the run fully completes.
- Require label removal for every completed run outcome, including approval, comment-only completion, and noop-style early exits after the workflow starts.
- Capture any permission or workflow-contract changes needed so the automation can mutate pull request labels safely.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: clear the `verify-openspec` label when a verification run completes, regardless of outcome

## Impact

- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- Pull request label lifecycle for OpenSpec verification runs
