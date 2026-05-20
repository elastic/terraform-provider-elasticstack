## Why

Change-factory and reproducer-factory workflows currently create pull requests as drafts (the default behaviour of the `create-pull-request` safe output). Draft pull requests cannot be reviewed until they are manually converted to ready-for-review, which defeats the purpose of the factory pipeline producing immediately actionable proposals and reproductions.

## What Changes

- Add `draft: false` to the `safe-outputs.create-pull-request` configuration in the `change-factory` issue-intake workflow source so that proposal PRs are immediately reviewable.
- Add `draft: false` to the `safe-outputs.create-pull-request` configuration in the `reproducer-factory` issue-intake workflow source so that reproduction PRs are immediately reviewable.
- Regenerate the compiled lock files for both workflows.
- Update the `ci-change-factory-issue-intake` and `ci-reproducer-factory-issue-intake` specs to require non-draft PR creation.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-change-factory-issue-intake`: proposal pull requests SHALL be created as non-draft (immediately ready for review)
- `ci-reproducer-factory-issue-intake`: reproduction pull requests SHALL be created as non-draft (immediately ready for review)

## Impact

- `.github/workflows/change-factory-issue.md` (add `draft: false` to `safe-outputs.create-pull-request`)
- `.github/workflows/reproducer-factory-issue.md` (add `draft: false` to `safe-outputs.create-pull-request`)
- Compiled lock files for both workflows must be regenerated
- `ci-change-factory-issue-intake` spec gains a new requirement scenario
- `ci-reproducer-factory-issue-intake` spec gains a new requirement scenario
