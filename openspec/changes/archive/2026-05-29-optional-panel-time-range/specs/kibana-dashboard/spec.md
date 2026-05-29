## MODIFIED Requirements

### Requirement: XY chart panel behavior and typed `vis` `time_range` (REQ-013)

For **typed** `vis` panels (those built through the provider's typed `*_config` blocks and the shared typed visualization write path, not panels managed solely through raw `config_json`), the resource SHALL expose `time_range` as an optional flat sibling attribute on every typed Lens chart block (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`). The attribute SHALL match the dashboard-level `time_range` shape: required `from` (string), required `to` (string), and optional `mode` enum (`absolute` | `relative`).

When the chart-level `time_range` is null in configuration and state, the provider SHALL omit `time_range` from the API payload entirely. The provider SHALL NOT inherit the dashboard-level `time_range` and SHALL NOT use any hardcoded fallback window. Kibana will apply its own default (global dashboard time range) for panels with no panel-level override.

When the chart-level `time_range` is set in configuration, the provider SHALL pass the configured values to the API verbatim, overriding the dashboard-level value for that panel only.

The `vis_config.by_reference` block SHALL expose `time_range` as an **optional** attribute (same shape: required `from`, required `to`, optional `mode`). When `time_range` is null in `by_reference` configuration, the provider SHALL omit it from the API payload. When set, the provider SHALL send it verbatim.

For XY chart `vis` panels specifically, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, and at least one `layers` entry. The axis object SHALL use `x`, optional primary `y`, and optional secondary `y2`; `axis.x.domain_json` SHALL represent the X-axis domain, and each configured Y axis SHALL require `domain_json`. Each layer SHALL represent either a data layer or a reference-line layer, not both. **`query` SHALL be optional** on the XY chart schema so that ES|QL XY panels (which carry no `query` in the API) are valid without a dummy query block.

REQ-025 governs raw `config_json` `vis` panels; the typed-vs-raw distinction is unchanged.

#### Scenario: Typed `vis` write omits time_range when chart time_range is null

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is null in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL NOT include `time_range` in the API payload for that panel

#### Scenario: Typed `vis` write uses configured chart-level time_range when set

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is set to `{ from = "now-30d", to = "now-1d" }` in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL set `time_range` on the API payload to the chart-level value `{ from = "now-30d", to = "now-1d" }`

#### Scenario: by_reference write omits time_range when not configured

- GIVEN a `vis_config.by_reference` panel on create or update where `time_range` is null in configuration
- WHEN the provider builds the API payload
- THEN it SHALL NOT include `time_range` in the by-reference config payload

#### Scenario: by_reference write sends time_range when configured

- GIVEN a `vis_config.by_reference` panel on create or update where `time_range` is set to `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the API payload
- THEN it SHALL include `time_range = { from = "now-7d", to = "now" }` in the by-reference config payload

#### Scenario: Read preserves null time_range when API returns none

- GIVEN a typed `vis` panel where `time_range` is null in Terraform state
- AND the Kibana API returns no `time_range` field for that panel on read
- WHEN the provider processes the read response
- THEN it SHALL keep `time_range` as null in state (no drift)

#### Scenario: XY panel requires layers

- GIVEN an XY chart panel configuration
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL require at least one layer and the fixed XY sub-blocks needed by the schema

#### Scenario: ES|QL XY panel omits query

- GIVEN an XY chart panel configured for ES|QL mode (no usable query expression)
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be accepted without a `query` block

### Requirement: Chart-level `time_range` null-preservation (REQ-040)

The resource SHALL preserve practitioner intent for the chart-level `time_range` block on every typed Lens chart reachable under `panels[].vis_config.by_value.<chart>_config` (for `type = "vis"`), using the same null-preservation pattern as REQ-009 for `time_range.mode`.

When the Kibana API response omits chart-level `time_range` (or returns an empty/zero-valued time range struct), the provider SHALL leave state's chart-level `time_range` as null. When the API returns a populated chart-level `time_range`, the provider SHALL populate state from the API response (subject to the `time_range.mode` null-preservation rule below).

The chart-level `time_range.mode` attribute SHALL follow the same null-preservation rule as the dashboard-level `time_range.mode` in REQ-009: when prior state has `mode = null` and the API response omits or returns no usable mode, state SHALL preserve null rather than overwriting with a default.

#### Scenario: Chart time_range stays null when API omits it

- GIVEN a `vis` panel with a typed Lens chart block under `vis_config.by_value` whose prior state has `time_range = null`
- AND the Kibana API response omits `time_range` on that chart root
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range = null` on the chart panel

#### Scenario: Chart time_range mode null-preservation

- GIVEN a typed Lens chart panel whose prior state has `time_range = { from = "now-7d", to = "now", mode = null }`
- AND the Kibana API response omits `mode` on the chart-root `time_range`
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range.mode = null`
