## 1. Shared schema helpers (prep)

- [x] 1.1 Add a `schema_shared_drilldowns.go` (or equivalent) factory function returning the shared `url_drilldown` nested-block schema with trigger validators
- [x] 1.2 Add a shared `time_range` nested-block schema helper mirroring the dashboard-root shape (`from`, `to`, `mode?`)
- [x] 1.3 Refactor `slo_burn_rate_config.drilldowns` and `slo_overview_config.drilldowns` schemas to consume the shared `url_drilldown` factory (no behavior change)
- [x] 1.4 Confirm `slo_burn_rate` / `slo_overview` unit + acceptance tests still pass unchanged after the refactor

## 2. Image panel (`type = "image"`)

- [x] 2.1 Add `schema_image_panel.go` with the `image_config` block, `src` discriminated sub-blocks, presentation attributes, and `drilldowns` discriminated union (`dashboard_drilldown` xor `url_drilldown`)
- [x] 2.2 Add validators: `src.file` xor `src.url` (exactly one); per-entry `dashboard_drilldown` xor `url_drilldown`; `object_fit` enum; drilldown trigger enums
- [x] 2.3 Add conflict validators (mutex with all other typed blocks and `config_json`)
- [x] 2.4 Update REQ-010 / panel-type-rejection table for `image` + `config_json`
- [x] 2.5 Create `models_image_panel.go` with the model and read/write helpers; apply REQ-009 null-preservation to optional fields and drilldown defaults (`use_filters`, `use_time_range`, `open_in_new_tab`, `encode_url`, `object_fit`)
- [x] 2.6 Wire into `mapPanelFromAPI` (`models_panels.go`) and the panel write dispatcher
- [x] 2.7 Unit tests in `models_image_panel_test.go`: both `src` variants, both drilldown variants, null-preservation, validator failures
- [x] 2.8 Description text under `internal/kibana/dashboard/descriptions/`

## 3. SLO alerts panel (`type = "slo_alerts"`)

- [x] 3.1 Add `slo_alerts_config` schema in `schema_slo_panel.go` reusing the shared `url_drilldown` block; require `slos`; add `len(slos) > 0` validator
- [x] 3.2 Add conflict validators (mutex with all other typed blocks and `config_json`)
- [x] 3.3 Update REQ-010 / panel-type-rejection table for `slo_alerts` + `config_json`
- [x] 3.4 Create `models_slo_alerts_panel.go` with the model and read/write helpers; apply REQ-009 null-preservation including `slo_instance_id = "*"` server-default
- [x] 3.5 Wire into `mapPanelFromAPI` and the panel write dispatcher
- [x] 3.6 Unit tests in `models_slo_alerts_panel_test.go`: round-trip, `slo_instance_id` null-preservation, drilldown round-trip, validator failures (empty `slos`); model-layer assertion that emitted API drilldown objects use `trigger = "on_open_panel_menu"`
- [x] 3.7 Description text under `internal/kibana/dashboard/descriptions/`

## 4. Discover session panel (`type = "discover_session"`)

- [x] 4.1 Add `schema_discover_session_panel.go` with `discover_session_config`, mutually exclusive `by_value` / `by_reference` sub-blocks, the single `tab` object with mutually exclusive `dsl` / `esql` sub-blocks, typed envelope fields, the shared `time_range` helper, the shared `url_drilldown` block, and `data_source_json` JSON attributes
- [x] 4.2 Add validators: `by_value` xor `by_reference`; `tab.dsl` xor `tab.esql`; `view_mode` enum; `density` enum; `header_row_height` (`"1".."5"|"auto"`) and `row_height` (`"1".."20"|"auto"`) string validators; numeric bounds on `rows_per_page` / `sample_size`; well-formed JSON for `data_source_json` (URL drilldown `trigger` is not practitioner-configurable when only one API trigger applies)
- [x] 4.3 Add semantic-equality plan modifiers on `data_source_json` (reuse existing JSON normalization helper)
- [x] 4.4 Add conflict validators (mutex with all other typed blocks and `config_json`)
- [x] 4.5 Update REQ-010 / panel-type-rejection table for `discover_session` + `config_json`
- [x] 4.6 Create `models_discover_session_panel.go` with the model, read/write helpers, and dashboard-time-range fallback for null panel-level `time_range` at write time
- [x] 4.7 Apply REQ-009 null-preservation to all optional fields, drilldown defaults, and panel-level `time_range`
- [x] 4.8 Compute `selected_tab_id` from API response when omitted in configuration
- [x] 4.9 Wire into `mapPanelFromAPI` and the panel write dispatcher
- [x] 4.10 Unit tests in `models_discover_session_panel_test.go`: `by_value` (both `dsl` and `esql` tabs), `by_reference` (with and without `selected_tab_id`), JSON normalization, validator failures, time_range inheritance from dashboard
- [x] 4.11 Description text under `internal/kibana/dashboard/descriptions/`

## 5. Spike: verify `references` requirement

- [x] 5.1 Manually create a `discover_session` `by_reference` panel against a live Kibana stack; confirm whether the dashboard request must include client-side `references` for the panel
- [x] 5.2 **N/A** — references not required for `discover_session` `by_reference` on Kibana 9.4.0; see `design.md` “Open questions” (also: top-level dashboard `references` rejected by API validation).
- [x] 5.3 Documented the finding in `design.md` “Open questions” (stack 9.4.0, methodology, HTTP outcomes).

## 6. Acceptance tests

- [x] 6.1 `acc_image_panels_test.go`: both `src` variants, at least one `dashboard_drilldown` and one `url_drilldown`
- [x] 6.2 `acc_slo_alerts_panels_test.go`: create an SLO via the SLO API in test setup, attach an `slo_alerts` panel referencing it, exercise `url_drilldown`
- [x] 6.3 `acc_discover_session_panels_test.go`: `by_value` with `dsl` tab; `by_value` with `esql` tab; `by_reference` (saved object created in test setup mirroring lens-by-reference fixture pattern)
- [x] 6.4 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...`, then `TF_ACC=1 go test ./internal/kibana/dashboard/...`

## 7. Examples and docs

- [x] 7.1 Add an image panel example under `examples/resources/elasticstack_kibana_dashboard/`
- [x] 7.2 Add an `slo_alerts` panel example
- [x] 7.3 Add a `discover_session` panel example covering both `by_value` (one tab variant) and `by_reference`
- [x] 7.4 Document in resource docs: `file_id` lifecycle (out of scope for this change, future resource), `data_source_json` shape pointers to the API schema, and the v1 single-`tab` constraint

## 8. Spec sync

- [x] 8.1 Run `openspec validate add-new-panels --strict`
- [x] 8.2 Run `make check-openspec`
