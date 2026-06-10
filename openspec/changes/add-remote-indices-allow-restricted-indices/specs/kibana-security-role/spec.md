## ADDED Requirements

### Requirement: remote_indices allow_restricted_indices schema (REQ-027)

The resource and data source SHALL expose `allow_restricted_indices` on each `elasticsearch.remote_indices` entry as an optional boolean attribute with the same security warning description used for restricted index access elsewhere in the provider.

#### Scenario: Resource schema includes remote allow_restricted_indices

- **GIVEN** an `elasticsearch.remote_indices` block on `elasticstack_kibana_security_role`
- **WHEN** the provider schema is inspected
- **THEN** `allow_restricted_indices` SHALL be available on that block

#### Scenario: Data source exposes remote allow_restricted_indices

- **GIVEN** a Kibana role with `elasticsearch.remote_indices` entries that set `allow_restricted_indices`
- **WHEN** `elasticstack_kibana_security_role` data source is read
- **THEN** `elasticsearch.remote_indices.*.allow_restricted_indices` SHALL reflect the API value

## MODIFIED Requirements

### Requirement: Elasticsearch remote index privilege mapping (REQ-021)

When `elasticsearch.remote_indices` entries are configured, the resource SHALL map `clusters`, `names`, `privileges`, `query`, `allow_restricted_indices`, and `field_security` to the corresponding API fields. When `query` is an empty string, the resource SHALL omit `query` from the API payload for that remote index entry. When `allow_restricted_indices` is unset, the resource SHALL omit `allow_restricted_indices` from the API payload for that entry.

#### Scenario: Remote index mapping

- **GIVEN** `remote_indices` entries are configured
- **WHEN** the API payload is built
- **THEN** `clusters`, `names`, `privileges`, and optional `query`, `allow_restricted_indices`, and `field_security` SHALL be populated correctly

#### Scenario: allow_restricted_indices round-trip

- **GIVEN** a `remote_indices` entry with `allow_restricted_indices = true`
- **WHEN** the role is created and subsequently read
- **THEN** state SHALL retain `allow_restricted_indices = true` for that entry

### Requirement: Read state — elasticsearch block (REQ-022–REQ-023)

When the API response contains an `elasticsearch` object, the resource SHALL set the `elasticsearch` block in state including `cluster`, `indices`, `remote_indices` (with `allow_restricted_indices` when present in the API), and `run_as`. When the API `cluster` list is empty or when `run_as` is empty, those fields SHALL be omitted from the flattened state (not stored as empty lists). When no `elasticsearch` object is present in the response, the resource SHALL store an empty list for the `elasticsearch` attribute.

#### Scenario: Empty cluster and run_as omitted

- **GIVEN** the API returns an empty `cluster` list
- **WHEN** state is updated from the API
- **THEN** `cluster` SHALL be omitted from the flattened elasticsearch block
