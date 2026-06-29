# `elasticstack_fleet_server_host` — Schema and Functional Requirements

Resource implementation: `internal/fleet/serverhost`

## Purpose

Define schema and behavior for the Fleet server host resource. The resource manages Fleet server hosts via the Fleet Fleet Server Hosts API, including Kibana space awareness.

## Schema

```hcl
resource "elasticstack_fleet_server_host" "example" {
  id      = <computed, string>          # internal identifier, mirrors host_id
  host_id = <optional+computed, string> # Fleet-assigned or user-supplied host ID

  name    = <required, string>
  hosts   = <required, list(string)>    # at least one entry
  default = <optional+computed, bool>  # defaults to false when omitted
  space_ids = <optional+computed, set(string)>
}
```
## Requirements
### Requirement: Fleet Server Host CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Fleet Create fleet server host API to create server hosts. The resource SHALL use the Fleet Get fleet server host API to read server hosts. The resource SHALL use the Fleet Update fleet server host API to update server hosts. The resource SHALL use the Fleet Delete fleet server host API to delete server hosts. When the Fleet API returns a non-success status for any create, update, or delete operation, the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Fleet API returns an error on create
- WHEN the resource create runs
- THEN diagnostics SHALL contain the API error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` attribute that mirrors `host_id`. On create, if `host_id` is not configured, the Fleet API SHALL assign the identifier; the resource SHALL store the API-assigned value in both `id` and `host_id`. When `host_id` is configured, the resource SHALL pass it to the create API as the desired identifier.

#### Scenario: Auto-assigned host_id

- GIVEN `host_id` is not set in config
- WHEN create completes successfully
- THEN `host_id` and `id` SHALL both be set to the API-returned identifier

### Requirement: Create behavior — omit id when unset

When `host_id` is not set in config (null or unknown), the resource SHALL omit the `id` field from the Fleet Create Fleet Server Host request body entirely, allowing Fleet to auto-generate an ID. The resource SHALL NOT send `"id": ""` to the API.

An explicit `host_id = ""` in config is rejected at plan time by the `host_id` validation requirement and therefore never reaches the create payload.

#### Scenario: host_id unset — id omitted from create body

- GIVEN `host_id` is not set in config
- WHEN create runs and the Fleet Create Fleet Server Host API is called
- THEN the request body SHALL NOT contain an `"id"` field
- AND Fleet SHALL auto-assign a host ID which is stored in `host_id` state

#### Scenario: host_id set — id sent in create body

- GIVEN `host_id = "my-host-id"` is set in config
- WHEN create runs and the Fleet Create Fleet Server Host API is called
- THEN the request body SHALL contain `"id": "my-host-id"`

### Requirement: host_id plan-time validation

The resource SHALL validate `host_id` at plan time when an explicit value is provided. A supplied `host_id` value SHALL satisfy all of the following constraints:

1. Length is between 1 and 255 characters (inclusive).
2. Does not contain `/` (path separator).
3. Does not contain `..` (traversal sequence).
4. Does not contain any of the reserved keys as a substring: `__proto__`, `constructor`, `prototype`.

When any constraint is violated, the resource SHALL surface a plan-time error diagnostic naming the violated constraint. The validator SHALL NOT produce an error for null or unknown values; those are treated as not set and the `id` field is omitted from the create payload. An explicit empty string (`host_id = ""`) SHALL be rejected as a length-0 violation.

#### Scenario: Valid explicit host_id passes validation

- GIVEN `host_id = "my-valid-host"` is set in config
- WHEN a plan is generated
- THEN no validation errors SHALL be produced

#### Scenario: Explicit empty host_id fails validation

- GIVEN `host_id = ""` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Unset host_id passes validation

- GIVEN `host_id` is not set in config (null or unknown)
- WHEN a plan is generated
- THEN the plan-time validator SHALL NOT produce an error
- AND the `id` field SHALL be omitted from the create request body

#### Scenario: host_id with path separator fails validation

- GIVEN `host_id = "my/host"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `/` constraint

#### Scenario: host_id with traversal sequence fails validation

- GIVEN `host_id = "my..host"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `..` constraint

#### Scenario: host_id exceeds maximum length fails validation

- GIVEN `host_id` is set to a string of 256 or more characters
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Reserved key host_id fails validation

- GIVEN `host_id = "__proto__"` (or `"constructor"` or `"prototype"`) is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the reserved-key constraint

#### Scenario: Reserved key as substring fails validation

- GIVEN `host_id = "my-__proto__-host"` (or any value containing `constructor` or `prototype` as a substring) is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the reserved-key constraint

### Requirement: Import (REQ-007)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<host_id>`, the resource SHALL set `host_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state, so that server hosts in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` that can be parsed as a composite ID — the resource SHALL treat the entire string as `host_id` and SHALL NOT set `space_ids` from the import ID. This preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL use the space from state to query the Fleet API and populate all remaining attributes.

#### Scenario: Import by composite space/host ID

- GIVEN an existing Fleet server host in Kibana space `"my-space"` with host ID `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Import by plain host ID (default space)

- GIVEN an existing Fleet server host in the default Kibana space with host ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL NOT be set from the import ID

### Requirement: Space-aware create (REQ-008)

On create, when `space_ids` is configured with at least one space ID, the resource SHALL pass the first space ID from `space_ids` to the Fleet create API as the space context. When `space_ids` is null or unknown, the resource SHALL call the create API without a space prefix (default space).

#### Scenario: Create in named space

- GIVEN `space_ids = ["my-space"]`
- WHEN create runs
- THEN the Fleet create API SHALL be called with `my-space` as the space context

### Requirement: Space-aware read and update using state (REQ-009)

On read and update, the resource SHALL derive the operational space from the `space_ids` stored in state (not the plan). If `space_ids` in state is null or empty, the resource SHALL query using the default space. Otherwise, the resource SHALL use the first space ID from state as the space context. On update, the resource SHALL send updated fields (including `space_ids` from the plan) in the request body so the Fleet API can adjust space membership.

#### Scenario: Read uses state space

- GIVEN state has `space_ids = ["space-a"]`
- WHEN read runs
- THEN the Fleet get API SHALL be called using `space-a` as the space context

### Requirement: Space-aware delete (REQ-010)

On delete, the resource SHALL use the first space ID from state as the space context (same logic as read). Deleting removes the server host from all spaces; to remove from specific spaces only, `space_ids` SHALL be updated instead.

#### Scenario: Delete uses state space

- GIVEN state has `space_ids = ["space-a"]`
- WHEN destroy runs
- THEN the Fleet delete API SHALL be called using `space-a` as the space context

### Requirement: Delete clears default before removal (REQ-010a)

The Fleet API rejects deletion of a server host that is currently marked as default. On delete, when state has `default = true`, the resource SHALL first issue an update to set `is_default = false` (using the same operational space as the delete) and then SHALL issue the delete request. If the pre-delete update returns an error, the resource SHALL surface it to diagnostics and SHALL NOT proceed with the delete.

#### Scenario: Destroying a default server host

- GIVEN state has `default = true`
- WHEN `terraform destroy` runs
- THEN the resource SHALL update the host with `is_default = false` before calling the Fleet delete API
- AND the delete SHALL succeed without manual intervention

### Requirement: Read — not found removes from state (REQ-011)

On read, if the Fleet API returns a nil response for the server host, the resource SHALL remove itself from state. When the Fleet API returns an error, the resource SHALL surface it to Terraform diagnostics.

#### Scenario: Server host deleted outside Terraform

- GIVEN the server host was manually deleted from Fleet
- WHEN read (refresh) runs
- THEN the resource SHALL be removed from state

### Requirement: State mapping (REQ-012)

On read, the resource SHALL map `id`, `host_id`, `name`, `hosts`, and `default` from the API response. The `default` attribute SHALL always carry a known boolean value in state — when the user omits it from configuration, it SHALL default to `false` so plan and post-apply state agree. `space_ids` is not returned by the Fleet API; if `space_ids` is unknown in state after the API call, the resource SHALL set it to explicit null. If `space_ids` has a configured value, the resource SHALL preserve it.

#### Scenario: default omitted from config

- GIVEN config does not set `default`
- WHEN apply completes
- THEN state SHALL have `default = false` (matching the Fleet API response) with no inconsistent-result error

#### Scenario: space_ids preserved after read

- GIVEN `space_ids = ["my-space"]` in state
- WHEN read runs
- THEN `space_ids` SHALL remain `["my-space"]` after the refresh

### Requirement: Create API body (REQ-013)

On create, the resource SHALL submit `host_urls` (from `hosts`), `name`, `is_default` (from `default`), and optionally `id` (from `host_id` when configured) in the create request body.

#### Scenario: Create with all fields

- GIVEN name, hosts, and default are all configured
- WHEN create runs
- THEN the Fleet API SHALL receive host_urls, name, and is_default in the request body

### Requirement: Update API body (REQ-014)

On update, the resource SHALL submit `host_urls`, `name`, and `is_default` in the update request body, using `host_id` from the plan as the resource identifier.

#### Scenario: Update name

- GIVEN a server host with a new `name` in plan
- WHEN update runs
- THEN the Fleet update API SHALL be called with the new name

### Requirement: Provider-level Fleet client by default with optional scoped override

The `elasticstack_fleet_server_host` resource SHALL use the provider-configured Fleet client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Fleet client for its API calls.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the scoped Fleet client derived from that block

### Requirement: `host_id` plan modifiers (REQ-015)

The `host_id` attribute SHALL carry `UseStateForUnknown()` and `RequiresReplace()` plan modifiers.

`UseStateForUnknown()` ensures that when `host_id` is absent from the practitioner's configuration, the Plugin Framework carries the prior-state value (the Fleet-assigned UUID) into the update plan so that the `Update` handler receives the correct resource identifier. Without this modifier, `host_id` is `null` in the plan, causing the `Update` handler to construct a request URL with an empty path segment and Kibana to return 404.

`RequiresReplace()` ensures that if a practitioner explicitly provides a new `host_id` value that differs from the one in state, the resource is destroyed and recreated rather than attempting an in-place update with a mismatched ID.

Both modifiers align `fleet_server_host.host_id` with the identical pattern already applied to every other fleet resource ID attribute (`fleet_output.output_id`, `fleet_agent_policy.policy_id`, `fleet_proxy.proxy_id`, `fleet_agent_download_source.source_id`).

#### Scenario: Update succeeds when host_id is not configured

- GIVEN a `fleet_server_host` resource created with `host_id` omitted from config (computed by Fleet)
- WHEN the practitioner changes `name` or `hosts` and runs `terraform apply`
- THEN the `Update` handler SHALL receive the prior-state `host_id` value in the plan (not null)
- AND the Fleet update API SHALL be called with the correct resource identifier
- AND the apply SHALL succeed without a 404 error

#### Scenario: Explicit host_id change triggers replacement

- GIVEN a `fleet_server_host` resource with a known `host_id` in state
- WHEN the practitioner sets a different explicit `host_id` value in config
- THEN Terraform SHALL plan a destroy-and-recreate rather than an in-place update

### Requirement: Acceptance test UPDATE coverage for computed host_id (REQ-016)

The acceptance test `TestAccResourceFleetServerHost_computedID` SHALL include at least one UPDATE `resource.TestStep` that:

1. Changes a mutable attribute (`name` or `hosts`) while omitting `host_id` from the test config.
2. Asserts that the apply succeeds.
3. Asserts that `host_id` remains set to a non-empty value in state after the update.

This guards the `UseStateForUnknown()` fix against future regressions.

#### Scenario: Acceptance test UPDATE step passes with computed host_id

- GIVEN a `fleet_server_host` acceptance test that created a host without an explicit `host_id`
- WHEN an UPDATE step is applied that changes `name` or `hosts` and still omits `host_id`
- THEN the apply SHALL succeed and state SHALL contain a non-empty `host_id`

