## Why

Targeted pull-request acceptance runs currently create both fixed shards even
when the selected package set only has work for one. Empty shards still consume
runner setup and disk-cleanup time, while the package-selection work is
repeated independently by every version-and-shard job.

## What Changes

- Prepare targeted acceptance-package selection once before pull-request
  acceptance jobs fan out, and use its result to create only populated shards.
- Keep the existing package-count threshold: a targeted selection below 30
  packages uses one shard; a selection of 30 or more uses two.
- Make the targeted selection tool produce a machine-readable shard plan by
  default, replacing its package-list output.
- Retain successful no-op acceptance jobs for a centrally verified
  zero-package pull request so the provider gate's current result contract
  remains unchanged.
- Preserve fixed two-shard, full-suite acceptance behavior for push,
  merge-queue, and manually dispatched runs.
- Exclude changes to the acceptance-test version matrix and its coverage
  policy.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `selective-acceptance-tests`: Centralize targeted package planning and
  expose dynamically populated shards as a machine-readable plan.
- `ci-build-lint-test`: Build the pull-request acceptance matrix from the
  centrally prepared shard plan while preserving full-suite behavior for
  non-pull-request events and the provider gate contract.

## Impact

- Affects the targeted acceptance-test selector, its local Make targets, and
  their tests.
- Affects the Provider CI matrix-preparation and acceptance-test workflow
  steps, plus workflow and OpenSpec validation tests.
- Reduces unnecessary GitHub Actions runner work for targeted pull requests
  without reducing version coverage or changing authoritative full-suite runs.
