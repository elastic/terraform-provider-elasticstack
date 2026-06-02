## ADDED Requirements

### Requirement: Data source reads Entity Store status (REQ-001)

The `elasticstack_kibana_security_entity_store_status` data source SHALL call
`GET /api/security/entity_store/status` and expose the result as computed attributes.

The data source SHALL enforce `EnforceMinVersion("9.4.0")` before calling the API.

The data source SHALL NOT modify any API state. It is read-only.

#### Schema

| Attribute | Type | Description |
|---|---|---|
| `space_id` | Optional + Computed `string` | Kibana space; defaults to `default`. |
| `include_components` | Optional `bool` | When `true`, passes `?include_components=true` to the API. |
| `installed` | Computed `bool` | `true` when `status != "not_installed"`. |
| `overall_status` | Computed `string` | The `status` field from the API response. |
| `engines` | Computed `list(object)` | Per-engine status details (see Engine object below). |
| `status_json` | Computed `string` | Normalized JSON of the full status response body. |
| `kibana_connection` | Optional block | Kibana connection configuration (injected by envelope). |

#### Scenario: Data source reads installed status

- GIVEN an installed Entity Store with two engines (`host`, `user`) in status `running`
- WHEN the data source is read
- THEN `installed` SHALL be `true`
- AND `overall_status` SHALL be `"running"` (or the equivalent API string)
- AND `engines` SHALL contain two engine objects with `type`, `status`, and `index_pattern`
- AND `status_json` SHALL contain the full status response as normalized JSON

#### Scenario: Data source reads not-installed status

- GIVEN an Entity Store that has been uninstalled
- WHEN the data source is read
- THEN `installed` SHALL be `false`
- AND `overall_status` SHALL be `"not_installed"`
- AND `engines` SHALL be an empty list

#### Scenario: Data source with include_components

- GIVEN an installed Entity Store
- AND `include_components = true` in the data source configuration
- WHEN the data source is read
- THEN the provider SHALL call `GET /api/security/entity_store/status?include_components=true`
- AND `engines[].components` SHALL include component-level detail for each engine

### Requirement: Data source is space-scoped (REQ-002)

The `space_id` attribute SHALL control which Kibana space is queried. When omitted, the data
source SHALL use the space configured in the `kibana_connection` block or the provider default.

#### Scenario: Space-specific status read

- GIVEN two Kibana spaces `space-a` and `space-b` each with an installed Entity Store
- WHEN a data source with `space_id = "space-a"` is read
- THEN the result SHALL reflect only the Entity Store state of `space-a`

## Schema

### Engine object

Each element of `engines` is an object with the following attributes:

| Attribute | Type | Description |
|---|---|---|
| `type` | `string` | The entity type managed by this engine. |
| `status` | `string` | Current status of the engine (e.g. `started`, `stopped`, `error`). |
| `index_pattern` | `string` | Index pattern used by the engine. |
| `field_history_length` | `int64` | Number of historical values kept per field. |
| `delay` | `string` | Delay used for log extraction, if configured. |
| `frequency` | `string` | Frequency used for log extraction, if configured. |
| `lookback_period` | `string` | Lookback period used for log extraction, if configured. |
| `filter` | `string` | Filter query applied to the engine, if any. |
| `timeout` | `string` | Timeout setting for the engine, if any. |
| `timestamp_field` | `string` | Timestamp field used by the engine, if any. |
| `error_action` | `string` | Action associated with the last engine error, if any. |
| `error_message` | `string` | Message describing the last engine error, if any. |
| `components` | `list(object)` | Component-level status (see Component object below), present when `include_components = true`. |

### Component object

Each element of `components` is an object with the following attributes:

| Attribute | Type | Description |
|---|---|---|
| `id` | `string` | Component identifier. |
| `installed` | `bool` | Whether the component is installed. |
| `resource` | `string` | Type of Elasticsearch or Kibana resource backing this component. |
| `health` | `string` | Health status of the component, if available. |
