# Code Review: dashboard-description-null-normalization

Branch: `dashboard-description-null-normalization` vs `main`
Files changed: `models.go`, `models_dashboard_description_test.go`, `acc_test.go`, two testdata fixtures, `tasks.md`.

## Correct (with evidence)

- **Core fix is logically correct across create/read/update.** `internal/kibana/dashboard/models.go:56-61` introduces an intent-preserving check. Verified the data flow: after Create/Update, `runKibanaWrite` (`internal/entitycore/kibana_resource_envelope.go:407`) calls `r.readFunc(...)` with `written.Model`, which is the **plan** model (`create.go` returns `planModel`). So `m.Description` during the post-write read reflects plan intent (null when omitted, `""` when explicit). During a pure Read, `base_envelope.go:84` decodes the model from `req.State`, so `m.Description` is prior state. Both intent sources are handled correctly.
  - Omitted description (plan null) + API `""` → `null` (the bug fix). ✔
  - Explicit `description = ""` (plan `""`) + API `""` → `""` preserved. ✔
  - Non-empty + any prior → API value. ✔
  - API `nil` + prior null → `null` (8.x/9.4 case). ✔
- **Write path is unchanged and consistent.** `dashboardToAPICreateRequest` / `dashboardToAPIUpdateRequest` (`models.go:163-166`, `222-225`) send `description` only when `IsKnown` (non-null): null → field omitted from request body; `""` → `&""` sent. This matches the read-side intent preservation.
- **Schema is compatible.** `description` is `Optional` (not Computed) in `schema.go:105-107`, so omitted → plan null, and there is no ModifyPlan that would force a value. No schema/migration changes needed.
- **Unit tests cover the four key matrix cases** (`models_dashboard_description_test.go`) and pass:
  ```
  go test ./internal/kibana/dashboard/ -run 'TestDashboardModel_populateFromAPI_descriptionNormalization' -v
  ok  github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard
  ```
  Full package unit suite also passes (`go test ./internal/kibana/dashboard/`).
- **Build / vet / gofmt clean.** `go build ./...`, `go vet ./internal/kibana/dashboard/`, and `gofmt -l` on the changed files all produced no output.
- **Acceptance test structure follows existing repo conventions.** `ConfigDirectory: acctest.NamedTestCaseDirectory(...)` with `testdata/<TestName>/{omitted,empty}/main.tf` matches the pattern in `acc_dashboard_root_filters_test.go`. `TestCheckNoResourceAttr("description")` is the correct assertion for a null Optional attribute (absent from state).
- **No staged files** (`git diff --cached` empty); only untracked `.pi-subagents/` runtime dir.

## Blocker
None.

## Notes (non-blocking)

- **Note (Low–Medium) — acceptance `empty` case may be fragile on pre-9.5 stacks.** `TestAccResourceDashboardDescriptionNormalization` is gated by `minDashboardAPISupport` = `9.4.0-SNAPSHOT` (`dashboardacctest.go:26`). The `empty` step asserts `description = ""` in state. If a 9.4 dashboard API returns `nil` (rather than echoing `""`) for an explicitly-set empty-string description, the read-path else-branch (`models.go:60`) sets state to `StringNull()`, which would fail `TestCheckResourceAttr("description", "")`. The fix targets the 9.5 `""`-echo behavior; whether 9.4 echoes `""` for an explicit `description = ""` should be confirmed against the 9.4 CI image. If 9.4 does not echo `""`, consider gating the `empty` sub-case to 9.5+ (e.g. a version-conditional check) or confirming 9.4 round-trips `""`.
- **Note (Low) — known limitation, pre-existing, not a regression.** The condition `apiDescription.ValueString() == ""` (`models.go:57`) conflates API-`nil` and API-`""`. Consequently an explicit `description = ""` cannot be round-tripped when the API returns `nil` (8.x / 9.4 omitted-field behavior): the else-branch stores `StringNull()`. This is identical to the pre-change behavior (`StringishPointerValue(nil)` → null) and is acceptable given the fix's scope, but is an inherent limitation of intent preservation against a non-echoing API.
- **Note (Low) — test helper duplicates generated anonymous struct.** `newDashboardAPIResponse` (`models_dashboard_description_test.go`) re-declares the `JSON200` anonymous struct shape inline (including a `//nolint:revive` for the `Id` field, which is appropriate). This is brittle to upstream generated-struct changes but fails at compile time, so it is safe and acceptable. No change needed.

## Recommended actions
1. Confirm the 9.4 dashboard API echoes `""` for an explicit `description = ""` before relying on the `empty` acceptance sub-case on the 9.4 CI lane; gate to 9.5+ if it does not.

No code edits were made (review-only).