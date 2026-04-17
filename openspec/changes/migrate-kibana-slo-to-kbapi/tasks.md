## 1. `kibanaoapi` SLO helpers

- [x] 1.1 Add `internal/clients/kibanaoapi/slo.go` implementing Get SLO (`GetSloOpWithResponse`), Create (`CreateSloOpWithResponse`), Update (`UpdateSloOpWithResponse`), Delete (`DeleteSloOpWithResponse`), and Find (`FindSlosOpWithResponse`) using the same error/status patterns as `alerting_rule.go`. Note: `SpaceAwarePathRequestEditor` is NOT used for SLO operations because the kbapi-generated SLO endpoints already embed the spaceId in the URL path (e.g. `/s/%s/api/observability/slos`); adding the editor would double the prefix and cause 404s.
- [x] 1.2 Ensure get-by-id returns `(nil, nil)` diagnostics on HTTP 404 where the existing `internal/clients/kibana/slo.go` contract expects "not found".
- [x] 1.3 Add unit tests for the helper layer (mocked `*http.Client` or response fixtures) covering at least one success path and one non-2xx path per operation where practical.

## 2. Types and model migration (`generated/slo` → `generated/kbapi`)

- [x] 2.1 Replace `generated/slo` imports in `internal/models/slo.go` with `generated/kbapi` equivalents (`SLOsSloWithSummaryResponse`, `SLOsSloWithSummaryResponse_Indicator`, `SLOsTimeWindow`, `SLOsObjective`, `SLOsBudgetingMethod`, `SLOsSettings`, etc.).
- [x] 2.2 Update `internal/kibana/slo/models.go`, all `models_*_indicator.go` files, tests, and `group_by` / settings mapping to use kbapi types end-to-end; fix indicator union construction to use kbapi `From*` / `As*` helpers or explicit switches matching current JSON. Key fixes: dispatch on aggregation value (not error-check) for metric/timeslice indicators; KQL indicator tries object variant (`*1`) before string variant (`*0`); float32ToFloat64 helper avoids precision artifacts in Terraform state.
- [x] 2.3 Reimplement `group_by` request/response transforms using `SLOsGroupBy` while preserving REQ-023 behavior driven by existing version flags (`supportsGroupByList` / stack version from resource).

## 3. Client wiring and legacy removal

- [x] 3.1 Refactor `internal/clients/kibana/slo.go` to resolve `*kibanaoapi.Client` via `KibanaScopedClient.GetKibanaOapiClient()` (or equivalent factory path) and call the new helpers; reimplement `responseIndicatorToCreateSloRequestIndicator` (and any update-body builder) against kbapi indicator unions.
- [x] 3.2 Update `internal/kibana/slo/create.go`, `read.go`, `update.go`, and `delete.go` to use the refactored client entry points only (no `GetSloClient` / `SetSloAuthContext`).
- [x] 3.3 Remove `GetSloClient`, `SetSloAuthContext`, `buildSloClient`, and `APIClient`/factory fields that exist solely for `generated/slo`; update `internal/clients/provider_client_factory_test.go` and related tests to cover `kibanaoapi` paths instead.
- [x] 3.4 Confirm no remaining `generated/slo` imports in the provider; remove or narrow Makefile `generate-slo-client` / `generated/slo` targets per repo policy if nothing else consumes that client.

## 4. Verification

- [x] 4.1 Run `make build` and fix compile or staticcheck issues.
- [x] 4.2 Run unit tests for `internal/clients/kibana`, `internal/clients/kibanaoapi`, and `internal/kibana/slo`.
- [x] 4.3 Run targeted acceptance tests for `elasticstack_kibana_slo` against a configured stack, including cases that exercise `group_by`, multiple `group_by`, `prevent_initial_backfill`, and `data_view_id` where the stack version supports them. All 29 acceptance test matrices pass in CI.
