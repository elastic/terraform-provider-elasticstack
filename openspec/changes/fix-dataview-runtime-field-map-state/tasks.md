## 1. Schema Fix

- [ ] 1.1 Add `Computed: true` to `data_view.runtime_field_map` in `internal/kibana/dataview/schema.go`
- [ ] 1.2 Run `make build` to verify the schema change compiles

## 2. Acceptance Test Fix

- [ ] 2.1 Update `internal/kibana/dataview/testdata/TestAccResourceDataView/basic_updated/data_view.tf` to include `field_attrs` with the same values as in `internal/kibana/dataview/testdata/TestAccResourceDataView/basic/data_view.tf`, so the step no longer triggers `RequiresReplace()`
- [ ] 2.2 Add `captureID` and `checkIDUnchanged` helper functions in `internal/kibana/dataview/acc_test.go` (same pattern used in `TestAccResourceDataViewNamespaces`)
- [ ] 2.3 Attach `captureID` to the `basic` step's Check and `checkIDUnchanged` to the `basic_updated` step's Check in `TestAccResourceDataView`
- [ ] 2.4 Update the `basic_updated` step assertions: replace `TestCheckNoResourceAttr(..., "data_view.runtime_field_map")` with `TestCheckResourceAttr(..., "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)")` to assert that Kibana preserves the runtime field
- [ ] 2.5 Update the import step to add `ImportStateVerifyIgnore: []string{"data_view.runtime_field_map"}` so that import verification succeeds when the imported state contains a preserved `runtime_field_map` that is absent from config

## 3. Requirements Update

- [ ] 3.1 Update REQ-011 in `openspec/specs/kibana-data-view/spec.md` to match the modified requirement in the change delta spec (mark `runtime_field_map` as `Computed`, add scenario for omitted config with persisted API value)

## 4. Validation

- [ ] 4.1 Run unit tests: `go test ./internal/kibana/dataview/...`
- [ ] 4.2 Run `make check-lint` to ensure lint passes
- [ ] 4.3 If a Kibana stack is available, run acceptance tests: `TF_ACC=1 go test -v -run TestAccResourceDataView ./internal/kibana/dataview/... -timeout 20m`
