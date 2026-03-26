## ADDED Requirements

### Requirement: Synthetics stats overview panel behavior (REQ-033)

For `type = "synthetics_stats_overview"` panels, the resource SHALL accept an optional `synthetics_stats_overview_config` block. All fields within the block are optional; the panel is valid with an entirely absent or empty config block, in which case it displays monitoring statistics for all Elastic Synthetics monitors visible within the space.

The `synthetics_stats_overview_config` block SHALL expose the following optional attributes:

- `title` (string): display title shown in the panel header.
- `description` (string): descriptive text for the panel.
- `hide_title` (bool): when true, suppresses the panel title in the dashboard.
- `hide_border` (bool): when true, suppresses the panel border in the dashboard.
- `drilldowns` (list of objects): URL drilldown actions attached to the panel. Each drilldown object SHALL contain:
  - `url` (string): the URL template for the drilldown action.
  - `label` (string): the human-readable label for the drilldown action.
  - `trigger` (string): the trigger event; the only supported value is `on_open_panel_menu`.
  - `type` (string): the drilldown type; the only supported value is `url_drilldown`.
  - `encode_url` (bool, optional): whether to URL-encode the drilldown target; defaults to `true` at the API level.
  - `open_in_new_tab` (bool, optional): whether to open the drilldown in a new browser tab; defaults to `true` at the API level.
- `filters` (nested block, optional): Synthetics-specific monitor filter constraints. Each filter category within the block is optional and accepts a `list(object({ label = string, value = string }))`:
  - `projects`: filter by Synthetics project.
  - `tags`: filter by monitor tag.
  - `monitor_ids`: filter by monitor ID (the API accepts up to 5000 entries).
  - `locations`: filter by monitor location.
  - `monitor_types`: filter by monitor type (e.g. `browser`, `http`).
  - `statuses`: filter by monitor status.

On write, the resource SHALL include only the attributes that are known and non-null in the Terraform configuration. Absent optional fields SHALL NOT be sent in the API request.

On read-back, when Kibana returns an empty or absent config object for a `synthetics_stats_overview` panel, the resource SHALL preserve `null` in state for the `synthetics_stats_overview_config` block rather than materializing an empty block. Individual optional fields absent from the API response SHALL remain null in state rather than being forced to default values.

On read-back, when Kibana returns a nil or empty `filters` object, the resource SHALL treat it as equivalent to an absent `filters` block and SHALL NOT populate the `filters` block in state.

The `synthetics_stats_overview_config` block SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`. The resource SHALL return an error diagnostic if `config_json` is used with `type = "synthetics_stats_overview"`.

#### Scenario: Panel with no config

- GIVEN a panel with `type = "synthetics_stats_overview"` and no `synthetics_stats_overview_config` block
- WHEN create or update runs
- THEN the resource SHALL send a valid panel API payload without a config body and SHALL NOT return an error

#### Scenario: Panel with filter constraints

- GIVEN a `synthetics_stats_overview_config` block with a `filters` sub-block containing one or more filter category lists
- WHEN create or update runs
- THEN the resource SHALL include those filter constraints in the panel config payload sent to Kibana

#### Scenario: Read-back with empty API config

- GIVEN a `synthetics_stats_overview` panel whose API response contains no config fields
- WHEN read runs
- THEN the resource SHALL keep `synthetics_stats_overview_config` null in state

#### Scenario: Read-back with empty filters

- GIVEN a `synthetics_stats_overview` panel whose API response contains a `filters` object with no entries
- WHEN read runs
- THEN the resource SHALL not populate the `filters` block in state

#### Scenario: Panel config block exclusivity

- GIVEN a panel with `type = "synthetics_stats_overview"` and `config_json` also set
- WHEN the provider builds the API request
- THEN it SHALL return an error diagnostic for unsupported `config_json` panel type

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

REQ-006 is extended to include:

- `synthetics_stats_overview_config` SHALL only be valid on panels with `type = "synthetics_stats_overview"`.
- `synthetics_stats_overview_config` SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`.

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

REQ-010 is updated to document that `config_json` write support is **not** extended to `synthetics_stats_overview`. The write-path dispatcher SHALL return an error diagnostic if `config_json` is set on a panel with `type = "synthetics_stats_overview"`. The error message SHALL explicitly name `synthetics_stats_overview` as an unsupported type for `config_json`.
