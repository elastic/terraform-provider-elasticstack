## Context

The matrix acceptance test workflow runs ~21 entries in parallel; the longest entry (`9.4.0-SNAPSHOT`) gates merge time. Within that entry, `make testacc` invokes:

```
TF_ACC=1 go tool gotestsum --format testname \
  --rerun-fails=$(RERUN_FAILS) --rerun-fails-max-failures=$(RERUN_FAILS_MAX_FAILURES) \
  --packages="./..." -- -v -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) \
  $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)
```

Two go-test scheduling knobs apply:

- `-parallel N` (currently `10` via `ACCTEST_PARALLELISM`): cap for `t.Parallel()` tests **inside one package**.
- `-p N` (currently unset → `GOMAXPROCS` → `4` on the standard 4‑vCPU public-runner): how many test **binaries** run concurrently.

`go test ./...` schedules packages roughly in source order, so on the snapshot matrix entry `internal/kibana/dashboard` (the package that carries 89.5 minutes of serial test CPU and the 561 s longest test) sat in the queue for ~10 minutes while four earlier-alphabetical packages occupied the four `-p` slots. The dashboard package only began at 05:35 of a test phase that started at 05:25 and ended at 05:48.

Inside the dashboard package, every `TestAcc*` function uses `resource.ParallelTest` (which calls `t.Parallel()`), so all 55 dashboard acceptance tests are eligible to run concurrently up to `-parallel 10`. The 89.5 min ÷ 10 floor is `~537 s`; the actual wall is `~690 s`. The remaining gap is the single longest test (`TestAccResourceDashboardXYChart`, 561 s, 10 sequential steps) — Amdahl's law made visible.

## Goals / Non-Goals

**Goals:**
- Cut snapshot-matrix wall-clock by addressing both head-of-line scheduling (`-p`) and the longest single test (split monoliths).
- Eliminate test bodies that re-create dashboards a sibling test has already created+destroyed.
- Keep total test coverage equivalent: every `ConfigDirectory` exercised today must still be exercised after the change.
- Make the package-parallelism choice explicit in the Makefile so it is reproducible across runner configurations.

**Non-Goals:**
- Changing the `kibana-dashboard` resource implementation or its spec.
- Reducing Elastic Stack matrix coverage or skipping versions.
- Caching docker images / pre-warming Kibana / changing the stack startup phase.
- Restructuring tests in any package other than `internal/kibana/dashboard`.
- Tuning `-parallel` (in-package) above 10 — kept as a follow-up once the new test shape lands.

## Decisions

### 1. Use `-p 8` rather than `-p 4` (default), `-p 12`, or "as high as possible"

Three options were considered:

- `-p 4` (status quo): preserves the head-of-line stall; we already know its wall-clock cost.
- `-p 8`: enough to ensure `internal/kibana/dashboard` is in the first parallel batch alongside the other long packages (`security_detection_rule` 990 s serial, `ingest` 732 s, `index/template` 467 s, `agentpolicy` 293 s). Peak in-flight tests rise from 40 → 80, sharing one Kibana + ES instance on 4 vCPUs.
- `-p 12` or higher: head-of-line is solved at `-p 8`; adding more slots beyond the long-package count only increases Kibana contention without scheduling benefit.

We choose `-p 8` because it is the smallest value that resolves the observed scheduling problem.

The value is exposed as a new Make variable `ACCTEST_PACKAGE_PARALLELISM ?= 8`, parallel to `ACCTEST_PARALLELISM` (which controls in-package `-parallel`). Contributors and CI may override it without editing the recipe.

### 2. Do not change `GOMAXPROCS`

`GOMAXPROCS` defaults to `runtime.NumCPU()`, which already matches the runner. Setting it explicitly would risk drift if Actions later changes the runner spec, and a lower value would starve the Go scheduler. The relevant decision is `-p`, not `GOMAXPROCS`. The proposal records this so future contributors don't conflate the two.

### 3. Splitting strategy: one function per `ConfigDirectory`, not per `Check`

The longest monolithic tests sequence multiple full create/replace cycles against the same resource address with **independent** config directories (e.g. `axis`, `decorations`, `filters`, …). Each cycle pays a full Kibana create + delete and is logically a different scenario. Splitting at the `ConfigDirectory` boundary keeps each new test small and parallelisable, and preserves a 1:1 mapping with the existing `testdata/` fixtures so reviewers can audit coverage by looking at directory listings.

We deliberately do **not** split at the `Check` level — that would force testdata duplication or shared fixtures and add review burden without wall-clock benefit.

ImportState steps for split tests stay paired with their producing test: each split function ends with its own ImportState step against the same `ConfigDirectory` so import is exercised against the exact shape that test produces. This matches the current pattern in `TestAccResourceDashboardXYChart`'s final step (which imports against `layers_reference`).

### 4. Initial scope: only the heaviest monoliths

Splitting every multi-step `ParallelTest` is unnecessary; the wall-clock floor is set by the single longest test, so we only need to bring the longest down to the level of the next-longest until diminishing returns kick in. The plan is iterative:

- Round 1 (this change): `XYChart` (561 s), `Panels` (242 s), `ESQLControl` (232 s) — three obvious wins, each with cleanly independent steps.
- Round 2 (this change, if scope allows): `LensDashboardAppByValue*` (~190 s each), `SloBurnRateDisplayOptions` (151 s), and any redundancy uncovered during the lens audit (§5).
- Future: tests below ~120 s — addressed when a new bottleneck emerges.

The acceptance criterion is "max single dashboard test ≤ 120 s on the snapshot matrix entry", not "no test has more than one step".

### 5. Duplicate-test deletions

Two known duplicates:

- `TestAccResourceDashboardSloBurnRateSloInstanceIDNullPreservation` (127 s) re-applies `required_only` then `PlanOnly: true`. Its assertion is `slo_instance_id` stays null after read-back. **Action:** delete the function, add a third step (`PlanOnly: true`, same `ConfigDirectory`) plus a `TestCheckNoResourceAttr` to `TestAccResourceDashboardSloBurnRate`'s existing Steps slice.
- `TestAccResourceDashboardSloErrorBudgetSloInstanceIDNullPreservation` (143 s): same pattern, same fix against `TestAccResourceDashboardSloErrorBudget*`.

The `acc_lens_dashboard_app_panels_test.go` audit (22 `ConfigDirectory` references across multiple functions) will surface any further duplicates of this shape; the action for each will be the same: append-as-step, delete the standalone function.

### 6. Audit method for `acc_lens_dashboard_app_panels_test.go`

Mechanical, not exploratory:

1. Tabulate every `(test_function, config_directory, step_intent)` triple in the file.
2. Group by `config_directory`. Any directory referenced by more than one test function is a candidate.
3. Within each candidate group, classify each occurrence as one of:
   - First creator (`Steps[0]` apply)
   - ImportState (uses `ResourceName` + `ImportState: true`)
   - PlanOnly assertion / null-preservation (uses `PlanOnly: true`)
   - Update-from-prior (relies on a prior step's state)
4. Fold (2)+(3) into the corresponding (1) test as additional Steps; delete the now-empty wrappers. Update steps that depend on prior state cannot be folded across function boundaries and stay where they are.

This audit MAY surface zero foldable duplicates — that is an acceptable outcome and is not a failure of the change. The goal is correctness, not a quota.

## Risks / Trade-offs

- **Kibana saturation at `-p 8`** — Mitigation: keep `-parallel 10` (no increase in per-package concurrency); rely on the existing `--rerun-fails=5 --rerun-fails-max-failures=20` budget to absorb transient failures during rollout; revert to `-p 4` is a one-line change.
- **More test functions to maintain** — Mitigation: splits sit next to their parents in the same `acc_*_panels_test.go` file; helper-extracting common preamble (`PreCheck`, `SkipFunc`, `ProtoV6ProviderFactories`, `ConfigVariables`) is a follow-up if duplication becomes painful, but is out of scope for this change to keep diffs reviewable.
- **`testdata/` directory churn** — None expected. Existing directories continue to be referenced from the new function names.
- **Merging the duplicate-deletion and split-monolith work in one change** — Reviewers will see a large diff in `acc_xy_panels_test.go`, `acc_panels_test.go`, `acc_esql_control_panels_test.go`, `acc_slo_burn_rate_panels_test.go`, `acc_slo_error_budget_panels_test.go`, and `acc_lens_dashboard_app_panels_test.go`. We accept that because the wins compound (one CI run validates both buckets).

## Migration Plan

1. Land the Makefile change (`ACCTEST_PACKAGE_PARALLELISM`, default `8`) and the `makefile-workflows` spec update first; this can be exercised on its own and produces an immediate ~10‑minute saving on the snapshot matrix entry.
2. Split `TestAccResourceDashboardXYChart` and the next two heaviest tests in a single follow-up commit; verify locally with `make testacc-vs-docker TESTARGS='-run ^TestAccResourceDashboardXY...'`.
3. Apply the duplicate-deletion fixes for `*SloInstanceIDNullPreservation`.
4. Run the lens audit and apply its findings.
5. Validate on a real PR run; capture the new snapshot matrix entry wall-clock for the change description.

## Open Questions

- **Is `-p 8` enough or should we go to `-p 6`?** The 6‑slot configuration would cover the four long packages (`dashboard`, `security_detection_rule`, `ingest`, `index/template`) plus two more, with peak in-flight = 60. Worth measuring after the split lands; out of scope for this change.
- **Should `-parallel` rise from 10 once the longest dashboard test is below ~120 s?** Possibly to `16`; deferred to a follow-up because tuning that knob is independent of this change's correctness story.
