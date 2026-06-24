## 1. Prep and discovery

- [x] 1.1 Verify that `OsqueryCreatePacks`, `OsqueryGetPacksDetails`, `OsqueryUpdatePacks`, `OsqueryDeletePacks` are present in `generated/kbapi/kibana.gen.go`; confirm their request/response type signatures (especially `queries`, `shards`, and the create response wrapper shape)
- [x] 1.2 Bump OAS ref in `generated/kbapi/Makefile` to a version that includes `schedule_type`/`rrule_schedule`/per-query `interval`/`timeout`/`splay`; run `make -C generated/kbapi all`; confirm these fields appear in the regenerated client. If they do not, scope the implementation to interval-only scheduling and update design.md accordingly.
- [x] 1.3 Confirm the minimum Kibana version for base packs CRUD (`/api/osquery/packs`) and for the full scheduling model; record confirmed versions in design.md Decision 11 and update `GetVersionRequirements`
- [x] 1.4 Confirm that Kibana generates a UUID for `pack_id` when omitted on Create; if not, escalate to Required and update design.md Decision 2
- [x] 1.5 Confirm whether the Create response wraps the pack in a `data` field (like `osquery_saved_query`) or is a direct object; update design.md Decision 14 if different
- [x] 1.6 Confirm the exact `shards` wire format on Create/Update (array `[{key, value}]` or already a map?); confirm read (`GetPacksDetails`) returns `map[string]float32`; update design.md Decision 8 if different

## 2. kibanaoapi client helper

- [ ] 2.1 Create `internal/clients/kibanaoapi/osquery_pack.go` with thin wrappers `CreateOsqueryPack`, `GetOsqueryPack`, `UpdateOsqueryPack`, `DeleteOsqueryPack` — each passing `kibanautil.SpaceAwarePathRequestEditor(spaceID)` and using `HandleGetTypedResponse` / `HandleMutateTypedResponse` / `diagutil.HandleStatusResponse` consistently with existing kibanaoapi wrappers (e.g. `synthetics_private_location.go`)
- [ ] 2.2 Map HTTP 404 on Get to a nil/sentinel result (resource removed from state); map HTTP 404 on Delete to a no-op success
- [ ] 2.3 Map non-2xx responses to provider diagnostics consistently with other kibanaoapi helpers
- [ ] 2.4 Normalize `shards` on read: convert `map[string]float32` from `GetPacksDetails` to `map[string]float64` (or `map[string]int64` pending task 1.6 resolution) for state

## 3. Resource skeleton and model

- [ ] 3.1 Create `internal/kibana/osquery_pack/` directory mirroring `internal/kibana/osquery_saved_query/`
- [ ] 3.2 Implement `models.go` with `osqueryPackModel` implementing `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`, `GetVersionRequirements` (two entries: base CRUD floor + scheduling floor from task 1.3)
- [ ] 3.3 Implement `rruleScheduleModel` shared nested struct covering `rrule`, `start_date` (`timetypes.RFC3339`), `end_date` (`timetypes.RFC3339`, optional), `splay` (`customtypes.DurationType`, optional), `timeout` (Int64, optional) plus `toAPIType()` and `fromAPIType()` converters
- [ ] 3.4 Implement `queryModel` nested struct covering `query`, `interval`, `rrule_schedule`, `platform`, `version`, `snapshot`, `removed`, `saved_query_id`, `ecs_mapping`; plus `toAPIType()` and `fromAPIType()`
- [ ] 3.5 Implement `ecsMappingModel` (reuse or mirror from `osquery_saved_query`) covering `field`, `value`, `values`
- [ ] 3.6 Implement `populateFromAPI` mapping the kbapi pack response to the model, including `shards` normalization and per-query field mapping

## 4. Resource schema

- [ ] 4.1 Implement `getSchema` covering: `id` (Computed), `pack_id` (Optional+Computed, RequiresReplace), `space_id` (Optional+Computed, default `"default"`, RequiresReplace), `kibana_connection` (Optional), `name` (Required string), `description` (Optional string), `enabled` (Optional bool), `policy_ids` (Optional list(string)), `shards` (Optional MapAttribute of numbers), `schedule_type` (Required string, enum validator), `interval` (Optional Int64), `rrule_schedule` (Optional SingleNestedAttribute), `queries` (Required MapNestedAttribute)
- [ ] 4.2 Implement the `rrule_schedule` nested schema: `rrule` (Required string, starts-with-FREQ= validator), `start_date` (`timetypes.RFC3339`, Required), `end_date` (`timetypes.RFC3339`, Optional), `splay` (`customtypes.DurationType`, Optional, ≤12h validator), `timeout` (Int64, Optional)
- [ ] 4.3 Implement the per-query nested schema inside `queries` map: `query` (Required string), `interval` (Optional Int64), `rrule_schedule` (Optional SingleNestedAttribute, same shape as pack-level), `platform` (Optional SetAttribute of strings with allowed-values validator), `version` (Optional string), `snapshot` (Optional+Computed bool), `removed` (Optional+Computed bool), `saved_query_id` (Optional string), `ecs_mapping` (Optional MapNestedAttribute)
- [ ] 4.4 Add the per-element `ConfigValidator` on each `ecs_mapping` enforcing exactly-one-of `field`/`value`/`values`
- [ ] 4.5 Implement `ConfigValidators` on the resource enforcing: (a) exactly-one-of `interval`/`rrule_schedule` at pack level; (b) per-query scheduling override mode (if set) must match pack's `schedule_type`
- [ ] 4.6 Add `RequiresReplace` plan modifiers on `pack_id` and `space_id`; add `UseStateForUnknown` on Optional+Computed fields

## 5. Resource CRUD and import

- [ ] 5.1 Implement `create.go` calling `POST /api/osquery/packs` (space-aware), unwrapping response (confirm `data` wrapper from task 1.5), calling `populateFromAPI`; error if `read_only=true`
- [ ] 5.2 Implement `read.go` calling `GET /api/osquery/packs/{id}` (space-aware); on HTTP 404, remove from state; error if `read_only=true`
- [ ] 5.3 Implement `update.go` calling `PUT /api/osquery/packs/{id}` (space-aware, full body); repopulate state from response
- [ ] 5.4 Implement `delete.go` calling `DELETE /api/osquery/packs/{id}` (space-aware); treat HTTP 404 as success
- [ ] 5.5 Implement `ImportState` accepting composite `"<space_id>/<pack_id>"` form
- [ ] 5.6 Register `osqueryPack.NewResource()` in `provider/plugin_framework.go`

## 6. Data source

- [ ] 6.1 Implement `internal/kibana/osquery_pack/datasource.go` with schema: `pack_id` (Required), `space_id` (Optional, default `"default"`), `kibana_connection` (Optional), plus all Computed fields matching the resource (`name`, `description`, `enabled`, `policy_ids`, `shards`, `schedule_type`, `interval`, `rrule_schedule`, `queries`, `read_only` as Computed bool)
- [ ] 6.2 Implement Read calling `GET /api/osquery/packs/{id}` — same kibanaoapi wrapper; on HTTP 404, return error diagnostic
- [ ] 6.3 Do NOT error on `read_only=true` in the data source
- [ ] 6.4 Register the data source in `provider/plugin_framework.go`

## 7. Acceptance tests

- [ ] 7.1 Add `acc_test.go` covering full resource lifecycle: create with all fields (including `ecs_mapping` with all three shapes, `policy_ids`, `shards`) → read → update `description` → destroy
- [ ] 7.2 Add resource lifecycle test with `pack_id` explicitly set; forces-new on change
- [ ] 7.3 Add resource lifecycle test with `pack_id` omitted (verify Kibana generates an ID; import by generated ID)
- [ ] 7.4 Add import test via composite `"<space_id>/<pack_id>"`
- [ ] 7.5 Add scheduling test: `interval` mode create → update; scheduling validator tests (both modes set → error; neither set → error)
- [ ] 7.6 Add `ecs_mapping` validator test: config with two fields set in same element → verify plan error
- [ ] 7.7 Add data source test: resource creates pack → data source reads same pack by ID → values match
- [ ] 7.8 Version-skip gate: skip tests against Kibana versions below the confirmed minimum

## 8. Documentation and examples

- [ ] 8.1 Add `examples/resources/elasticstack_kibana_osquery_pack/resource.tf` with `interval` scheduling mode and `ecs_mapping` example
- [ ] 8.2 Add `examples/resources/elasticstack_kibana_osquery_pack/import.sh` showing composite ID import
- [ ] 8.3 Add `examples/data-sources/elasticstack_kibana_osquery_pack/data-source.tf`
- [ ] 8.4 Generate provider docs via the existing `make` target
- [ ] 8.5 Add a CHANGELOG entry following the repo's existing format

## 9. Validation and cleanup

- [ ] 9.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 9.2 Run `make check-openspec` — confirm this change validates
- [ ] 9.3 Run targeted acceptance tests against a real Kibana at or above the confirmed minimum version
- [ ] 9.4 Verify generated docs render correctly
- [ ] 9.5 Self-review with the `requirements-verification` skill against this change's specs
