## 1. Package scaffold

- [ ] 1.1 Create `internal/kibana/dashboard/panel/apmservicemap/` directory with `schema.go`, `model.go`, `api.go`, `acc_test.go`, and `testdata/` subdirectory

## 2. Schema

- [ ] 2.1 Implement `SchemaAttribute()` in `schema.go` using `panelkit.PanelPresentationAttributes()` as the base attribute map, adding:
  - `environment` — optional `StringAttribute`
  - `service_name` — optional `StringAttribute`
  - `service_group_id` — optional `StringAttribute`
  - `kuery` — optional `StringAttribute`
  - `map_orientation` — optional `StringAttribute` with `stringvalidator.OneOf("horizontal", "vertical")`
  - `sync_with_dashboard_filters` — optional `BoolAttribute`
  - `alert_status_filter` — optional `SetAttribute(stringvalidator.OneOf("active", "delayed", "recovered", "untracked"))` on element type
  - `anomaly_severity_filter` — optional `SetAttribute(stringvalidator.OneOf("low", "warning", "minor", "major", "critical", "unknown"))` on element type
  - `connection_filter` — optional `SetAttribute(stringvalidator.OneOf("connected", "orphaned"))` on element type
  - `slo_status_filter` — optional `SetAttribute(stringvalidator.OneOf("degrading", "healthy", "noData", "violated"))` on element type
  - `time_range` — optional `SingleNestedAttribute` with `from` and `to` string attributes (reuse existing `panelkit.TimeRangeAttributes()` or equivalent helper)
- [ ] 2.2 Wrap all attributes in `panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{BlockName: "apm_service_map_config", PanelType: "apm_service_map", ...})`

## 3. Model

- [ ] 3.1 Define `ApmServiceMapConfigModel` in `model.go` with fields mirroring the schema attributes (use `types.String`, `types.Bool`, `types.Set`, and a nested `TimeRangeModel` or equivalent for `time_range`)
- [ ] 3.2 Add `ApmServiceMapConfigModel` field to `models.PanelModel` in the shared models package

## 4. API conversion

- [ ] 4.1 Implement `BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics` in `api.go`:
  - Map `title`, `description`, `hide_title`, `hide_border` via `typeutils.IsKnown` guards (pointer assignment)
  - Map `environment`, `service_name`, `service_group_id`, `kuery` as optional string pointers
  - Map `map_orientation` as `*KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientation` when set
  - Map `sync_with_dashboard_filters` as optional bool pointer
  - Map each filter set to `*[]Enum` (convert `types.Set` elements to the appropriate enum type); omit from payload when set is null/empty
  - Map `time_range` to `*KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema` when set
- [ ] 4.2 Implement `PopulateFromAPI(pm, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics` in `api.go`:
  - Apply REQ-009 null-preservation: when prior state had a field null and the API returns a value, keep state null
  - For filter sets: reconstruct `types.Set` from the API slice; preserve null if prior state had it null
  - For `time_range`: preserve null in state if prior state had it null
- [ ] 4.3 Implement `Handler.FromAPI` and `Handler.ToAPI` using `panelkit.SimpleFromAPI` / `panelkit.SimpleToAPI`

## 5. Registry and schema wiring

- [ ] 5.1 Register the `apmservicemap.Handler{}` in `panelHandlers` in `internal/kibana/dashboard/registry.go`
- [ ] 5.2 Add `apm_service_map_config` to the panel schema in `internal/kibana/dashboard/schema.go` and the mutual-exclusion validation list
- [ ] 5.3 Confirm `config_json` is rejected for `apm_service_map` panel type (handled automatically by the registry guard added in REQ-044A)

## 6. Unit tests

- [ ] 6.1 Write unit tests for `BuildConfig` covering:
  - All optional fields set → verify API payload fields
  - Nil config → verify empty/zero API payload
  - Each filter set with multiple enum values → verify slice contents
- [ ] 6.2 Write unit tests for `PopulateFromAPI` covering:
  - Import (no prior): API returns values → state populated
  - Import (no prior): API returns nothing → state remains nil
  - Round-trip: prior null fields stay null (null-preservation)
  - Filter set re-ordering: API returns values in different order from prior state → no plan change
- [ ] 6.3 Write unit test for invalid `map_orientation` enum value rejected at plan time

## 7. Acceptance tests

- [ ] 7.1 `TestAccDashboardPanelApmServiceMap_environmentOnly` — create with `environment` set, verify state and plan stability
- [ ] 7.2 `TestAccDashboardPanelApmServiceMap_serviceNameOnly` — create with `service_name` set
- [ ] 7.3 `TestAccDashboardPanelApmServiceMap_serviceGroupIdOnly` — create with `service_group_id` set
- [ ] 7.4 `TestAccDashboardPanelApmServiceMap_combinedSelectors` — create with all three service selectors set simultaneously
- [ ] 7.5 `TestAccDashboardPanelApmServiceMap_allFilters` — create with all four filter sets populated with multiple enum values; verify no drift on re-read (order independence)
- [ ] 7.6 `TestAccDashboardPanelApmServiceMap_full` — create a panel with every attribute set; verify state round-trip and plan stability

## 8. Build and lint

- [ ] 8.1 Run `make build` to confirm compilation
- [ ] 8.2 Run `go vet ./internal/kibana/dashboard/...`
- [ ] 8.3 Run `go test ./internal/kibana/dashboard/...` (unit tests, no `TF_ACC`)

## 9. Spec sync

- [ ] 9.1 Run `make check-openspec` to confirm the delta spec is valid and all tasks are accounted for
