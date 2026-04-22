## Why

The prepare-release workflow creates a PR that bumps the `VERSION` in `Makefile`, but this PR should not itself appear as a changelog entry. Without the `no-changelog` label, the changelog generator may attempt to generate a changelog entry for the release-preparation PR itself, which is incorrect.

## What Changes

- The `prep-release.yml` workflow will apply the `no-changelog` label to the release PR when creating it, and will also apply it to any existing (reused) PR.

## Capabilities

### New Capabilities

<!-- None — this is a behaviour fix to an existing workflow -->

### Modified Capabilities

- `ci-release-pr-preparation`: The release PR creation and reuse steps must apply the `no-changelog` label to the PR.

## Impact

- `.github/workflows/prep-release.yml` — `gh pr create` call gains `--label no-changelog`; a new step is added to label an existing PR when it is reused.
- `openspec/specs/ci-release-pr-preparation/spec.md` — new requirement and scenario for the `no-changelog` label.
- No API, dependency, or other system changes required.
