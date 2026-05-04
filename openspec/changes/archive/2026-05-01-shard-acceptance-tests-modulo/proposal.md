## Why

After `speed-up-dashboard-acceptance-tests` (PR #2539) the matrix snapshot entry runs in ~25 minutes, with `internal/kibana/dashboard` as the undisputed critical path at ~11.5 minutes of wall clock. The test splits and `-p 6` change reduce total time from ~27 min to ~25 min, but further gains within a single runner are bounded: `kibana/dashboard` sets an ~11.5 min floor that no amount of in-process parallelism can move.

Package-level timing data from the PR #2539 CI run (9.4.0-SNAPSHOT, `-p 6`):

| Package group | Packages | Sum | Est. wall-clock (with `-p 6`) |
|---|---|---|---|
| `internal/kibana/...` | 28 | 38.6 m | **11.6 m** (critical path) |
| `internal/elasticsearch/...` + other | 48 | 25.0 m | 4.2 m |
| `internal/fleet/...` | 12 | 15.6 m | 6.1 m |

Running all 88 packages serially through `-p 6` slots forces the ~4 m elasticsearch work and ~6 m fleet work to queue behind the 11.6 m kibana work. Splitting the suite across two concurrent runners with independent stacks breaks this coupling and delivers roughly a **halving of wall-clock**: max(11.6, 6.1) = **~12 min** (down from ~25 min), using the same runner hardware already provisioned per-version by the matrix strategy.

There are also two consistently flaky packages — `internal/fleet/integration` and `internal/fleet/integration_policy` — that fail on first run and require gotestsum reruns in nearly every matrix version. Today both land on the same runner; separating them onto different shards gives each its own rerun budget and prevents one package's retries from delaying the other's shard completion.

## What Changes

- Add two new Makefile variables — `ACCTEST_TOTAL_SHARDS` (default `1`) and `ACCTEST_SHARD_INDEX` (default `0`) — that control which subset of packages the `testacc` target runs.
- When `ACCTEST_TOTAL_SHARDS=1` (the default), behaviour is identical to today: all packages run, no filtering applied. Contributors and existing CI jobs that do not set these variables are unaffected.
- When `ACCTEST_TOTAL_SHARDS > 1`, the `testacc` recipe narrows `--packages` to those packages whose position in the sorted `go list ./...` output satisfies `index % ACCTEST_TOTAL_SHARDS == ACCTEST_SHARD_INDEX`. Sorting is alphabetical, which is stable across runs.
- Update the GitHub Actions acceptance-test workflow to add a `shard` dimension (`[0, 1]`) to the version matrix. Each (version, shard) combination starts its own runner, spins up its own isolated Elastic stack, and runs `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`.
- Update the `makefile-workflows` capability spec to require sharding support in `testacc`.

This change is purely test-infrastructure; no resource implementation or provider spec is touched.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `makefile-workflows`: the `testacc` target gains shard-selection inputs (`ACCTEST_TOTAL_SHARDS`, `ACCTEST_SHARD_INDEX`) that are no-ops at their defaults and enable modulo-based package sharding when set.

## Impact

- `Makefile`: two new `?=`-defaulted variables; `testacc` recipe gains a conditional `--packages` filter using `go list ./... | awk`.
- `.github/workflows/`: the matrix acceptance-test workflow gains a `shard: [0, 1]` matrix dimension, doubling the number of runners per stack version.
- CI wall-clock for the snapshot matrix entry: expected reduction from ~25 min to ~12 min (~52% improvement).
- Runner cost: doubles the runner-minutes per version (2 runners × ~12 min vs 1 runner × ~25 min = ~24 min vs ~25 min), near cost-neutral.
- Fleet flakiness isolation: `internal/fleet/integration` (alphabetical index 42, shard 0) and `internal/fleet/integration_policy` (index 43, shard 1) land on separate runners. Each can rerun independently without delaying the other shard.
- Coverage guarantee: the union of all shards is exactly `go list ./...`. No new package can be silently excluded — any package added to the module is automatically assigned to a shard by the modulo rule.
