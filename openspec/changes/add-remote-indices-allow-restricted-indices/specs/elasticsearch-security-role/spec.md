## ADDED Requirements

### Requirement: remote_indices allow_restricted_indices schema (REQ-028)

The resource and data source SHALL expose `allow_restricted_indices` on each `remote_indices` entry. On the resource, the attribute SHALL be optional and computed with a `UseStateForUnknown` plan modifier, matching the existing `indices.allow_restricted_indices` definition and description. On the data source, the attribute SHALL be computed.

#### Scenario: Resource schema includes remote allow_restricted_indices

- **GIVEN** a `remote_indices` block on `elasticstack_elasticsearch_security_role`
- **WHEN** the provider schema is inspected
- **THEN** `allow_restricted_indices` SHALL be available on that block with the same semantics as `indices.allow_restricted_indices`

#### Scenario: Data source exposes remote allow_restricted_indices

- **GIVEN** a role in Elasticsearch with `remote_indices` entries that set `allow_restricted_indices`
- **WHEN** `elasticstack_elasticsearch_security_role` data source is read
- **THEN** `remote_indices.*.allow_restricted_indices` SHALL reflect the API value

### Requirement: remote_indices allow_restricted_indices API mapping (REQ-029)

When `remote_indices.allow_restricted_indices` is known on the resource, the provider SHALL include it in the Put role API payload for that remote index entry. When the attribute is unset or null, the provider SHALL omit `allow_restricted_indices` from the API payload. When reading a role from Elasticsearch, the provider SHALL map `allow_restricted_indices` from each `remote_indices` API entry into Terraform state.

#### Scenario: Write true to API

- **GIVEN** a `remote_indices` entry with `allow_restricted_indices = true`
- **WHEN** create or update runs
- **THEN** the Put role payload SHALL include `"allow_restricted_indices": true` for that entry

#### Scenario: Read from API into state

- **GIVEN** Elasticsearch returns a remote index entry with `allow_restricted_indices: false`
- **WHEN** the resource or data source reads the role
- **THEN** state SHALL store `allow_restricted_indices = false` for that entry

## MODIFIED Requirements

### Requirement: Data source attribute mapping (DS-REQ-007)

The data source SHALL map the Get Role API response into the following computed attributes: `description`, `cluster`, `run_as`, `global` (as normalized JSON string), `metadata` (as normalized JSON string), `applications` (set of objects), `indices` (set of objects with nested `field_security` list), and `remote_indices` (set of objects with nested `field_security` list and `allow_restricted_indices`). `cluster` privileges SHALL be mapped as strings.

#### Scenario: All attributes mapped

- **GIVEN** a successful API response with all role fields present
- **WHEN** read completes
- **THEN** every computed attribute SHALL reflect the corresponding API value

### Requirement: Unknown values in plan (REQ-023–REQ-024)

When `indices.allow_restricted_indices` is unknown during planning, the resource SHALL preserve the prior state value for that field. When `remote_indices.allow_restricted_indices` is unknown during planning, the resource SHALL preserve the prior state value for that field. When `indices.field_security.except` is unknown during planning, the resource SHALL preserve the prior state value for that field.

#### Scenario: Deferred unknowns

- **GIVEN** those attributes are unknown at plan time
- **WHEN** planning applies preservation rules
- **THEN** prior state values SHALL be kept for those fields
