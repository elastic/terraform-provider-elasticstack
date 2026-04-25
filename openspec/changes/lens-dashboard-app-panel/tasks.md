# Tasks: `lens-dashboard-app` Panel Support

## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [x] 1.2 On completion, sync delta into canonical spec or archive

## 2. Schema

- [x] 2.1 Add `lens_dashboard_app_config` schema block to `internal/kibana/dashboard/schema.go` with `by_value` and `by_reference` sub-blocks
- [x] 2.2 Add required `by_reference.time_range` nested block with required `from` / `to` and optional `mode`
- [x] 2.3 Add schema validators to enforce that exactly one of `by_value` or `by_reference` is set within `lens_dashboard_app_config`
- [x] 2.4 Add schema validation that `lens_dashboard_app_config` is only valid when `type = "lens-dashboard-app"` (REQ-006 extension)
- [x] 2.5 Add schema validation that `lens_dashboard_app_config` is mutually exclusive with all other panel config blocks (REQ-006 extension)
- [x] 2.6 Add `by_value.config_json` as a required normalized JSON string
- [x] 2.7 Add by-reference attributes: required `ref_id`, optional `references_json`, optional `title`, optional `description`, optional `hide_title`, optional `hide_border`, and optional `drilldowns_json`

## 3. Models

- [x] 3.1 Extend `panelModel` struct in `models_panels.go` with a `LensDashboardAppConfig` field
- [x] 3.2 Create `models_lens_dashboard_app_panel.go` with model structs for `lensDashboardAppConfigModel`, `lensDashboardAppByValueModel`, `lensDashboardAppByReferenceModel`, and `lensDashboardAppTimeRangeModel`

## 4. Converters

- [x] 4.1 Implement write converter for `by_value` mode by decoding `by_value.config_json` and assigning it directly to `KbnDashboardPanelTypeLensDashboardApp.Config`
- [x] 4.2 Implement write converter for `by_reference` mode using `KbnDashboardPanelTypeLensDashboardAppConfig1` (`ref_id`, required `time_range`, optional `references`, display fields, and `drilldowns`)
- [x] 4.3 Implement read converter with mode detection: root chart `type` string (by-value) first; else by-reference when no chart discriminator and `ref_id` plus `time_range.from`/`to`; else if ambiguous and prior had `by_reference`, preserve it per REQ-009; else populate `by_value.config_json` from the full config
- [x] 4.4 Implement read converter population of by-reference optional fields (`references_json`, `title`, `description`, `hide_title`, `hide_border`, `drilldowns_json`)
- [x] 4.5 Update the panel write-path dispatcher in `models_panels.go` to handle `lens-dashboard-app` type via `lens_dashboard_app_config`
- [x] 4.6 Update the panel read-path dispatcher in `models_panels.go` to populate `lens_dashboard_app_config` on read-back

## 5. Validation

- [x] 5.1 Update panel-level `config_json` write-path error message in `models_panels.go` to explicitly name `lens-dashboard-app` as unsupported (REQ-025 update)
- [x] 5.2 Add validator or plan modifier to enforce mutual exclusivity of `by_value` and `by_reference` sub-blocks at plan time (done in task 2.3: `lensDashboardAppConfigModeValidator` on `lens_dashboard_app_config`)
- [x] 5.3 Validate `by_reference.time_range.mode` accepts only `absolute` or `relative` when set (done in task 2.2: `stringvalidator.OneOf` on `time_range.mode`)
- [x] 5.4 Update resource descriptions and documentation for the new block and its attributes

## 6. Testing

- [x] 6.1 Add acceptance tests for `lens-dashboard-app` panel creation in by-reference mode with required `ref_id` and `time_range`; include coverage that sets optional `references_json` for a typical saved-object reference-wiring case (see REQ-035). **Also:** `TestAccResourceDashboardLensDashboardAppByReference` includes a second apply (title / time range / flags) and a follow-up import with `ImportStateVerifyIgnore` for `references_json` / `panels.0.id` where read order can differ.
- [x] 6.2 Add acceptance tests for `lens-dashboard-app` panel creation in by-value mode with required `config_json` — `TestAccResourceDashboardLensDashboardAppByValue` (metric inline chart with `jsonencode`). Read-back stability uses `preservePriorLensByValueConfigJSON` in `models_lens_dashboard_app_converters.go`: when Kibana’s stored `config` is a value-expansion of the practitioner’s object (plus unit tests for embed and write/read). If a Kibana version rewrites a user-authored field (e.g. renames a chart `type` string) so it is not a value-subset, plan may still show drift; see file comment in `acc_lens_dashboard_app_panels_test.go`.
- [x] 6.3 Add acceptance tests for by-reference panel with optional `title`, `description`, `hide_title`, and `hide_border`
- [x] 6.4 Add acceptance or unit coverage for by-reference `drilldowns_json`
- [x] 6.5 Add acceptance tests for by-reference `time_range.mode` (valid `absolute` / `relative` values) — `TestAccResourceDashboardLensDashboardAppByReferenceAbsoluteTimeMode` includes import after the initial apply (same ignore list as the main by-ref test for `references_json` / `panels.0.id` when unneeded on this fixture).
- [x] 6.6 Add acceptance or plan tests for REQ-006: reject `lens_dashboard_app_config` when panel `type` is not `lens-dashboard-app` — plan fixtures: `wrong_type` (`type = markdown`), `wrong_type_vis` (`type = vis` with only `lens_dashboard_app_config` per delta spec scenario; diagnostic may be `Missing vis panel…` and/or `can only be set when…`)
- [x] 6.7 Add acceptance or plan tests for REQ-006: reject `lens_dashboard_app_config` set together with any other panel config block on the same panel (e.g. `markdown_config`, `config_json` where disallowed, or a `vis` chart block)
- [x] 6.8 Add acceptance or plan tests for REQ-006: reject `type = "lens-dashboard-app"` when `lens_dashboard_app_config` is missing (`RequiredIf` on the block)
- [x] 6.9 Add acceptance or plan tests for plan-time rejection of invalid `by_reference.time_range.mode` (value other than `absolute` or `relative` when set)
- [x] 6.10 Add acceptance tests for plan-time validation rejection when both `by_value` and `by_reference` are set simultaneously
- [x] 6.11 Add acceptance tests for plan-time validation rejection when neither `by_value` nor `by_reference` is set
- [x] 6.12 Add optional unit tests for `lensDashboardAppConfigModeValidator` (both set, neither set, unknown branches)
- [x] 6.13 Add unit tests for the `by_value` write converter ensuring `config_json` is sent directly as API `config`
- [x] 6.14 Add unit tests for the `by_reference` write converter ensuring `ref_id`, `references`, required `time_range`, display fields, and `drilldowns` map to API `config`
- [x] 6.15 Add unit tests for the read converter mode detection and field population (by-value path, including ambiguous fallback to `by_value` when there is no prior `by_reference`)
- [x] 6.16 Add unit tests for the read converter mode detection and field population (by-reference path, including ambiguous API response with prior `by_reference` preserved per REQ-009)
- [x] 6.17 Add coverage that practitioner-authored **panel-level** `config_json` with `type = "lens-dashboard-app"` is rejected: **apply/write-path** error (`Unsupported panel type for config_json` with REQ-025 message to use `lens_dashboard_app_config`) when `lens_dashboard_app_config` is absent; **plan-time** rejection remains via the `config_json` type allowlist + conflicts when applicable (distinguish from the write-path explicit diagnostic, which targets bypass/state edge cases)
