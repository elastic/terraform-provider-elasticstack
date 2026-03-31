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
  overwrite            = <optional, bool>    # automatically resolve conflicts by overwriting
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

- The resource uses the legacy Kibana SDK client (`GetKibanaClient`), not the Kibana OpenAPI client.
- Read and Delete are intentional no-ops; the resource is not refreshed from Kibana on plan.
- The resource does not support Terraform import.
- The resource does not declare a custom state upgrader.

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

### Requirement: Provider-level Kibana client only (REQ-004)

The resource SHALL use the provider's configured Kibana client for create and update. The resource SHALL NOT support a resource-local connection override in its schema or request path.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN create or update runs
- THEN all API operations SHALL use the provider-level Kibana client

### Requirement: Create and update share the same import logic (REQ-005)

Create and update SHALL both invoke the same import logic: reading the plan, calling the Kibana Saved Objects Import API, and writing the result to state. Changing `file_contents`, `overwrite`, or `space_id` SHALL trigger an in-place update that re-runs the import.

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

## Traceability

| Area | Primary files |
|------|---------------|
| Schema / Metadata / Configure | `internal/kibana/import_saved_objects/schema.go` |
| Create / Update / Import logic | `internal/kibana/import_saved_objects/create.go`, `internal/kibana/import_saved_objects/update.go` |
| Read (no-op) | `internal/kibana/import_saved_objects/read.go` |
| Delete (no-op) | `internal/kibana/import_saved_objects/delete.go` |
