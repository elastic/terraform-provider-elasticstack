## 1. Schema and Models

- [x] 1.1 Extend `lensDashboardAppByValueModel` with optional fields for each supported typed by-value Lens chart block.
- [x] 1.2 Extend `getLensDashboardAppConfigSchema()` so `by_value` includes `config_json` plus supported typed Lens chart blocks using the existing chart schema helper functions where possible.
- [x] 1.3 Add or update validators so `lens_dashboard_app_config` still requires exactly one of `by_value` or `by_reference`.
- [x] 1.4 Add a by-value source validator so `lens_dashboard_app_config.by_value` requires exactly one source: `config_json` or one typed chart block.
- [x] 1.5 Update schema descriptions to explain that typed by-value chart blocks send `lens-dashboard-app` API config directly and do not create `type = "vis"` panels.

## 2. Converter Adapter

- [x] 2.1 Add helper logic that identifies which typed by-value chart block is set on a `lensDashboardAppByValueModel`.
- [x] 2.2 Add a build adapter that creates a scratch `panelModel`, reuses the matching existing `lensVisualizationConverter.buildAttributes`, and bridges `KbnDashboardPanelTypeVisConfig0` JSON into `KbnDashboardPanelTypeLensDashboardAppConfig0`.
- [x] 2.3 Update `lensDashboardAppByValueToAPI` to use raw `config_json` when selected and the typed chart adapter when a typed by-value source is selected.
- [x] 2.4 Add a read adapter that bridges `lens-dashboard-app` by-value config JSON into `KbnDashboardPanelTypeVisConfig0`, detects the Lens chart type, and reuses the matching converter to populate the selected typed by-value chart block.
- [x] 2.5 Update `populateLensDashboardAppByValueFromAPI` so prior typed by-value state is preserved when read-back can be represented by the same typed chart block, otherwise falling back to `config_json`.
- [x] 2.6 Keep raw `by_value.config_json` preservation behavior unchanged for configurations that selected raw JSON.

## 3. Validation and Drift Behavior

- [x] 3.1 Add plan/unit coverage that rejects `by_value` with both `config_json` and a typed chart block.
- [x] 3.2 Add plan/unit coverage that rejects `by_value` with no source.
- [x] 3.3 Add unit coverage that raw `by_value.config_json` still maps directly to API config and preserves practitioner JSON subset behavior.
- [x] 3.4 Add unit coverage that typed by-value read-back keeps the typed representation when the returned chart can be decoded by the matching converter.
- [x] 3.5 Add unit coverage that typed by-value read-back falls back to `config_json` when the API response cannot be represented by the prior typed chart block.

## 4. Chart Coverage

- [x] 4.1 Add adapter unit tests for representative no-ESQL and ES|QL chart families, at minimum metric, XY, pie, and waffle.
- [x] 4.2 Verify each exposed typed by-value chart block is backed by a chart struct present in both `KbnDashboardPanelTypeVisConfig0` and `KbnDashboardPanelTypeLensDashboardAppConfig0`.
- [x] 4.3 Add acceptance coverage for at least one typed by-value `lens-dashboard-app` chart and require a second apply with an empty plan.
- [x] 4.4 Add import or read-back coverage confirming typed by-value state does not populate panel-level `config_json`.

## 5. Documentation and Spec Alignment

- [x] 5.1 Update generated/resource documentation for `lens_dashboard_app_config.by_value` typed chart sources.
- [x] 5.2 Update examples or acceptance fixtures to show a typed by-value `lens-dashboard-app` chart.
- [x] 5.3 Run OpenSpec validation for `typed-lens-dashboard-app-by-value`.
- [x] 5.4 Run targeted dashboard unit tests for schema validators and lens-dashboard-app converters.
- [x] 5.5 Run targeted dashboard acceptance tests if a Kibana test stack is available; otherwise document that acceptance tests were not run.
