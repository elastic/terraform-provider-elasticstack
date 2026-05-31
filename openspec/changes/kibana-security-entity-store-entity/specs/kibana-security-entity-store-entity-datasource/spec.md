## ADDED Requirements

### Requirement: Data source provides single-entity lookup by entity_id (REQ-DS-001)

The system SHALL implement a Terraform data source `elasticstack_kibana_security_entity_store_entity` that retrieves a single entity record from the Kibana Security Entity Store by filtering the list endpoint on `entity.id`.

#### Scenario: Successful single-entity lookup

- **GIVEN** a data source configuration with `entity_id = "host:web-01"` and `entity_type = "host"`
- **WHEN** `terraform plan` or `terraform apply` executes the data source read
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with `filter = entity.id:"host:web-01"` and `entity_types = ["host"]`
- **AND** if exactly one record is returned, SHALL populate `document_json` with the normalized JSON of that record

#### Scenario: Entity not found returns error

- **GIVEN** a data source configuration with an `entity_id` that does not exist in the Entity Store
- **WHEN** the data source read executes
- **THEN** the provider SHALL return an error diagnostic indicating the entity was not found
- **AND** SHALL NOT produce a warning-only or partial-state result

#### Scenario: Lookup without entity_type

- **GIVEN** a data source configuration with `entity_id` set but no `entity_type`
- **WHEN** the data source read executes
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with `filter = entity.id:"<entity_id>"` only (no `entity_types` filter)
- **AND** SHALL use the first returned record if the result is unambiguous

---

### Requirement: Data source schema (REQ-DS-002)

The `elasticstack_kibana_security_entity_store_entity` data source SHALL expose the following attributes.

- `id` — computed string; composite `<space_id>/<entity_id>`; set by the data source read.
- `space_id` — optional computed string; default `"default"`; `UseStateForUnknown`.
- `entity_id` — required string; the entity record identifier to look up.
- `entity_type` — optional string; when set, narrows the list endpoint filter to the specified type.
- `document_json` — computed string; canonical JSON (sorted keys) of the full entity document returned by Kibana.

#### Scenario: Computed id reflects composite identity

- **WHEN** a data source read completes successfully with `space_id = "default"` and `entity_id = "host:web-01"`
- **THEN** the computed `id` SHALL be `"default/host:web-01"`

---

### Requirement: Version gating at Elastic Stack 9.1.0 (REQ-DS-003)

The data source SHALL enforce a minimum Elastic Stack version of `9.1.0` via `GetVersionRequirements()` on the data source model. When the connected Kibana server is below this minimum, the envelope SHALL short-circuit the read with a descriptive error diagnostic.

#### Scenario: Read blocked below minimum version

- **WHEN** the Kibana server is below version `9.1.0`
- **THEN** the provider SHALL return an error diagnostic describing the version requirement
- **AND** SHALL NOT make an API call

---

### Requirement: Acceptance test coverage (REQ-DS-004)

Acceptance tests for this data source SHALL cover the following scenarios. All tests SHALL be skipped when the test Elastic Stack is below version `9.1.0`.

#### Scenario: Data source reads entity created by resource

- **GIVEN** an entity resource that has been created via `terraform apply`
- **WHEN** a data source referencing the same `entity_id` and `space_id` is read
- **THEN** `document_json` SHALL be non-empty and SHALL contain the same `entity.id` value

#### Scenario: Data source errors on non-existent entity

- **GIVEN** a data source configuration with an `entity_id` that does not exist in the Entity Store
- **WHEN** the data source read executes
- **THEN** the provider SHALL return an error diagnostic and SHALL NOT produce partial state
