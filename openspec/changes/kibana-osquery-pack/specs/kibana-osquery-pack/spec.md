## ADDED Requirements

### Requirement: Resource identity and composite ID

The `elasticstack_kibana_osquery_pack` resource SHALL set its `id` to `pack_id` after every Create and Update. `pack_id` SHALL be Optional + Computed with `RequiresReplace`: when omitted from config, the API-assigned ID SHALL be populated into state; when supplied, the API SHALL be called with that ID. `space_id` SHALL be Optional + Computed, defaulting to `"default"`, and SHALL force replacement on change.

#### Scenario: Create with explicit pack_id
- **WHEN** `pack_id = "linux-processes"` is set and the resource is created
- **THEN** the API SHALL be called with `id: "linux-processes"`
- **AND** `id` in state SHALL equal `"linux-processes"`

#### Scenario: Create with server-generated pack_id
- **WHEN** `pack_id` is not set in config and the resource is created
- **THEN** `pack_id` SHALL be populated from the API-assigned ID
- **AND** `id` SHALL equal that API-assigned ID

#### Scenario: pack_id change forces replacement
- **WHEN** `pack_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Schema attributes

The resource SHALL expose the following attributes:

- `id` — Computed string; mirrors `pack_id`
- `pack_id` — Optional + Computed string with RequiresReplace
- `space_id` — Optional + Computed string, default `"default"`, RequiresReplace
- `kibana_connection` — Optional block (provided by entitycore envelope)
- `name` — Required string; human-readable pack name
- `description` — Optional string
- `enabled` — Optional bool
- `policy_ids` — Optional list of strings; Fleet agent policy IDs this pack is deployed to
- `shards` — Optional map(string → number); percent (1–100) of hosts per policy ID to receive the pack
- `schedule_type` — Required string; must be `"interval"` or `"rrule"`
- `interval` — Optional Int64; pack-level execution interval in seconds; exactly-one-of with `rrule_schedule`
- `rrule_schedule` — Optional SingleNestedAttribute; pack-level RRULE schedule; exactly-one-of with `interval`
- `queries` — Required MapNestedAttribute; at least one query must be provided

#### Scenario: Required name attribute enforced
- **WHEN** a resource is configured without `name`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Required queries attribute enforced
- **WHEN** a resource is configured without `queries`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Invalid schedule_type rejected
- **WHEN** `schedule_type = "cron"` is set
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: Scheduling — exactly-one-of interval / rrule_schedule at pack level

The resource SHALL enforce that exactly one of `interval` or `rrule_schedule` is set (not both, not neither).

#### Scenario: Both interval and rrule_schedule set — plan error
- **WHEN** both `interval = 3600` and `rrule_schedule = { ... }` are set in config
- **THEN** Terraform SHALL reject the plan with a validation error indicating exactly one must be set

#### Scenario: Neither interval nor rrule_schedule set — plan error
- **WHEN** `schedule_type = "interval"` and neither `interval` nor `rrule_schedule` is set
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: interval matches schedule_type interval
- **WHEN** `schedule_type = "interval"` and `interval = 3600` (and no `rrule_schedule`)
- **THEN** the plan SHALL be accepted

#### Scenario: rrule_schedule matches schedule_type rrule
- **WHEN** `schedule_type = "rrule"` and `rrule_schedule = { rrule = "FREQ=DAILY;COUNT=5", start_date = "2025-01-01T00:00:00Z" }`
- **THEN** the plan SHALL be accepted

### Requirement: rrule_schedule nested attribute

The `rrule_schedule` attribute (at pack level and as per-query override) SHALL be a SingleNestedAttribute with the following fields:

- `rrule` — Required string; validator: must start with `FREQ=` (shallow RFC 5545 check)
- `start_date` — Required `timetypes.RFC3339`; schedule anchor
- `end_date` — Optional `timetypes.RFC3339`; provider SHALL validate that `end_date > start_date` when both are set
- `splay` — Optional `customtypes.DurationType`; provider SHALL validate that `splay ≤ 12h` (43200s)
- `timeout` — Optional Int64; execution timeout in seconds

#### Scenario: rrule missing FREQ= prefix rejected
- **WHEN** `rrule_schedule = { rrule = "INTERVAL=1", start_date = "2025-01-01T00:00:00Z" }`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: end_date before start_date rejected
- **WHEN** `rrule_schedule.end_date < rrule_schedule.start_date`
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: splay exceeding 12h rejected
- **WHEN** `rrule_schedule.splay = "13h"`
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: queries MapNestedAttribute

The `queries` attribute SHALL be a MapNestedAttribute where map keys are query names (canonical query identifier in Kibana; the inner `id` field is NOT exposed). Each element SHALL be a SingleNestedAttribute with the following fields:

- `query` — Required string; SQL query text
- `interval` — Optional Int64; per-query scheduling override; only allowed when pack `schedule_type = "interval"`
- `rrule_schedule` — Optional SingleNestedAttribute; per-query scheduling override; same shape as pack-level; only allowed when pack `schedule_type = "rrule"`
- `platform` — Optional SetAttribute of strings; allowed values: `"linux"`, `"darwin"`, `"windows"`; on write, sorted and joined to comma-separated string; on read, split back to set
- `version` — Optional string
- `snapshot` — Optional + Computed bool
- `removed` — Optional + Computed bool
- `saved_query_id` — Optional string; references an `elasticstack_kibana_osquery_saved_query`
- `ecs_mapping` — Optional MapNestedAttribute; same shape as the `ecs_mapping` on `elasticstack_kibana_osquery_saved_query`

#### Scenario: Per-query interval override allowed when pack uses interval mode
- **WHEN** `schedule_type = "interval"` and a query sets `interval = 1800`
- **THEN** the plan SHALL be accepted and the override interval SHALL be sent to the API

#### Scenario: Per-query interval override rejected when pack uses rrule mode
- **WHEN** `schedule_type = "rrule"` and a query sets `interval = 1800`
- **THEN** Terraform SHALL reject the plan with a validation error indicating schedule mode mismatch

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

The `shards` attribute SHALL be stored as `map(string → number)` in Terraform state. The canonical source for state is the `GetPacksDetails` response, which returns `map[string]float32`. The provider SHALL normalize this to numeric values in state. On write, the provider SHALL convert to the API-expected wire format (confirmed during implementation per task 1.6).

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

The resource SHALL call `POST /api/osquery/packs` (space-aware via `SpaceAwarePathRequestEditor`) with all managed fields. On success, the provider SHALL populate state from the response; on `read_only = true`, the provider SHALL return an error diagnostic.

#### Scenario: Successful create with interval scheduling
- **WHEN** a resource with `schedule_type = "interval"`, `interval = 3600`, and one query is applied
- **THEN** `POST /api/osquery/packs` SHALL be called (space-aware)
- **AND** state SHALL be populated from the response

### Requirement: Read

The resource SHALL call `GET /api/osquery/packs/{id}` (space-aware). On HTTP 404, the resource SHALL remove itself from state without error. On `read_only = true`, the resource SHALL return an error diagnostic.

#### Scenario: Resource removed from state on 404
- **WHEN** the pack no longer exists in Kibana (HTTP 404)
- **THEN** the resource SHALL be removed from Terraform state without returning an error

#### Scenario: Prebuilt pack detected on read returns error
- **WHEN** the API returns `read_only = true` during a refresh
- **THEN** the resource SHALL return an error diagnostic and SHALL NOT update state

### Requirement: Update

The resource SHALL call `PUT /api/osquery/packs/{id}` (space-aware, full body replacement). Server-managed fields SHALL be omitted from the request. After a successful update, state SHALL be repopulated from the response.

#### Scenario: Successful update of description
- **WHEN** `description` is changed in config and `terraform apply` is run
- **THEN** `PUT /api/osquery/packs/{id}` SHALL be called with the updated description
- **AND** state SHALL reflect the new description after the apply

### Requirement: Delete

The resource SHALL call `DELETE /api/osquery/packs/{id}` (space-aware). HTTP 404 SHALL be treated as idempotent success.

#### Scenario: Successful delete
- **WHEN** `terraform destroy` is run for the resource
- **THEN** `DELETE /api/osquery/packs/{id}` SHALL be called
- **AND** the resource SHALL be removed from state

#### Scenario: Delete of already-removed pack succeeds
- **WHEN** the pack has already been deleted externally and `terraform destroy` is run
- **THEN** the HTTP 404 response SHALL be treated as success
- **AND** the resource SHALL be removed from state without error

### Requirement: Import

The resource SHALL support composite import with format `"<space_id>/<pack_id>"` (e.g., `"default/linux-processes"`).

#### Scenario: Import via composite ID
- **WHEN** `terraform import elasticstack_kibana_osquery_pack.example default/linux-processes` is run
- **THEN** state SHALL be populated for the pack with `pack_id = "linux-processes"` in space `"default"`

### Requirement: Minimum Kibana version

The resource SHALL enforce a minimum Kibana version via `GetVersionRequirements`. Two version requirements SHALL be registered: one for base packs CRUD (initially `8.5.0`, confirmed in task 1.3) and one for the full scheduling model (`schedule_type`/`rrule_schedule`) at a higher version (TBD in task 1.3).

#### Scenario: Kibana below minimum version returns a version error
- **WHEN** the configured Kibana instance is below the declared minimum version
- **THEN** the provider SHALL return an error diagnostic indicating the unsatisfied version requirement before making any API calls
