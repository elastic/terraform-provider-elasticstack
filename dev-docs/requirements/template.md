# `<RESOURCE_NAME>` — Schema and Functional Requirements

Resource implementation: `<GO_PACKAGE_OR_DIR>`

## Schema

<!-- Example 
```hcl
resource "<PROVIDER_RESOURCE_TYPE>" "example" {
  # Required identity/config
  name = <required, string>

  # Optional arguments
  description = <optional, string> # note any server version requirements

  # JSON (normalized) strings
  metadata = <optional | optional+computed, json string>

  # Sets of strings
  cluster = <optional, set(string)>
}
```
-->

## Requirements

<!-- Examples 
- **[REQ-001] (API)**: The resource shall use the `<API_NAME>` API to create and update `<OBJECT_PLURAL>` ([docs](<LINK>)).
- **[REQ-002] (API)**: The resource shall use the `<API_NAME>` API to read `<OBJECT_PLURAL>` ([docs](<LINK>)).
- **[REQ-003] (API)**: The resource shall use the `<API_NAME>` API to delete `<OBJECT_PLURAL>` ([docs](<LINK>)).
- **[REQ-004] (API)**: When the API returns a non-success status for create, update, read, or delete requests (other than “not found” on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-005] (Import)**: The resource shall support import by accepting an `id` in the format `<ID_FORMAT>` and persisting it to state.
-->
