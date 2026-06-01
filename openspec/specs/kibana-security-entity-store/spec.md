## Purpose

Define the behavior of the `elasticstack_kibana_security_entity_store` Terraform resource,
which manages the lifecycle of the Elastic Security Entity Store within a Kibana space.

## Requirements

### Requirement: Resource manages Entity Store lifecycle (REQ-001)

The `elasticstack_kibana_security_entity_store` resource SHALL manage the full lifecycle of the
Elastic Security Entity Store within a single Kibana space:

- **Create**: call `POST /api/security/entity_store/install` with the desired entity types, optional
  `historySnapshot.frequency`, and optional `logExtraction` configuration. Accept HTTP 200
  (already installed) and HTTP 201 (newly installed) as success.
- **Read**: call `GET /api/security/entity_store/status`. When the overall status is
  `not_installed`, the provider SHALL remove the resource from state (treat as deleted externally).
- **Update**: call `PUT /api/security/entity_store` for `logExtraction` changes; call
  `POST /api/security/entity_store/install` to expand entity types; call
  `POST /api/security/entity_store/uninstall` to shrink entity types (subject to the shrink guard);
  call `PUT /api/security/entity_store/start` or `PUT /api/security/entity_store/stop` to reconcile
  desired engine state.
- **Delete**: call `POST /api/security/entity_store/uninstall` for the entity types held in state.

The resource SHALL enforce `EnforceMinVersion("9.1.0")` in Create, Read, and Update.

#### Scenario: Create installs the Entity Store

- GIVEN a valid `elasticstack_kibana_security_entity_store` configuration with `entity_types = ["host", "user"]`
- WHEN `terraform apply` runs
- THEN the provider SHALL call `POST /api/security/entity_store/install` with `entityTypes = ["host", "user"]`
- AND accept HTTP 200 or HTTP 201 as success
- AND populate `id = "<space_id>/entity_store"` in state

#### Scenario: Read removes resource when not installed

- GIVEN a resource in state with installed entity types
- WHEN an external actor calls `POST /api/security/entity_store/uninstall` for all types
- AND `terraform refresh` or a subsequent apply runs
- THEN the provider SHALL call `GET /api/security/entity_store/status`
- AND detect `status = "not_installed"` in the response
- AND remove the resource from state without returning an error

#### Scenario: Delete uninstalls the Entity Store

- GIVEN a resource in state with `entity_types = ["host", "user"]`
- WHEN `terraform destroy` runs
- THEN the provider SHALL call `POST /api/security/entity_store/uninstall` with the entity types from state
- AND the resource SHALL be removed from state

### Requirement: Entity-type shrink guard (REQ-002)

The resource SHALL guard against implicit destruction of entity engines when `entity_types` is
reduced in the Terraform configuration.

When `allow_entity_type_shrink = false` (the default) and the plan removes one or more entity
types that are currently in state, the provider SHALL return a clear error diagnostic and SHALL NOT
call any API to remove engines.

When `allow_entity_type_shrink = true`, the provider SHALL call
`POST /api/security/entity_store/uninstall` with only the removed entity types and continue with
the update.

`allow_entity_type_shrink` SHALL NOT be sent to the API; it is a Terraform-only guard flag.

#### Scenario: Shrink blocked when guard is false

- GIVEN a resource in state with `entity_types = ["host", "user", "service"]`
- AND a plan that sets `entity_types = ["host"]`
- AND `allow_entity_type_shrink = false` (default)
- WHEN `terraform apply` runs
- THEN the provider SHALL return an error diagnostic
- AND SHALL NOT call `POST /api/security/entity_store/uninstall`

#### Scenario: Shrink allowed when guard is true

- GIVEN a resource in state with `entity_types = ["host", "user", "service"]`
- AND a plan that sets `entity_types = ["host"]`
- AND `allow_entity_type_shrink = true`
- WHEN `terraform apply` runs
- THEN the provider SHALL call `POST /api/security/entity_store/uninstall` with `entityTypes = ["user", "service"]`
- AND the state SHALL reflect `entity_types = ["host"]` after apply

### Requirement: Desired engine start/stop state (REQ-003)

The resource SHALL expose a `started` attribute (Optional + Computed `bool`, default `true`) that
controls whether engines should be running after create or update.

When `started = true`, the provider SHALL call `PUT /api/security/entity_store/start` if any
engine is not running after install or update. When `started = false`, the provider SHALL call
`PUT /api/security/entity_store/stop`.

On Read, `started` SHALL be set to `true` if at least one engine reports status `running`, otherwise
`false`.

#### Scenario: Create with started = false stops engines

- GIVEN a configuration with `started = false`
- WHEN `terraform apply` runs
- THEN the provider SHALL call `POST /api/security/entity_store/install`
- AND then call `PUT /api/security/entity_store/stop`
- AND `started` in state SHALL be `false`

### Requirement: Log-extraction configuration (REQ-004)

The resource SHALL expose a `log_extraction` Optional single-nested block with the following
attributes, all Optional:

| Attribute | Type | API field |
|---|---|---|
| `additional_index_patterns` | `list(string)` | `additionalIndexPatterns` |
| `excluded_index_patterns` | `list(string)` | `excludedIndexPatterns` |
| `delay` | `string` | `delay` |
| `docs_limit` | `int64` | `docsLimit` |
| `field_history_length` | `int64` | `fieldHistoryLength` |
| `frequency` | `string` | `frequency` |
| `lookback_period` | `string` | `lookbackPeriod` |
| `max_logs_per_page` | `int64` | `maxLogsPerPage` |
| `max_logs_per_window` | `int64` | `maxLogsPerWindow` |
| `max_logs_per_window_cap_behavior` | `string` (`drop` or `defer`) | `maxLogsPerWindowCapBehavior` |
| `max_time_window_size` | `string` | `maxTimeWindowSize` |

These fields are sent in the `logExtraction` body of both `POST /api/security/entity_store/install`
and `PUT /api/security/entity_store`.

On Read, the provider SHALL populate `log_extraction` from the first engine's extraction settings
returned by `GET /api/security/entity_store/status`.

#### Scenario: Update log_extraction without replacement

- GIVEN a resource in state with `log_extraction.delay = "2m"`
- AND a plan that changes `log_extraction.delay = "5m"`
- WHEN `terraform apply` runs
- THEN the provider SHALL call `PUT /api/security/entity_store` with `logExtraction.delay = "5m"`
- AND the resource SHALL NOT be destroyed and recreated

### Requirement: history_snapshot triggers replacement (REQ-005)

The resource SHALL expose a `history_snapshot` Optional single-nested block with a single attribute
`frequency` (Optional `string`). Changes to `history_snapshot.frequency` SHALL trigger replacement
(`RequiresReplace` plan modifier), because the install endpoint accepts this field but the update
endpoint does not support it.

#### Scenario: history_snapshot change triggers replacement

- GIVEN a resource in state with `history_snapshot.frequency = "1h"`
- AND a plan that changes `history_snapshot.frequency = "2h"`
- WHEN `terraform plan` runs
- THEN the plan SHALL show the resource being destroyed and recreated

### Requirement: Space-scoped resource with import support (REQ-006)

The resource SHALL support the standard Kibana space pattern:

- `space_id` is Optional + Computed, defaulting to `default` (as resolved by the Kibana client).
- `space_id` changes SHALL force replacement.
- `id` is Computed, formatted as `<space_id>/entity_store`.

The resource SHALL implement `resource.ResourceWithImportState`. Import ID format SHALL be
`<space_id>/entity_store`. On import, the provider SHALL parse `space_id` from the ID, call Read to
populate all remaining fields, and leave `allow_entity_type_shrink` at `false` and
`history_snapshot` at null.

#### Scenario: Import by ID

- GIVEN an installed Entity Store in space `my-space`
- WHEN `terraform import elasticstack_kibana_security_entity_store.example my-space/entity_store`
- THEN the provider SHALL call `GET /api/security/entity_store/status` for space `my-space`
- AND populate `entity_types`, `started`, `log_extraction`, and `status_json` from the response
- AND `allow_entity_type_shrink` SHALL be `false`

### Requirement: status_json computed field (REQ-007)

The resource SHALL expose a `status_json` Computed `string` attribute containing the normalized
JSON representation of the most recent `GET /api/security/entity_store/status` response, for use
in `output` blocks or external tooling.

The value SHALL be refreshed on every Read.

#### Scenario: status_json reflects current status on read

- GIVEN an installed Entity Store resource in state
- WHEN Terraform refreshes the resource
- THEN the provider SHALL call `GET /api/security/entity_store/status`
- AND `status_json` in state SHALL contain the normalized JSON of the full response body
- AND the value SHALL differ from a previous read if the API response changed
