## MODIFIED Requirements

### Requirement: Unknown values in plan (REQ-023–REQ-024)

When `indices.field_security.except` is unknown during planning, the resource SHALL preserve the prior state value for that field. `indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` are not covered by this requirement; they use a schema-level default (REQ-039).

#### Scenario: Deferred unknowns

- GIVEN `indices.field_security.except` is unknown at plan time
- WHEN planning applies preservation rules
- THEN the prior state value SHALL be kept for that field

### Requirement: remote_indices allow_restricted_indices schema (REQ-037)

The resource and data source SHALL expose `allow_restricted_indices` on each `remote_indices` entry. On the resource, the attribute SHALL be optional and computed with a schema-level `Default: booldefault.StaticBool(false)`, matching `indices.allow_restricted_indices` (REQ-039). On the data source, the attribute SHALL be computed.

#### Scenario: Resource schema includes remote allow_restricted_indices

- GIVEN a `remote_indices` block on `elasticstack_elasticsearch_security_role`
- WHEN the provider schema is inspected
- THEN `allow_restricted_indices` SHALL be available on that block with the same default-based semantics as `indices.allow_restricted_indices`

#### Scenario: Data source exposes remote allow_restricted_indices

- GIVEN a role in Elasticsearch with `remote_indices` entries that set `allow_restricted_indices`
- WHEN `elasticstack_elasticsearch_security_role` data source is read
- THEN `remote_indices.*.allow_restricted_indices` SHALL reflect the API value

### Requirement: remote_indices allow_restricted_indices API mapping (REQ-038)

When `remote_indices.allow_restricted_indices` is known on the resource (including `false` after the schema default resolves an omitted config value), the provider SHALL include it in the Put role API payload for that remote index entry. When the attribute is still unknown at write time, the provider SHALL omit `allow_restricted_indices` from the API payload. When reading a role from Elasticsearch, the provider SHALL map `allow_restricted_indices` from each `remote_indices` API entry into Terraform state.

#### Scenario: Write true to API

- GIVEN a `remote_indices` entry with `allow_restricted_indices = true`
- WHEN create or update runs
- THEN the Put role payload SHALL include `"allow_restricted_indices": true` for that entry

#### Scenario: Omitted config writes false

- GIVEN a `remote_indices` entry that omits `allow_restricted_indices`
- WHEN create or update runs
- THEN the planned value SHALL be `false`
- AND the Put role payload SHALL include `"allow_restricted_indices": false` for that entry

#### Scenario: Read from API into state

- GIVEN Elasticsearch returns a remote index entry with `allow_restricted_indices: false`
- WHEN the resource or data source reads the role
- THEN state SHALL store `allow_restricted_indices = false` for that entry

## ADDED Requirements

### Requirement: allow_restricted_indices schema default (REQ-039)

`indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` SHALL each be optional and computed with a schema-level `Default: booldefault.StaticBool(false)`, not a `UseStateForUnknown` plan modifier. Whenever the configured value is null, the plan SHALL use the known value `false`, whether the set element is new (`Create`, or a newly-appended `Set` element during `Update`) or pre-existing. Explicit `true` or `false` in configuration SHALL be planned as configured. The capability Schema sketch SHALL document both fields as `<optional, computed, bool, default false>`.

When the planned value is known, `toAPIModel` SHALL include `allow_restricted_indices` in the Put role payload (so an omitted HCL field is sent as `false`). Mapping functions in `models.go` SHALL NOT special-case the default; Framework planning applies it.

#### Scenario: New indices element appended during update omits allow_restricted_indices

- GIVEN an `elasticstack_elasticsearch_security_role` resource already applied with one `indices` element
- WHEN a `dynamic "indices"` block adds a second `indices` element that omits `allow_restricted_indices`, and `terraform apply` runs
- THEN the plan for the new element's `allow_restricted_indices` SHALL be the known value `false`
- AND the apply SHALL succeed without a "Provider produced inconsistent result after apply" error

#### Scenario: New remote_indices element appended during update omits allow_restricted_indices

- GIVEN an `elasticstack_elasticsearch_security_role` resource already applied with one `remote_indices` element
- WHEN a `dynamic "remote_indices"` block adds a second `remote_indices` element that omits `allow_restricted_indices`, and `terraform apply` runs
- THEN the plan for the new element's `allow_restricted_indices` SHALL be the known value `false`
- AND the apply SHALL succeed without a "Provider produced inconsistent result after apply" error

#### Scenario: Create omits allow_restricted_indices

- GIVEN a new `indices` entry that omits `allow_restricted_indices`
- WHEN create runs
- THEN the planned value SHALL be `false`
- AND the apply SHALL succeed
- AND state SHALL store `allow_restricted_indices = false`

#### Scenario: Explicit true is unchanged

- GIVEN an `indices` or `remote_indices` entry with `allow_restricted_indices = true`
- WHEN create or update runs
- THEN the planned value SHALL be `true`
- AND the Put role payload SHALL include `"allow_restricted_indices": true` for that entry
