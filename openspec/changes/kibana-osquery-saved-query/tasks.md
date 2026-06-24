## 1. Prep and discovery

- [x] 1.1 Verify that `OsqueryCreateSavedQuery`, `OsqueryGetSavedQueryDetails`, `OsqueryUpdateSavedQuery`, `OsqueryDeleteSavedQuery` are present in `generated/kbapi/kibana.gen.go` and confirm their request/response type signatures match the design (in particular: `SecurityOsqueryAPICreateSavedQueryRequestBody` and the `data`-wrapped create response)
- [x] 1.2 Confirm the minimum Kibana version that ships `/api/osquery/saved_queries` CRUD; record the documented/conservative floor (`8.5.0`) in design.md (implementation of `GetVersionRequirements` is task 3.2; live confirmation is task 7.9)
- [x] 1.3 Confirm that Kibana generates a UUID for `saved_query_id` when it is omitted on Create; if it does not, escalate `saved_query_id` to Required and update design.md Decision 2
- [x] 1.4 Confirm validator approach for the three-way `field`/`value`/`values` constraint inside `MapNestedAttribute` values; plan `ExactlyOneOfNestedAttrsValidator` on `NestedObject.Validators` (fallback: custom inline `ValidateObject` only if map nested validation fails at implementation time)

## 2. kibanaoapi client helper

- [x] 2.1 Create `internal/clients/kibanaoapi/osquery_saved_query.go` with thin wrappers `CreateOsquerySavedQuery`, `GetOsquerySavedQuery`, `UpdateOsquerySavedQuery`, `DeleteOsquerySavedQuery` — each passing `kibanautil.SpaceAwarePathRequestEditor(spaceID)` and using `HandleGetTypedResponse` / `HandleMutateTypedResponse` / `HandleStatusResponse` consistently with `maintenance_window.go`
- [x] 2.2 Map HTTP 404 on Get to a nil/sentinel result (resource removed from state); map HTTP 404 on Delete to a no-op success
- [x] 2.3 Map non-2xx responses to provider diagnostics consistently with other kibanaoapi helpers
- [x] 2.4 Unwrap the `data` field from the Create (and Update, if applicable) response before returning the typed entity

## 3. Resource skeleton and model

- [x] 3.1 Create `internal/kibana/osquery_saved_query/` directory mirroring `internal/kibana/maintenance_window/`
- [x] 3.2 Implement `models.go` with `osquerySavedQueryModel`, implementing `GetID` (composite `<space_id>/<saved_query_id>`), `GetResourceID` (`saved_query_id` for API lookup), `GetSpaceID`, `GetKibanaConnection`, `GetVersionRequirements` (declare `8.5.0` floor from task 1.2)
- [x] 3.3 Implement `ecsMapping` nested model covering `field`, `value` (string), `values` (set of string), plus the `toAPIType()` and `fromAPIType()` converters handling the `string | []string` union
- [x] 3.4 Implement per-operation `populateFromAPI` mappers from kibanaoapi entity types (`OsquerySavedQueryCreateEntity`, `OsquerySavedQueryGetEntity`, `OsquerySavedQueryUpdateEntity`) to the model: Create and GET use union types for `interval`/`version` (`AsXxx0()/AsXxx1()`); Update entity types `version` as plain `*string` while `interval` remains a union; handle `platform` comma-split
- [x] 3.5 Add unit tests for `populateFromAPI` and converters covering all three kibanaoapi entity shapes: interval int and string union arms on Create/GET entities, version int and string union arms on Create/GET entities, Update plain `*string` version, platform comma join/split (kibanaoapi `.Data` unwrap mappers covered in task 2 — see `osquery_saved_query_test.go`)

## 4. Resource schema

- [x] 4.1 Implement `getSchema` covering: `id` (Computed, composite `<space_id>/<saved_query_id>`), `saved_query_id` (Required, RequiresReplace), `space_id` (Optional+Computed, default `"default"`, RequiresReplace), `kibana_connection` (Optional, from `entitycore`), `query` (Required string), `description` (Optional string), `platform` (Optional SetAttribute of strings with allowed-values validator), `interval` (Optional Int64), `version` (Optional string), `snapshot` (Optional+Computed bool), `removed` (Optional+Computed bool), `ecs_mapping` (Optional MapNestedAttribute)
- [x] 4.2 Add `RequiresReplace` plan modifiers on `saved_query_id` and `space_id`
- [x] 4.3 Add `UseStateForUnknown` plan modifiers on Optional+Computed fields (`space_id`, `snapshot`, `removed`, and other computed-only attributes as needed)
- [x] 4.4 Implement the `ecs_mapping` element schema as a `MapNestedAttribute` nested object with `field` (Optional string), `value` (Optional string), `values` (Optional SetAttribute of strings)
- [x] 4.5 Attach `validators.ExactlyOneOfNestedAttrsValidator` to `ecs_mapping` `MapNestedAttribute.NestedObject.Validators` enforcing exactly-one-of `field`, `value`, `values` per element (proven on nested/list objects, not yet on map values — task 7.6 validates); if map nested validation fails at implementation time, fall back to a custom inline `ValidateObject` per task 1.4

## 5. Resource CRUD and import

- [x] 5.1 Implement `create.go` via `CreateOsquerySavedQuery`, map the returned entity with `populateFromAPI`, and return a prebuilt error diagnostic if `prebuilt == true`
- [x] 5.2 Implement `read.go` via `GetOsquerySavedQuery` (lookup via `saved_query_id`); on HTTP 404, remove from state without error; on success, call `populateFromAPI`; return a prebuilt error diagnostic if `prebuilt == true`
- [x] 5.3 Implement `update.go` via `UpdateOsquerySavedQuery` (managed field set from plan/state, omitting server-managed fields and null/unset optional keys), then repopulate state from the returned entity
- [x] 5.4 Implement `delete.go` calling `DELETE /api/osquery/saved_queries/{id}` (space-aware); treat HTTP 404 as success
- [x] 5.5 Implement `ImportState` for composite `"<space_id>/<saved_query_id>"`: prefer `ImportStatePassthroughID` on `id`; if Required `saved_query_id` must be seeded before Read, use thin custom parser (as `alerting_rule`) to set `space_id`, `saved_query_id`, and `id` from import string
- [x] 5.6 Register `osquerySavedQuery.NewResource()` in the resource slice in `provider/plugin_framework.go`

## 6. Data source

- [x] 6.1 Implement `internal/kibana/osquery_saved_query/datasource.go` (or `data_source.go`) with schema: `saved_query_id` (Required), `space_id` (Optional, default `"default"`), `kibana_connection` (Optional), plus all the same Computed fields as the resource (`query`, `description`, `platform`, `interval`, `version`, `snapshot`, `removed`, `ecs_mapping`, and `prebuilt` as Computed bool); shared model or datasource model implements `GetVersionRequirements` with `8.5.0` floor
- [x] 6.2 Implement Read via `GetOsquerySavedQuery` (same kibanaoapi wrapper as the resource); on HTTP 404, return an error diagnostic rather than removing from state (data sources error on missing)
- [x] 6.3 Do NOT error on `prebuilt == true` in the data source — prebuilt queries are a primary use case for the data source
- [x] 6.4 Register the data source in `provider/plugin_framework.go`

## 7. Acceptance tests

- [x] 7.1 Add `acc_test.go` covering full resource lifecycle: create with all fields (including `ecs_mapping` with all three shapes) → read → update `query` and `description` → destroy
- [x] 7.2 Add resource lifecycle test with `saved_query_id` explicitly set (forces-new on change)
- [x] 7.3 Add plan/validation test: config without `saved_query_id` → verify plan-time error (Required attribute)
- [x] 7.4 Add import test via composite `"<space_id>/<saved_query_id>"`
- [x] 7.5 Add `platform` test: create with `["linux", "darwin"]` → verify state and round-trip
- [x] 7.6 Add `ecs_mapping` validator tests: config with two fields set in same element → plan error; config with empty element `{}` → plan error
- [x] 7.10 Add resource test: import or read of a prebuilt query (by known prebuilt ID, skip if none in test env) → verify prebuilt error diagnostic and no state write
- [x] 7.7 Add data source test: resource creates query → data source reads same query by ID → values match
- [x] 7.8 Add data source test: read a prebuilt query by ID (skip if none available in test environment); add data source pre-minimum version gate test aligned with task 7.9
- [x] 7.9 Version-skip gate: skip tests against Kibana versions below the documented minimum (`8.5.0`); confirms floor against a live stack when available

## 8. Documentation and examples

- [ ] 8.1 Add `examples/resources/elasticstack_kibana_osquery_saved_query/resource.tf` with `ecs_mapping` example (covering `field`, `value`, and `values` forms)
- [ ] 8.2 Add `examples/resources/elasticstack_kibana_osquery_saved_query/import.sh` showing composite ID import
- [ ] 8.3 Add `examples/data-sources/elasticstack_kibana_osquery_saved_query/data-source.tf`
- [ ] 8.4 Generate provider docs (`docs/resources/kibana_osquery_saved_query.md`, `docs/data-sources/kibana_osquery_saved_query.md`) via the existing `make` target
- [ ] 8.5 Add a CHANGELOG entry following the repo's existing format

## 9. Validation and cleanup

- [ ] 9.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 9.2 Run `make check-openspec` — confirm this change validates
- [ ] 9.3 Run targeted acceptance tests against a real Kibana at or above the confirmed minimum version (per `dev-docs/high-level/testing.md`)
- [ ] 9.4 Verify generated docs render correctly
- [ ] 9.5 Self-review with the `requirements-verification` skill against this change's specs
