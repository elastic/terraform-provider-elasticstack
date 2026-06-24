## 1. Prep and discovery

- [x] 1.1 Verify that `OsqueryCreatePacks`, `OsqueryGetPacksDetails`, `OsqueryUpdatePacks`, `OsqueryDeletePacks` are present in `generated/kbapi/kibana.gen.go`; confirm their request/response type signatures (especially `queries`, `shards`, and the create response wrapper shape)
- [x] 1.2 Bump OAS ref in `generated/kbapi/Makefile` to a version that includes `schedule_type`/`rrule_schedule`/per-query `interval`/`timeout`/`splay`; run `make -C generated/kbapi all`; confirm these fields appear in the regenerated client. **Fallback (confirmed):** kbapi bump blocked by `transform_schema.go`; v1 scoped to packs CRUD with **no scheduling attributes** in schema or API mapping; design.md updated accordingly.
- [x] 1.3 Confirm and record minimum Kibana versions for base packs CRUD and for the full scheduling model in design.md Decision 11 (base **8.5.0**, scheduling **9.5.0** deferred). Code implementation of `GetVersionRequirements` is task 3.2.
- [x] 1.4 Confirm that Kibana generates a UUID for `pack_id` when omitted on Create; if not, escalate to Required and update design.md Decision 2
- [x] 1.5 Confirm whether the Create response wraps the pack in a `data` field (like `osquery_saved_query`) or is a direct object; update design.md Decision 14 if different
- [x] 1.6 Confirm the exact `shards` wire format on Create/Update (array `[{key, value}]` or already a map?); confirm read (`GetPacksDetails`) returns `map[string]float32`; update design.md Decision 8 if different

## 2. kibanaoapi client helper

- [x] 2.1 Create `internal/clients/kibanaoapi/osquery_pack.go` with thin wrappers `CreateOsqueryPack`, `GetOsqueryPack`, `UpdateOsqueryPack`, `DeleteOsqueryPack` — each passing `kibanautil.SpaceAwarePathRequestEditor(spaceID)` and using `HandleGetTypedResponse` / `HandleMutateTypedResponse` / `diagutil.HandleStatusResponse` consistently with existing kibanaoapi wrappers (e.g. `synthetics_private_location.go`)
- [x] 2.2 Map HTTP 404 on Get to a nil/sentinel result (resource removed from state); map HTTP 404 on Delete to a no-op success
- [x] 2.3 Map non-2xx responses to provider diagnostics consistently with other kibanaoapi helpers
- [x] 2.4 Normalize `shards` on read: convert `map[string]float32` from `GetPacksDetails` to `map[string]float64` for state; handle create-response array `[{key,value}]` quirk (re-read via GET or convert inline)

## 3. Resource skeleton and model

- [x] 3.1 Create `internal/kibana/osquery_pack/` directory mirroring `internal/kibana/osquery_saved_query/`
- [x] 3.2 Implement `models.go` with `osqueryPackModel` implementing `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`, `GetVersionRequirements` (v1: single entry `8.5.0` base CRUD floor)
- [x] 3.3 Implement `queryModel` nested struct covering `query`, `platform`, `version`, `snapshot`, `removed`, `saved_query_id`, `ecs_mapping` (pinned kbapi fields only); plus `toAPIType()` and `fromAPIType()`
- [x] 3.4 Implement `ecsMappingModel` (reuse or mirror from `osquery_saved_query`) covering `field`, `value`, `values`
- [x] 3.5 Implement `populateFromAPI` accepting unwrapped detail payload (from `response.JSON200.Data`); map `saved_object_id` → `pack_id`, shards normalization, and per-query field mapping
- [x] 3.6 Add unit tests for `populateFromAPI` / converters: shards map normalization, ECS mapping three shapes, platform comma-string ↔ set, version round-trip

## 4. Resource schema

- [ ] 4.1 Implement `getSchema` covering: `id` (Computed), `pack_id` (Computed), `space_id` (Optional+Computed, default `"default"`, RequiresReplace), `kibana_connection` (Optional), `name` (Required string), `description` (Optional string), `enabled` (Optional bool), `policy_ids` (Optional list(string)), `shards` (Optional MapAttribute of numbers), `queries` (Required MapNestedAttribute)
- [ ] 4.2 Implement the per-query nested schema inside `queries` map: `query` (Required string), `platform` (Optional SetAttribute of strings with allowed-values validator), `version` (Optional string), `snapshot` (Optional+Computed bool), `removed` (Optional+Computed bool), `saved_query_id` (Optional string), `ecs_mapping` (Optional MapNestedAttribute)
- [ ] 4.3 Add the per-element `ConfigValidator` on each `ecs_mapping` enforcing exactly-one-of `field`/`value`/`values`
- [ ] 4.4 Add `RequiresReplace` plan modifier on `space_id` only; add `UseStateForUnknown` on Optional+Computed fields (`pack_id` is Computed-only, no RequiresReplace)

## 5. Resource CRUD and import

- [ ] 5.1 Implement `create.go` calling `POST /api/osquery/packs` (space-aware), unwrap `response.JSON200.Data` for initial fields, then read-after-write GET (unwrap `.Data`) to populate full state and enforce prebuilt guard — Create POST response omits `read_only`
- [ ] 5.2 Implement `read.go` calling `GET /api/osquery/packs/{id}` (space-aware) using `pack_id`, unwrap `response.JSON200.Data`; on HTTP 404, remove from state; error if detail payload has `read_only=true`
- [ ] 5.3 Implement `update.go` calling `PUT /api/osquery/packs/{id}` (space-aware, full body); unwrap update response `.Data` or read-after-write GET; repopulate state
- [ ] 5.4 Implement `delete.go` calling `DELETE /api/osquery/packs/{id}` (space-aware); treat HTTP 404 as success
- [ ] 5.5 Implement `ImportState` accepting composite `"<space_id>/<pack_id>"` form (pack_id is saved_object_id UUID); Import refresh uses GET detail and prebuilt guard
- [ ] 5.6 Register `osqueryPack.NewResource()` in `provider/plugin_framework.go`

## 6. Data source

- [ ] 6.1 Implement `internal/kibana/osquery_pack/datasource.go` with schema: `pack_id` (Required), `space_id` (Optional, default `"default"`), `kibana_connection` (Optional), plus Computed fields matching v1 resource (`name`, `description`, `enabled`, `policy_ids`, `shards`, `queries`, `read_only` as Computed bool)
- [ ] 6.2 Implement Read calling `GET /api/osquery/packs/{id}` — same kibanaoapi wrapper, unwrap `response.JSON200.Data`; on HTTP 404, return error diagnostic
- [ ] 6.3 Do NOT error on `read_only=true` in the data source
- [ ] 6.4 Register the data source in `provider/plugin_framework.go`

## 7. Acceptance tests

- [ ] 7.1 Add `acc_test.go` covering full resource lifecycle: create with all v1 fields (including `ecs_mapping` with all three shapes, `policy_ids`, `shards`) → read → update `description` → destroy
- [ ] 7.2 Add import test via composite `"<space_id>/<pack_id>"` using server-generated UUID from create
- [ ] 7.3 Add `ecs_mapping` validator test: config with two fields set in same element → verify plan error
- [ ] 7.4 Add data source test: resource creates pack → data source reads same pack by `pack_id` → values match
- [ ] 7.5 Version-skip gate: skip tests against Kibana versions below `8.5.0`
- [ ] 7.6 Add prebuilt-pack test: data source reads a known prebuilt pack with `read_only=true` successfully; resource read/import against same pack returns error diagnostic
- [ ] 7.7 Add invalid `platform` plan-error test (e.g., `"ios"`)
- [ ] 7.8 Add non-default `space_id` resource test: create and read in a non-default space
- [ ] 7.9 Add data source non-default `space_id` test: read pack in non-default space
- [ ] 7.10 Add resource read 404 test: external delete → refresh removes resource from state without error
- [ ] 7.11 Add delete 404 idempotency test: destroy after external delete succeeds

## 8. Documentation and examples

- [ ] 8.1 Add `examples/resources/elasticstack_kibana_osquery_pack/resource.tf` with queries and `ecs_mapping` example (v1 — no scheduling fields)
- [ ] 8.2 Add `examples/resources/elasticstack_kibana_osquery_pack/import.sh` showing composite UUID import
- [ ] 8.3 Add `examples/data-sources/elasticstack_kibana_osquery_pack/data-source.tf`
- [ ] 8.4 Generate provider docs via the existing `make` target
- [ ] 8.5 Add a CHANGELOG entry following the repo's existing format

## 9. Validation and cleanup

- [ ] 9.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 9.2 Run `make check-openspec` — confirm this change validates
- [ ] 9.3 Run targeted acceptance tests against a real Kibana at or above `8.5.0`
- [ ] 9.4 Verify generated docs render correctly
- [ ] 9.5 Self-review with the `requirements-verification` skill against this change's specs

## Deferred (post-kbapi bump follow-up — not in v1)

- kbapi regeneration: fix `transform_schema.go` for Fleet `$ref` responses, bump OAS to ≥ `9dc7627253d0`, run `make -C generated/kbapi all`
- Scheduling schema: `schedule_type`, pack-level `interval`/`rrule_schedule`, per-query `interval`/`timeout`, RRULE validators, cross-mode ConfigValidators; interval Int64 normalization (former Decision 15)
- `GetVersionRequirements` second entry: scheduling floor `9.5.0`
- Acceptance tests for scheduling modes
