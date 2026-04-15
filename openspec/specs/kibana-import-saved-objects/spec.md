# `elasticstack_kibana_import_saved_objects` — Schema and Functional Requirements

Resource implementation: `internal/kibana/import_saved_objects`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_import_saved_objects` resource, which imports a set of Kibana saved objects from an export file using the Kibana Saved Objects Import API. The resource is intentionally write-only: read and delete are no-ops. Import results including success counts and per-object errors are captured in computed attributes.

## Schema

```hcl
resource "elasticstack_kibana_import_saved_objects" "example" {
  id                   = <computed, string>  # UUID generated on create; UseStateForUnknown
  space_id             = <optional, string>  # target Kibana space; uses default space when omitted
  file_contents        = <required, string>  # NDJSON content from Kibana export API
  overwrite            = <optional, bool>    # automatically resolve conflicts by overwriting; conflicts with create_new_copies
  create_new_copies    = <optional, bool>    # regenerate IDs and reset origin; conflicts with overwrite and compatibility_mode
  compatibility_mode   = <optional, bool>    # adjust objects for cross-version compatibility; conflicts with create_new_copies
  ignore_import_errors = <optional, bool>    # if true, import errors do not fail apply

  # Computed result attributes
  success       = <computed, bool>           # true if all objects imported successfully
  success_count = <computed, int64>          # number of successfully imported objects
  errors        = <computed, list(object({
    id    = string
    type  = string
    title = string
    error = object({ type = string })
    meta  = object({ icon = string, title = string })
  }))>
  success_results = <computed, list(object({
    id             = string
    type           = string
    destination_id = string
    meta           = object({ icon = string, title = string })
  }))>
}
```

Notes:

- The resource uses the generated `kbapi` Kibana OpenAPI client (`GetKibanaOapiClient`) via the `kibanaoapi.ImportSavedObjects` helper.
- Read and Delete are intentional no-ops; the resource is not refreshed from Kibana on plan.
- The resource does not support Terraform import.
- The resource does not declare a custom state upgrader.
- `create_new_copies` conflicts with both `overwrite` and `compatibility_mode` (enforced via `ResourceWithConfigValidators`).

## Requirements

### Requirement: Kibana Saved Objects Import API (REQ-001)

The resource SHALL import saved objects through the Kibana Saved Objects Import API ([Kibana saved objects API docs](https://www.elastic.co/guide/en/kibana/current/saved-objects-api-import.html)).

#### Scenario: Create calls Import API

- GIVEN a valid `file_contents` and optional `space_id`
- WHEN create runs
- THEN the provider SHALL call the Kibana Saved Objects Import API with the file contents, overwrite flag, and space ID

### Requirement: API and client error surfacing (REQ-002)

When the provider cannot obtain the Kibana client, create and update operations SHALL return an error diagnostic. Transport or client errors from the Import API SHALL also be surfaced as error diagnostics.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana client from provider configuration
- WHEN create or update runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Import API call failure

- GIVEN a network or server error when calling the Import API
- WHEN create or update runs
- THEN the provider SHALL surface an error diagnostic

### Requirement: Computed `id` (REQ-003)

The resource SHALL generate a UUID as the computed `id` on the first create and SHALL preserve it across subsequent updates using `UseStateForUnknown`.

#### Scenario: UUID assigned on create

- GIVEN a new resource with no prior state
- WHEN create runs successfully
- THEN `id` SHALL be set to a newly generated UUID

#### Scenario: ID preserved on update

- GIVEN a resource already in state with a UUID `id`
- WHEN update runs
- THEN `id` SHALL remain unchanged

### Requirement: Effective Kibana client selection (REQ-004)

The resource SHALL use the provider's configured Kibana OpenAPI client (`GetKibanaOapiClient`) by default for create and update. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for create and update.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all import saved objects API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all import saved objects API operations SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Create and update share the same import logic (REQ-005)

Create and update SHALL both invoke the same import logic: reading the plan, calling the Kibana Saved Objects Import API, and writing the result to state. Changing `file_contents`, `overwrite`, `space_id`, `create_new_copies`, or `compatibility_mode` SHALL trigger an in-place update that re-runs the import.

#### Scenario: Update re-imports objects

- GIVEN an existing resource with `file_contents` changed in configuration
- WHEN update runs
- THEN the provider SHALL call the Kibana Saved Objects Import API again with the new file contents

### Requirement: Read is a no-op (REQ-006)

Read SHALL NOT call any Kibana API and SHALL NOT modify Terraform state. The resource is not refreshable from Kibana because the saved objects import is an imperative action, not a declarative state.

#### Scenario: Read does not call Kibana

- GIVEN a resource recorded in Terraform state
- WHEN Terraform refreshes state
- THEN the provider SHALL NOT call any Kibana API and SHALL leave state unchanged

### Requirement: Delete is a no-op (REQ-007)

Delete SHALL NOT call any Kibana API. Destroying the resource does not remove the previously imported saved objects from Kibana.

#### Scenario: Destroy does not call Kibana

- GIVEN a resource recorded in Terraform state
- WHEN destroy runs
- THEN the provider SHALL NOT call any Kibana API

### Requirement: Import result attributes (REQ-008)

After each create or update, the resource SHALL store the import result in state: `success` (whether all objects imported without error), `success_count` (number of successfully imported objects), `errors` (per-object error list), and `success_results` (per-object success list).

#### Scenario: Success result stored in state

- GIVEN all objects in `file_contents` import successfully
- WHEN create or update completes
- THEN `success` SHALL be `true`, `success_count` SHALL equal the number of imported objects, `errors` SHALL be empty, and `success_results` SHALL list each imported object

#### Scenario: Partial failure stored in state

- GIVEN some objects in `file_contents` fail to import
- WHEN create or update completes
- THEN `success` SHALL be `false`, `errors` SHALL list the failed objects, and `success_count` SHALL reflect the number of successfully imported objects

### Requirement: Error handling with ignore_import_errors (REQ-009)

When the Kibana Import API returns `success: false` and `ignore_import_errors` is `false` (the default), the resource SHALL:
- surface an error diagnostic (and no warning) when no objects imported successfully (`success_count == 0`)
- surface a warning diagnostic when at least one object imported successfully but others failed

When `ignore_import_errors` is `true`, the resource SHALL NOT fail and SHALL NOT emit a warning regardless of the import result; it SHALL still record the result attributes in state.

#### Scenario: All imports fail without ignore flag

- GIVEN `ignore_import_errors = false` (or unset)
- AND the Kibana API returns `success: false` with `success_count == 0`
- WHEN create or update runs
- THEN the provider SHALL surface an error diagnostic referencing the `errors` attribute

#### Scenario: Partial imports fail without ignore flag

- GIVEN `ignore_import_errors = false` (or unset)
- AND the Kibana API returns `success: false` with `success_count > 0`
- WHEN create or update runs
- THEN the provider SHALL surface a warning diagnostic and SHALL NOT fail apply

#### Scenario: Import errors ignored

- GIVEN `ignore_import_errors = true`
- AND the Kibana API returns `success: false`
- WHEN create or update runs
- THEN the provider SHALL NOT emit any error or warning diagnostic and SHALL record the result attributes in state

### Requirement: Overwrite flag (REQ-010)

The resource SHALL pass the `overwrite` attribute value to the Kibana Saved Objects Import API. When `overwrite` is `true`, the API SHALL automatically resolve conflict errors by overwriting existing saved objects.

#### Scenario: Overwrite conflicts

- GIVEN `overwrite = true` and existing conflicting saved objects in Kibana
- WHEN create or update runs
- THEN the provider SHALL send `overwrite: true` to the Import API, resolving conflicts automatically

### Requirement: create_new_copies flag (REQ-011)

The resource SHALL expose an optional `create_new_copies` boolean attribute. When set to `true`, the provider SHALL send `createNewCopies=true` on the Import API request. When unset or `false`, the provider SHALL omit the query parameter.

#### Scenario: create_new_copies forwarded to API

- GIVEN `create_new_copies = true`, a valid export in `file_contents`, and no conflicting `overwrite` or `compatibility_mode` flags in configuration
- WHEN create or update runs
- THEN the provider SHALL include `createNewCopies=true` on the Import API request

### Requirement: compatibility_mode flag (REQ-012)

The resource SHALL expose an optional `compatibility_mode` boolean attribute. When set to `true`, the provider SHALL send `compatibilityMode=true` on the Import API request. When unset or `false`, the provider SHALL omit the query parameter.

#### Scenario: compatibility_mode forwarded to API

- GIVEN `compatibility_mode = true`, a valid export in `file_contents`, and `create_new_copies` is not set to `true` in configuration
- WHEN create or update runs
- THEN the provider SHALL include `compatibilityMode=true` on the Import API request

### Requirement: Mutually exclusive import flags (REQ-013)

The resource SHALL reject, at plan time, combinations of `overwrite`, `create_new_copies`, and `compatibility_mode` that the Kibana Saved Objects Import API rejects. Specifically, `create_new_copies` is incompatible with `overwrite` and with `compatibility_mode`. These constraints SHALL be enforced via `ResourceWithConfigValidators` before any API call is made.

#### Scenario: create_new_copies conflicts with overwrite

- GIVEN both `create_new_copies = true` and `overwrite = true`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a configuration error diagnostic explaining the conflict

#### Scenario: create_new_copies conflicts with compatibility_mode

- GIVEN both `create_new_copies = true` and `compatibility_mode = true`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a configuration error diagnostic explaining the conflict

## Traceability

| Area | Primary files |
|------|---------------|
| Schema / Metadata / Configure / ConfigValidators | `internal/kibana/import_saved_objects/schema.go` |
| Create / Update / Import logic | `internal/kibana/import_saved_objects/create.go`, `internal/kibana/import_saved_objects/update.go` |
| Read (no-op) | `internal/kibana/import_saved_objects/read.go` |
| Delete (no-op) | `internal/kibana/import_saved_objects/delete.go` |
| kbapi multipart helper | `internal/clients/kibanaoapi/saved_objects_import.go` |
| Unit tests for helper | `internal/clients/kibanaoapi/saved_objects_import_test.go` |
