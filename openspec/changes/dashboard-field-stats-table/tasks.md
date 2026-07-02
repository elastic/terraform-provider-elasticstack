## 1. Handler package scaffold

- [ ] 1.1 Create `internal/kibana/dashboard/panel/fieldstatstable/` directory with initial files: `schema.go`, `model.go`, `api.go`, `fromapi.go`
- [ ] 1.2 Add a `descriptions/` sub-directory with embedded description markdown files for `field_stats_table_config`, `by_dataview`, and `by_esql` blocks
- [ ] 1.3 Implement the `discoverSession`-style exactly-one-of validator for `field_stats_table_config` (ensures exactly one of `by_dataview` / `by_esql` is set)

## 2. Schema

- [ ] 2.1 Define `FieldStatsTableSchema()` in `schema.go` returning the `field_stats_table_config` attribute with nested `by_dataview` and `by_esql` optional object blocks
- [ ] 2.2 `by_dataview` block: required `data_view_id` (string), optional `show_distributions` (bool), optional `title` (string), optional `description` (string), optional `hide_title` (bool), optional `hide_border` (bool), optional `time_range` object (`from` required, `to` required, `mode` optional per REQ-009 semantics)
- [ ] 2.3 `by_esql` block: required `query` (string, the ES|QL query text mapped to `query.esql` on the wire), optional `show_distributions` (bool), optional `title` / `description` / `hide_title` / `hide_border` / `time_range` matching `by_dataview`
- [ ] 2.4 Register the exactly-one-of validator on the `field_stats_table_config` object attribute
- [ ] 2.5 Add `field_stats_table_config` to the panel schema in `internal/kibana/dashboard/schema.go`

## 3. Model

- [ ] 3.1 Define Go structs `FieldStatsTablePanelConfig`, `FieldStatsTableByDataviewConfig`, and `FieldStatsTableByEsqlConfig` in `model.go` using `types.*` fields for Terraform framework compatibility
- [ ] 3.2 Define `attrTypes()` helpers for each struct to support `types.ObjectValueFrom` / `types.ObjectAs` usage

## 4. Write path (ToAPI)

- [ ] 4.1 Implement `ToAPI(ctx, config)` in `api.go` that builds the `KibanaHTTPAPIsDataVisualizerFieldStats` union value from the active branch
- [ ] 4.2 For `by_dataview`: set `view_type = "dataview"`, `data_view_id`, and optional fields
- [ ] 4.3 For `by_esql`: set `view_type = "esql"`, `query.esql`, and optional fields
- [ ] 4.4 Apply panelkit helpers for `title`, `description`, `hide_title`, `hide_border`, `time_range` passthrough

## 5. Read path (FromAPI)

- [ ] 5.1 Implement `FromAPI(ctx, raw)` in `fromapi.go` that detects the `view_type` discriminator and populates the matching branch, leaving the other null
- [ ] 5.2 Apply REQ-009 null-preservation for `time_range` and other optional fields: keep null in state when prior state had them null even if Kibana returns defaults
- [ ] 5.3 Use `panelkit.SimpleFromAPI` (or equivalent) for common passthrough fields

## 6. Registry and config_json guard

- [ ] 6.1 Register the `fieldstatstable` panel handler in `panelHandlers` in `internal/kibana/dashboard/registry.go`
- [ ] 6.2 Extend the `config_json` rejection guard (REQ-010) to include `field_stats_table` in the error-producing panel type list

## 7. Tests

- [ ] 7.1 Unit tests in `internal/kibana/dashboard/panel/fieldstatstable/api_test.go`: ToAPI round-trip for `by_dataview`, ToAPI round-trip for `by_esql`, null optional fields omitted from payload
- [ ] 7.2 Unit tests in `fromapi_test.go` (or combined): FromAPI correctly detects each branch; optional fields null when API omits them; `time_range` null-preservation
- [ ] 7.3 Validator unit tests: both branches set → error; neither branch set → error; exactly one set → passes
- [ ] 7.4 Acceptance tests in `acc_test.go`:
  - Create/read/update round-trip for `by_dataview` branch with `show_distributions = true` and `time_range`
  - Create/read/update round-trip for `by_esql` branch with `show_distributions = false`
  - Validator rejection: both `by_dataview` and `by_esql` set simultaneously → plan-time error
  - Drift detection: Kibana returns branch data intact → no diff on subsequent plan
- [ ] 7.5 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...` (unit); `TF_ACC=1 go test` for acceptance tests

## 8. Spec sync

- [ ] 8.1 Verify `make check-openspec` passes after merging the delta spec into the main spec
