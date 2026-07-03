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

The resource SHALL enforce `EnforceMinVersion("9.4.0")` in Create, Read, and Update.

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

On Read, `started` SHALL be set to `true` if at least one engine reports status `started`, otherwise
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

### Requirement: Delete waits for uninstall completion (REQ-WAIT-001)

After calling `POST /api/security/entity_store/uninstall`, the provider SHALL poll
`GET /api/security/entity_store/status` at a fixed interval (5 seconds) until the response contains
`status == "not_installed"` or the Delete operation deadline is exceeded. The deadline SHALL be
derived from the resource's `timeouts` block (Delete), defaulting to the entitycore Delete timeout
when unset, and SHALL NOT be a separate hardcoded value.

While polling, transient network errors on the status request SHALL be retried. A deadline expiry
SHALL produce a clear error diagnostic and SHALL NOT silently remove the resource from state.

#### Scenario: Delete waits for not_installed before returning

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN `POST /api/security/entity_store/uninstall` returns 200
- AND the store status is `"uninstalling"` for the first two polls
- THEN the provider SHALL continue polling
- AND SHALL remove the resource from state only after status transitions to `"not_installed"`

#### Scenario: Delete times out if uninstall does not complete

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN the store status never reaches `"not_installed"` within the Delete timeout
- THEN the provider SHALL return an error diagnostic describing the timeout
- AND SHALL NOT remove the resource from state

### Requirement: Read waits for the store to leave installing state before reading engines (REQ-WAIT-002)

The provider SHALL first read `GET /api/security/entity_store/status` synchronously; when the
overall status is `"installing"` it SHALL poll every 3 seconds until the status is `"running"`,
`"stopped"`, `"error"`, or `"not_installed"`, or the Read operation deadline is exceeded. The
deadline SHALL be derived from the resource's `timeouts` block (Read), defaulting to the
entitycore Read timeout when unset, and SHALL NOT be a separate hardcoded value.

When the status transitions to `"not_installed"` during polling, the provider SHALL apply the
existing "remove from state" path (REQ-001 Scenario: Read removes resource when not installed).

When the deadline is exceeded while the status is still `"installing"`, the provider SHALL emit a
warning diagnostic and proceed with the partial data available (degraded-read behavior, not a
hard error), to avoid breaking `terraform refresh` on a slow stack.

#### Scenario: Read polls while store is installing

- GIVEN a resource in state and `GET /api/security/entity_store/status` returns `"installing"`
- WHEN the provider reads the resource
- THEN the provider SHALL retry the status call until status is no longer `"installing"` or until the Read timeout is exceeded
- AND SHALL NOT return a partial `entity_types` list unless the timeout is exceeded

#### Scenario: Read emits warning after timeout and returns partial data

- GIVEN a resource in state and the store remains in `"installing"` beyond the Read timeout
- WHEN the provider reads the resource
- THEN the provider SHALL emit a warning diagnostic
- AND SHALL continue reading with whatever engine data is available
- AND SHALL NOT fail the read with a hard error

### Requirement: Install retries on HTTP 500 within the configured timeout (REQ-WAIT-003)

The provider SHALL retry `POST /api/security/entity_store/install` (issued during Create, and during Update when new entity types are added) when it returns HTTP 500, treating this as a transient initialization error and retrying at a fixed poll interval (5 seconds) until the install succeeds or the operation deadline is exceeded. The deadline SHALL be derived from the resource's `timeouts` block (Create/Update), defaulting to the entitycore timeout when unset, and SHALL NOT be a separate hardcoded wall-clock budget.

HTTP 500 is the only status that triggers retry. All other non-2xx responses SHALL be treated as
fatal and returned immediately as error diagnostics without retrying. This mirrors the entity-link
Create retry behavior (`kibana-security-entity-store-entity-link` REQ-ESL-RETRY-001) and is
especially important because the test-isolation cleanup uninstalls the singleton store between
tests, so a subsequent install can race the store's background teardown.

#### Scenario: Install succeeds after retrying a transient 500

- GIVEN the entity store was recently uninstalled and background teardown is still in progress
- AND `POST /api/security/entity_store/install` returns HTTP 500 for the first attempts
- AND returns HTTP 200 on a later attempt
- WHEN the provider creates (or updates) the entity store resource
- THEN the provider SHALL succeed once the install returns 2xx

#### Scenario: Install does not retry on non-500 errors

- GIVEN `POST /api/security/entity_store/install` returns HTTP 400
- WHEN the provider creates the entity store resource
- THEN the provider SHALL immediately return an error diagnostic without retrying

### Requirement: Acceptance tests enforce isolation between entity store tests (REQ-TEST-ISOLATION-001)

Every acceptance test that manages the entity store resource SHALL register a `t.Cleanup`
function that uninstalls the entity store and waits for `not_installed` (same timeout/interval
as the provider Delete wait) before the test run concludes. This cleanup MUST be idempotent:
calling it when the store is already `not_installed` SHALL succeed without error.

#### Scenario: Cleanup runs after test body regardless of test outcome

- GIVEN an acceptance test registers `t.Cleanup(cleanupEntityStore)`
- WHEN the test body panics or fails
- THEN `cleanupEntityStore` SHALL still run and leave the space in `not_installed` state

#### Scenario: Cleanup on already-uninstalled store is a no-op

- GIVEN the entity store is already in `not_installed` state when the cleanup runs
- WHEN `cleanupEntityStore` is called
- THEN the function SHALL return without error

### Requirement: Acceptance tests use superset assertions for entity_types (REQ-TEST-SUPERSET-001)

Acceptance tests SHALL NOT assert exact cardinality of `entity_types` when the test installs a
strict subset of the possible entity types. Assertions SHALL verify that each expected type is
**present** in the returned set (superset containment), not that the set is exactly equal to the
installed subset. This tolerates the Kibana `install` merge-on-install behavior without masking
actual type-presence regressions.

#### Scenario: Superset assertion passes when extra types are present

- GIVEN a test installs `entity_types = ["host"]`
- AND a prior test left `"generic"` installed in the singleton store
- WHEN the provider reads and returns `entity_types = ["generic", "host"]`
- THEN a `TestCheckTypeSetElemAttr` assertion on `"host"` SHALL pass
- AND an exact-count assertion on `entity_types.# == 1` SHALL be absent from the test

