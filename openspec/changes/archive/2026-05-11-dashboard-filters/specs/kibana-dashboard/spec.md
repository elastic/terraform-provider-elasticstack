## ADDED Requirements

### Requirement: Dashboard-level saved filters round-trip (REQ-037)

The resource SHALL expose dashboard-level saved filters at the dashboard root as `filters = list(object({ filter_json = string }))`. Each `filter_json` value SHALL be a JSON-encoded object that conforms to the Kibana Dashboard API `kbn-as-code-filters-schema_*` discriminated union (or DSL/spatial filter shape) and SHALL be normalized for diff comparison using the same semantic JSON equality applied to per-panel `filter_json` values.

On create and update, the resource SHALL include each filter's JSON object in the dashboard API request `filters` array in the order given in configuration. On read, the resource SHALL repopulate `filters` from the API response in the order returned. When `filters` is unset in configuration and the API returns either no `filters` field or an empty list, the resource SHALL keep the Terraform attribute unset (not coerce it to an empty list).

#### Scenario: Saved filters round-trip across create and read

- GIVEN `filters = [{ filter_json = jsonencode({ operator = "is", field = "host.name", value = "web-01" }) }]`
- WHEN create runs and the post-apply read returns the same filter
- THEN state SHALL contain a single `filters` entry whose `filter_json` is semantically equal to the configured JSON

#### Scenario: Filter JSON normalization avoids spurious diffs

- GIVEN a `filter_json` value whose key ordering or whitespace differs from the API response
- WHEN refresh runs
- THEN the provider SHALL not produce a diff for that filter

#### Scenario: Unset filters preserved when API returns empty

- GIVEN `filters` is unset in configuration
- WHEN read runs and the API returns no `filters` field or `filters: []`
- THEN the Terraform `filters` attribute SHALL remain unset rather than being set to an empty list

#### Scenario: Multiple filters preserved in order

- GIVEN `filters = [a, b, c]` with three distinct filter JSON values
- WHEN create or update runs and the post-apply read returns the same three filters in the same order
- THEN state SHALL contain those three filters in the same order
