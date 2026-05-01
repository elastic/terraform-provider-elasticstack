## Context

`speed-up-dashboard-acceptance-tests` (PR #2539) reduced the snapshot matrix entry from ~27 min to ~25 min by adding explicit cross-package parallelism (`-p 6`) and splitting monolithic test functions. The remaining bottleneck is `internal/kibana/dashboard` (~11.6 min), which sets a floor that no amount of in-process parallelism can move.

Package-level timing data from PR #2539 (9.4.0-SNAPSHOT, `-p 6`, 88 first-run packages):

| Group | Pkgs | Sum | Est. wall-clock (`-p 6`) |
|---|---|---|---|
| `internal/kibana/...` | 28 | 38.6 m | **11.6 m** (critical path) |
| `internal/elasticsearch/...` + other | 48 | 25.0 m | 4.2 m |
| `internal/fleet/...` | 12 | 15.6 m | 6.1 m |

The ~4 m elasticsearch and ~6 m fleet work queues behind the 11.6 m kibana work on the same runner. Splitting the suite across two concurrent runners with independent stacks breaks this coupling.

Two consistently flaky packages — `internal/fleet/integration` and `internal/fleet/integration_policy` — fail on first run and require gotestsum reruns in nearly every matrix version. Today both run on the same runner; separating them allows each to rerun independently.

## Goals / Non-Goals

**Goals**:
- Halve the snapshot matrix wall-clock from ~25 min to ~12 min by running two parallel shards per version.
- Separate the two known-flaky fleet packages onto different shards.
- Guarantee that all packages are covered without hardcoding package names.
- Keep `make testacc` (no shard variables) fully backwards-compatible for local use.

**Non-Goals**:
- Further splitting beyond two shards (a third shard does not reduce the critical path).
- Duration-aware shard balancing (alphabetical modulo is sufficient; the imbalance is bounded by the longest single package).
- Changing per-package parallelism (`-parallel`) or rerun settings.

## Decisions

### 1. Modulo bucketing over sorted `go list ./...` output, not named shards

Two approaches were compared using actual timing data:

- **Named shards** (`internal/kibana/...` | `internal/elasticsearch/...` | `internal/fleet/...`): deterministic but requires updating shard filter rules when new product namespaces are added. Produces imbalanced shards (11.6 m / 4.2 m / 6.1 m at N=3). Does not separate the two flaky fleet packages — they would both land in the fleet shard.
- **Modulo on alphabetical `go list ./...`**: zero-maintenance. Coverage is mathematically guaranteed (every index 0–N appears in exactly one shard, and all N shards together cover the full list). Alphabetically, `internal/fleet/integration` (index 42) and `internal/fleet/integration_policy` (index 43) are adjacent, so modulo-2 puts them in different shards — a useful side effect.

Modulo is chosen because it requires no ongoing maintenance and improves fleet flakiness isolation compared to named shards.

### 2. N=2 shards to start

Measured wall-clock estimates from PR #2539 CI data:

| N shards | Shard 0 | Shard 1 | Shard 2 | Critical path |
|---|---|---|---|---|
| 1 (today) | 21.4 m | — | — | 21.4 m |
| **2 (modulo-2)** | **11.6 m** | **6.1 m** | — | **11.6 m** |
| 3 (modulo-3) | 11.6 m | 6.1 m | 3.2 m | 11.6 m |

`kibana/dashboard` (11.6 m) sets the floor regardless. A third shard moves work that is already finishing before the kibana shard completes; it adds runner cost without reducing wall-clock. Two shards is the minimum effective split; three can be revisited if the non-critical shard later exceeds ~10 min due to new packages.

### 3. Makefile implementation

```makefile
ACCTEST_TOTAL_SHARDS ?= 1
ACCTEST_SHARD_INDEX  ?= 0

.PHONY: testacc
testacc:
	TF_ACC=1 go tool gotestsum \
	  --format testname \
	  --rerun-fails=$(RERUN_FAILS) \
	  --rerun-fails-max-failures=$(RERUN_FAILS_MAX_FAILURES) \
	  --packages="$(shell go list ./... \
	    | sort \
	    | awk '(NR-1) % $(ACCTEST_TOTAL_SHARDS) == $(ACCTEST_SHARD_INDEX)')" \
	  -- -p $(ACCTEST_PACKAGE_PARALLELISM) \
	     -v \
	     -count $(ACCTEST_COUNT) \
	     -parallel $(ACCTEST_PARALLELISM) \
	     $(TESTARGS) \
	     -timeout $(ACCTEST_TIMEOUT)
```

When `ACCTEST_TOTAL_SHARDS=1` (default), `awk 'NR % 1 == 0'` matches every line — identical to `./...` behaviour today. `$(shell go list ./...)` is evaluated at recipe expansion time, so it reflects the actual module state at invocation (~1 s overhead on warm cache).

The existing `TEST ?= ./...` variable is deliberately not repurposed for sharding; `TEST` is already used by contributors to run individual packages, and overloading it would create confusing semantics.

### 4. GitHub Actions workflow change

A `shard: [0, 1]` dimension is added to the existing version matrix so each `(version, shard)` pair runs as an independent job. The `make testacc` step becomes:

```yaml
make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}
```

Stack startup, fleet setup, and teardown are unchanged and run per-shard (each shard has its own isolated Elastic stack). Changes MUST be authored in the `workflows-src/` template system and compiled via `make workflow-generate`.

### 5. Coverage guarantee

For `ACCTEST_TOTAL_SHARDS=N`, the set of packages in shard `k` is `{ p_i : i % N == k }`. The union over k=0…N−1 equals `{ p_i : 0 ≤ i < len }` = all packages. CI requires all shards to pass; branch protection enforces this, so a missing shard job leaves the workflow non-green.

## Risks / Trade-offs

- **Runner cost**: 2 runners × ~12 min ≈ 24 runner-min vs 1 runner × ~25 min ≈ 25 runner-min. Near cost-neutral.
- **Index drift**: adding or removing packages shifts modulo assignments by one. Not a correctness concern (coverage is always complete); worst-case impact is a temporarily imbalanced run.
- **`go list` in recipe**: ~1 s on warm cache, ~5 s cold. Acceptable overhead for an acceptance-test target.
- **Stack startup cost paid twice**: each shard starts its own ES + Kibana + Fleet stack (~2 min). This is already inside the 25-min baseline and not on the critical path.

## Migration Plan

1. Land Makefile changes (`ACCTEST_TOTAL_SHARDS`, `ACCTEST_SHARD_INDEX`) with default no-op values; verify local backwards compatibility.
2. Update the `makefile-workflows` spec and run `make check-lint`.
3. Update the workflow source template to add the shard matrix dimension; regenerate and verify with `make check-workflows`.
4. Open PR; capture actual shard wall-clocks from the first CI run.
