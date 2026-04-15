## MODIFIED Requirements

### Requirement: Synthetics Parameters API (REQ-001)

The resource SHALL manage Synthetics parameters through Kibana's Synthetics Parameters API: create via `POST /api/synthetics/params`, read via `GET /api/synthetics/params/{id}`, update via `PUT /api/synthetics/params/{id}`, and delete via `DELETE /api/synthetics/params/{id}` using the same generated Kibana OpenAPI (`kbapi`) client used for the other operations.

#### Scenario: CRUD uses Synthetics Parameters APIs

- GIVEN a managed Synthetics parameter
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Parameters API operation

### Requirement: API and client error surfacing (REQ-002)

The resource SHALL fail with an error diagnostic when it cannot obtain the Kibana OpenAPI client. Transport errors and unexpected API responses for create, read, update, and delete SHALL be surfaced as error diagnostics. On read, a 404 response SHALL cause the resource to be removed from state rather than returning an error.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Read returns 404

- GIVEN a parameter that no longer exists in Kibana
- WHEN read calls the API and receives a 404 response
- THEN the provider SHALL remove the resource from state

#### Scenario: Create transport or API error

- GIVEN a create request
- WHEN the API returns a transport error or an unexpected response
- THEN the provider SHALL surface an error diagnostic

### Requirement: Provider-level Kibana client by default (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI (`kbapi`) client by default for all parameter API operations (create, read, update, and delete). When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all of those operations.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all parameter API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all parameter API operations SHALL use the scoped Kibana OpenAPI client derived from that block

## ADDED Requirements

### Requirement: Create and update request bodies encoded with manual JSON (REQ-011)

For create and update, the resource SHALL serialize the parameter request DTO with `encoding/json` and send it with `Content-Type: application/json` through the generated client's `WithBody` request methods, rather than relying on the generated union request types alone, until oapi-codegen correctly encodes the oneOf request body for this API.

#### Scenario: Create uses marshalled JSON body

- GIVEN a parameter create
- WHEN the provider issues the POST request
- THEN the request body SHALL be produced by JSON-marshalling the request DTO and the call SHALL use the OpenAPI client's body-based POST method for parameters

#### Scenario: Update uses marshalled JSON body

- GIVEN a parameter update
- WHEN the provider issues the PUT request
- THEN the request body SHALL be produced by JSON-marshalling the request DTO and the call SHALL use the OpenAPI client's body-based PUT method for parameters
