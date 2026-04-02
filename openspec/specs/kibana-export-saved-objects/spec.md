# `elasticstack_kibana_export_saved_objects` — Schema and Functional Requirements

Data source implementation: `internal/kibana/exportsavedobjects`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_export_saved_objects` data source, which exports a specified set of Kibana saved objects and makes the resulting NDJSON payload available in Terraform state. This data source is read-only and calls the Kibana Export Saved Objects API on every refresh.

## Schema

```hcl
data "elasticstack_kibana_export_saved_objects" "example" {
  space_id = <optional, string>  # defaults to "default" when unset

  objects = <required, list(object({  # at least 1 entry required
    type = <required, string>  # saved object type
    id   = <required, string>  # saved object id
  }))>

  exclude_export_details  = <optional, bool>  # defaults to true when unset
  include_references_deep = <optional, bool>  # defaults to true when unset

  # Computed outputs
  id               = <computed, string>  # composite "<space_id>/export"
  exported_objects = <computed, string>  # raw NDJSON export payload
}
```

## Requirements

### Requirement: Export Saved Objects API (REQ-001)

The data source SHALL use the Kibana Post Saved Objects Export API (`POST /api/saved_objects/_export`) to export saved objects. When the API returns a non-200 status code, the data source SHALL surface an error diagnostic including the HTTP status code and response body and SHALL NOT write any state.

#### Scenario: Successful export

- GIVEN a valid list of objects and optional parameters
- WHEN Terraform refreshes the data source
- THEN the provider SHALL call `POST /api/saved_objects/_export` with the constructed request body

#### Scenario: Non-200 API response

- GIVEN the Kibana API returns a non-200 HTTP status code
- WHEN read handles the response
- THEN the data source SHALL surface an error diagnostic with the status code and response body and SHALL NOT write state

#### Scenario: HTTP transport error

- GIVEN the API call itself fails with a transport error
- WHEN read handles the error
- THEN the data source SHALL surface an error diagnostic and SHALL NOT write state

### Requirement: Identity (REQ-002)

The data source SHALL expose a computed `id` in the format `<space_id>/export`, constructed using the resolved `space_id` (after applying the default of `"default"` when unset) and the fixed string `"export"`.

#### Scenario: Computed id uses resolved space_id

- GIVEN `space_id` is set to `"my-space"`
- WHEN read completes
- THEN `id` SHALL equal `"my-space/export"`

#### Scenario: Default space_id in id

- GIVEN `space_id` is not configured
- WHEN read completes
- THEN `id` SHALL equal `"default/export"`

### Requirement: Connection (REQ-003)

The data source SHALL use the provider's configured Kibana OAPI client by default. The data source does not support a resource-level connection override.

#### Scenario: Provider-level Kibana client

- GIVEN no resource-level connection override
- WHEN any API call runs
- THEN the provider-level Kibana OAPI client SHALL be used

### Requirement: Default values for optional boolean arguments (REQ-004)

When `exclude_export_details` is not set in configuration (null or unknown), the data source SHALL default it to `true` for the API request and SHALL record `true` in state. When `include_references_deep` is not set in configuration (null or unknown), the data source SHALL default it to `true` for the API request and SHALL record `true` in state.

#### Scenario: exclude_export_details defaults to true

- GIVEN `exclude_export_details` is not configured
- WHEN read runs
- THEN the API request SHALL include `exclude_export_details: true` and state SHALL record `true`

#### Scenario: include_references_deep defaults to true

- GIVEN `include_references_deep` is not configured
- WHEN read runs
- THEN the API request SHALL include `include_references_deep: true` and state SHALL record `true`

### Requirement: Default space_id (REQ-005)

When `space_id` is not configured (null or unknown), the data source SHALL use `"default"` as the Kibana space identifier for the API request and SHALL record `"default"` in state.

#### Scenario: Unset space_id resolves to default

- GIVEN `space_id` is not configured
- WHEN read runs
- THEN the API request SHALL target the `"default"` space and state SHALL record `space_id = "default"`

### Requirement: Objects list validation (REQ-006)

The `objects` argument MUST contain at least one entry. The data source schema SHALL enforce a minimum list size of 1 via a validator, rejecting configurations with an empty `objects` list before any API call is made.

#### Scenario: Empty objects list rejected

- GIVEN `objects` is configured as an empty list
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned without calling the Kibana API

### Requirement: Read and state mapping (REQ-007)

When the API returns HTTP 200, the data source SHALL store the raw response body (NDJSON payload) as the `exported_objects` computed attribute. The data source SHALL copy the configured `objects` list from configuration to state unchanged. All resolved values (`space_id`, `exclude_export_details`, `include_references_deep`) SHALL be written to state.

#### Scenario: Exported objects stored in state

- GIVEN a successful HTTP 200 response from the Kibana export API
- WHEN read completes
- THEN `exported_objects` SHALL contain the raw NDJSON response body and all other attributes SHALL be recorded in state
