## ADDED Requirements

### Requirement: APM service map panel support (REQ-047)

The `elasticstack_kibana_dashboard` resource SHALL support `type = "apm_service_map"` panels through a typed `apm_service_map_config` block. The block exposes the full flat configuration surface of `KibanaHTTPAPIsApmServiceMapEmbeddable`.

#### Schema attributes

All attributes within `apm_service_map_config` are optional unless stated otherwise.

**Service selectors** (all optional strings — freely combinable, no mutual exclusion):
- `environment` — APM service environment (e.g. `"production"`).
- `service_name` — Focus the map on a specific service.
- `service_group_id` — Reference to a saved APM service group (opaque string; no foreign-key validation by the provider).

**Query**:
- `kuery` — KQL query string (plain `StringAttribute`; always KQL, not an object).

**Layout**:
- `map_orientation` — String enum: `horizontal` or `vertical`. The resource SHALL return an error diagnostic at plan time when a value outside this set is supplied.
- `sync_with_dashboard_filters` — Boolean; when null, the attribute is omitted from the API payload.

**Filter lists** (each a set of validated strings; order does not affect plan stability):
- `alert_status_filter` — Set of strings; allowed values: `active`, `delayed`, `recovered`, `untracked`.
- `anomaly_severity_filter` — Set of strings; allowed values: `low`, `warning`, `minor`, `major`, `critical`, `unknown`.
- `connection_filter` — Set of strings; allowed values: `connected`, `orphaned`.
- `slo_status_filter` — Set of strings; allowed values: `degrading`, `healthy`, `noData`, `violated`.

Invalid values for any filter set attribute SHALL produce an error diagnostic at plan time.

**Presentation passthroughs** (reuse `panelkit.PanelPresentationAttributes()`):
- `title`, `description`, `hide_title`, `hide_border`

**Time range**:
- `time_range` — Optional sub-block `{ from: string, to: string, mode: optional string ("absolute" | "relative") }`.

#### Write (ToAPI) behaviour

When `apm_service_map_config` is set, the provider SHALL:
- Set `type` to `"apm_service_map"` in the API panel payload.
- Map each set attribute that is non-null and non-empty to the corresponding slice in the `config` object; omit null/empty sets from the payload.
- Map scalar optional attributes (strings, bools) only when non-null.
- Map `time_range` only when the block is non-null.

`config_json` SHALL NOT be accepted for `apm_service_map` panels; the registry guard (REQ-044A) SHALL return an error diagnostic if `config_json` is set on a panel with `type = "apm_service_map"`. The `apm_service_map` panel type SHALL be managed exclusively through the typed `apm_service_map_config` block.

#### Read (FromAPI) behaviour and null-preservation

On read, the provider SHALL apply REQ-009 null-preservation for every optional field:
- When prior state had a field null, the provider SHALL keep it null in state even if the API returns a value.
- When prior state had a field set, the provider SHALL update it from the API response.
- For filter set attributes: when prior state had the attribute null, the provider SHALL keep it null regardless of the API response. When prior state had the attribute set (including empty set), the provider SHALL reconstruct the `types.Set` from the API slice; the set implementation guarantees that element order is ignored for plan comparison, so re-ordered API responses SHALL produce no plan diff.
- `time_range` — when prior state had it null and the API echoes a value (e.g. the dashboard-level time range), state SHALL remain null.
- On import (no prior state): when the API returns a non-empty config object, populate all non-null API fields into state; when the API returns a nil or empty config, leave `apm_service_map_config` null in state.

The `apm_service_map_config` block SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`. The registry-driven mutual-exclusion guard (REQ-044A) enforces this.

#### Scenarios

##### Scenario: Create apm_service_map panel with environment selector

- GIVEN a dashboard configuration with a panel of `type = "apm_service_map"` and `apm_service_map_config = { environment = "production" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "environment": "production" }` in the panel body
- AND a subsequent plan SHALL show no changes

##### Scenario: Create apm_service_map panel with service_name selector

- GIVEN a dashboard configuration with `apm_service_map_config = { service_name = "checkout" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "service_name": "checkout" }` in the panel body
- AND a subsequent plan SHALL show no changes

##### Scenario: Create apm_service_map panel with service_group_id selector

- GIVEN a dashboard configuration with `apm_service_map_config = { service_group_id = "group-abc" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "service_group_id": "group-abc" }` in the panel body

##### Scenario: Create apm_service_map panel with all three service selectors combined

- GIVEN a dashboard configuration with `apm_service_map_config = { environment = "staging", service_name = "checkout", service_group_id = "group-abc" }`
- WHEN the resource creates the dashboard
- THEN all three fields SHALL appear in the API payload
- AND no mutual-exclusion error SHALL be returned

##### Scenario: Filter sets with multiple values and order independence

- GIVEN a dashboard with `apm_service_map_config.alert_status_filter = ["active", "delayed"]`
- WHEN the resource reads back the dashboard and the API returns `["delayed", "active"]` (reversed order)
- THEN the provider SHALL produce no plan diff
- AND state SHALL contain a set with values `"active"` and `"delayed"`

##### Scenario: All filter sets populated

- GIVEN a panel with all four filter attributes set with multiple valid enum values
- WHEN the resource creates the dashboard and reads it back
- THEN each filter set in state SHALL contain the expected values
- AND a subsequent plan SHALL show no changes

##### Scenario: Invalid alert_status_filter value rejected

- GIVEN a panel with `apm_service_map_config = { alert_status_filter = ["invalid_value"] }`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value is not an allowed enum member

##### Scenario: Invalid map_orientation value rejected

- GIVEN a panel with `apm_service_map_config = { map_orientation = "diagonal" }`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be `horizontal` or `vertical`

##### Scenario: config_json rejected for apm_service_map panel

- GIVEN a panel with `type = "apm_service_map"` and `config_json` also set
- WHEN Terraform plans the configuration
- THEN the resource SHALL return an error diagnostic indicating `config_json` is unsupported for the `apm_service_map` panel type

##### Scenario: Null-preservation on optional scalars

- GIVEN a prior state where `apm_service_map_config.environment` is null
- WHEN the API read returns an `environment` value
- THEN state SHALL keep `environment` null
- AND the subsequent plan SHALL show no changes

##### Scenario: Import null-preservation

- GIVEN an existing dashboard with `apm_service_map` panels that have API-side defaults for optional fields
- WHEN the resource imports the dashboard
- THEN optional fields not explicitly configured SHALL remain null in state
- AND a subsequent plan against a configuration that omits those fields SHALL show no changes

##### Scenario: Full configuration round-trip

- GIVEN an `apm_service_map_config` block with every attribute populated
- WHEN the resource creates the dashboard and reads it back
- THEN all attribute values SHALL appear in state
- AND a subsequent plan SHALL show no changes

## MODIFIED Requirements

### Requirement: Panel type routing and config_json guard (REQ-010 extension)

The list of panel types that SHALL NOT accept practitioner-authored `config_json` (REQ-010) is extended to include `apm_service_map`. The `apm_service_map` panel type SHALL be managed exclusively through the `apm_service_map_config` block. This extension follows the same enforcement pattern as existing entries in REQ-044A (the registry-driven simple panel handler architecture).
#### Scenario: apm_service_map_config routed by type discriminant

- GIVEN a dashboard API response containing a panel with `"type": "apm_service_map"` and a non-empty config object
- WHEN the resource reads the dashboard
- THEN the provider SHALL populate `apm_service_map_config` in state from the API response
- AND SHALL NOT fall back to `config_json` for that panel
