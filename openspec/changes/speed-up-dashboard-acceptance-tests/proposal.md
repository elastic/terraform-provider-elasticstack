## Why

The matrix acceptance job for the snapshot stack version (currently the matrix critical path) takes ~27 minutes, of which ~22 minutes is the test phase. Profiling the [`Matrix Acceptance Test (9.4.0-SNAPSHOT)` job in run 25035397726](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/25035397726/job/73325928002) shows two compounding root causes that both live in our control:

1. **The `internal/kibana/dashboard` package waits ~10 minutes before its first test runs.** `make testacc` invokes `go test` without an explicit `-p`, so the test scheduler defaults to `GOMAXPROCS` (4 on the standard 4‑vCPU runner) and `kibana/dashboard` queues behind earlier alphabetical packages. Once it does start, it runs for ~11.5 minutes and gates wall-clock for the whole job.
2. **Inside the dashboard package, a few monolithic `ParallelTest` functions dominate the floor.** `TestAccResourceDashboardXYChart` alone is 561 s with 10 sequential steps, each fully replacing the dashboard with an independent config (basic, axis, decorations, filters, fitting, two legend variants, two layer variants, ImportState). Several other tests follow the same shape (`TestAccResourceDashboardPanels` 242 s, `TestAccResourceDashboardESQLControl` 232 s, `LensDashboardAppByValue*` ~190 s, `SloBurnRateDisplayOptions` 151 s).

There are also two pure duplicate-work test functions whose body could be a single extra step in an existing test: `TestAccResourceDashboardSloBurnRateSloInstanceIDNullPreservation` (127 s) and `TestAccResourceDashboardSloErrorBudgetSloInstanceIDNullPreservation` (143 s) re-create the same `required_only` dashboard that their corresponding base tests have already created and torn down.

`acc_lens_dashboard_app_panels_test.go` carries 22 `ConfigDirectory` references across multiple test functions and is the most likely place for the same redundancy patterns to recur.

## What Changes

- Set an explicit package parallelism (`-p 8`) for the `testacc` Make target so the dashboard package starts in the first parallel batch on 4‑vCPU GitHub-hosted runners instead of queueing behind alphabetical predecessors. Document the value via a `?=`-style Makefile variable so contributors can override it locally without editing the recipe.
- Split the longest monolithic `resource.ParallelTest` functions in `internal/kibana/dashboard/` into per-facet `TestAcc*` functions so each test exercises a single configuration shape rather than 4–10 sequential, independent shapes. Initial scope:
  - `TestAccResourceDashboardXYChart` → 1 function per facet (basic, axis, decorations, filters, fitting, legend_outside, legend_inside, layers, layers_reference, ImportState).
  - `TestAccResourceDashboardPanels` → 1 function per facet (basic, multiple_panels, with_sections, multi_sections_*, panels_and_sections).
  - `TestAccResourceDashboardESQLControl` → 1 function per facet equivalent to its sequential steps.
- Delete the duplicated `*SloInstanceIDNullPreservation` test functions and instead append a `PlanOnly: true` re-apply step to their corresponding base tests.
- Audit `acc_lens_dashboard_app_panels_test.go` and apply the same two patterns (split monoliths, fold duplicates) wherever they appear.
- Update the `makefile-workflows` capability spec to require that the `testacc` target sets package parallelism explicitly rather than relying on `GOMAXPROCS`.

This change is purely test- and tooling-side; the `kibana-dashboard` resource implementation and its spec are not touched. Test coverage is preserved — every `ConfigDirectory` exercised today will still be exercised after the split.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `makefile-workflows`: the `testacc` target gains an explicit, contributor-overridable package parallelism input distinct from the existing `ACCTEST_PARALLELISM` (which controls `-parallel`, the in-package `t.Parallel()` cap).

## Impact

- `Makefile`: new `ACCTEST_PACKAGE_PARALLELISM` variable (defaulting to `8`), threaded into `go tool gotestsum -- ... -p $(ACCTEST_PACKAGE_PARALLELISM)`.
- `internal/kibana/dashboard/`: net more test functions, fewer steps per function. No new `testdata/` config directories are added; existing ones are re-targeted from sequential steps to parallel functions.
- CI wall-clock for the snapshot matrix entry: expected reduction of ~10 minutes from the `-p` change alone, with a further ~5–8 minutes once the longest dashboard tests are split. Other matrix entries skip dashboard tests via `minDashboardAPISupport` and are not affected.
- Risk: peak in-flight tests rise from `4 × 10 = 40` to `8 × 10 = 80`, increasing load on the colocated single Kibana + Elasticsearch instance. Mitigated by leaving `-parallel` at 10 and watching the existing `--rerun-fails` budget. No spec contract change for any resource.
