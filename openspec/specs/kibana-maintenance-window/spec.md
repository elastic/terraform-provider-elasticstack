# `elasticstack_kibana_maintenance_window` — Schema and Functional Requirements

Resource implementation: `internal/kibana/maintenance_window`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_maintenance_window` resource, including Kibana Maintenance Window API usage, identity and import support, provider-level Kibana client usage, version compatibility requirements, and mapping between Terraform state and Kibana maintenance window responses.

## Schema

```hcl
resource "elasticstack_kibana_maintenance_window" "example" {
  id       = <computed, string>                    # Kibana maintenance window ID; UseStateForUnknown
  space_id = <optional, computed, string>          # default "default"; RequiresReplace
  title    = <required, string>                    # minimum length 1
  enabled  = <optional, computed, bool>            # default false

  custom_schedule = {                              # required
    start    = <required, string>                  # ISO 8601 datetime (UTC)
    duration = <required, string>                  # alerting duration format
    timezone = <optional, computed, string>        # IANA timezone; defaults to UTC

    recurring = {                                  # required
      end         = <optional, string>             # ISO 8601 datetime (UTC)
      every       = <optional, string>             # maintenance window interval frequency
      occurrences = <optional, int32>              # minimum 1
      on_week_day = <optional, list(string)>       # e.g. MO, TU, +1MO, -3FR
      on_month_day = <optional, list(int32)>       # 1–31
      on_month    = <optional, list(int32)>        # 1–12
    }
  }

  scope = <optional, object({                      # narrows affected alerts
    alerting = {                                   # required when scope is set
      kql = <required, string>                     # KQL filter expression
    }
  })>
}
```

Notes:

- The resource uses the provider-level Kibana OpenAPI client; there is no resource-level Kibana connection override block.
- This resource does not declare a custom state upgrader or custom plan modifier beyond schema-level defaults listed above.

## Requirements

### Requirement: Kibana Maintenance Window APIs (REQ-001)

The resource SHALL manage maintenance windows through Kibana's Maintenance Window HTTP APIs: create, get, update, and delete ([Kibana maintenance windows API docs](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-maintenance-window)).

#### Scenario: CRUD uses Maintenance Window APIs

- GIVEN a managed Kibana maintenance window
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Maintenance Window API operation

### Requirement: API and client error surfacing (REQ-002)

For create, read, update, and delete, when the provider cannot obtain the Kibana OpenAPI client the operation SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses for create, read, and update SHALL be surfaced as error diagnostics.

#### Scenario: Missing Kibana OpenAPI client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

### Requirement: Version compatibility (REQ-003)

The resource SHALL verify that the target Elastic Stack server version is at least `9.1.0` before creating a maintenance window. If the server version is lower and the server is not a serverless Kibana instance, the resource SHALL fail with an "Unsupported server version" error and SHALL NOT call the Maintenance Window API.

#### Scenario: Server version below minimum on create

- GIVEN a non-serverless Elastic Stack server at version 8.x
- WHEN create runs
- THEN the provider SHALL fail with an "Unsupported server version" error and SHALL NOT call the Maintenance Window API

#### Scenario: Serverless bypasses version check

- GIVEN a serverless Kibana instance
- WHEN create runs
- THEN the provider SHALL NOT fail the version check and SHALL proceed to call the Maintenance Window API

#### Scenario: Read and update validate server version

- GIVEN a non-serverless Elastic Stack server at version 8.x
- WHEN read or update runs
- THEN the provider SHALL fail the version check and SHALL NOT call the Maintenance Window API

### Requirement: Identity and `id` (REQ-004)

The resource SHALL store a computed `id` equal to the Kibana-generated maintenance window identifier. After create and after read-back following update, the provider SHALL set `id` from the API response. On read, the provider SHALL derive the maintenance window ID and space ID from state using `getMaintenanceWindowIDAndSpaceID`, which falls back to interpreting `id` as a composite `<space_id>/<maintenance_window_id>` for backward compatibility.

#### Scenario: Computed id set after create

- GIVEN a successful create of a maintenance window
- WHEN the provider records Terraform state
- THEN `id` SHALL equal the maintenance window ID returned by Kibana

### Requirement: Import support (REQ-005)

The resource SHALL support Terraform import using `ImportStatePassthroughID`, passing the `id` attribute through directly. The accepted import `id` is the Kibana maintenance window ID, or the composite `<space_id>/<maintenance_window_id>` for backward compatibility.

#### Scenario: Import by maintenance window ID

- GIVEN an import `id` equal to the Kibana maintenance window ID
- WHEN import runs
- THEN the provider SHALL set `id` in state to that value and subsequent read SHALL populate the remaining attributes

### Requirement: Space ID lifecycle and default (REQ-006)

If `space_id` is omitted, the resource SHALL default it to `"default"`. Changes to `space_id` SHALL require resource replacement rather than an in-place update.

#### Scenario: Default space

- GIVEN configuration that does not set `space_id`
- WHEN Terraform plans the resource
- THEN `space_id` SHALL default to `"default"`

#### Scenario: Replace on space change

- GIVEN an existing managed maintenance window
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Provider-level Kibana client only (REQ-007)

The resource SHALL use the provider's configured Kibana OpenAPI client for all create, read, update, and delete operations. The resource SHALL NOT support a resource-local connection override in its schema or request path.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the resource performs CRUD
- THEN all API operations SHALL use the provider-level Kibana OpenAPI client

### Requirement: Create flow with read-back (REQ-008)

On create, the resource SHALL call the Kibana create maintenance window API with the fields from plan, then perform an immediate read of the created window by its returned ID. The resource SHALL set state from the read response to avoid drift.

#### Scenario: Create followed by read-back

- GIVEN a valid maintenance window configuration
- WHEN create runs
- THEN the provider SHALL call the create API, then call the get API with the returned ID, and SHALL set Terraform state from the get response

#### Scenario: Created window not found after create

- GIVEN the create API returns a success response
- AND the subsequent get returns nil (not found)
- WHEN the provider processes the read-back
- THEN the provider SHALL remove the resource from state

### Requirement: Update flow with read-back (REQ-009)

On update, the resource SHALL call the Kibana update maintenance window API (PATCH) with fields from plan, then perform an immediate read of the updated window. The resource SHALL set state from the read response to avoid drift.

#### Scenario: Update followed by read-back

- GIVEN an existing managed maintenance window with changed configuration
- WHEN update runs
- THEN the provider SHALL call the update API, then call the get API, and SHALL set Terraform state from the get response

#### Scenario: Updated window not found after update

- GIVEN the update API returns a success response
- AND the subsequent get returns nil (not found)
- WHEN the provider processes the read-back
- THEN the provider SHALL remove the resource from state

### Requirement: Read behavior and missing resource handling (REQ-010)

On read, the resource SHALL obtain the maintenance window by calling the Kibana get API using the maintenance window ID and space ID from state. If the API returns nil (not found), the resource SHALL remove itself from Terraform state. Otherwise it SHALL repopulate state from the API response.

#### Scenario: Read of deleted maintenance window

- GIVEN a resource recorded in Terraform state
- WHEN read calls Kibana and receives not found (nil response)
- THEN the provider SHALL remove the resource from state

### Requirement: Schedule mapping (REQ-011)

The resource SHALL map the `custom_schedule` block between Terraform configuration and the Kibana API. The required `start` (ISO 8601) and `duration` (alerting duration format) SHALL always be sent. The optional `timezone` SHALL be sent only when it is known and non-null. The `recurring` sub-block fields (`end`, `every`, `occurrences`, `on_week_day`, `on_month_day`, `on_month`) SHALL be sent only when they are set to known values.

#### Scenario: Minimal schedule without recurring

- GIVEN `custom_schedule` with only `start` and `duration` set and `recurring` fields all null
- WHEN create or update builds the API request body
- THEN the request SHALL include `start` and `duration` and SHALL NOT include `timezone`, `recurring.end`, `recurring.every`, `recurring.occurrences`, `recurring.on_week_day`, `recurring.on_month_day`, or `recurring.on_month`

#### Scenario: Full recurring schedule

- GIVEN `custom_schedule` with all `recurring` fields populated
- WHEN create or update builds the API request body
- THEN the request SHALL include all recurring fields mapped from plan

### Requirement: Scope mapping (REQ-012)

The optional `scope` block SHALL be sent to the Kibana API when configured and SHALL be omitted when null. When `scope` is configured, the `scope.alerting.kql` KQL filter expression SHALL be mapped into the API request's `scope.alerting.query.kql` field. When mapping the API response back to state, the resource SHALL set `scope` from the response when the response contains a non-nil scope, and SHALL leave `scope` null when the response scope is nil.

#### Scenario: Scope included in request when configured

- GIVEN `scope.alerting.kql = "some-kql-expression"`
- WHEN create or update builds the API request body
- THEN the request SHALL include `scope.alerting.query.kql = "some-kql-expression"`

#### Scenario: Scope absent when not configured

- GIVEN no `scope` block in configuration
- WHEN create or update builds the API request body
- THEN the request SHALL NOT include a `scope` field

### Requirement: Enabled default (REQ-013)

If `enabled` is omitted, the resource SHALL default it to `false`.

#### Scenario: Default enabled state

- GIVEN configuration that does not set `enabled`
- WHEN Terraform plans the resource
- THEN `enabled` SHALL default to `false`

## Traceability

| Area | Primary files |
|------|---------------|
| Schema | `internal/kibana/maintenance_window/schema.go` |
| Metadata / Configure / Import | `internal/kibana/maintenance_window/resource.go` |
| Version compatibility | `internal/kibana/maintenance_window/version_utils.go` |
| CRUD orchestration | `internal/kibana/maintenance_window/create.go`, `internal/kibana/maintenance_window/read.go`, `internal/kibana/maintenance_window/update.go`, `internal/kibana/maintenance_window/delete.go` |
| Model mapping | `internal/kibana/maintenance_window/models.go` |
| Response types | `internal/kibana/maintenance_window/response_types.go` |
