## MODIFIED Requirements

### Requirement: Kibana Saved Objects Import API (REQ-001)

The resource SHALL import saved objects through the Kibana Saved Objects Import HTTP API (`POST /api/saved_objects/_import`, space-aware when `space_id` is set) as documented in [Kibana saved objects API docs](https://www.elastic.co/guide/en/kibana/current/saved-objects-api-import.html). The provider SHALL send the export NDJSON as `multipart/form-data` with a `file` part whose body is the `file_contents` value. Query parameters on that request SHALL reflect the resource’s `overwrite`, `create_new_copies`, and `compatibility_mode` attributes as defined by the Kibana API.

#### Scenario: Create calls Import API

- **WHEN** create runs with a valid `file_contents` and optional `space_id`
- **THEN** the provider SHALL issue a multipart Import API request with the file contents and configured query parameters on the correct space-aware path

#### Scenario: Import uses OpenAPI client stack

- **WHEN** create or update runs the import operation with a valid Kibana connection
- **THEN** the request SHALL be executed through `generated/kbapi` (`PostSavedObjectsImportWithBodyWithResponse`) via an `internal/clients/kibanaoapi` helper, not through the legacy go-kibana-rest `KibanaSavedObject.Import` method

### Requirement: API and client error surfacing (REQ-002)

When the provider cannot obtain the Kibana OpenAPI client, create and update operations SHALL return an error diagnostic. Network errors, non-success HTTP status codes from the Import API, failures to build the multipart request, or failures to parse a successful import JSON body SHALL be surfaced as error diagnostics.

#### Scenario: Missing Kibana client

- **WHEN** create or update runs and the resource cannot obtain a Kibana client from provider configuration
- **THEN** the operation SHALL fail with an error diagnostic

#### Scenario: Import API call failure

- **WHEN** create or update runs and a network or transport error occurs when calling the Import API
- **THEN** the provider SHALL surface an error diagnostic

#### Scenario: Import API returns client error response

- **WHEN** create or update runs and the Import API responds with an HTTP status other than 200, or responds with HTTP 200 but with a body that cannot be interpreted as a successful import result
- **THEN** the provider SHALL surface an error diagnostic with available detail from the response

### Requirement: Effective Kibana client selection (REQ-004)

The resource SHALL use the provider's configured Kibana OpenAPI client by default for create and update. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for create and update.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all import saved objects API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all import saved objects API operations SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Create and update share the same import logic (REQ-005)

Create and update SHALL both invoke the same import logic: reading the plan, calling the Kibana Saved Objects Import API, and writing the result to state. Changing `file_contents`, `overwrite`, `space_id`, `create_new_copies`, or `compatibility_mode` SHALL trigger an in-place update that re-runs the import.

#### Scenario: Update re-imports objects

- **WHEN** update runs for an existing resource whose `file_contents` changed in configuration
- **THEN** the provider SHALL call the Kibana Saved Objects Import API again with the new file contents

#### Scenario: Update re-imports when import flags change

- **WHEN** update runs for an existing resource whose `overwrite`, `create_new_copies`, or `compatibility_mode` changed in configuration
- **THEN** the provider SHALL call the Kibana Saved Objects Import API again with the updated parameters

### Requirement: Overwrite flag (REQ-010)

When `overwrite` is `true` and configuration validation permits the combination with other import flags, the resource SHALL send `overwrite: true` on the Kibana Saved Objects Import request so that conflict errors are resolved by overwriting existing saved objects. When `overwrite` is `false` or unset, the provider SHALL omit the overwrite query parameter (Kibana default).

#### Scenario: Overwrite conflicts

- **WHEN** create or update runs with `overwrite = true`, conflicting saved objects exist in Kibana, and `create_new_copies` is not set to `true` in configuration
- **THEN** the provider SHALL send `overwrite: true` to the Import API, resolving conflicts automatically

## ADDED Requirements

### Requirement: create_new_copies flag (REQ-011)

The resource SHALL expose an optional `create_new_copies` boolean attribute. When set to `true`, the provider SHALL send `createNewCopies=true` on the Import API request. When unset or `false`, the provider SHALL omit the query parameter.

#### Scenario: create_new_copies forwarded to API

- **WHEN** create or update runs with `create_new_copies = true`, a valid export in `file_contents`, and no conflicting `overwrite` or `compatibility_mode` flags in configuration
- **THEN** the provider SHALL include `createNewCopies=true` on the Import API request

### Requirement: compatibility_mode flag (REQ-012)

The resource SHALL expose an optional `compatibility_mode` boolean attribute. When set to `true`, the provider SHALL send `compatibilityMode=true` on the Import API request. When unset or `false`, the provider SHALL omit the query parameter.

#### Scenario: compatibility_mode forwarded to API

- **WHEN** create or update runs with `compatibility_mode = true`, a valid export in `file_contents`, and `create_new_copies` is not set to `true` in configuration
- **THEN** the provider SHALL include `compatibilityMode=true` on the Import API request

### Requirement: Mutually exclusive import flags (REQ-013)

The resource SHALL reject, at plan time, the same invalid combinations of `overwrite`, `create_new_copies`, and `compatibility_mode` that the Kibana Saved Objects Import API rejects (`create_new_copies` incompatible with `overwrite` and with `compatibility_mode`).

#### Scenario: create_new_copies conflicts with overwrite

- **WHEN** Terraform validates configuration for a resource with `create_new_copies = true` and `overwrite = true`
- **THEN** the provider SHALL emit a configuration error diagnostic explaining the conflict

#### Scenario: create_new_copies conflicts with compatibility_mode

- **WHEN** Terraform validates configuration for a resource with `create_new_copies = true` and `compatibility_mode = true`
- **THEN** the provider SHALL emit a configuration error diagnostic explaining the conflict
