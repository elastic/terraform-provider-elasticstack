## Context

Full-suite Provider CI uses two alphabetical package shards per configured Stack version. Dashboard and Lens XY acceptance packages land together and share their job's Stack with unrelated packages. Existing successful runs show a representative four-markdown-panel dashboard test taking 19–22 minutes, but do not identify whether provider CPU work, API latency, or contention is responsible. This change isolates that workload so its elapsed cost is independently visible.

The reproducible baseline, hotspot package totals, and slow-test investigation are recorded in [`measurement.md`](./measurement.md). The baseline shows that isolating the entire dashboard subtree would consolidate work currently split across both general Stacks, so the selected suite is limited to the root dashboard and Lens XY hotspot packages.

## Goals / Non-Goals

**Goals:**

- Give the two measured dashboard hotspot packages their own Stack in full-suite runs for Stack 9.4 and later.
- Preserve complete package coverage, two general shards, older Stack execution, local defaults, targeted PR selection, and required-check semantics.
- Make general and dashboard suite durations distinguishable in CI.

**Non-Goals:**

- Prove or fix the root cause of slow individual tests, tune test concurrency, or rewrite dashboard tests.
- Change provider behavior, add dependencies, change Stack version discovery, or expand targeted PR shard planning.

## Decisions

Use the existing acceptance job's setup, readiness checks, test execution, diagnostics, and teardown for every suite. Prepare explicit per-version matrix entries before fan-out: general shard 0, general shard 1, and, for non-PR versions with numeric major/minor at least 9.4, one unsharded dashboard entry. The dashboard entry is a separate GitHub Actions job with its own containers, not a second invocation on a general shard's Stack. A suite field distinguishes it in the job name. Dynamic entries add suite scheduling only; runner, Fleet image, and synthetics flags continue to come from the existing version-flag preparation.

Full-suite package selection has a single deterministic owner shared by CI and the Makefile. Baseline logs identify `internal/kibana/dashboard` and `internal/kibana/dashboard/panel/lensxy` as the two dominant packages on shard 0, so the dashboard suite contains exactly those import paths. Other dashboard descendants remain in the general package pool. On supported versions, general suites exclude the two hotspots before dividing the remaining sorted packages between two shards. Older versions retain the existing all-package split. Local `make testacc` defaults to the existing all-package suite; explicit suite selection is opt-in. Invalid suite or shard input fails before test execution.

Keep targeted PR selection and its two-shard package outputs unchanged. PR matrix entries consume those prepared packages without extra suite filtering or dashboard duplication. Full-suite matrix creation does not run diff selection. Preparation failures remain blocking, and the existing Provider Gate observes the aggregate result of the same acceptance job ID, covering the dedicated dashboard jobs without a second gate path. Existing snapshot failure policy and unconditional teardown apply to every suite.

Use deterministic package-plan and matrix tests for partition coverage, non-overlap, path boundaries, numeric version boundaries including snapshots, invalid input, PR zero-package behavior, and unchanged targeted routing. Workflow contract tests and `actionlint` validate the prepared entries and their execution wiring. Build validation remains required by the repository. Contributor testing documentation explains suite selection and compares dashboard/general elapsed times and summed job minutes across equivalent full-suite runs. No local implementation result will claim a performance improvement without those runs.

## Risks / Trade-offs

- Additional Stack jobs increase runner consumption → add them only for full-suite versions supporting dashboards and compare summed job minutes alongside critical-path duration.
- An incomplete partition could silently drop tests → verify that general shards plus the dashboard suite cover the original package set exactly once and that non-hotspot descendants remain general.
- Version parsing could misroute snapshots or two-digit minors → reuse existing numeric version handling and test the 9.4 snapshot boundary and a two-digit minor.
- Dashboard tests still share their dedicated Stack with each other → preserve current concurrency initially; isolation is an experiment, not an asserted fix for individual test latency.
- Extra job visibility could accidentally weaken required checks → retain the existing acceptance job ID and gate aggregation, including failure, cancellation, and snapshot handling.
- Rollback returns packages to the original general suites → revert matrix scheduling and suite selection together so dashboard coverage is never omitted.
