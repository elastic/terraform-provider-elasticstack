## 0. contracttest harness

- [x] 0.1 Create `dashboard/panelkit/contracttest/parse.go` — raw JSON string → `kbapi.DashboardPanelItem` adapter
- [x] 0.2 Create `dashboard/panelkit/contracttest/roundtrip.go` — `FromAPI → ToAPI` stability assertion with JSON diff output on failure
- [x] 0.3 Create `dashboard/panelkit/contracttest/nullpreserve.go` — walk `handler.SchemaAttribute()` to collect optional leaves; generate three sub-tests per field (prior null, prior known, prior nil)
- [x] 0.4 Create `dashboard/panelkit/contracttest/reflection.go` — `HasConfig`/`ClearConfig` post-condition assertions using `panelkit.HasConfig`
- [x] 0.5 Create `dashboard/panelkit/contracttest/schema.go` — structural assertions: outer attribute is Optional, Required leaves are present in fixture, `ValidatePanelConfig` rejects zeroed Required fields
- [x] 0.6 Create `dashboard/panelkit/contracttest/harness.go` — `Run(t, handler, config)` entry point orchestrating the above; `Config` struct with `FullAPIResponse string` and optional `SkipFields []string`
- [x] 0.7 Verify harness compiles and `contracttest.Run` fails with a clear error when given a no-op stub handler

## 1. Infrastructure

- [x] 1.1 Create `dashboard/panel/iface/iface.go` with `Handler` and `PinnedHandler` interfaces
- [x] 1.2 Create `dashboard/panelkit/grid.go` with `GridFromAPI`, `GridToAPI`, `IDFromAPI`, `IDToAPI`
- [x] 1.3 Create `dashboard/panelkit/nullpreserve.go` with `PreserveString`, `PreserveBool`, `PreserveList`, `PreserveFloat64`
- [x] 1.4 Create `dashboard/panelkit/reflection.go` with `HasConfig`, `ClearConfig`, `SetConfig` using tfsdk tag matching
- [x] 1.5 Verify `panelkit` reflection panics at init if any registered handler has a block name with no matching tfsdk tag
- [x] 1.6 Create `dashboard/registry.go` with `panelHandlers` slice, `panelTypeToHandler` map, and `derivedPanelConfigNames` (typed `*_config` names plus `config_json`) populated at `init`
- [x] 1.7 Create `dashboard/router.go` with `mapPanelFromAPIViaRegistry` and `panelModelToAPIViaRegistry` delegating registry lookup to `iface.Handler` conversions (alongside existing dashboard mapping until handlers are registered)

## 2. Migrate simple panels to Handler implementations

### SLO panels
- [x] 2.1 Create `panel/sloburnrate/api.go` with `Handler` and `FromAPI`/`ToAPI`
- [x] 2.2 Create `panel/sloburnrate/model.go` with `populateFromAPI`, `buildConfig` (exported)
- [x] 2.3 Create `panel/sloburnrate/schema.go` with `SchemaAttribute()`
- [x] 2.4 Create `panel/sloburnrate/api_test.go` with `contracttest.Run` call + any additional drilldown-specific assertions
- [x] 2.5 Remove SLO burn rate cases from `models_panels.go` switch and `toAPI()` cascade
- [x] 2.6 Register `sloburnrate.Handler{}` in `registry.go`
- [x] 2.7 Repeat 2.1–2.6 for `slo_overview` (include `contracttest.Run` in `api_test.go`)
- [x] 2.8 Repeat 2.1–2.6 for `slo_error_budget` (include `contracttest.Run` in `api_test.go`)

### Synthetics panels
- [x] 2.9 Repeat 2.1–2.6 for `synthetics_stats_overview` (include `contracttest.Run` in `api_test.go`)
- [x] 2.10 Repeat 2.1–2.6 for `synthetics_monitors` (include `contracttest.Run` in `api_test.go`)

### Control panels
- [x] 2.11 Create `panel/timeslider/` with Handler and PinnedHandler
- [x] 2.12 Create `panel/optionslist/` with Handler and PinnedHandler
- [x] 2.13 Create `panel/rangeslider/` with Handler and PinnedHandler
- [x] 2.14 Create `panel/esqlcontrol/` with Handler and PinnedHandler
- [x] 2.15 Remove control panel cases from `models_panels.go` and `toAPI()` cascade
- [x] 2.16 Register all 4 control handlers in `registry.go`
- [x] 2.17 Migrate `pinned_panels_mapping.go` to delegate to `handler.PinnedHandler()` for controls
- [x] 2.18 Remove hard-coded control cases from `pinned_panels_mapping.go`

### Markdown panel
- [x] 2.19 Create `panel/markdown/api.go` with Handler supporting both typed config and config_json paths
- [x] 2.20 Create `panel/markdown/model.go` with `populateFromAPI`, `buildConfigByValue`, `buildConfigByReference`
- [x] 2.21 Create `panel/markdown/schema.go` with `SchemaAttribute()`
- [x] 2.22 Create `panel/markdown/api_test.go` with `contracttest.Run` call covering by_value path; additional handwritten tests for by_reference and config_json fallback branch classification
- [x] 2.23 Remove markdown cases from `models_panels.go` switch and `toAPI()` cascade
- [x] 2.24 Register `markdown.Handler{}` in `registry.go`

### SLO alerts panel
- [x] 2.25 Create `panel/sloalerts/api.go` with Handler and `FromAPI`/`ToAPI`
- [x] 2.26 Create `panel/sloalerts/model.go` with `populateFromAPI`, `buildConfig`
- [x] 2.27 Create `panel/sloalerts/schema.go` with `SchemaAttribute()` (uses `panelkit.URLDrilldownSchema()` for drilldowns)
- [x] 2.28 Create `panel/sloalerts/api_test.go` with `contracttest.Run` call
- [x] 2.29 Remove `slo_alerts` cases from `models_panels.go` switch and `toAPI()` cascade
- [x] 2.30 Register `sloalerts.Handler{}` in `registry.go`

### Image panel
- [x] 2.31 Add `panelkit.ImageDrilldownSchema()` factory to `panelkit/schema.go` for the dashboard+URL drilldown variant
- [x] 2.32 Create `panel/image/api.go` with Handler and `FromAPI`/`ToAPI`
- [x] 2.33 Create `panel/image/model.go` with `populateFromAPI`, `buildConfig`
- [x] 2.34 Create `panel/image/schema.go` with `SchemaAttribute()` (uses `panelkit.ImageDrilldownSchema()`)
- [x] 2.35 Create `panel/image/api_test.go` with `contracttest.Run` call + handwritten tests for file vs URL src variants and dashboard vs URL drilldown variants
- [x] 2.36 Remove `image` cases from `models_panels.go` switch and `toAPI()` cascade
- [x] 2.37 Register `image.Handler{}` in `registry.go`

## 3. Refactor validators and schema assembly

- [x] 3.1 Refactor `panel_config_validator.go` to remove all simple panel cases from `panelConfigValidateDiags`
- [x] 3.2 Add registry dispatch loop: iterate `registry.AllHandlers()`, call `ValidatePanelConfig` per handler
- [x] 3.3 Refactor `getPanelSchema()` in `schema.go` to assemble panel config attributes from `registry.AllHandlers()`
- [x] 3.4 Remove hard-coded `panelConfigNames` slice; derive from `registry.ConfigNames()`
- [x] 3.5 Ensure conflict validators (`objectvalidator.ConflictsWith`) still reference correct sibling paths via derived names

## 4. State alignment delegation

- [x] 4.1 Refactor `alignPanelStateFromPlan` to call `handler.AlignStateFromPlan` for registered handlers
- [x] 4.2 Move handler-specific alignment functions out of `models_plan_state_alignment.go` into panel packages (or delete if no-op)

## 5. Config JSON defaulting

- [x] 5.1 Refactor `populatePanelConfigJSONDefaults` to dispatch to `handler.ClassifyJSON` and `handler.PopulateJSONDefaults`
- [x] 5.2 Remove hard-coded markdown special-casing from `panel_config_defaults.go` (Lens `attributes.*` defaulting unchanged; pending `dashboard-lens-contract`)

## 6. Verification

- [x] 6.1 `go build ./internal/kibana/dashboard/...` passes
- [x] 6.2 `go vet ./...` passes
- [x] 6.3 `go test ./internal/kibana/dashboard/panelkit/contracttest/...` passes (harness self-tests against stub handler)
- [x] 6.4 `go test ./internal/kibana/dashboard/panel/...` passes (all `contracttest.Run` calls pass for every migrated panel)
- [x] 6.5 `go test ./internal/kibana/dashboard/...` passes (all remaining unit tests)
- [x] 6.6 All migrated panel acceptance tests pass (`slo_*`, `synthetics_*`, control panel, markdown, image, slo_alerts)
- [x] 6.7 `make build` passes
