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
  default = <optional, bool>
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

### Requirement: Import (REQ-007)

The resource SHALL support import via `ImportStatePassthroughID` using `host_id` as the import path. When importing, the provided ID value SHALL be stored directly as `host_id`.

#### Scenario: Import by host_id

- GIVEN an existing Fleet server host with id `my-host`
- WHEN `terraform import ... my-host` runs
- THEN `host_id` SHALL be `my-host` and a read cycle SHALL refresh all other attributes

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

### Requirement: Read — not found removes from state (REQ-011)

On read, if the Fleet API returns a nil response for the server host, the resource SHALL remove itself from state. When the Fleet API returns an error, the resource SHALL surface it to Terraform diagnostics.

#### Scenario: Server host deleted outside Terraform

- GIVEN the server host was manually deleted from Fleet
- WHEN read (refresh) runs
- THEN the resource SHALL be removed from state

### Requirement: State mapping (REQ-012)

On read, the resource SHALL map `id`, `host_id`, `name`, `hosts`, and `default` from the API response. `space_ids` is not returned by the Fleet API; if `space_ids` is unknown in state after the API call, the resource SHALL set it to explicit null. If `space_ids` has a configured value, the resource SHALL preserve it.

#### Scenario: space_ids preserved after read

- GIVEN `space_ids = ["my-space"]` in state
- WHEN read runs
- THEN `space_ids` SHALL remain `["my-space"]` after the refresh

### Requirement: Create API body (REQ-013)

On create, the resource SHALL submit `host_urls` (from `hosts`), `name`, `is_default` (from `default`), and optionally `id` (from `output_id` when configured) in the create request body.

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
