## Why

When a maintainer applies the `verify-openspec` label to a pull request that does not meet the eligibility criteria (for example, no OpenSpec change files, multiple change IDs, or an unsupported file status), the agent job is silently skipped. The PR author receives no feedback and does not know what to fix.

## What Changes

- Add a deterministic `comment_ineligible` step to the `pre_activation` inject-steps block in `openspec-verify-label.md`, immediately after `classify_and_select`, that posts a PR comment when `label_verified == 'true'` and `selection_status == 'ineligible'`.
- Add a new script `.github/scripts/workflows/openspec-verify/comment-ineligible.js` that reads the `selection_reason` output from the prior step and calls the GitHub API to post a comment on the triggering pull request.
- The comment body includes the specific ineligibility reason and a "How to fix" section linking to the OpenSpec docs and explaining the required file path pattern.
- Add unit tests for the new script, consistent with existing test coverage in `.github/scripts/workflows/`.
- Recompile `.github/workflows/openspec-verify-label.lock.yml` to include the new step.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: add a deterministic ineligible-PR comment step to `pre_activation` that posts PR feedback when the `verify-openspec` label was applied to an ineligible pull request

## Impact

- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- `.github/scripts/workflows/openspec-verify/comment-ineligible.js` (new file)
- `.github/scripts/workflows/openspec-verify/comment-ineligible.test.mjs` (new file)
