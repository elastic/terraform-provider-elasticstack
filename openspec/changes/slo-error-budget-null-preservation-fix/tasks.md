## 1. Fix `PopulateFromAPI`

- [x] 1.1 In `internal/kibana/dashboard/panel/sloerrorbudget/model.go`, replace the four
      `if (prior == nil || typeutils.IsKnown(existing.Title)) && apiConfig.Title != nil { ... }`-shaped
      conditionals (for `Title`, `Description`, `HideTitle`, `HideBorder`) with:
      - Unconditional adoption from the API pointer (e.g.
        `existing.Title = types.StringPointerValue(apiConfig.Title)`, and similarly for the other
        three), which is correct for creation/import (`prior == nil`).
      - A single `panelkit.NullPreservePresentationFromPrior(...)` call, gated on
        `prior != nil && prior.SloErrorBudgetConfig != nil`, passing
        `prior.SloErrorBudgetConfig.Title` / `.Description` / `.HideTitle` / `.HideBorder` and
        `&existing.Title` / `&existing.Description` / `&existing.HideTitle` / `&existing.HideBorder`.
- [x] 1.2 Confirm the `panelkit` import in this file already covers `NullPreservePresentationFromPrior`
      (it is in the same package as `BuildPresentationConfig`, already imported) — no new import
      needed.
- [x] 1.3 Leave `SloID`, `SloInstanceID`, and `Drilldowns` handling unchanged.

## 2. Unit test coverage

- [x] 2.1 In `internal/kibana/dashboard/panel/sloerrorbudget/model_test.go`, add a test asserting
      that on a same-type update (non-nil `prior` with `prior.SloErrorBudgetConfig` set and
      `title`/`description`/`hide_title`/`hide_border` known), `PopulateFromAPI` adopts the API's
      values for all four fields — the scenario currently broken. Follow the existing helper style
      (`sebWithSloTitle`, `sebWithSloDescription`, `sebWithHideTitle`, `sebWithHideBorder` already
      exist in this file) and existing test naming
      (`Test_populateSloErrorBudgetFromAPI_<behavior>`).
- [x] 2.2 Add a test asserting null-preservation: when `prior.SloErrorBudgetConfig`'s
      `title`/`description`/`hide_title`/`hide_border` are null (practitioner never configured them)
      and the API returns concrete values for them, `PopulateFromAPI` keeps them null in the
      resulting state — mirroring `Test_populateSloErrorBudgetFromAPI_sloInstanceID_nullPreservation`
      for the presentation fields.
- [x] 2.3 Confirm `Test_populateSloErrorBudgetFromAPI_import_populatesAll` (creation/import,
      `prior == nil`) still passes unchanged — it already exercises unconditional adoption and must
      keep doing so.
- [x] 2.4 Run `go test ./internal/kibana/dashboard/panel/sloerrorbudget/...` and confirm all tests
      pass.

## 3. Acceptance test coverage

- [x] 3.1 Add acceptance-test fixtures under
      `internal/kibana/dashboard/panel/sloerrorbudget/testdata/` (new test case, e.g.
      `TestAccResourceDashboardSloErrorBudgetWithDisplayFields`) whose `main.tf` sets `title`,
      `description`, `hide_title`, and `hide_border` inside `slo_error_budget_config` alongside the
      required `slo_id`.
- [x] 3.2 Add the corresponding `TestAcc...` function in
      `internal/kibana/dashboard/panel/sloerrorbudget/acc_test.go`, following the existing
      `TestAccResourceDashboardSloErrorBudgetMinimal` / `...WithDrilldowns` structure: apply, assert
      state matches configuration, then run a subsequent plan/apply step and assert no diff (no
      "Provider produced inconsistent result after apply").
- [x] 3.3 Acceptance test written; not executed here — Elastic stack was not available on the
      worktree ports from `.env` (`ELASTICSEARCH_ENDPOINTS` / `KIBANA_ENDPOINT` unreachable).
      Run with `source .env && TF_ACC=1 go test -v -run TestAccResourceDashboardSloErrorBudgetWithDisplayFields ./internal/kibana/dashboard/panel/sloerrorbudget/...` when the stack is up.

## 4. Spec and validation

- [ ] 4.1 Confirm the `openspec/changes/slo-error-budget-null-preservation-fix/specs/kibana-dashboard/spec.md`
      delta in this change directory accurately reflects the implemented behavior once 1-3 are done
      (update wording if the implementation deviates from what is currently drafted).
- [ ] 4.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate
      slo-error-budget-null-preservation-fix --type change` and resolve any reported issues.
