## ADDED Requirements

### Requirement: Resource provides full lifecycle management of an Entity Store entity record (REQ-001)

The system SHALL implement a Terraform managed resource `elasticstack_kibana_security_entity_store_entity` that creates, reads, updates, deletes, and imports individual entity records in the Kibana Security Entity Store via the Kibana API.

#### Scenario: Create a host entity with typed blocks

- **GIVEN** a configuration with `entity_type = "host"`, `entity_id = "host:web-01"`, and a typed `entity` block
- **WHEN** `terraform apply` is executed
- **THEN** the provider SHALL call `POST /api/security/entity_store/entities/host` with the assembled JSON body
- **AND** treat HTTP 200 as a successful create response
- **AND** perform an authoritative read-after-write to populate all computed attributes in state

#### Scenario: Create returns 409 when entity already exists

- **GIVEN** a configuration referencing an `entity_id` that already exists in the Entity Store
- **WHEN** `terraform apply` executes the create
- **THEN** the provider SHALL return an error diagnostic indicating the entity already exists
- **AND** SHALL NOT silently overwrite the existing record

#### Scenario: Read uses list endpoint with KQL filter

- **GIVEN** an entity record managed by Terraform with `entity_id = "host:web-01"` and `entity_type = "host"`
- **WHEN** Terraform refreshes state
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with `filter = entity.id:"host:web-01"` and `entity_types = ["host"]`
- **AND** if exactly one record is returned, SHALL map its fields to state
- **AND** if no record is returned, SHALL remove the resource from state

#### Scenario: Update with force flag

- **GIVEN** an existing entity resource with `force = true` in configuration
- **WHEN** `terraform apply` executes an update
- **THEN** the provider SHALL call `PUT /api/security/entity_store/entities/{entity_type}?force=true`
- **AND** SHALL NOT pass `force` to create or delete operations

#### Scenario: Delete uses JSON body with entityId

- **GIVEN** an entity resource in state
- **WHEN** `terraform destroy` is executed
- **THEN** the provider SHALL call `DELETE /api/security/entity_store/entities/` with JSON body `{"entityId": "<entity_id>"}`

#### Scenario: Import using composite ID

- **GIVEN** an existing entity record in Kibana with space `production` and entity_id `host:web-01`
- **WHEN** `terraform import elasticstack_kibana_security_entity_store_entity.example "production/host:web-01"`
- **THEN** the provider SHALL parse the composite ID `<space_id>/<entity_id>` and populate `space_id` and `entity_id` in state
- **AND** SHALL perform a read to populate all other attributes

---

### Requirement: Resource schema — identity attributes (REQ-002)

The `elasticstack_kibana_security_entity_store_entity` resource SHALL expose the following identity attributes.

- `id` — computed string; composite `<space_id>/<entity_id>`; `UseStateForUnknown`; set by the provider after create/read.
- `space_id` — optional computed string; default `"default"`; `RequiresReplace` on change; `UseStateForUnknown`.
- `entity_type` — required string; must be one of `"user"`, `"host"`, `"service"`, `"generic"`; `RequiresReplace` on change.
- `entity_id` — required string; `RequiresReplace` on change; must match the `entity.id` field in the typed `entity` block or `entity_json` when supplied.

#### Scenario: entity_id triggers replacement on change

- **WHEN** `entity_id` changes between plan and apply
- **THEN** the provider SHALL destroy the old entity and create a new one (RequiresReplace semantics)

---

### Requirement: Resource schema — top-level body attributes (REQ-003)

The resource SHALL expose typed attributes and JSON fallback escape hatches for each top-level API body section. Every typed block SHALL conflict with its matching `_json` fallback attribute.

| Typed attribute | JSON fallback | Notes |
|---|---|---|
| `entity` (nested block) | `entity_json` (string) | Required for create |
| `host` (nested block) | `host_json` (string) | Optional; relevant when `entity_type = "host"` |
| `user` (nested block) | `user_json` (string) | Optional; relevant when `entity_type = "user"` |
| `service` (nested block) | `service_json` (string) | Optional; relevant when `entity_type = "service"` |
| `cloud` (nested block) | `cloud_json` (string) | Optional |
| `asset` (nested block) | `asset_json` (string) | Optional |
| `orchestrator` (nested block) | `orchestrator_json` (string) | Optional |
| `labels` (map of string) | `labels_json` (string) | Optional; JSON fallback supports non-string values |

Additional attributes:

- `timestamp` — optional string; maps to `@timestamp` in the API body.
- `tags` — optional set of string.
- `force` — optional bool; default `false`; passed as `?force=true` query parameter on PUT only.

#### Scenario: ConflictsWith enforced at plan time

- **WHEN** both `entity` (typed block) and `entity_json` are set in the same configuration
- **THEN** the provider SHALL produce a plan-time error indicating the two attributes conflict
- **AND** SHALL NOT proceed to apply

#### Scenario: entity_id matches entity.id

- **WHEN** `entity_id = "host:web-01"` is set and the typed `entity` block has `id = "host:web-02"`
- **THEN** the provider SHALL produce a plan-time or validate-time error indicating the mismatch

---

### Requirement: Resource schema — computed outputs (REQ-004)

The resource SHALL expose the following computed-only output attributes.

- `document_json` — computed string; canonical JSON (sorted keys) containing the full entity document as read back from Kibana on the most recent Read. Updated on every successful read.
- `response_json` — computed string; raw API response body serialized as normalized JSON; for troubleshooting only; not used for drift detection.

#### Scenario: document_json populated after create

- **GIVEN** a successful create + authoritative read
- **WHEN** `terraform apply` completes
- **THEN** `document_json` SHALL contain a non-empty JSON string representing the entity as Kibana returned it

---

### Requirement: Version gating at Elastic Stack 9.1.0 (REQ-005)

The resource SHALL enforce a minimum Elastic Stack version of `9.1.0` by implementing `GetVersionRequirements()` on the resource model. When the connected Kibana server reports a version below this minimum, the envelope SHALL short-circuit Create, Read, and Update with a descriptive error diagnostic. Delete is NOT version-gated (deletion should succeed if the resource was previously created, regardless of version drift).

#### Scenario: Create blocked below minimum version

- **WHEN** the Kibana server is below version `9.1.0`
- **THEN** the provider SHALL return an error diagnostic describing the version requirement
- **AND** SHALL NOT make an API call to create the entity

---

### Requirement: Acceptance test coverage (REQ-006)

Acceptance tests SHALL cover the following scenarios. All acceptance tests for this resource SHALL be skipped when the test Elastic Stack is below version `9.1.0`.

#### Scenario: Create update destroy host entity

- **GIVEN** a Terraform configuration managing a `host` entity with typed `entity` and `host` blocks
- **WHEN** `terraform apply` is run followed by an update to `host.ip` followed by `terraform destroy`
- **THEN** each phase SHALL succeed without error and state SHALL reflect the applied configuration

#### Scenario: Typed block and JSON fallback conflict at plan time

- **GIVEN** a Terraform configuration that sets both the typed `entity` block and `entity_json`
- **WHEN** `terraform plan` is executed
- **THEN** the provider SHALL produce a plan-time error and SHALL NOT proceed to apply

#### Scenario: Import restores full state

- **GIVEN** an entity record exists in Kibana with composite ID `"default/host:web-01"`
- **WHEN** `terraform import elasticstack_kibana_security_entity_store_entity.example "default/host:web-01"`
- **THEN** the provider SHALL read the entity and populate all computed attributes in state without error
