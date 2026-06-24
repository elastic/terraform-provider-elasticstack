## ADDED Requirements

### Requirement: Resource identity and composite ID

The `elasticstack_kibana_osquery_pack` resource SHALL set its `id` to `pack_id` after every Create and Update. `pack_id` SHALL be **Computed-only** (maps to API `saved_object_id`); the Create request body does not accept a client-supplied pack ID and Kibana always generates a UUID on Create. `space_id` SHALL be Optional + Computed, defaulting to `"default"`, and SHALL force replacement on change.

#### Scenario: Create populates server-generated pack_id
- **WHEN** a resource is created successfully
- **THEN** `pack_id` SHALL be populated from the API `saved_object_id` in the Create response
- **AND** `id` SHALL equal `pack_id`

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Schema attributes

The resource SHALL expose the following attributes (v1 scope — pinned kbapi client, no scheduling fields):

- `id` — Computed string; mirrors `pack_id`
- `pack_id` — Computed string (API `saved_object_id`; not settable in config)
- `space_id` — Optional + Computed string, default `"default"`, RequiresReplace
- `kibana_connection` — Optional block (provided by entitycore envelope)
- `name` — Required string; human-readable pack name
- `description` — Optional string
- `enabled` — Optional bool
- `policy_ids` — Optional list of strings; Fleet agent policy IDs this pack is deployed to
- `shards` — Optional map(string → number); percent (1–100) of hosts per policy ID to receive the pack
- `queries` — Required MapNestedAttribute; at least one query must be provided

Scheduling attributes (`schedule_type`, pack-level `interval`, `rrule_schedule`, per-query `interval`/`timeout`) are **out of v1 scope** because the pinned `SecurityOsqueryAPIObjectQueriesItem` and create/update request types do not include them. A follow-up change after kbapi regeneration may add scheduling.

#### Scenario: Required name attribute enforced
- **WHEN** a resource is configured without `name`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Required queries attribute enforced
- **WHEN** a resource is configured without `queries`
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: queries MapNestedAttribute

The `queries` attribute SHALL be a MapNestedAttribute where map keys are query names (canonical query identifier in Kibana; the inner `id` field is NOT exposed). Each element SHALL be a SingleNestedAttribute with the following fields (aligned with pinned `SecurityOsqueryAPIObjectQueriesItem`):

- `query` — Required string; SQL query text
- `platform` — Optional SetAttribute of strings; allowed values: `"linux"`, `"darwin"`, `"windows"`; on write, sorted and joined to comma-separated string; on read, split back to set
- `version` — Optional string
- `snapshot` — Optional + Computed bool
- `removed` — Optional + Computed bool
- `saved_query_id` — Optional string; references an `elasticstack_kibana_osquery_saved_query`
- `ecs_mapping` — Optional MapNestedAttribute; same shape as the `ecs_mapping` on `elasticstack_kibana_osquery_saved_query`

#### Scenario: Invalid platform value rejected
- **WHEN** a query sets `platform = ["ios"]`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Multiple queries in a pack
- **WHEN** `queries` contains two entries (e.g., `"find_procs"` and `"find_users"`)
- **THEN** both queries SHALL be sent to the API on Create and both SHALL be read back into state

### Requirement: ECS mapping with three-way exactly-one-of constraint

The `ecs_mapping` attribute on each query SHALL be a MapNestedAttribute where each key maps to a SingleNestedAttribute with three Optional fields: `field` (string), `value` (string), `values` (set of strings). A ConfigValidator SHALL enforce that exactly one of `field`, `value`, or `values` is set per element.

On write, the mapping SHALL be converted to the API `{Field, Value: string|[]string}` shape: `field` → `{field: "..."}`, `value` → `{value: "abc"}` (string arm), `values` → `{value: ["a", "b"]}` (array arm).

On read, the API `Value` field SHALL be inspected for string vs array type to determine whether to populate `value` or `values` in state.

#### Scenario: ecs_mapping with field reference
- **WHEN** `ecs_mapping = { "process.name" = { field = "cmdline" } }` is set on a query
- **THEN** the API SHALL be sent the corresponding ECS mapping

#### Scenario: Multiple ecs_mapping fields on same element rejected
- **WHEN** `ecs_mapping = { "k" = { field = "col", value = "literal" } }` is set
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: shards normalization

The `shards` attribute SHALL be stored as `map(string → number)` in Terraform state. On write (Create/Update), the provider SHALL send `map[string]float32` (`SecurityOsqueryAPIShards`). On read, the canonical source is `GetPacksDetails`, which returns `map[string]float32`. The Create response may return shards as an array of `{key, value}` pairs; the provider SHALL normalize to map form in state (prefer re-read via GET when the create response uses array form).

#### Scenario: shards round-trip
- **WHEN** `shards = { "policy-abc" = 75 }` is set
- **THEN** state SHALL contain `shards = { "policy-abc" = 75 }` after Create and Read

### Requirement: Prebuilt pack protection

The `elasticstack_kibana_osquery_pack` resource SHALL return an error diagnostic and refuse to write state when the API response indicates `read_only = true`. This guard applies on Read and after Create. The error diagnostic SHALL direct users to the `elasticstack_kibana_osquery_pack` data source.

#### Scenario: Prebuilt pack detected on read
- **WHEN** the API returns `read_only = true` for a pack
- **THEN** the resource SHALL return an error diagnostic
- **AND** SHALL NOT write the prebuilt pack into state

### Requirement: Create

The resource SHALL call `POST /api/osquery/packs` (space-aware via `SpaceAwarePathRequestEditor`) with all managed fields. On success, the provider SHALL unwrap the `data` wrapper from the response and populate state from `saved_object_id` and other fields; on `read_only = true`, the provider SHALL return an error diagnostic.

#### Scenario: Successful create
- **WHEN** a resource with `name`, `queries`, and optional managed fields is applied
- **THEN** `POST /api/osquery/packs` SHALL be called (space-aware)
- **AND** state SHALL be populated with Computed `pack_id` from `saved_object_id`

### Requirement: Read

The resource SHALL call `GET /api/osquery/packs/{id}` (space-aware) using `pack_id` as `{id}`. On HTTP 404, the resource SHALL remove itself from state without error. On `read_only = true`, the resource SHALL return an error diagnostic.

#### Scenario: Resource removed from state on 404
- **WHEN** the pack no longer exists in Kibana (HTTP 404)
- **THEN** the resource SHALL be removed from Terraform state without returning an error

#### Scenario: Prebuilt pack detected on read returns error
- **WHEN** the API returns `read_only = true` during a refresh
- **THEN** the resource SHALL return an error diagnostic and SHALL NOT update state

### Requirement: Update

The resource SHALL call `PUT /api/osquery/packs/{id}` (space-aware, full body replacement) using `pack_id` as `{id}`. Server-managed fields SHALL be omitted from the request. After a successful update, state SHALL be repopulated from the response.

#### Scenario: Successful update of description
- **WHEN** `description` is changed in config and `terraform apply` is run
- **THEN** `PUT /api/osquery/packs/{id}` SHALL be called with the updated description
- **AND** state SHALL reflect the new description after the apply

### Requirement: Delete

The resource SHALL call `DELETE /api/osquery/packs/{id}` (space-aware) using `pack_id` as `{id}`. HTTP 404 SHALL be treated as idempotent success.

#### Scenario: Successful delete
- **WHEN** `terraform destroy` is run for the resource
- **THEN** `DELETE /api/osquery/packs/{id}` SHALL be called
- **AND** the resource SHALL be removed from state

#### Scenario: Delete of already-removed pack succeeds
- **WHEN** the pack has already been deleted externally and `terraform destroy` is run
- **THEN** the HTTP 404 response SHALL be treated as success
- **AND** the resource SHALL be removed from state without error

### Requirement: Import

The resource SHALL support composite import with format `"<space_id>/<pack_id>"` where `pack_id` is the API `saved_object_id` (UUID).

#### Scenario: Import via composite ID
- **WHEN** `terraform import elasticstack_kibana_osquery_pack.example default/3c42c847-eb30-4452-80e0-728584042334` is run
- **THEN** state SHALL be populated for the pack with `pack_id = "3c42c847-eb30-4452-80e0-728584042334"` in space `"default"`

### Requirement: Minimum Kibana version

The resource SHALL enforce minimum Kibana version **8.5.0** via `GetVersionRequirements` (base packs CRUD). A second scheduling-floor requirement (9.5.0) is deferred until kbapi regeneration and scheduling scope land in a follow-up.

#### Scenario: Kibana below minimum version returns a version error
- **WHEN** the configured Kibana instance is below `8.5.0`
- **THEN** the provider SHALL return an error diagnostic indicating the unsatisfied version requirement before making any API calls
