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

The data source SHALL use the Kibana Get All Spaces API to retrieve all spaces ([docs](https://www.elastic.co/guide/en/kibana/master/spaces-api-get-all.html)) via the generated OpenAPI kbapi client and `internal/clients/kibanaoapi` helpers. When the API returns an error, the data source SHALL surface the error to Terraform diagnostics and SHALL NOT write any state.

#### Scenario: API call on read

- **GIVEN** the data source is referenced in configuration
- **WHEN** Terraform refreshes state
- **THEN** the provider SHALL call the Kibana Get All Spaces API through kbapi to retrieve the full list of spaces

#### Scenario: API error surfaces to diagnostics

- **GIVEN** the Kibana API returns an error
- **WHEN** read runs
- **THEN** the error SHALL appear in Terraform diagnostics and no state SHALL be written

### Requirement: Identity (REQ-002)

The data source SHALL expose a computed `id` with the fixed value `"spaces"` to satisfy Terraform's identity requirements for data sources. The `id` does not represent any Kibana-side identifier.

#### Scenario: Fixed id after read

- GIVEN a successful read
- WHEN state is written
- THEN `id` SHALL equal `"spaces"`

### Requirement: Connection (REQ-003)

The data source SHALL use the provider's configured Kibana connection resolved to the generated OpenAPI Kibana client (`generated/kbapi`) and `internal/clients/kibanaoapi` spaces helpers by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use that scoped client for the Get All Spaces API call.

#### Scenario: Provider-level Kibana client

- **WHEN** `kibana_connection` is not configured on the data source
- **THEN** the provider-level Kibana connection mapped to kbapi SHALL be used

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the data source
- **THEN** the scoped Kibana client derived from that block SHALL be used for kbapi Spaces list calls

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

### Requirement: Kibana OpenAPI client for listing spaces (REQ-006)

The `elasticstack_kibana_spaces` implementation SHALL retrieve spaces using the generated OpenAPI Kibana client (`generated/kbapi`) and `internal/clients/kibanaoapi` list-spaces helpers. The implementation SHALL NOT call the legacy `go-kibana-rest` `KibanaSpaces.List` method after this migration.

#### Scenario: Read uses kbapi list transport

- **GIVEN** the data source read runs successfully
- **WHEN** the provider fetches all spaces from Kibana
- **THEN** the HTTP call SHALL be executed through the kbapi client configured for the effective Kibana connection

### Requirement: Typed list items aligned with legacy KibanaSpace (REQ-007)

The kbapi types used for each entry in the Get All Spaces response SHALL decode the same logical fields as `github.com/disaster37/go-kibana-rest/v8/kbapi.KibanaSpace` for provider mapping (`id`, `name`, `description`, `disabledFeatures`, `initials`, `color`, `imageUrl`, `solution`, and `_reserved` when returned by Kibana), produced after `generated/kbapi/transform_schema.go` updates and regeneration.

#### Scenario: Nested spaces attributes unchanged for a fixed API payload

- **GIVEN** a list response payload identical to one handled before the migration
- **WHEN** read completes
- **THEN** each nested `spaces` entry SHALL carry the same Terraform-visible values for `id`, `name`, `description`, `disabled_features`, `initials`, `color`, `image_url`, and `solution` as the pre-migration mapping

