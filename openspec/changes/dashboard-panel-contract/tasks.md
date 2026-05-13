## 0. contracttest harness

- [ ] 0.1 Create `dashboard/panelkit/contracttest/parse.go` — raw JSON string → `kbapi.DashboardPanelItem` adapter
- [ ] 0.2 Create `dashboard/panelkit/contracttest/roundtrip.go` — `FromAPI → ToAPI` stability assertion with JSON diff output on failure
- [ ] 0.3 Create `dashboard/panelkit/contracttest/nullpreserve.go` — walk `handler.SchemaAttribute()` to collect optional leaves; generate three sub-tests per field (prior null, prior known, prior nil)
- [ ] 0.4 Create `dashboard/panelkit/contracttest/reflection.go` — `HasConfig`/`ClearConfig` post-condition assertions using `panelkit.HasConfig`
- [ ] 0.5 Create `dashboard/panelkit/contracttest/schema.go` — structural assertions: outer attribute is Optional, Required leaves are present in fixture, `ValidatePanelConfig` rejects zeroed Required fields
- [ ] 0.6 Create `dashboard/panelkit/contracttest/harness.go` — `Run(t, handler, config)` entry point orchestrating the above; `Config` struct with `FullAPIResponse string` and optional `SkipFields []string`
- [ ] 0.7 Verify harness compiles and `contracttest.Run` fails with a clear error when given a no-op stub handler

## 1. Infrastructure

- [ ] 1.1 Create `dashboard/panel/iface/iface.go` with `Handler` and `PinnedHandler` interfaces
- [ ] 1.2 Create `dashboard/panelkit/grid.go` with `GridFromAPI`, `GridToAPI`, `IDFromAPI`, `IDToAPI`
- [ ] 1.3 Create `dashboard/panelkit/nullpreserve.go` with `PreserveString`, `PreserveBool`, `PreserveList`, `PreserveFloat64`
- [ ] 1.4 Create `dashboard/panelkit/reflection.go` with `HasConfig`, `ClearConfig`, `SetConfig` using tfsdk tag matching
- [ ] 1.5 Verify `panelkit` reflection panics at init if any registered handler has a block name with no matching tfsdk tag
- [ ] 1.6 Create `dashboard/registry.go` with `panelHandlers` slice, `panelTypeToHandler` map, `panelConfigNames` derivation
- [ ] 1.7 Create `dashboard/router.go` with `mapPanelFromAPI` and `panelModelToAPI` delegating to registry

## 2. Migrate simple panels to Handler implementations

### SLO panels
- [ ] 2.1 Create `panel/sloburnrate/api.go` with `Handler` and `FromAPI`/`ToAPI`
- [ ] 2.2 Create `panel/sloburnrate/model.go` with `populateFromAPI`, `buildConfig` (exported)
- [ ] 2.3 Create `panel/sloburnrate/schema.go` with `SchemaAttribute()`
- [ ] 2.4 Create `panel/sloburnrate/api_test.go` with `contracttest.Run` call + any additional drilldown-specific assertions
- [ ] 2.5 Remove SLO burn rate cases from `models_panels.go` switch and `toAPI()` cascade
- [ ] 2.6 Register `sloburnrate.Handler{}` in `registry.go`
- [ ] 2.7 Repeat 2.1–2.6 for `slo_overview` (include `contracttest.Run` in `api_test.go`)
- [ ] 2.8 Repeat 2.1–2.6 for `slo_error_budget` (include `contracttest.Run` in `api_test.go`)

### Synthetics panels
- [ ] 2.9 Repeat 2.1–2.6 for `synthetics_stats_overview` (include `contracttest.Run` in `api_test.go`)
- [ ] 2.10 Repeat 2.1–2.6 for `synthetics_monitors` (include `contracttest.Run` in `api_test.go`)

### Control panels
- [ ] 2.11 Create `panel/timeslider/` with Handler and PinnedHandler
- [ ] 2.12 Create `panel/optionslist/` with Handler and PinnedHandler
- [ ] 2.13 Create `panel/rangeslider/` with Handler and PinnedHandler
- [ ] 2.14 Create `panel/esqlcontrol/` with Handler and PinnedHandler
- [ ] 2.15 Remove control panel cases from `models_panels.go` and `toAPI()` cascade
- [ ] 2.16 Register all 4 control handlers in `registry.go`
- [ ] 2.17 Migrate `pinned_panels_mapping.go` to delegate to `handler.PinnedHandler()` for controls
- [ ] 2.18 Remove hard-coded control cases from `pinned_panels_mapping.go`

### Markdown panel
- [ ] 2.19 Create `panel/markdown/api.go` with Handler supporting both typed config and config_json paths
- [ ] 2.20 Create `panel/markdown/model.go` with `populateFromAPI`, `buildConfigByValue`, `buildConfigByReference`
- [ ] 2.21 Create `panel/markdown/schema.go` with `SchemaAttribute()`
- [ ] 2.22 Create `panel/markdown/api_test.go` with `contracttest.Run` call covering by_value path; additional handwritten tests for by_reference and config_json fallback branch classification
- [ ] 2.23 Remove markdown cases from `models_panels.go` switch and `toAPI()` cascade
- [ ] 2.24 Register `markdown.Handler{}` in `registry.go`

## 3. Refactor validators and schema assembly

- [ ] 3.1 Refactor `panel_config_validator.go` to remove all simple panel cases from `panelConfigValidateDiags`
- [ ] 3.2 Add registry dispatch loop: iterate `registry.AllHandlers()`, call `ValidatePanelConfig` per handler
- [ ] 3.3 Refactor `getPanelSchema()` in `schema.go` to assemble panel config attributes from `registry.AllHandlers()`
- [ ] 3.4 Remove hard-coded `panelConfigNames` slice; derive from `registry.ConfigNames()`
- [ ] 3.5 Ensure conflict validators (`objectvalidator.ConflictsWith`) still reference correct sibling paths via derived names

## 4. State alignment delegation

- [ ] 4.1 Refactor `alignPanelStateFromPlan` to call `handler.AlignStateFromPlan` for registered handlers
- [ ] 4.2 Move handler-specific alignment functions out of `models_plan_state_alignment.go` into panel packages (or delete if no-op)

## 5. Config JSON defaulting

- [ ] 5.1 Refactor `populatePanelConfigJSONDefaults` to dispatch to `handler.ClassifyJSON` and `handler.PopulateJSONDefaults`
- [ ] 5.2 Remove hard-coded markdown and lens special-casing from `panel_config_defaults.go`

## 6. Verification

- [ ] 6.1 `go build ./internal/kibana/dashboard/...` passes
- [ ] 6.2 `go vet ./...` passes
- [ ] 6.3 `go test ./internal/kibana/dashboard/panelkit/contracttest/...` passes (harness self-tests against stub handler)
- [ ] 6.4 `go test ./internal/kibana/dashboard/panel/...` passes (all `contracttest.Run` calls pass for every migrated panel)
- [ ] 6.5 `go test ./internal/kibana/dashboard/...` passes (all remaining unit tests)
- [ ] 6.6 All migrated panel acceptance tests pass (`slo_*`, `synthetics_*`, control panel, markdown)
- [ ] 6.7 `make build` passes
