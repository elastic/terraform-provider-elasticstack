## Why

Dashboard acceptance tests compete with unrelated packages for the same Elastic Stack during full-suite CI runs, contributing to uneven shard durations and limited merge-queue timeout headroom. Isolating that workload will make its cost visible and let it run without consuming the general suites' Stack capacity; any timing improvement must still be measured.

## What Changes

- Run the measured dashboard hotspots (`internal/kibana/dashboard` and `internal/kibana/dashboard/panel/lensxy`) against their own Stack in full-suite CI for Stack 9.4 and later, including snapshots of supported versions.
- Keep those hotspot packages out of the two general full-suite shards when the dedicated dashboard suite runs, preserving complete package coverage without duplicate execution.
- Preserve older Stack coverage, targeted pull-request selection, local full-suite defaults, and existing failure and cleanup behavior.
- Document suite selection and the timing measurements needed to evaluate isolation.
- Non-goals: provider performance optimization, dashboard test rewrites, new dependencies, changes to targeted PR selection, or changes to Terraform resource behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-build-lint-test`: Isolate dashboard full-suite acceptance work on supported Stack versions while retaining two general shards and existing PR selection behavior.

## Impact

Provider CI matrix preparation and execution, full-suite package selection, deterministic CI tests, and contributor testing documentation. CI uses one additional Stack job for each configured version supporting the dashboard suite; no provider APIs, credentials, or dependencies change. Tracks issue #5469 and follows the timeout-headroom change in #5468.
