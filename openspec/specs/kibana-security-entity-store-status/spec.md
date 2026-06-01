## Purpose

Define the behavior of the `elasticstack_kibana_security_entity_store_status` Terraform data
source, which reads the status of the Elastic Security Entity Store for a Kibana space.

## Requirements

### Requirement: Data source reads Entity Store status (REQ-001)

The `elasticstack_kibana_security_entity_store_status` data source SHALL call
`GET /api/security/entity_store/status` and expose the result as computed attributes.

The data source SHALL enforce `EnforceMinVersion("9.1.0")` before calling the API.

The data source SHALL NOT modify any API state. It is read-only.

#### Schema

| Attribute | Type | Description |
|---|---|---|
| `space_id` | Optional + Computed `string` | Kibana space; defaults to `default`. |
| `include_components` | Optional `bool` | When `true`, passes `?include_components=true` to the API. |
| `installed` | Computed `bool` | `true` when `status != "not_installed"`. |
| `overall_status` | Computed `string` | The `status` field from the API response. |
| `engines_json` | Computed `string` | Normalized JSON of the `engines` array. |
| `status_json` | Computed `string` | Normalized JSON of the full status response body. |
| `kibana_connection` | Optional block | Kibana connection configuration (injected by envelope). |

#### Scenario: Data source reads installed status

- GIVEN an installed Entity Store with two engines (`host`, `user`) in status `running`
- WHEN the data source is read
- THEN `installed` SHALL be `true`
- AND `overall_status` SHALL be `"running"` (or the equivalent API string)
- AND `engines_json` SHALL contain a JSON array with two engine objects
- AND `status_json` SHALL contain the full status response as normalized JSON

#### Scenario: Data source reads not-installed status

- GIVEN an Entity Store that has been uninstalled
- WHEN the data source is read
- THEN `installed` SHALL be `false`
- AND `overall_status` SHALL be `"not_installed"`
- AND `engines_json` SHALL be `"[]"`

#### Scenario: Data source with include_components

- GIVEN an installed Entity Store
- AND `include_components = true` in the data source configuration
- WHEN the data source is read
- THEN the provider SHALL call `GET /api/security/entity_store/status?include_components=true`
- AND `engines_json` SHALL include component-level detail for each engine

### Requirement: Data source is space-scoped (REQ-002)

The `space_id` attribute SHALL control which Kibana space is queried. When omitted, the data
source SHALL use the space configured in the `kibana_connection` block or the provider default.

#### Scenario: Space-specific status read

- GIVEN two Kibana spaces `space-a` and `space-b` each with an installed Entity Store
- WHEN a data source with `space_id = "space-a"` is read
- THEN the result SHALL reflect only the Entity Store state of `space-a`
