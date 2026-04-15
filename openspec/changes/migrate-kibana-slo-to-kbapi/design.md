## Context

The `elasticstack_kibana_slo` resource today issues SLO HTTP calls through a dedicated OpenAPI client under `generated/slo`, wired from `internal/clients/kibana_scoped_client.go` and `internal/clients/api_client.go` via `GetSloClient()` / `buildSloClient()` and `SetSloAuthContext()` (context-based credential injection tailored to that client). Domain types on the wire (`SloWithSummaryResponse`, `CreateSloRequest`, `UpdateSloRequest`, indicator unions, `GroupBy`, `Settings`, etc.) are imported from `generated/slo` in `internal/models/slo.go` and throughout `internal/kibana/slo`.

The repository already generates a unified Kibana client in `generated/kbapi` (`kibana.gen.go`) with SLO operations and models such as `SLOsSloWithSummaryResponse`, `SLOsCreateSloRequest`, `SLOsUpdateSloRequest`, `FindSlosOp`, `GetSloOp`, `CreateSloOp`, `UpdateSloOp`, `DeleteSloOp`, plus `SLOsGroupBy` (string vs string-array union). Other Kibana entities use thin wrappers in `internal/clients/kibanaoapi` around `*kbapi.ClientWithResponses`, `SpaceAwarePathRequestEditor`, and consistent HTTP status handling.

## Goals / Non-Goals

**Goals:**

- Add `internal/clients/kibanaoapi` helpers for SLO **find** (paginated search via `FindSlosOpWithResponse` or equivalent), **get**, **create**, **update**, and **delete**, mirroring patterns in `alerting_rule.go` / `connector.go` (typed responses, `SpaceAwarePathRequestEditor`, `kbn-xsrf` header where required by the generated client, clear diagnostics on unknown status codes).
- Retarget `internal/clients/kibana/slo.go` to obtain `*kibanaoapi.Client` from `KibanaScopedClient.GetKibanaOapiClient()` (or the same factory path other PF resources use) and delegate to those helpers instead of `slo.SloAPI` + `SetSloAuthContext`.
- Replace all `github.com/elastic/terraform-provider-elasticstack/generated/slo` imports in `internal/models/slo.go` and `internal/kibana/slo/**` with `generated/kbapi` types, including discriminated indicator unions and `SLOsGroupBy` encode/decode aligned with REQ-023.
- Remove the `generated/slo` client surface from the provider wiring (`GetSloClient`, `SetSloAuthContext`, `buildSloClient`, `APIClient.slo` field, factory tests that only exist for the legacy client) once nothing references it.
- Preserve **exact** user-visible behavior for version gates and wire formats: REQ-014 (`group_by` ≥ 8.10.0), REQ-015 (multi `group_by` ≥ 8.14.0), REQ-016 (`prevent_initial_backfill` ≥ 8.15.0), REQ-017 (`data_view_id` ≥ 8.15.0), and REQ-023 (`group_by` string vs JSON array by version). Version resolution and scoped `kibana_connection` behavior stay as today (REQ-007, scoped-client scenarios).

**Non-Goals:**

- Terraform schema changes, attribute renames, or new SLO indicator types.
- Regenerating or editing the standalone SLO OpenAPI spec under `generated/slo` unless required purely to delete the tree; primary work is **consumer** migration to `kbapi`.
- Broader refactors of unrelated Kibana resources or the Plugin SDK.

## Decisions

1. **Single transport: `kbapi.ClientWithResponses` via `kibanaoapi.Client`** — SLO CRUD and find use the same HTTP stack, TLS, debug transport, and auth as other `kibanaoapi` consumers (`transport` round-tripper in `kibanaoapi.NewClient`). **Rationale:** Eliminates parallel auth (`SetSloAuthContext`) and duplicate base URLs. **Alternative considered:** keep `generated/slo` for SLO only — rejected per consolidation goal.

2. **Rename-aligned types, not parallel DTOs** — `models.Slo` and `tfModel` indicator helpers use `kbapi` structs (`SLOsSloWithSummaryResponse`, `SLOsSloWithSummaryResponse_Indicator`, etc.) directly. **Rationale:** One set of generated types; compile-time coverage. **Alternative considered:** adapter structs mapping slo→kbapi — rejected as extra maintenance.

3. **Indicator conversion layer** — Keep a focused conversion from read/response indicator unions to create/update indicator unions (today `responseIndicatorToCreateSloRequestIndicator` in `internal/clients/kibana/slo.go`); reimplement against `SLOsSloWithSummaryResponse_Indicator` → `SLOsCreateSloRequest_Indicator` / `SLOsUpdateSloRequest_Indicator` using kbapi `As*` / `From*` helpers where available, with explicit switches for parity. **Rationale:** API asymmetry between GET and PUT/POST bodies remains; behavior must match current implementation.

4. **`group_by` mapping** — Reimplement `transformGroupBy` / `transformGroupByFromResponse` using `SLOsGroupBy` (union of string vs `[]string`) with the same version-driven rules as today (`supportsGroupByList` / stack version flags passed from the resource). **Rationale:** REQ-023 is normative; only the backing type changes.

5. **Find helper scope** — Implement `FindSlos` (or similar) in `kibanaoapi` with parameters sufficient to locate an SLO by id within a space (e.g. KQL / filter supported by the API). The current resource read path may continue to use **get-by-id** only; find is for consistency with the requested helper surface and for tests or future diagnostics. **Rationale:** Matches user-requested API surface without forcing a read-path change unless beneficial.

6. **Legacy client removal order** — Migrate call sites and types first, then delete `GetSloClient` / `buildSloClient` / tests, then Makefile `generate-slo-client` / `generated/slo` references in a final cleanup sub-task if no other package imports `generated/slo`. **Rationale:** Keeps intermediate builds green.

## Risks / Trade-offs

- **[Risk] Union / JSON shape drift between `generated/slo` and `kbapi`** — Field names or nested types may differ. **Mitigation:** Compare structs side-by-side for each indicator; add unit tests with golden JSON fixtures from existing tests; run acceptance tests for `elasticstack_kibana_slo`.

- **[Risk] Subtle plan noise** — Nil vs empty slices on `tags` / `group_by` after decode. **Mitigation:** Reuse existing normalization in `internal/kibana/slo/models.go` read paths; do not change schema defaults.

- **[Risk] Auth regression** — SLO previously used context-injected credentials. **Mitigation:** Rely on `kibanaoapi` transport already used by other resources for the same `kibana_connection`; extend factory tests if a gap appears.

- **[Risk] Version gate bypass** — If refactor accidentally skips pre-flight checks. **Mitigation:** Leave version checks in `internal/kibana/slo` create/update/read; add a regression test that mocks low version and expects diagnostics before HTTP (where feasible).

## Migration Plan

1. Implement `internal/clients/kibanaoapi/slo.go` (name may vary) with get/create/update/delete/find, using `SpaceAwarePathRequestEditor(spaceID)` and response classification (200 vs 404 vs error bodies).
2. Replace `generated/slo` types in `internal/models/slo.go` and all `internal/kibana/slo` files and tests; fix compile errors indicator-by-indicator.
3. Switch `internal/clients/kibana/slo.go` to call helpers with `*kibanaoapi.Client`; update `responseIndicatorToCreateSloRequestIndicator` and `group_by` transforms for kbapi types.
4. Update resource create/read/update/delete to resolve `GetKibanaOapiClient()` instead of any SLO-specific client accessor.
5. Remove dead `generated/slo` wiring from `api_client.go`, `provider_client_factory.go`, `kibana_scoped_client.go`, and associated tests; adjust Makefile/spec references if the SLO generator becomes unused.
6. Run `make build`, unit tests for `internal/clients/kibana` and `internal/kibana/slo`, and targeted acceptance `TestAccResourceKibanaSlo` (or current SLO acc test name).

## Open Questions

- Whether `FindSlos` needs to be invoked from production code in the first iteration, or only implemented and unit-tested for parity (read currently uses get-by-id).
- Exact `FindSlosOp` query parameter for filtering by SLO id — confirm against `kbapi` parameter names during implementation.
- Whether any **other** package still imports `generated/slo` after SLO migration; if yes, either migrate those callers or defer deleting `generate-slo-client` until a follow-up change.
