# `elasticstack_kibana_spaces` — Schema and Functional Requirements

Data source implementation: `internal/kibana/spaces`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_spaces` data source, which lists all existing Kibana spaces and exposes their attributes in Terraform state. This data source is read-only and does not manage individual spaces.

## Schema

```hcl
data "elasticstack_kibana_spaces" "example" {
  # No required arguments — always retrieves all spaces

  # Computed outputs
  id     = <computed, string> # fixed value "spaces"
  spaces = <computed, list(object({
    id                = <computed, string>
    name              = <required, string>            # display name; required in nested schema
    description       = <optional, string>
    disabled_features = <computed, list(string)>
    initials          = <computed, string>
    color             = <computed, string>
    image_url         = <optional, string>
    solution          = <computed, string>
  }))>
}
```

## Requirements

### Requirement: List Spaces API (REQ-001)

The data source SHALL use the Kibana Get All Spaces API to retrieve all spaces ([docs](https://www.elastic.co/guide/en/kibana/master/spaces-api-get-all.html)). When the API returns an error, the data source SHALL surface the error to Terraform diagnostics and SHALL NOT write any state.

#### Scenario: API call on read

- GIVEN the data source is referenced in configuration
- WHEN Terraform refreshes state
- THEN the provider SHALL call the Kibana Get All Spaces API to retrieve the full list of spaces

#### Scenario: API error surfaces to diagnostics

- GIVEN the Kibana API returns an error
- WHEN read runs
- THEN the error SHALL appear in Terraform diagnostics and no state SHALL be written

### Requirement: Identity (REQ-002)

The data source SHALL expose a computed `id` with the fixed value `"spaces"` to satisfy Terraform's identity requirements for data sources. The `id` does not represent any Kibana-side identifier.

#### Scenario: Fixed id after read

- GIVEN a successful read
- WHEN state is written
- THEN `id` SHALL equal `"spaces"`

### Requirement: Connection (REQ-003)

The data source SHALL use the provider's configured Kibana client (`KibanaSpaces` API client) by default. The data source does not support a resource-level connection override.

#### Scenario: Provider-level Kibana client

- GIVEN no resource-level connection override
- WHEN any API call runs
- THEN the provider-level Kibana client SHALL be used

### Requirement: Read and state mapping (REQ-004–REQ-005)

When the Kibana API returns successfully, the data source SHALL map each returned space to an entry in the `spaces` list. For each space, the data source SHALL populate `id`, `name`, `description`, `disabled_features`, `initials`, `color`, `image_url`, and `solution` from the corresponding API response fields. If converting the `disabled_features` list to a Terraform list type produces diagnostics, the data source SHALL surface those diagnostics and SHALL NOT write state.

#### Scenario: Full space list mapped to state

- GIVEN the Kibana API returns a list of spaces
- WHEN read completes successfully
- THEN each space SHALL appear as an entry in the `spaces` computed attribute with all fields populated from the API response

#### Scenario: disabled_features conversion error

- GIVEN the `disabled_features` field for a space cannot be converted to a Terraform list
- WHEN read runs
- THEN the data source SHALL surface the conversion diagnostics and SHALL NOT write state
