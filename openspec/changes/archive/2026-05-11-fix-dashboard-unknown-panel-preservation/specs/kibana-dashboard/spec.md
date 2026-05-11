## MODIFIED Requirements

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The resource SHALL support top-level `panels`, section-contained `panels`, and `sections` in the order returned by the API and the order given in configuration when building requests. For panel reads, it SHALL distinguish sections from top-level panels and map each panel's `type`, `grid`, optional **`id`**, and configuration. For typed panel mappings, the resource SHALL seed from prior state or plan so that optional panel attributes omitted by Kibana on read can be preserved. When a panel is managed through `config_json` only, the resource SHALL preserve that JSON-centric representation and SHALL NOT populate typed configuration blocks from the API for that panel.

On write, practitioner-authored panel-level `config_json` SHALL be supported only for `markdown` and `vis` panel types; using practitioner-authored panel-level `config_json` with any other panel type, including `slo_burn_rate`, `slo_error_budget`, `esql_control`, and `lens-dashboard-app`, or omitting all panel configuration blocks, SHALL return an error diagnostic. Exception: when a panel's `config_json` value was populated by the provider during a prior read to preserve an unknown panel type (see "Unknown panel-type preservation" below), the provider SHALL re-emit that preserved payload on write without error, because the value originates from the API rather than from the practitioner. The `esql_control` panel type SHALL be managed exclusively through the typed `esql_control_config` block. The `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.

`config_json` SHALL NOT be supported for `options_list_control` panels; the `options_list_control` panel type SHALL be managed exclusively through the typed `options_list_control_config` block; using `config_json` with `type = "options_list_control"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `synthetics_monitors` panels; the `synthetics_monitors` panel type SHALL be managed exclusively through the typed `synthetics_monitors_config` block; using `config_json` with `type = "synthetics_monitors"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `synthetics_stats_overview` panels; the write-path dispatcher SHALL return an error diagnostic if `config_json` is set on a panel with `type = "synthetics_stats_overview"`. The error message SHALL indicate that `config_json` is unsupported for the configured panel type.

**Unknown panel-type preservation**: When the read path encounters a panel whose `type` does not match any typed configuration block, the resource SHALL preserve the panel's `id`, `grid`, `type`, and the panel's full raw API configuration payload in state. The preserved payload SHALL be stored in the panel's existing `config_json` attribute (which is `Optional: true, Computed: true`), reusing that attribute for storage rather than introducing a new private field. This is an intentional implementation decision: it avoids introducing a new unexposed attribute and leverages the existing semantic-equality normalization path already in place for `config_json`. The `config_json` value in state is provider-populated, not practitioner-authored. On subsequent writes, the resource SHALL re-emit the preserved payload verbatim through the `config_json` write codepath. If a configuration declares a panel whose `type` matches no typed block and no preserved payload exists in `config_json` state, the resource SHALL return the "unsupported panel type" error diagnostic, clarifying that the type is not yet supported and was not preserved from the API.

**Panel and section identity**: The Terraform attributes **`panels[].id`** and **`sections[].id`** SHALL align directly with the generated Kibana API field **`id`** for panels and sections.

#### Scenario: Panel id round-trip

- GIVEN a panel with `id = "panel-a"` in configuration
- WHEN create or update runs
- THEN the API request SHALL include the generated API `id` field (or equivalent panel identity) consistent with `panel-a` for that panel

#### Scenario: config_json rejected for options_list_control panel type

- GIVEN a panel with `type = "options_list_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `options_list_control`

#### Scenario: config_json rejected for synthetics_monitors panel type

- GIVEN a panel with `type = "synthetics_monitors"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `synthetics_monitors`

#### Scenario: config_json rejected for synthetics_stats_overview panel type

- GIVEN a panel with `type = "synthetics_stats_overview"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic indicating that `config_json` is not supported for the configured panel type

#### Scenario: Unknown panel type preserved on read

- GIVEN a dashboard managed by the resource that contains a panel of a type the resource does not type today (for example `image`, `discover_session`, or `slo_alerts` before those typed blocks ship)
- WHEN refresh, import, or post-apply read runs
- THEN the resource SHALL store the panel's `id`, `grid`, `type`, and full raw API config payload in state without populating any typed config block, and a subsequent plan against unchanged configuration SHALL produce no diff

#### Scenario: Unknown panel type round-trips on write

- GIVEN state contains a panel with an unknown `type` and a preserved raw API payload from a prior read
- WHEN create or update runs and the user has not modified the panel
- THEN the provider SHALL re-emit the preserved payload verbatim in the API request body

#### Scenario: Practitioner cannot author unknown panel types

- GIVEN a Terraform configuration declaring `panels[].type = "<not-typed-and-no-prior-state>"` with no typed config block and no `config_json`
- WHEN validate or plan runs
- THEN the resource SHALL return an error diagnostic indicating the panel type is unsupported
