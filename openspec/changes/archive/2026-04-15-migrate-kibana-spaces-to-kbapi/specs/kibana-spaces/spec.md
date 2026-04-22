## ADDED Requirements

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

## MODIFIED Requirements

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

### Requirement: Connection (REQ-003)

The data source SHALL use the provider's configured Kibana connection resolved to the generated OpenAPI Kibana client (`generated/kbapi`) and `internal/clients/kibanaoapi` spaces helpers by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use that scoped client for the Get All Spaces API call.

#### Scenario: Provider-level Kibana client

- **WHEN** `kibana_connection` is not configured on the data source
- **THEN** the provider-level Kibana connection mapped to kbapi SHALL be used

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the data source
- **THEN** the scoped Kibana client derived from that block SHALL be used for kbapi Spaces list calls
