# kibana-osquery-saved-query Specification

## Purpose
TBD - created by archiving change kibana-osquery-saved-query. Update Purpose after archive.
## Requirements
### Requirement: Resource identity and composite ID

The `elasticstack_kibana_osquery_saved_query` resource SHALL set its Computed `id` to the composite `<space_id>/<saved_query_id>` after every Create, Update, and Read. `GetResourceID()` SHALL return `saved_query_id` for user-facing identity. Read, Update, Delete, import read-after-import, and data source reads SHALL resolve the Kibana `saved_object_id` for API detail/update/delete calls by paginating `OsqueryFindSavedQueries` and matching `data[].id == saved_query_id`; when no match exists, resource Read removes state, data source Read returns not found, and Delete succeeds idempotently. `saved_query_id` SHALL be **Required** with `RequiresReplace`: the API does not assign an ID when `id` is omitted on create (see design Decision 2). `space_id` SHALL be Optional + Computed, defaulting to `"default"`, and SHALL force replacement on change.

#### Scenario: Create with explicit saved_query_id
- **WHEN** `saved_query_id = "list_all_processes"` and `space_id = "default"` and the resource is created
- **THEN** the API SHALL be called with `id: "list_all_processes"` (space-aware)
- **AND** `id` in state SHALL equal `"default/list_all_processes"`
- **AND** `saved_query_id` SHALL equal `"list_all_processes"`

#### Scenario: saved_query_id is required
- **WHEN** `saved_query_id` is not set in config
- **THEN** Terraform SHALL report a configuration validation error at plan time

#### Scenario: saved_query_id change forces replacement
- **WHEN** `saved_query_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Schema attributes

The resource SHALL expose the following attributes:

- `id` — Computed string; composite `<space_id>/<saved_query_id>`
- `saved_object_id` — Computed string; Kibana saved object UUID used by the detail/update/delete APIs
- `saved_query_id` — Required string with RequiresReplace; API lookup key
- `space_id` — Optional + Computed string, default `"default"`, RequiresReplace
- `kibana_connection` — Optional block (provided by entitycore envelope)
- `query` — Required string; the SQL query text
- `description` — Optional string
- `platform` — Optional SetAttribute of strings; allowed values: `"linux"`, `"darwin"`, `"windows"`
- `interval` — Required Int64; query execution interval in seconds (required by live Kibana create/update API despite OpenAPI marking it optional)
- `version` — Optional string
- `snapshot` — Optional + Computed bool; no static default
- `removed` — Optional + Computed bool; no static default
- `ecs_mapping` — Optional MapNestedAttribute; maps query column names to ECS field paths

#### Scenario: Required query attribute enforced
- **WHEN** a resource is configured without `query`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Required interval attribute enforced
- **WHEN** a resource is configured without `interval`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Invalid platform value rejected
- **WHEN** `platform = ["ios"]` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error naming the disallowed value

### Requirement: Platform wire format

On write, `platform` SHALL be sorted and joined to a comma-separated string (e.g. `"darwin,linux"`). On read, the API comma-string SHALL be split back into a set in state. Sorting ensures deterministic plan output.

#### Scenario: Platform round-trip
- **WHEN** `platform = ["linux", "darwin"]` is set in config
- **THEN** the API SHALL be sent `platform: "darwin,linux"` (sorted)
- **AND** after read, state SHALL contain `platform = ["darwin", "linux"]`

### Requirement: ECS mapping with three-way exactly-one-of constraint

The `ecs_mapping` attribute SHALL be a MapNestedAttribute where each key maps to a SingleNestedAttribute with three Optional fields: `field` (string), `value` (string), `values` (set of strings). `ExactlyOneOfNestedAttrsValidator` from `internal/utils/validators` SHALL be attached to `MapNestedAttribute.NestedObject.Validators` to enforce that exactly one of `field`, `value`, or `values` is set per element. This validator is proven on nested/list objects but not yet directly proven on MapNestedAttribute map values; if map nested validation fails during implementation, a custom inline `ValidateObject` MAY be used instead.

On write, the mapping SHALL be converted to the API `{Field, Value: string|[]string}` shape: `field` → `{field: "..."}`, `value` → `{value: "abc"}` (string arm), `values` → `{value: ["a", "b"]}` (array arm).

On read, the API `Value` field SHALL be inspected for string vs array type to determine whether to populate `value` or `values` in state.

#### Scenario: ecs_mapping with field reference
- **WHEN** `ecs_mapping = { "process.name" = { field = "cmdline" } }` is set
- **THEN** the API SHALL be sent `ecs_mapping: { "process.name": { field: "cmdline" } }`

#### Scenario: ecs_mapping with static scalar value
- **WHEN** `ecs_mapping = { "event.category" = { value = "process" } }` is set
- **THEN** the API SHALL be sent `ecs_mapping: { "event.category": { value: "process" } }`

#### Scenario: ecs_mapping with static array values
- **WHEN** `ecs_mapping = { "event.category" = { values = ["process", "network"] } }` is set
- **THEN** the API SHALL be sent `ecs_mapping: { "event.category": { value: ["process", "network"] } }`

#### Scenario: Multiple ecs_mapping fields on same element rejected
- **WHEN** `ecs_mapping = { "k" = { field = "col", value = "literal" } }` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error indicating exactly one of `field`, `value`, `values` must be set

#### Scenario: Empty ecs_mapping element rejected
- **WHEN** `ecs_mapping = { "k" = {} }` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error indicating exactly one of `field`, `value`, `values` must be set

### Requirement: `interval` and `version` response normalisation

`interval` SHALL be stored as Int64 in Terraform state. On read from Create and GET kibanaoapi entities, the API `interval` field is a `json.RawMessage` union (`int | string`); the provider SHALL use the integer accessor first (falling back to parsing the string arm as int64). The Update kibanaoapi entity uses the same union type for `interval`. On write, the Int64 value SHALL be sent as a stringified integer.

`version` SHALL be stored as a string in Terraform state. On read from Create and GET kibanaoapi entities, the API `version` field is a `json.RawMessage` union (`int | string`); the provider SHALL stringify the value regardless of which arm is populated. The Update kibanaoapi entity types `version` as plain `*string` — dereference directly without union accessors. On write, the string is sent verbatim.

#### Scenario: interval round-trip as integer
- **WHEN** `interval = 3600` is set in config
- **THEN** the API SHALL be sent `interval: "3600"` (or the numeric form accepted by the API)
- **AND** state SHALL contain `interval = 3600` after read

#### Scenario: version round-trip as string
- **WHEN** `version = "1.0.0"` is set in config
- **THEN** state SHALL contain `version = "1.0.0"` after read

### Requirement: Create

The resource SHALL call `POST /api/osquery/saved_queries` (space-aware via `SpaceAwarePathRequestEditor`) with `id` (saved_query_id), `query`, `interval`, and any other optional fields that are set. The API response wraps the entity in a `data` field; the kibanaoapi helper returns an unwrapped `OsquerySavedQueryCreateEntity`. The resource SHALL map that entity to state via `populateFromAPI`. If `prebuilt == true` on the entity, the provider SHALL return an error diagnostic and not write to state.

#### Scenario: Successful create
- **WHEN** a resource with `query = "SELECT * FROM processes"` is applied
- **THEN** `POST /api/osquery/saved_queries` SHALL be called (space-aware)
- **AND** state SHALL be populated from the unwrapped `OsquerySavedQueryCreateEntity` returned by `CreateOsquerySavedQuery` via `populateFromAPI`

#### Scenario: Create of a prebuilt query is refused
- **WHEN** the Create API response contains `prebuilt: true`
- **THEN** the provider SHALL return an error diagnostic referencing the data source
- **AND** state SHALL NOT be written

### Requirement: Read

The resource SHALL call `GET /api/osquery/saved_queries/{saved_object_id}` (space-aware) via `GetOsquerySavedQuery`. When computed `saved_object_id` is present in state, the provider SHALL use it directly. When `saved_object_id` is absent (for example immediately after import or in older state), the provider SHALL resolve it from `saved_query_id` via `FindOsquerySavedObjectID` before calling GET. The API response wraps the entity in a `data` field; the kibanaoapi helper returns an unwrapped `OsquerySavedQueryGetEntity` for model mapping. When find returns no match or GET returns HTTP 404, the resource SHALL be removed from state without error. On success, if `prebuilt == true`, the provider SHALL return an error diagnostic explaining that the query is prebuilt and cannot be managed by this resource.

#### Scenario: Resource deleted out of band
- **WHEN** the API returns HTTP 404 on Read
- **THEN** the resource SHALL be removed from state without error

#### Scenario: Read encounters a prebuilt query
- **WHEN** the API returns `prebuilt: true` on Read
- **THEN** the provider SHALL return an error diagnostic: "Prebuilt Osquery saved queries are managed by the osquery_manager integration package and cannot be managed by this resource. Use the elasticstack_kibana_osquery_saved_query data source to read this query."

### Requirement: Update

The resource SHALL call `PUT /api/osquery/saved_queries/{saved_object_id}` (space-aware) via `UpdateOsquerySavedQuery`. When computed `saved_object_id` is present in state, the provider SHALL use it directly; otherwise it SHALL resolve it from `saved_query_id` via `FindOsquerySavedObjectID`. The PUT body SHALL contain `id` (saved_query_id), `query`, `interval`, and the managed optional field set from plan/state (`description`, `platform`, `version`, `snapshot`, `removed`, `ecs_mapping`); server-managed fields (`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, `saved_object_id`) SHALL be omitted. Optional attributes null/unset in plan omit the corresponding JSON keys except `id`, `query`, and `interval`, which are always sent; when a previously-set string/map optional is removed from configuration, the provider SHALL send an empty string (`description`, `platform`, `version`) or empty map (`ecs_mapping`) so Kibana clears the remote field. Empty API strings for optional fields SHALL be normalised back to null in state. The API response wraps the entity in a `data` field; the kibanaoapi helper returns an unwrapped `OsquerySavedQueryUpdateEntity` for model mapping.

#### Scenario: Update query text
- **WHEN** `query` is changed in config
- **THEN** `PUT /api/osquery/saved_queries/{id}` SHALL be called with the new query
- **AND** state SHALL reflect the new query

#### Scenario: Update ecs_mapping
- **WHEN** an ecs_mapping entry is added or modified
- **THEN** `PUT /api/osquery/saved_queries/{id}` SHALL be called with the complete updated ecs_mapping
- **AND** state SHALL reflect the new mapping

### Requirement: Delete

The resource SHALL call `DELETE /api/osquery/saved_queries/{saved_object_id}` (space-aware). When computed `saved_object_id` is present in state, the provider SHALL use it directly; otherwise it SHALL resolve it from `saved_query_id` via `FindOsquerySavedObjectID`. When find returns no match, delete SHALL succeed without error. HTTP 404 on delete SHALL also be treated as success (idempotent delete).

#### Scenario: Successful delete
- **WHEN** the resource is destroyed
- **THEN** `DELETE /api/osquery/saved_queries/{id}` SHALL be called
- **AND** no error SHALL be returned on HTTP 200

#### Scenario: Already-deleted resource
- **WHEN** the API returns HTTP 404 on Delete
- **THEN** the resource SHALL be removed from state without error

### Requirement: Import

The resource SHALL support import via the composite ID `"<space_id>/<saved_query_id>"`. Prefer `ImportStatePassthroughID` on `id` (entitycore parses the composite on Read); if Required `saved_query_id` must be seeded before Read, use a thin custom `ImportState` (as in `alerting_rule`) to set `space_id` and `saved_query_id` from the composite string and set `id` to the import string. Read after import populates remaining attributes. If `prebuilt == true` on the read-after-import, the import SHALL fail with the prebuilt error diagnostic.

#### Scenario: Import by composite ID
- **WHEN** `terraform import elasticstack_kibana_osquery_saved_query.x "default/list_all_processes"` is run
- **THEN** `id` SHALL equal `"default/list_all_processes"`
- **AND** `saved_query_id` SHALL be set to `"list_all_processes"`
- **AND** `space_id` SHALL be set to `"default"`
- **AND** all other attributes SHALL be populated from the API GET response

#### Scenario: Import of prebuilt query fails
- **WHEN** import targets a query whose GET response has `prebuilt: true`
- **THEN** the provider SHALL return the prebuilt error diagnostic
- **AND** state SHALL NOT be written

### Requirement: Connection override and version gating

The resource SHALL obtain its Kibana client via the resource-level `kibana_connection` block when provided, otherwise via the provider-level Kibana configuration. Space-aware requests SHALL use `space_id` via `kibanautil.SpaceAwarePathRequestEditor`. The resource SHALL declare a `GetVersionRequirements` entry that fails with a helpful error against Kibana versions older than `8.5.0` — the documented/conservative floor from discovery; live confirmation against a running stack is expected via acceptance tests.

#### Scenario: Resource-level kibana_connection override
- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all API calls SHALL use that connection instead of the provider-level Kibana connection

#### Scenario: Pre-minimum Kibana version
- **WHEN** the resource is planned against a Kibana version older than the configured minimum
- **THEN** Terraform SHALL fail with an error message stating the minimum required version

