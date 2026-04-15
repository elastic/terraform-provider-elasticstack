## MODIFIED Requirements

### Requirement: Synthetics Private Locations API (REQ-001)

The resource SHALL manage private locations through Kibana's Synthetics Private Locations HTTP API using the generated OpenAPI client (`generated/kbapi`) invoked via provider `kibanaoapi` helpers. Create SHALL use the POST private locations endpoint, read SHALL use GET by id, and delete SHALL use DELETE by id. The provider SHALL pass the effective Kibana space derived from `space_id` (per REQ-010) and from composite import identifiers (per REQ-004) into these operations so that requests use the correct space-scoped API paths.

#### Scenario: CRUD uses Private Locations API

- **WHEN** create, read, or delete runs for a managed Synthetics private location
- **THEN** the provider SHALL use the corresponding Synthetics Private Location OpenAPI operation with the effective space from `space_id` and import resolution rules

### Requirement: API and client error surfacing (REQ-002)

The resource SHALL fail with an error diagnostic when it cannot obtain the Kibana OpenAPI client (`kibanaoapi` / `generated/kbapi`) required for private location operations. Transport errors and unexpected HTTP status codes or unusable response bodies for create, read, and delete SHALL be surfaced as error diagnostics. On read, an HTTP 404 response SHALL cause the resource to be removed from state rather than returning an error.

#### Scenario: Missing Kibana client

- **GIVEN** the resource cannot obtain the scoped Kibana OpenAPI client from provider configuration
- **WHEN** any CRUD operation runs
- **THEN** the operation SHALL fail with an error diagnostic

#### Scenario: Read returns 404

- **GIVEN** a private location that no longer exists in Kibana
- **WHEN** read calls the API and receives a 404 response
- **THEN** the provider SHALL remove the resource from state

#### Scenario: Create API error

- **GIVEN** a create request
- **WHEN** the API returns a transport error or unexpected response
- **THEN** the provider SHALL surface an error diagnostic

### Requirement: Provider-level default Kibana legacy client with optional scoped override (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client by default for create, read, and delete. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for create, read, and delete.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all private location API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all private location API operations SHALL use the scoped Kibana OpenAPI client derived from that block

## ADDED Requirements

### Requirement: Generated private location model mapping (REQ-012)

The resource SHALL map API responses onto Terraform state using `kbapi.SyntheticsGetPrivateLocation` as the canonical decoded representation for successful GET responses and for normalized POST success payloads. The mapping layer SHALL populate `id`, `label`, `agent_policy_id`, optional `tags`, and optional `geo` without silent loss of API-returned values, including values deserialized into `AdditionalProperties` when they are not first-class fields on the generated struct. The implementation SHALL include automated tests (for example JSON fixture tests or table-driven mapper tests) that assert round-trip fidelity for representative payloads.

#### Scenario: Tags round-trip when present in API JSON

- **GIVEN** a GET (or normalized POST) JSON body that includes a `tags` array for the private location
- **WHEN** the provider maps the body into state
- **THEN** the Terraform `tags` attribute SHALL match the API values

#### Scenario: Geo round-trip

- **GIVEN** a GET JSON body that includes `geo` coordinates
- **WHEN** the provider maps the body into state
- **THEN** the Terraform `geo` object SHALL reflect the API `lat` and `lon` values within float32 precision
