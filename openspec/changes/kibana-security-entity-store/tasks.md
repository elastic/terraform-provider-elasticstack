## 1. Provider registration

- [ ] 1.1 Register `elasticstack_kibana_security_entity_store` in the provider's resource list
  in `provider/plugin_framework.go`.
- [ ] 1.2 Register `elasticstack_kibana_security_entity_store_status` in the provider's data
  source list in `provider/plugin_framework.go`.

## 2. Package scaffold

- [ ] 2.1 Create directory `internal/kibana/security_entity_store/`.
- [ ] 2.2 Create the following empty files (stubs) so the package compiles:
  `resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`,
  `data_source.go`, `data_source_schema.go`, `data_source_models.go`, `data_source_read.go`.

## 3. Resource schema — `elasticstack_kibana_security_entity_store`

- [ ] 3.1 In `schema.go`, define `getSchema()` returning a `schema.Schema` with:
  - `id` — Computed `StringAttribute`.
  - `space_id` — Optional + Computed `StringAttribute`, `RequiresReplace` plan modifier.
  - `entity_types` — Optional + Computed `SetAttribute` with `ElementType: types.StringType`;
    validators: set elements must be one of `user`, `host`, `service`, `generic`.
  - `allow_entity_type_shrink` — Optional `BoolAttribute`, default `false`.
  - `started` — Optional + Computed `BoolAttribute`, default `true`.
  - `history_snapshot` — Optional `SingleNestedAttribute` containing:
    - `frequency` — Optional `StringAttribute`, `RequiresReplace` plan modifier on the block.
  - `log_extraction` — Optional `SingleNestedAttribute` containing:
    - `additional_index_patterns` — Optional `ListAttribute{ElementType: types.StringType}`.
    - `excluded_index_patterns` — Optional `ListAttribute{ElementType: types.StringType}`.
    - `delay` — Optional `StringAttribute`.
    - `docs_limit` — Optional `Int64Attribute`.
    - `field_history_length` — Optional `Int64Attribute`.
    - `frequency` — Optional `StringAttribute`.
    - `lookback_period` — Optional `StringAttribute`.
    - `max_logs_per_page` — Optional `Int64Attribute`.
    - `max_logs_per_window` — Optional `Int64Attribute`.
    - `max_logs_per_window_cap_behavior` — Optional `StringAttribute`; validator: one of `drop`, `defer`.
    - `max_time_window_size` — Optional `StringAttribute`.
  - `status_json` — Computed `StringAttribute`.
  - `kibana_connection` — injected by the Kibana resource envelope (do not add manually if using
    `entitycore.KibanaResource`).

## 4. Resource models — `models.go`

- [ ] 4.1 Define `tfModel` struct with `tfsdk` tags matching all schema attributes.
- [ ] 4.2 Define nested structs `historySnapshotModel` and `logExtractionModel`.
- [ ] 4.3 Implement `KibanaResourceModel` interface methods on `tfModel`:
  - `GetID() types.String`
  - `GetResourceID() types.String` (returns `types.StringValue("entity_store")`)
  - `GetSpaceID() types.String`
  - `GetKibanaConnection() types.List`

## 5. Resource constructor — `resource.go`

- [ ] 5.1 Declare `var MinVersion = version.Must(version.NewVersion("9.1.0"))`.
- [ ] 5.2 Construct the resource via `entitycore.NewKibanaResource[tfModel]` with callbacks:
  `Schema: getSchema`, `Create: createEntityStore`, `Read: readEntityStore`,
  `Update: updateEntityStore`, `Delete: deleteEntityStore`.
- [ ] 5.3 Assert `resource.Resource`, `resource.ResourceWithConfigure`, and
  `resource.ResourceWithImportState` are satisfied.

## 6. Create callback — `create.go`

- [ ] 6.1 Enforce `MinVersion` at 9.1.0 before any API call.
- [ ] 6.2 Build `PostSecurityEntityStoreInstallJSONRequestBody` from plan:
  - Map `entity_types` set elements to `[]PostSecurityEntityStoreInstallJSONBodyEntityTypes`.
  - Map `history_snapshot.frequency` to `HistorySnapshot.Frequency` pointer.
  - Map `log_extraction.*` fields to `LogExtraction.*` fields (int64 → int conversion).
- [ ] 6.3 Call `PostSecurityEntityStoreInstallWithResponse`; accept HTTP 200 and 201 as success;
  return error diagnostic for other codes.
- [ ] 6.4 If `started == false`, call `PutSecurityEntityStoreStopWithResponse`.
- [ ] 6.5 Compute `id` as `<space_id>/entity_store` and set in state.
- [ ] 6.6 Call the Read callback to populate all computed fields.

## 7. Read callback — `read.go`

- [ ] 7.1 Call `GetSecurityEntityStoreStatusWithResponse` with `IncludeComponents: false` (or true if
  needed for extraction settings).
- [ ] 7.2 If response status is `not_installed`, call `resp.State.RemoveResource(ctx)` and return.
- [ ] 7.3 Collect engine `Type` values and populate `entity_types` in state.
- [ ] 7.4 Set `started = true` if any engine has status `running`, else `false`.
- [ ] 7.5 Populate `log_extraction` from the first engine's extraction settings (delay, frequency,
  lookback_period, field_history_length, max_logs_per_page); leave nil fields as null.
- [ ] 7.6 Serialize the full status response body to `status_json` (normalized via `json.Marshal`
  after unmarshalling, or use the raw `Body` bytes stored via `json.RawMessage`).

## 8. Update callback — `update.go`

- [ ] 8.1 Enforce `MinVersion` at 9.1.0.
- [ ] 8.2 **Log extraction reconciliation**: if `log_extraction` changed between prior state and
  plan, call `PutSecurityEntityStoreWithResponse` with the new `LogExtraction` block. The API
  requires `LogExtraction` to be set (not a pointer) — populate all known fields.
- [ ] 8.3 **Entity type expansion**: compute added types (in plan but not in state). If non-empty,
  call `PostSecurityEntityStoreInstallWithResponse` with the full desired `entity_types`.
- [ ] 8.4 **Entity type shrink guard**: compute removed types (in state but not in plan).
  - If `allow_entity_type_shrink == false` and any removed types exist, append an error diagnostic
    explaining the shrink guard and return without calling any API.
  - If `allow_entity_type_shrink == true`, call `PostSecurityEntityStoreUninstallWithResponse`
    with only the removed entity types.
- [ ] 8.5 **Start/stop reconciliation**: if `started` changed from true to false, call
  `PutSecurityEntityStoreStopWithResponse`; from false to true, call
  `PutSecurityEntityStoreStartWithResponse`.
- [ ] 8.6 Call the Read callback to refresh state.

## 9. Delete callback — `delete.go`

- [ ] 9.1 Call `PostSecurityEntityStoreUninstallWithResponse` with the entity types from state
  (nil body uninstalls all, which is equivalent to passing all installed types).
- [ ] 9.2 Accept HTTP 200 as success. Remove resource from state implicitly (framework does this).

## 10. Data source — `data_source.go`, `data_source_schema.go`, `data_source_models.go`, `data_source_read.go`

- [ ] 10.1 In `data_source_schema.go`, define `getDataSourceSchema()` with:
  - `space_id` — Optional + Computed `StringAttribute`.
  - `include_components` — Optional `BoolAttribute`.
  - `installed` — Computed `BoolAttribute`.
  - `overall_status` — Computed `StringAttribute`.
  - `engines_json` — Computed `StringAttribute`.
  - `status_json` — Computed `StringAttribute`.
  - `kibana_connection` — injected by the data source envelope pattern.
- [ ] 10.2 In `data_source_models.go`, define `dsModel` struct.
- [ ] 10.3 In `data_source_read.go`, implement the read callback:
  - Enforce `MinVersion` at 9.1.0 before any API call.
  - Call `GetSecurityEntityStoreStatusWithResponse`, passing `IncludeComponents` from the config.
  - Populate `installed` from `status != "not_installed"`.
  - Populate `overall_status` from `status` field.
  - Serialize `engines` to `engines_json` and the full response to `status_json`.
- [ ] 10.4 In `data_source.go`, register the data source via the provider's data source list.

## 11. Import support

- [ ] 11.1 Implement `ImportState` on the resource by delegating to
  `resource.ImportStatePassthroughID` with path `path.Root("id")`, or by parsing `<space_id>/entity_store`
  from the import ID and setting `space_id` then calling Read.

## 12. Acceptance tests — `acc_test.go`

- [ ] 12.1 `TestAccResourceKibanaSecurityEntityStore_basic` — create with default entity types,
  verify plan is clean after apply.
- [ ] 12.2 `TestAccResourceKibanaSecurityEntityStore_singleType` — create with `entity_types = ["host"]`.
- [ ] 12.3 `TestAccResourceKibanaSecurityEntityStore_updateLogExtraction` — update `log_extraction`
  settings and verify no replacement.
- [ ] 12.4 `TestAccResourceKibanaSecurityEntityStore_import` — import the resource using its ID.
- [ ] 12.5 `TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails` — verify that reducing
  `entity_types` without `allow_entity_type_shrink = true` produces an error.
- [ ] 12.6 `TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag` — verify that reducing
  `entity_types` with `allow_entity_type_shrink = true` succeeds.
- [ ] 12.7 `TestAccResourceKibanaSecurityEntityStore_startedFalse` — create with `started = false`
  and verify engines are not running.
- [ ] 12.8 `TestAccDataSourceKibanaSecurityEntityStoreStatus_basic` — verify the data source reads
  status with and without `include_components`.

## 13. OpenSpec spec update

- [ ] 13.1 Apply the delta specs for capabilities `kibana-security-entity-store` and
  `kibana-security-entity-store-status` to the canonical specs in `openspec/specs/` when this
  change is merged (handled by the `openspec-sync-specs` workflow step).

## 14. Build and validation

- [ ] 14.1 `make build` — provider compiles without errors.
- [ ] 14.2 `go test ./internal/kibana/security_entity_store/... -v` — unit tests pass.
- [ ] 14.3 `make check-openspec` — spec validation passes.
