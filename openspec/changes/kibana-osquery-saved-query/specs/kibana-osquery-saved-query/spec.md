## ADDED Requirements

### Requirement: Resource identity and composite ID

The `elasticstack_kibana_osquery_saved_query` resource SHALL set its `id` to `saved_query_id` after every Create and Update. `saved_query_id` SHALL be Optional + Computed with `RequiresReplace`: when omitted from config, the API-assigned ID SHALL be populated into state; when supplied, the API SHALL be called with that ID. `space_id` SHALL be Optional + Computed, defaulting to `"default"`, and SHALL force replacement on change.

#### Scenario: Create with explicit saved_query_id
- **WHEN** `saved_query_id = "list_all_processes"` is set and the resource is created
- **THEN** the API SHALL be called with `id: "list_all_processes"`
- **AND** `id` in state SHALL equal `"list_all_processes"`

#### Scenario: Create with server-generated saved_query_id
- **WHEN** `saved_query_id` is not set in config and the resource is created
- **THEN** `saved_query_id` SHALL be populated from the API-assigned ID
- **AND** `id` SHALL equal that API-assigned ID

#### Scenario: saved_query_id change forces replacement
- **WHEN** `saved_query_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Schema attributes

The resource SHALL expose the following attributes:

- `id` — Computed string; mirrors `saved_query_id`
- `saved_query_id` — Optional + Computed string with RequiresReplace
- `space_id` — Optional + Computed string, default `"default"`, RequiresReplace
- `kibana_connection` — Optional block (provided by entitycore envelope)
- `query` — Required string; the SQL query text
- `description` — Optional string
- `platform` — Optional SetAttribute of strings; allowed values: `"linux"`, `"darwin"`, `"windows"`
- `interval` — Optional Int64; query execution interval in seconds
- `version` — Optional string
- `snapshot` — Optional + Computed bool; no static default
- `removed` — Optional + Computed bool; no static default
- `ecs_mapping` — Optional MapNestedAttribute; maps query column names to ECS field paths

#### Scenario: Required query attribute enforced
- **WHEN** a resource is configured without `query`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Invalid platform value rejected
- **WHEN** `platform = ["ios"]` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error naming the disallowed value

### Requirement: ECS mapping with three-way exactly-one-of constraint

The `ecs_mapping` attribute SHALL be a MapNestedAttribute where each key maps to a SingleNestedAttribute with three Optional fields: `field` (string), `value` (string), `values` (set of strings). A ConfigValidator SHALL enforce that exactly one of `field`, `value`, or `values` is set per element.

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

### Requirement: `interval` and `version` union-type normalisation

`interval` SHALL be stored as Int64 in Terraform state. On read, the API `interval` field is a `json.RawMessage` union (`int | string`); the provider SHALL use the integer accessor first (falling back to parsing the string arm as int64). On write, the Int64 value SHALL be sent as a stringified integer.

`version` SHALL be stored as a string in Terraform state. On read, the API `version` field is a `json.RawMessage` union (`int | string`); the provider SHALL stringify the value regardless of which arm is populated.

#### Scenario: interval round-trip as integer
- **WHEN** `interval = 3600` is set in config
- **THEN** the API SHALL be sent `interval: "3600"` (or the numeric form accepted by the API)
- **AND** state SHALL contain `interval = 3600` after read

#### Scenario: version round-trip as string
- **WHEN** `version = "1.0.0"` is set in config
- **THEN** state SHALL contain `version = "1.0.0"` after read

### Requirement: Create

The resource SHALL call `POST /api/osquery/saved_queries` (space-aware via `SpaceAwarePathRequestEditor`) with `query` (required) and any optional fields that are set. The Create response wraps the entity in a `data` field; the provider SHALL unwrap `data` before populating state. After unwrapping, if `prebuilt == true`, the provider SHALL return an error diagnostic and not write to state.

#### Scenario: Successful create
- **WHEN** a resource with `query = "SELECT * FROM processes"` is applied
- **THEN** `POST /api/osquery/saved_queries` SHALL be called (space-aware)
- **AND** state SHALL be populated from the response `data` object

#### Scenario: Create of a prebuilt query is refused
- **WHEN** the Create API response contains `prebuilt: true`
- **THEN** the provider SHALL return an error diagnostic referencing the data source
- **AND** state SHALL NOT be written

### Requirement: Read

The resource SHALL call `GET /api/osquery/saved_queries/{id}` (space-aware). On HTTP 404 the resource SHALL be removed from state without error. On success, if `prebuilt == true`, the provider SHALL return an error diagnostic explaining that the query is prebuilt and cannot be managed by this resource.

#### Scenario: Resource deleted out of band
- **WHEN** the API returns HTTP 404 on Read
- **THEN** the resource SHALL be removed from state without error

#### Scenario: Read encounters a prebuilt query
- **WHEN** the API returns `prebuilt: true` on Read
- **THEN** the provider SHALL return an error diagnostic: "Prebuilt Osquery saved queries are managed by the osquery_manager integration package and cannot be managed by this resource. Use the elasticstack_kibana_osquery_saved_query data source to read this query."

### Requirement: Update

The resource SHALL call `PUT /api/osquery/saved_queries/{id}` (space-aware, full body replacement) with all managed fields that are set. Server-managed fields (`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, `saved_object_id`) SHALL be omitted from the PUT body. After Update, state SHALL be repopulated from the PUT response.

#### Scenario: Update query text
- **WHEN** `query` is changed in config
- **THEN** `PUT /api/osquery/saved_queries/{id}` SHALL be called with the new query
- **AND** state SHALL reflect the new query

#### Scenario: Update ecs_mapping
- **WHEN** an ecs_mapping entry is added or modified
- **THEN** `PUT /api/osquery/saved_queries/{id}` SHALL be called with the complete updated ecs_mapping
- **AND** state SHALL reflect the new mapping

### Requirement: Delete

The resource SHALL call `DELETE /api/osquery/saved_queries/{id}` (space-aware). HTTP 404 SHALL be treated as success (idempotent delete).

#### Scenario: Successful delete
- **WHEN** the resource is destroyed
- **THEN** `DELETE /api/osquery/saved_queries/{id}` SHALL be called
- **AND** no error SHALL be returned on HTTP 200

#### Scenario: Already-deleted resource
- **WHEN** the API returns HTTP 404 on Delete
- **THEN** the resource SHALL be removed from state without error

### Requirement: Import

The resource SHALL support import via the composite ID `"<space_id>/<saved_query_id>"`. On import, the provider SHALL parse the composite ID to derive `space_id` and `saved_query_id`, then call Read. If `prebuilt == true` on the read-after-import, the import SHALL fail with the prebuilt error diagnostic.

#### Scenario: Import by composite ID
- **WHEN** `terraform import elasticstack_kibana_osquery_saved_query.x "default/list_all_processes"` is run
- **THEN** `saved_query_id` SHALL be set to `"list_all_processes"`
- **AND** `space_id` SHALL be set to `"default"`
- **AND** all other attributes SHALL be populated from the API GET response

### Requirement: Connection override and version gating

The resource SHALL obtain its Kibana client via the resource-level `kibana_connection` block when provided, otherwise via the provider-level Kibana configuration. Space-aware requests SHALL use `space_id` via `kibanautil.SpaceAwarePathRequestEditor`. The resource SHALL declare a `GetVersionRequirements` entry that fails with a helpful error against Kibana versions older than the confirmed minimum version (initial assumption: `8.5.0`; verify during implementation).

#### Scenario: Resource-level kibana_connection override
- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all API calls SHALL use that connection instead of the provider-level Kibana connection

#### Scenario: Pre-minimum Kibana version
- **WHEN** the resource is planned against a Kibana version older than the configured minimum
- **THEN** Terraform SHALL fail with an error message stating the minimum required version
