## MODIFIED Requirements

### Requirement: Unknown values in plan (REQ-023–REQ-024)

When `indices.field_security.except` is unknown during planning, the resource SHALL preserve the prior state value for that field.

`indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` SHALL each default to `false` whenever the configured value is null, via a schema-level default rather than a `UseStateForUnknown` plan modifier. This SHALL apply uniformly whether the element is new (e.g. `Create`, or a newly-appended `Set` element during `Update`) or pre-existing, so that a new `indices` or `remote_indices` element added to an existing `Set` (for example via a `dynamic` block during `Update`) that omits `allow_restricted_indices` plans a known `false` value rather than a literal `null`, matching the value Elasticsearch's Put Role API normalizes the omitted field to on write.

#### Scenario: Deferred unknowns

- GIVEN `indices.field_security.except` is unknown at plan time
- WHEN planning applies preservation rules
- THEN the prior state value SHALL be kept for that field

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

### Requirement: remote_indices allow_restricted_indices schema (REQ-037)

The resource and data source SHALL expose `allow_restricted_indices` on each `remote_indices` entry. On the resource, the attribute SHALL be optional and computed with a schema-level `Default: booldefault.StaticBool(false)`, matching the existing `indices.allow_restricted_indices` definition and description. On the data source, the attribute SHALL be computed.

#### Scenario: Resource schema includes remote allow_restricted_indices

- GIVEN a `remote_indices` block on `elasticstack_elasticsearch_security_role`
- WHEN the provider schema is inspected
- THEN `allow_restricted_indices` SHALL be available on that block with the same default-based semantics as `indices.allow_restricted_indices`

#### Scenario: Data source exposes remote allow_restricted_indices

- GIVEN a role in Elasticsearch with `remote_indices` entries that set `allow_restricted_indices`
- WHEN `elasticstack_elasticsearch_security_role` data source is read
- THEN `remote_indices.*.allow_restricted_indices` SHALL reflect the API value
