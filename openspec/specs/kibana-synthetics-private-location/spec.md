# `elasticstack_kibana_synthetics_private_location` — Schema and Functional Requirements

Resource implementation: `internal/kibana/synthetics/privatelocation`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_synthetics_private_location` resource, which manages Kibana Synthetics private locations. Private locations represent on-premises or cloud agent deployments that run synthetic monitors. The resource covers the Synthetics Private Locations API, composite identity and import, provider-level Kibana legacy client usage, a deliberate absence of in-place update support (all changes require replacement), and optional geographic coordinate configuration.

## Schema

```hcl
resource "elasticstack_kibana_synthetics_private_location" "example" {
  id              = <computed, string>          # Kibana-generated id; UseStateForUnknown; RequiresReplace
  label           = <required, string>          # Unique label; UseStateForUnknown; RequiresReplace
  agent_policy_id = <required, string>          # Fleet agent policy id; UseStateForUnknown; RequiresReplace
  tags            = <optional, list(string)>    # UseStateForUnknown; RequiresReplace

  geo = <optional, object({                     # Geographic coordinates (WGS84)
    lat = <required, float64>
    lon = <required, float64>
  })>
}
```

Notes:

- The `geo` block does not carry `RequiresReplace` itself; however, because all attributes that identify the private location carry `RequiresReplace`, the effective behavior is that any configuration change triggers replacement.
- The resource uses the provider-level Kibana legacy client for all operations (create, read, and delete). There is no OpenAPI-based client used for this resource.
- Update is explicitly not supported: calling Update returns an error diagnostic and does not modify any state.
- There is no schema version or state upgrade defined for this resource.
## Requirements
### Requirement: Synthetics Private Locations API (REQ-001)

The resource SHALL manage private locations through Kibana's Synthetics Private Locations API: create via the legacy Kibana client's `KibanaSynthetics.PrivateLocation.Create`, read via `KibanaSynthetics.PrivateLocation.Get`, and delete via `KibanaSynthetics.PrivateLocation.Delete`. The provider SHALL pass the effective Kibana space derived from `space_id` (per REQ-010) into these operations so that requests use the correct space-scoped API paths.

#### Scenario: CRUD uses Private Locations API

- GIVEN a managed Synthetics private location
- WHEN create, read, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Private Location API operation with the effective space from `space_id`

### Requirement: API and client error surfacing (REQ-002)

The resource SHALL fail with an error diagnostic when it cannot obtain the Kibana legacy client. Transport errors and unexpected API responses for create, read, and delete SHALL be surfaced as error diagnostics. On read, a 404 response SHALL cause the resource to be removed from state rather than returning an error.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana legacy client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Read returns 404

- GIVEN a private location that no longer exists in Kibana
- WHEN read calls the API and receives a 404 response
- THEN the provider SHALL remove the resource from state

#### Scenario: Create API error

- GIVEN a create request
- WHEN the API returns a transport error or unexpected response
- THEN the provider SHALL surface an error diagnostic

### Requirement: Identity and computed `id` (REQ-003)

The resource SHALL expose a computed `id` set from the private location id returned by Kibana after a successful create. The `id` SHALL preserve prior state using `UseStateForUnknown`. Because `id` also carries `RequiresReplace`, any operation that changes the private location's Kibana identity SHALL trigger replacement.

#### Scenario: `id` set after create

- GIVEN a successful create of a private location
- WHEN Kibana returns the new private location object
- THEN the provider SHALL store Kibana's location id in state

### Requirement: Import identifiers (REQ-004)

The resource SHALL support Terraform import using `ImportStatePassthroughID`. On import, if the identifier contains no `/`, the full value SHALL be treated as the Kibana private location id in the default space. If the identifier contains a `/`, the provider SHALL parse it as a composite id in the format `<space_id>/<private_location_id>`, use the space segment as the effective Kibana space for subsequent API calls, and use only the private location id segment as the resource id. If the identifier contains `/` but is not a valid two-segment composite id, the provider SHALL return an error diagnostic describing the required format.

#### Scenario: Import with bare id

- GIVEN an import id that does not contain `/`
- WHEN import runs
- THEN the provider SHALL use the full value as the private location id for a subsequent read in the default Kibana space

#### Scenario: Import with composite id

- GIVEN an import id in the format `<space_id>/<private_location_id>`
- WHEN import runs and read is performed
- THEN the provider SHALL extract the `space_id` and `private_location_id` segments and use them to call the Synthetics Private Location API

#### Scenario: Import with malformed composite id

- GIVEN an import id containing `/` but not in a valid composite id format
- WHEN import runs
- THEN the provider SHALL return an error diagnostic describing the required format

### Requirement: Provider-level Kibana legacy client only (REQ-005)

The resource SHALL use the provider's configured Kibana legacy client for create, read, and delete. The resource SHALL NOT support a resource-level connection override block.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the resource performs create, read, or delete
- THEN all API operations SHALL use the provider-level Kibana legacy client

### Requirement: Update not supported (REQ-006)

The resource SHALL NOT support in-place update of a private location. When an Update is called (which only occurs if plan modifiers are not sufficient to force replacement), the resource SHALL return an error diagnostic with a message stating that update is not supported and that only unused locations can be deleted.

#### Scenario: Update returns error

- GIVEN a managed private location
- WHEN the Terraform framework calls the Update method
- THEN the provider SHALL return an error diagnostic stating that update is not supported

### Requirement: All mutable fields require replacement (REQ-007)

Changes to `id`, `label`, `agent_policy_id`, `tags`, or `space_id` SHALL each require resource replacement rather than an in-place update. The `geo` block does not carry `RequiresReplace` independently but changes to it in practice trigger replacement through the interaction with REQ-006 (update not supported).

#### Scenario: Replace on `label` change

- GIVEN an existing managed private location
- WHEN `label` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `agent_policy_id` change

- GIVEN an existing managed private location
- WHEN `agent_policy_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `tags` change

- GIVEN an existing managed private location
- WHEN `tags` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed private location
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Geographic coordinates mapping (REQ-008)

When `geo` is configured, the resource SHALL include geographic coordinates (`lat` and `lon`) in the create request. When `geo` is omitted, the create request SHALL NOT include geographic coordinates. When reading a private location from Kibana, if the API response includes geographic coordinates, the provider SHALL store them in the `geo` block in state; if the API response omits geographic coordinates, the provider SHALL store `geo` as null in state.

#### Scenario: `geo` included in create request

- GIVEN configuration includes `geo` with `lat` and `lon`
- WHEN the create request is built
- THEN the request SHALL include the geographic coordinates

#### Scenario: `geo` omitted from create request

- GIVEN configuration does not include `geo`
- WHEN the create request is built
- THEN the request SHALL NOT include geographic coordinates

#### Scenario: `geo` null when API response has no coordinates

- GIVEN Kibana returns a private location with no geographic coordinates
- WHEN the provider maps the response to state
- THEN `geo` SHALL be null in state

### Requirement: `label` as unique identifier (REQ-009)

The `label` attribute SHALL serve as the human-readable unique identifier for a private location within Kibana. The resource SHALL always persist `label` from the API response to state. Consumers of the private location (e.g. synthetics monitors) SHALL reference the private location by its `label`, not its `id`.

#### Scenario: `label` persisted from API response

- GIVEN a successful create or read of a private location
- WHEN the provider maps the response to state
- THEN `label` SHALL be set from the API response value

### Requirement: `space_id` attribute (REQ-010)

The resource SHALL expose an optional `space_id` string attribute that selects the Kibana space used for create, read, and delete. When `space_id` is omitted or set to an empty string, the provider SHALL use the default Kibana space. When `space_id` is set to a non-empty value, the provider SHALL use that Kibana space for all Synthetics Private Location API calls. The attribute SHALL use plan modifiers such that changing `space_id` requires resource replacement. The provider SHALL persist `space_id` in state from configuration (and reflect it on read as applicable).

#### Scenario: Default space when `space_id` omitted

- GIVEN configuration does not set `space_id` or sets it to an empty string
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests for the default Kibana space

#### Scenario: Non-default space

- GIVEN configuration sets `space_id` to a non-empty Kibana space identifier
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests scoped to that space

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed private location
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Import with non-default space (REQ-011)

When the practitioner imports a private location that exists in a non-default Kibana space, they SHALL use a composite import identifier in the format `<space_id>/<private_location_id>` so the provider can read from the correct Kibana space. After a successful read, the provider SHALL persist `space_id` in state.

#### Scenario: Import requires matching composite `space_id` for non-default space

- GIVEN a private location exists only in a non-default Kibana space
- WHEN the practitioner runs import with a bare Kibana id or a composite id whose `space_id` segment does not match the location's space
- THEN subsequent read MAY receive 404 and the provider SHALL apply existing 404 handling (remove from state) or fail as appropriate

## Traceability

| Area | Primary files |
|------|---------------|
| Schema and model | `internal/kibana/synthetics/privatelocation/schema.go` |
| Metadata / Configure / Import / Update | `internal/kibana/synthetics/privatelocation/resource.go` |
| Create | `internal/kibana/synthetics/privatelocation/create.go` |
| Read | `internal/kibana/synthetics/privatelocation/read.go` |
| Delete | `internal/kibana/synthetics/privatelocation/delete.go` |
| Shared client helpers | `internal/kibana/synthetics/api_client.go` |
| Shared utilities | `internal/kibana/synthetics/schema.go` |
