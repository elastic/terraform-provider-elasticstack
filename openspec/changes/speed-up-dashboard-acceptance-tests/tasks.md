## 1. Make `testacc` package-parallelism explicit

- [x] 1.1 Add `ACCTEST_PACKAGE_PARALLELISM ?= 6` near the existing `ACCTEST_PARALLELISM` block in `Makefile`. (Initial design said `8`; reduced to `6` after CI fleet HTTP 400 failures at 80 peak in-flight tests.)
- [x] 1.2 Thread the variable into the `testacc` recipe as `-p $(ACCTEST_PACKAGE_PARALLELISM)`, placed alongside `-parallel $(ACCTEST_PARALLELISM)`.
- [x] 1.3 Update the `makefile-workflows` capability spec to require explicit, contributor-overridable package parallelism for `testacc`, with a `MODIFIED` requirement and accompanying scenarios.
- [x] 1.4 Verify locally that `make testacc TESTARGS='-run ^TestAccResourceDashboardEmptyDashboard$$' ACCTEST_PACKAGE_PARALLELISM=2` honors the override (smoke test only; no behavioral assertion needed beyond observing the flag in the printed command).

## 2. Split the longest monolithic dashboard tests

- [x] 2.1 Replace `TestAccResourceDashboardXYChart` with one `TestAccResourceDashboardXYChart_<facet>` function per existing `ConfigDirectory` (`basic`, `axis`, `decorations`, `filters`, `fitting`, `legend_outside`, `legend_inside`, `layers`, `layers_reference`), each carrying its own `Check` block from the original Steps slice. Pair `layers_reference` with the existing ImportState step inside its own function. Keep `TestAccResourceDashboardXYChartMinimalConfig` as-is.
- [x] 2.2 Replace `TestAccResourceDashboardPanels` with one `TestAccResourceDashboardPanels_<facet>` function per existing `ConfigDirectory` (`basic`, `multiple_panels`, `with_sections`, `multi_sections_single_panel_each`, `multi_sections_multi_panels_each`, `panels_and_sections`).
- [x] 2.3 Apply the same split to `TestAccResourceDashboardESQLControl` (the 232 s test in `acc_esql_control_panels_test.go`).
- [x] 2.4 If wall-clock measurements after 2.1–2.3 still show a single dashboard test ≥ 120 s, apply the same split to `TestAccResourceDashboardLensDashboardAppByValue`, `TestAccResourceDashboardLensDashboardAppByValueTypedMetric`, and `TestAccResourceDashboardSloBurnRateDisplayOptions`; otherwise mark this task done and stop.

## 3. Delete duplicate dashboard tests

- [x] 3.1 Append a `PlanOnly: true` re-apply step (with a `TestCheckNoResourceAttr` for `panels.0.slo_burn_rate_config.slo_instance_id`) to `TestAccResourceDashboardSloBurnRate`'s Steps slice; delete `TestAccResourceDashboardSloBurnRateSloInstanceIDNullPreservation` from `acc_slo_burn_rate_panels_test.go`.
- [x] 3.2 Append the equivalent `PlanOnly: true` re-apply step to `TestAccResourceDashboardSloErrorBudget` (the base test in `acc_slo_error_budget_panels_test.go`); delete `TestAccResourceDashboardSloErrorBudgetSloInstanceIDNullPreservation`.

## 4. Audit `acc_lens_dashboard_app_panels_test.go`

- [x] 4.1 Tabulate every `(test_function, ConfigDirectory, step_intent)` triple in `acc_lens_dashboard_app_panels_test.go`. Record the table in the PR description.
- [x] 4.2 For each `ConfigDirectory` referenced by more than one function, classify each occurrence as one of: first creator, ImportState, PlanOnly assertion, or update-from-prior. Note any cases where two functions create-then-destroy the same `ConfigDirectory`. **Result**: `"basic"` appears as a `NamedTestCaseDirectory` arg in three functions, but each resolves to a distinct physical path (keyed by test function name via `config.TestNameDirectory()`). No shared physical paths — no true duplicates.
- [x] 4.3 Fold ImportState-only and PlanOnly-only duplicate functions into their first creator as additional Steps; delete the now-empty wrappers. Leave update-from-prior cases untouched. **Result**: Zero foldable duplicates found; no changes needed.
- [x] 4.4 If the audit also surfaces a monolithic ParallelTest in this file with ≥ 4 independent `ConfigDirectory` steps and a wall-clock ≥ 120 s in the most recent CI run, apply the §2 split to it. Otherwise note the file is already acceptably structured. **Result**: `TestAccResourceDashboardLensDashboardAppPlan` has 12 plan-error steps but all are `PlanOnly: true` with `ExpectError` — they fail before any API call, running in milliseconds. Not a split candidate. The file is already acceptably structured. Also fixed a latent bug from task 2.4: testdata directories `TestAccResourceDashboardLensDashboardAppByValue/` and `TestAccResourceDashboardLensDashboardAppByValueTypedMetric/` were renamed to `TestAccResourceDashboardLensDashboardAppByValue_basic/` and `TestAccResourceDashboardLensDashboardAppByValueTypedMetric_basic/` to match the `_basic`-suffixed function names introduced in 2.4.

## 5. Verify and document

- [x] 5.1 Run `make testacc TESTARGS='-run ^TestAccResourceDashboard'` against a local 9.4 stack (`testacc-vs-docker`) and confirm the suite passes with the new shape.
- [x] 5.2 Run `openspec validate --specs` and `openspec validate --change speed-up-dashboard-acceptance-tests` (when available) to confirm the spec delta is well-formed.
- [x] 5.3 Open the PR with the latest snapshot matrix wall-clock recorded in the description (before/after, captured from the failing-CI run linked in `proposal.md` and the new CI run on this branch). Confirm peak `--rerun-fails` usage did not increase materially.
