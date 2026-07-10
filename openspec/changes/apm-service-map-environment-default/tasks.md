## 1. Provider: suppress ENVIRONMENT_ALL default on read path

- [x] 1.1 In `internal/kibana/dashboard/panel/apmservicemap/model.go`, confirm the existing read-path null-preservation already keeps `apm_service_map_config.environment` null when the prior state is null/unknown (including when Kibana returns `"ENVIRONMENT_ALL"`). No read-path change should be needed for this issue; focus on import-path suppression (task 2.1) and add/import-focused unit tests.

- [x] 1.2 Define a package-level constant `const environmentServerDefault = "ENVIRONMENT_ALL"` in
  `model.go` (or a dedicated `defaults.go`) to avoid magic strings in the suppression logic.

## 2. Provider: suppress ENVIRONMENT_ALL default on import path

- [x] 2.1 In `apmServiceMapConfigFromAPIImport`, after setting `Environment:
  types.StringPointerValue(cfg.Environment)`, add a guard: if
  `result.Environment.ValueString() == environmentServerDefault`, set
  `result.Environment = types.StringNull()`. This ensures the import path produces the same
  null for an unconfigured `environment` as the normal read path.

## 3. Unit tests

- [x] 3.1 In `internal/kibana/dashboard/panel/apmservicemap/model_test.go`, add sub-tests for
  `PopulateFromAPI` covering:
  - Prior `environment = null`, API `environment = "ENVIRONMENT_ALL"` → state `environment = null`.
  - Prior `environment = "production"`, API `environment = "ENVIRONMENT_ALL"` → state
    `environment = "ENVIRONMENT_ALL"` (explicit value preserved).
  - Prior `environment = "ENVIRONMENT_ALL"` (explicitly set), API returns same → state
    `environment = "ENVIRONMENT_ALL"` (no spurious suppression).

- [x] 3.2 Add sub-tests for `apmServiceMapConfigFromAPIImport`:
  - API `environment = "ENVIRONMENT_ALL"` → returned config `environment = null`.
  - API `environment = nil` → returned config `environment = null` (unchanged behaviour).
  - API `environment = "production"` → returned config `environment = "production"`.

## 4. Acceptance tests

- [ ] 4.1 Re-run the four failing acceptance tests against a 9.5.0-SNAPSHOT stack to confirm the
  provider-side suppression alone resolves the `ImportStateVerify` mismatch:
  - `TestAccDashboardPanelApmServiceMap_allFilters`
  - `TestAccDashboardPanelApmServiceMap_noConfig`
  - `TestAccDashboardPanelApmServiceMap_serviceGroupIdOnly`
  - `TestAccDashboardPanelApmServiceMap_serviceNameOnly`

- [ ] 4.2 If any test still fails after suppression (e.g. import initialization still surfaces
  `environment`), add `ImportStateVerifyIgnore: []string{"panels.0.apm_service_map_config.environment"}`
  to the import step as a backstop — but prefer fixing the import-path suppression first (task 2.1).

## 5. Build and validate

- [x] 5.1 Run `make build` and confirm the provider compiles without errors.
- [x] 5.2 Run `go vet ./internal/kibana/dashboard/panel/apmservicemap/...` to confirm no static
  analysis issues.
- [x] 5.3 Run unit tests:
  `go test ./internal/kibana/dashboard/panel/apmservicemap/... -run TestUnit` (or equivalent unit
  test filter) to confirm all unit tests pass.

## 6. Spec sync

- [x] 6.1 Verify `make check-openspec` passes after the delta spec is applied.
