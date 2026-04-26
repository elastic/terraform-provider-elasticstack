## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `options_list_control_config` is valid only for panels with `type = "options_list_control"`, is mutually exclusive with all other panel configuration blocks and with `config_json`, and that `search_technique` is restricted to `prefix`, `wildcard`, or `exact` when set. Schema validation SHALL also enforce that `synthetics_monitors_config` is valid only for panels with `type = "synthetics_monitors"` and is mutually exclusive with all other panel configuration blocks and with `config_json`. Schema validation SHALL enforce that `synthetics_stats_overview_config` is valid only for panels with `type = "synthetics_stats_overview"` and is mutually exclusive with all other typed panel config blocks and with `config_json`. Schema validation SHALL enforce that `lens_dashboard_app_config` is valid only for panels with `type = "lens-dashboard-app"`, is mutually exclusive with all other panel configuration blocks, and that exactly one of `by_value` or `by_reference` is set. When `by_value` is set, schema validation SHALL enforce that exactly one by-value source is set: either `config_json` or one supported typed Lens chart block.

REQ-006 is extended to include:

- `options_list_control_config` SHALL be valid only for panels with `type = "options_list_control"`.
- `options_list_control_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- The `search_technique` attribute within `options_list_control_config` SHALL be restricted to the values `prefix`, `wildcard`, and `exact` when set; any other value SHALL be rejected at plan time.
- `synthetics_monitors_config` SHALL be valid only for panels with `type = "synthetics_monitors"`.
- `synthetics_monitors_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- `synthetics_stats_overview_config` SHALL only be valid on panels with `type = "synthetics_stats_overview"`.
- `synthetics_stats_overview_config` SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`.
- `lens_dashboard_app_config` SHALL be valid only for panels with `type = "lens-dashboard-app"`.
- `lens_dashboard_app_config` SHALL be mutually exclusive with all other panel configuration blocks.
- Within `lens_dashboard_app_config`, exactly one of `by_value` or `by_reference` SHALL be set; setting both or neither SHALL be rejected at plan time.
- Within `lens_dashboard_app_config.by_value`, exactly one by-value source SHALL be set: either `config_json` or one supported typed Lens chart block.
- `by_value.config_json` SHALL be valid only as the selected by-value source.
- `by_reference.ref_id` and `by_reference.time_range` SHALL be required when `by_reference` is set.
- `by_reference.time_range.mode` SHALL be restricted to `absolute` or `relative` when set.

#### Scenario: options_list_control_config rejected for non-options_list_control panel

- GIVEN a panel with `type = "vis"` and `options_list_control_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: synthetics_monitors_config rejected for non-synthetics_monitors panel

- GIVEN a panel with `type = "vis"` and `synthetics_monitors_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: synthetics_monitors_config conflicts with other typed blocks

- GIVEN a panel entry with `type = "synthetics_monitors"` that sets both `synthetics_monitors_config` and any other typed config block (e.g. `markdown_config`)
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic indicating the conflicting blocks are mutually exclusive

#### Scenario: synthetics_stats_overview_config rejected for non-synthetics_stats_overview panel

- GIVEN a panel with `type = "lens"` and `synthetics_stats_overview_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: lens_dashboard_app_config rejected for non-lens-dashboard-app panel

- GIVEN a panel whose `type` is not `lens-dashboard-app` and `lens_dashboard_app_config` is set (such as `type = "markdown"`, or `type = "vis"` with no other valid `vis` panel configuration / missing vis chart)
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call (for example `Missing vis panel configuration` when `type = "vis"` and the vis panel is incomplete, and/or schema-level `Invalid Configuration` when `lens_dashboard_app_config` is not allowed for the current `type`)

#### Scenario: Both sub-blocks set simultaneously

- GIVEN a `lens_dashboard_app_config` block with both `by_value` and `by_reference` set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that `by_value` and `by_reference` are mutually exclusive

#### Scenario: Neither sub-block set

- GIVEN a `lens_dashboard_app_config` block with neither `by_value` nor `by_reference` set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one of `by_value` or `by_reference` must be set

#### Scenario: Multiple by-value sources set simultaneously

- GIVEN a `lens_dashboard_app_config.by_value` block with both `config_json` and a typed Lens chart block set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one by-value source must be set

#### Scenario: No by-value source set

- GIVEN a `lens_dashboard_app_config.by_value` block with no `config_json` and no typed Lens chart block
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one by-value source must be set

### Requirement: `lens-dashboard-app` panel behavior (REQ-035)

For `type = "lens-dashboard-app"` panels, the resource SHALL accept `lens_dashboard_app_config` with exactly one of the `by_value` or `by_reference` sub-blocks set. Within `by_value`, practitioners SHALL configure exactly one by-value source: either `config_json` containing a JSON object that maps directly to the generated by-value `KbnDashboardPanelTypeLensDashboardApp.config` union, or one supported typed Lens chart block. Within `by_reference`, the `ref_id` and `time_range` attributes are required. The optional by-reference attributes `references_json`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns_json` MAY be set.

**On write (create and update):**

For by-value panels authored through `config_json`, the resource SHALL map `by_value.config_json` directly to the panel `config` object without wrapping it in an `attributes` object and without splitting out references. The JSON object SHALL be expected to match one of the current generated by-value Lens chart schemas, including that schema's required fields such as chart `type` and `time_range` where applicable.

For by-value panels authored through a supported typed Lens chart block, the resource SHALL convert that typed chart model into the matching generated by-value Lens chart schema and SHALL send the resulting object directly as the panel API `config`. The provider SHALL NOT wrap the chart object in an `attributes` object and SHALL NOT change the dashboard panel discriminator to `vis`.

For by-reference panels, the resource SHALL set the API `config.ref_id` field from `by_reference.ref_id`, set the API `config.time_range` object from `by_reference.time_range`, and include `references`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns` only when their corresponding Terraform attributes are set. `references_json` SHALL map to the API `references` array of `{ id, name, type }` objects. A saved Lens visualization reference SHALL be represented through `references_json`, typically with a reference whose `name` matches `ref_id`, whose `type` is `lens`, and whose `id` is the saved object ID.

**On read:**

The resource SHALL classify the API `config` JSON object in this order (relying on the raw object, not only generated union decode, which does not enforce a true oneOf on the wire): (1) **By-value:** if the object has a non-empty string at top-level `type` (the by-value Lens chart discriminator), the resource SHALL leave `by_reference` unset and populate `by_value` from the API read, including when `ref_id` and `time_range` are also present. When prior plan or state selected a supported typed by-value chart block and the API response can be represented by that same typed chart block, the resource SHALL repopulate that typed chart block. Otherwise, the resource SHALL populate `by_value.config_json` from the API read, including the practitioner string preservation rule in the next paragraph when plan or state includes a prior `by_value.config_json` object. (2) **By-reference:** otherwise, if the object omits that chart discriminator and has non-empty `ref_id` and a `time_range` with non-empty `from` and `to`, the resource SHALL populate `by_reference` and leave `by_value` unset. (3) **Neither (1) nor (2):** if prior plan or state had `by_reference`, the resource SHALL preserve that prior `by_reference` block per REQ-009 and SHALL NOT silently mode-flip to `by_value`. (4) Otherwise, the resource SHALL populate `by_value.config_json` from the API read (and the same preservation rule when applicable). Fields absent from the API response SHALL not be forced into state from the API response alone. Optional by-reference attributes SHALL also follow REQ-009 panel read seeding and alignment so prior practitioner intent is preserved when the API omits or differs on optional values.

`by_value.config_json`, `by_reference.references_json`, and `by_reference.drilldowns_json` SHALL use semantic JSON equality for plan comparison. API-injected field ordering SHALL NOT create spurious plan diffs. For `by_value.config_json`, when a read of the Kibana `config` returns additional key paths and values the practitioner’s object did not set (Kibana default or enrichment) while every value path the practitioner’s `config_json` object sets is still present in the API object with the same value, the provider SHALL preserve the practitioner’s `by_value.config_json` string in state; the implementation may treat top-level `styling` as rewritable by Kibana, optional empty `filters` (including `null` or omission), a default KQL `query` (only `language` and/or `expression: ""` matching API omission), and related cases consistent with a non-destructive next write; if the response changes a value the user set, or the prior object cannot be read as a value-subset of the API in this sense, the provider SHALL use the read-back value to avoid a destructive next write. For ordered JSON arrays on that value-subset path, the API may only **append** after the practitioner’s last index; reordered or prepended content relative to the practitioner’s array is not treated as a safe enrichment match.

The `lens-dashboard-app` panel type is distinct from the existing `vis` Lens panel path. Typed by-value Lens chart blocks under `lens_dashboard_app_config.by_value` SHALL use the `lens-dashboard-app` panel discriminator and `KbnDashboardPanelTypeLensDashboardApp.config` by-value shape. Existing top-level typed Lens panel blocks such as `xy_chart_config` and `metric_chart_config` SHALL remain valid only for `type = "vis"` panels. Existing `type = "vis"` Lens panel behavior SHALL remain unchanged.

#### Scenario: Creation of a by-reference lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_reference.ref_id = "panel_0"`
  - `lens_dashboard_app_config.by_reference.references_json = "[{\"id\":\"abc-123\",\"name\":\"panel_0\",\"type\":\"lens\"}]"`
  - `lens_dashboard_app_config.by_reference.time_range.from = "now-15m"`
  - `lens_dashboard_app_config.by_reference.time_range.to = "now"`
  - `lens_dashboard_app_config.by_reference.title = "My Shared Visualization"`
- WHEN the resource is created
- THEN the provider SHALL send a panel payload with `config.ref_id = "panel_0"`, the references array, the time range object, and `title = "My Shared Visualization"` to the Kibana dashboard API
- AND the panel SHALL appear in state with `by_reference.ref_id = "panel_0"` and `by_value` as null
- AND the provider SHALL NOT populate panel-level `config_json` for this panel in state

#### Scenario: Creation of a raw by-value lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_value.config_json = "<valid generated API Lens chart config JSON>"`
- WHEN the resource is created
- THEN the provider SHALL send the decoded JSON object directly as the panel API `config`
- AND the panel SHALL appear in state with `by_value.config_json` populated and `by_reference` as null

#### Scenario: Creation of a typed by-value lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - one supported typed Lens chart block under `lens_dashboard_app_config.by_value`
- WHEN the resource is created
- THEN the provider SHALL convert the typed chart block into the matching generated by-value Lens chart object
- AND the provider SHALL send that chart object directly as the panel API `config`
- AND the panel SHALL appear in state with that typed by-value chart block populated and `by_reference` as null
- AND the provider SHALL NOT populate panel-level `config_json` for this panel in state

#### Scenario: by-reference panel with required time_range and optional drilldowns_json

- GIVEN a `lens-dashboard-app` panel in by-reference mode with:
  - `lens_dashboard_app_config.by_reference.ref_id = "panel_0"`
  - `lens_dashboard_app_config.by_reference.time_range.from = "now-7d"`
  - `lens_dashboard_app_config.by_reference.time_range.to = "now"`
  - `lens_dashboard_app_config.by_reference.time_range.mode = "relative"`
  - `lens_dashboard_app_config.by_reference.drilldowns_json = "[{\"type\":\"url_drilldown\",\"trigger\":\"on_click_value\",\"label\":\"Open\",\"url\":\"https://example.com\"}]"`
- WHEN the resource is created or updated
- THEN the provider SHALL include the `time_range` object and `drilldowns` array in the API payload
- AND on read-back the provider SHALL repopulate both from the API response

#### Scenario: Invalid mixed configuration — both sub-blocks set

- GIVEN a `lens_dashboard_app_config` block with both `by_value` and `by_reference` configured
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating mutual exclusivity

#### Scenario: Invalid mixed by-value configuration

- GIVEN a `lens_dashboard_app_config.by_value` block with both `config_json` and a typed Lens chart block configured
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one by-value source must be set

#### Scenario: Read-back detects by-reference mode from API response

- GIVEN a managed `lens-dashboard-app` panel authored in by-reference mode
- WHEN Kibana returns the panel config with `ref_id` and `time_range`
- THEN the provider SHALL populate `by_reference` in state and leave `by_value` as null
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back preserves absent optional by-reference fields

- GIVEN a managed `lens-dashboard-app` panel in by-reference mode that omits `description`, `hide_title`, and `hide_border`
- WHEN Kibana returns the panel without those optional fields
- THEN the provider SHALL keep those optional fields null/unset in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back preserves typed by-value representation

- GIVEN a managed `lens-dashboard-app` panel authored with a supported typed Lens chart block under `by_value`
- WHEN Kibana returns a by-value chart config with the same chart discriminator and the response can be represented by that typed chart block
- THEN the provider SHALL populate that typed chart block in state
- AND the provider SHALL NOT replace it with `by_value.config_json`

#### Scenario: Read-back falls back to by_value config_json when prior typed by-value block cannot be preserved

- GIVEN a managed `lens-dashboard-app` panel with prior state that selected a supported typed Lens chart block under `by_value` (and not `by_value.config_json`)
- WHEN Kibana returns a by-value chart `config` that cannot be represented in that same typed chart block
- THEN the provider SHALL populate `by_value.config_json` from the API read
- AND the provider SHALL NOT keep the prior typed chart block in state when the response cannot be round-tripped to it

#### Scenario: Read-back preserves raw by-value representation

- GIVEN a managed `lens-dashboard-app` panel authored with `by_value.config_json`
- WHEN Kibana returns a by-value chart config with a top-level chart `type`
- THEN the provider SHALL populate `by_value.config_json` in state
- AND the provider SHALL NOT convert it to a typed by-value chart block
