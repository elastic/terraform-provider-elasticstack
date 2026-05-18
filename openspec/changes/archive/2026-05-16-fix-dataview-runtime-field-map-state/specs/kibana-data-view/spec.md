## MODIFIED Requirements

### Requirement: State mapping for empty collections (REQ-011)

When mapping API responses back to Terraform state, empty `source_filters`, `field_attrs`, `runtime_field_map`, and `field_formats` returned by Kibana SHALL preserve a prior null value instead of forcing an empty list or map into state. If a field format entry has no `params`, the resource SHALL store `params` as a null object in state. `runtime_field_map` SHALL be marked `Optional` and `Computed` so that it remains user-settable while Terraform also accepts a persisted non-empty value when the attribute is omitted from configuration.

#### Scenario: Empty API collection preserves null
- GIVEN prior state where `data_view.source_filters` is null
- WHEN Kibana returns an empty `source_filters` collection
- THEN the provider SHALL keep `data_view.source_filters` null in state

#### Scenario: Omitted runtime_field_map with persisted API value
- GIVEN prior state where `data_view.runtime_field_map` contains entries
- WHEN the configuration removes `data_view.runtime_field_map` and update runs
- AND Kibana preserves the existing runtime fields in its response
- THEN the provider SHALL accept the persisted `runtime_field_map` into state without raising a consistency error
