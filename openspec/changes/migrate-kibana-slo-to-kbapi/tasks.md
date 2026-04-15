## 1. `kibanaoapi` SLO helpers

- [ ] 1.1 Add `internal/clients/kibanaoapi/slo.go` implementing Get SLO (`GetSloOpWithResponse`), Create (`CreateSloOpWithResponse`), Update (`UpdateSloOpWithResponse`), Delete (`DeleteSloOpWithResponse`), and Find (`FindSlosOpWithResponse`) using `SpaceAwarePathRequestEditor(spaceID)` and the same error/status patterns as `alerting_rule.go` (including `kbn-xsrf` if required by generated signatures).
- [ ] 1.2 Ensure get-by-id returns `(nil, nil)` diagnostics on HTTP 404 where the existing `internal/clients/kibana/slo.go` contract expects “not found”.
- [ ] 1.3 Add unit tests for the helper layer (mocked `*http.Client` or response fixtures) covering at least one success path and one non-2xx path per operation where practical.

## 2. Types and model migration (`generated/slo` → `generated/kbapi`)

- [ ] 2.1 Replace `generated/slo` imports in `internal/models/slo.go` with `generated/kbapi` equivalents (`SLOsSloWithSummaryResponse`, `SLOsSloWithSummaryResponse_Indicator`, `SLOsTimeWindow`, `SLOsObjective`, `SLOsBudgetingMethod`, `SLOsSettings`, etc.).
- [ ] 2.2 Update `internal/kibana/slo/models.go`, all `models_*_indicator.go` files, tests, and `group_by` / settings mapping to use kbapi types end-to-end; fix indicator union construction to use kbapi `From*` / `As*` helpers or explicit switches matching current JSON.
- [ ] 2.3 Reimplement `group_by` request/response transforms using `SLOsGroupBy` while preserving REQ-023 behavior driven by existing version flags (`supportsGroupByList` / stack version from resource).

## 3. Client wiring and legacy removal

- [ ] 3.1 Refactor `internal/clients/kibana/slo.go` to resolve `*kibanaoapi.Client` via `KibanaScopedClient.GetKibanaOapiClient()` (or equivalent factory path) and call the new helpers; reimplement `responseIndicatorToCreateSloRequestIndicator` (and any update-body builder) against kbapi indicator unions.
- [ ] 3.2 Update `internal/kibana/slo/create.go`, `read.go`, `update.go`, and `delete.go` to use the refactored client entry points only (no `GetSloClient` / `SetSloAuthContext`).
- [ ] 3.3 Remove `GetSloClient`, `SetSloAuthContext`, `buildSloClient`, and `APIClient`/factory fields that exist solely for `generated/slo`; update `internal/clients/provider_client_factory_test.go` and related tests to cover `kibanaoapi` paths instead.
- [ ] 3.4 Confirm no remaining `generated/slo` imports in the provider; remove or narrow Makefile `generate-slo-client` / `generated/slo` targets per repo policy if nothing else consumes that client.

## 4. Verification

- [ ] 4.1 Run `make build` and fix compile or staticcheck issues.
- [ ] 4.2 Run unit tests for `internal/clients/kibana`, `internal/clients/kibanaoapi`, and `internal/kibana/slo`.
- [ ] 4.3 Run targeted acceptance tests for `elasticstack_kibana_slo` against a configured stack, including cases that exercise `group_by`, multiple `group_by`, `prevent_initial_backfill`, and `data_view_id` where the stack version supports them.
