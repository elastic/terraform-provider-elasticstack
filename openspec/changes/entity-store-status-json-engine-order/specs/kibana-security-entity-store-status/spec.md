## MODIFIED Requirements

### Requirement: Data source reads Entity Store status (REQ-001)

The `elasticstack_kibana_security_entity_store_status` data source SHALL call
`GET /api/security/entity_store/status` and expose the result as computed attributes.

The data source SHALL enforce `EnforceMinVersion("9.4.0")` before calling the API.

The data source SHALL NOT modify any API state. It is read-only.

Before the response body is stored as `status_json`, the provider SHALL normalize the order of the
`engines` array within the response by stable-sorting its elements by each element's `type` field,
so that two reads returning the same logical engine set, but with `engines` in a different order,
produce byte-identical (and therefore semantically-equal) `status_json` values.

#### Schema

| Attribute | Type | Description |
|---|---|---|
| `space_id` | Optional + Computed `string` | Kibana space; defaults to `default`. |
| `include_components` | Optional `bool` | When `true`, passes `?include_components=true` to the API. |
| `installed` | Computed `bool` | `true` when `status != "not_installed"`. |
| `overall_status` | Computed `string` | The `status` field from the API response. |
| `engines` | Computed `list(object)` | Per-engine status details (see Engine object below). |
| `status_json` | Computed `string` | Normalized JSON of the full status response body, with the `engines` array order stable-sorted by `type` so order alone never causes a diff. |
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

#### Scenario: status_json is stable across reads despite engine reordering

- GIVEN an installed Entity Store with engines for entity types `generic` and `user`, both in
  status `running`
- AND one `GET /api/security/entity_store/status` response returns `engines` in the order
  `[generic, user]`
- AND a later response for the same logical state returns the same two engines in the order
  `[user, generic]`
- WHEN the data source builds `status_json` from each response
- THEN both resulting `status_json` values SHALL be equal
- AND `terraform plan` SHALL NOT report a `status_json` difference solely due to the engine order
  change
