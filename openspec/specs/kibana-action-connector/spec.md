# `elasticstack_kibana_action_connector` — Schema and Functional Requirements

Resource implementation: `internal/kibana/connectors`
Data source implementation: `internal/kibana/connector_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_action_connector` resource and data source. The resource creates and manages Kibana action connectors via the Kibana Connectors API, supporting user-supplied connector IDs, space scoping, and type-specific configuration. The data source locates a Kibana action connector by name, space, and optional type and exposes its metadata as computed attributes.

## Schema

### Resource

```hcl
resource "elasticstack_kibana_action_connector" "example" {
  name             = <required, string>
  connector_type_id = <required, string>  # RequiresReplace

  connector_id = <optional+computed, string>  # UUID v1 or v4; RequiresReplace, UseStateForUnknown
  space_id     = <optional+computed, string>  # default "default"; RequiresReplace
  config       = <optional+computed, string>  # custom JSON type (ConfigType) with contextual defaults
  secrets      = <optional, sensitive, json string>

  # Computed
  id                = <computed, string>    # <space_id>/<connector_id>; UseStateForUnknown
  is_deprecated     = <computed, bool>
  is_missing_secrets = <computed, bool>
  is_preconfigured  = <computed, bool>

  # Resource-level Kibana connection override
  kibana_connection {
    endpoints    = <optional, list(string)>
    username     = <optional, string>
    password     = <optional, string>
    api_key      = <optional, string>
    bearer_token = <optional, string>
    insecure     = <optional, bool>
    ca_file      = <optional, string>
    ca_data      = <optional, string>
    cert_file    = <optional, string>
    cert_data    = <optional, string>
    key_file     = <optional, string>
    key_data     = <optional, string>
    headers      = <optional, map(string)>
  }
}
```

### Data source

```hcl
data "elasticstack_kibana_action_connector" "example" {
  name              = <required, string>
  space_id          = <optional, string, default: "default">
  connector_type_id = <optional, string>

  # Computed outputs
  connector_id        = <computed, string>
  config              = <computed, string>  # JSON string of connector configuration
  is_deprecated       = <computed, bool>
  is_missing_secrets  = <computed, bool>
  is_preconfigured    = <computed, bool>
}
```

## Requirements

### Requirement: Connector CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Kibana Create Connector API to create connectors. The resource SHALL use the Kibana Get Connector API to read a connector by ID. The resource SHALL use the Kibana Update Connector API to update connectors. The resource SHALL use the Kibana Delete Connector API to delete connectors.

#### Scenario: Lifecycle uses documented APIs

- GIVEN a connector managed by this resource
- WHEN create, update, read, or delete runs
- THEN the provider SHALL call the appropriate Kibana Connectors API endpoint

### Requirement: API error surfacing (REQ-004)

When the Kibana API returns a non-success response for create, update, read, or delete (other than "not found" on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Non-success API response

- GIVEN the Kibana API returns an error on create, update, read, or delete (except connector not found on read)
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` in the format `<space_id>/<connector_id>`. When creating or updating a connector, the resource SHALL compute `id` using the space ID and the connector ID returned from the API.

#### Scenario: Computed id after apply

- GIVEN a successful create or update
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<connector_id>`

### Requirement: Import (REQ-006)

The resource SHALL support import by accepting a connector's `id` attribute value (in the format `<space_id>/<connector_id>`) and setting it on the state attribute `id` directly via `response.State.SetAttribute`.

#### Scenario: Import by composite id

- GIVEN an import is requested with `<space_id>/<connector_id>`
- WHEN import runs
- THEN the id SHALL be set in state and a subsequent read SHALL populate all other attributes

### Requirement: Lifecycle — replacement on connector_type_id, connector_id, and space_id changes (REQ-007)

When `connector_type_id`, `connector_id`, or `space_id` changes, the resource SHALL require replacement (destroy/recreate) rather than an in-place update.

#### Scenario: Changing connector type

- GIVEN a configuration change to `connector_type_id`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Connection (REQ-008–REQ-009)

The resource SHALL use the provider's configured client by default. When the `kibana_connection` block is configured on the resource, the resource SHALL use that connection to create a Kibana client for all API calls of that instance.

#### Scenario: Resource-scoped Kibana connection

- GIVEN `kibana_connection` is set on the resource
- WHEN any API call runs for that instance
- THEN the client SHALL be built from that block

### Requirement: Compatibility — connector_id requires Kibana >= 8.8.0 (REQ-010)

When `connector_id` is configured (non-empty), the resource SHALL verify the Elastic Stack server version is at least 8.8.0. If the server version is lower, the resource SHALL fail with an "Unsupported Elastic Stack version" error and SHALL NOT call the Create Connector API.

#### Scenario: Preconfigured connector ID on older cluster

- GIVEN `connector_id` is set and the Elastic Stack version is below 8.8.0
- WHEN create runs
- THEN the resource SHALL fail with an "Unsupported Elastic Stack version" error

### Requirement: Create and update behavior (REQ-011–REQ-013)

When creating a connector, the resource SHALL call the Create Connector API and then read the connector back to populate all computed fields in state. If the connector cannot be read immediately after a successful create, the resource SHALL return an error diagnostic indicating the connector was not found after creation. When updating a connector, the resource SHALL call the Update Connector API using the connector ID parsed from `id`, then read the connector back to populate computed fields; if the connector is not found after update, the resource SHALL return an error diagnostic.

#### Scenario: Post-create read

- GIVEN a successful Create Connector response
- WHEN the provider refreshes state
- THEN it SHALL read the connector and populate state, or return an error if the connector is missing

### Requirement: Read and refresh (REQ-014–REQ-015)

When refreshing state, the resource SHALL parse `id` to extract the space ID and connector ID, then call the Get Connector API. If the connector is not found during refresh, the resource SHALL remove itself from Terraform state.

#### Scenario: Connector removed in Kibana

- GIVEN refresh runs and the connector no longer exists
- WHEN the API returns nil without an error
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-016)

When destroying, the resource SHALL parse `id` to extract the connector ID and space ID and then delete the connector via the Delete Connector API.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN the provider SHALL call Delete Connector for the connector ID and space ID parsed from `id`

### Requirement: Config JSON mapping (REQ-017–REQ-018)

When `config` is configured and known, the resource SHALL sanitize it (removing null-valued keys) before sending it to the Kibana API. When reading a connector from the API, if `config` is non-empty the resource SHALL re-populate `config` using a type-aware value that applies connector-type-specific contextual defaults, preserving semantic equivalence between plan and state.

#### Scenario: Config with null values

- GIVEN `config` contains null-valued keys
- WHEN the API payload is built
- THEN those null keys SHALL be stripped before sending to the Kibana API

### Requirement: Secrets omitted from read (REQ-019)

The `secrets` attribute is write-only by nature. The resource SHALL NOT attempt to read back `secrets` from the Kibana API (the API does not return secret values). The `secrets` value in state SHALL reflect only what the practitioner configured.

#### Scenario: Secrets not returned by API

- GIVEN a connector with `secrets` configured
- WHEN read or refresh runs
- THEN `secrets` SHALL retain its configured value in state (not be overwritten by the API response)

### Requirement: State upgrade v0 to v1 (REQ-020–REQ-022)

The resource SHALL support upgrading prior state schema version 0 to schema version 1. During v0→v1 upgrade, if `config` in the prior state is an empty string, the resource SHALL remove the `config` key from the upgraded state. During v0→v1 upgrade, if `secrets` in the prior state is an empty string, the resource SHALL remove the `secrets` key from the upgraded state. If the prior state JSON cannot be parsed during upgrade, the resource SHALL return a "Failed to unmarshal state" error diagnostic and SHALL NOT produce upgraded state.

#### Scenario: Empty config in prior state

- GIVEN prior state has `config` set to an empty string
- WHEN v0→v1 upgrade runs
- THEN `config` SHALL be absent from the upgraded state

#### Scenario: Corrupt prior state

- GIVEN prior state JSON cannot be parsed
- WHEN upgrade runs
- THEN a "Failed to unmarshal state" diagnostic SHALL be returned

---

## Data source requirements

### Requirement: Read API (REQ-DS-001)

The data source SHALL use the Kibana Get Connectors API (`GET /api/actions/connectors`) to list all connectors for the given `space_id` and then filter the results client-side by `name` and optionally by `connector_type_id`. When the Kibana API returns a non-success status, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN the Kibana Get Connectors API returns a non-success status
- WHEN the read runs
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Connector selection (REQ-DS-002–REQ-DS-004)

The data source SHALL match connectors by exact name equality. When `connector_type_id` is set to a non-empty string, the data source SHALL additionally filter connectors by that connector type. If no connector matches the name and optional type filter, the data source SHALL return an error diagnostic indicating the connector was not found. If more than one connector matches, the data source SHALL return an error diagnostic indicating that multiple connectors were found.

#### Scenario: No matching connector

- GIVEN no connector in the space matches the provided name (and optional type)
- WHEN read runs
- THEN the data source SHALL fail with a "not found" error

#### Scenario: Multiple matching connectors

- GIVEN more than one connector in the space shares the provided name and type
- WHEN read runs
- THEN the data source SHALL fail with a "multiple connectors found" error

### Requirement: Identity (REQ-DS-005)

The data source SHALL set its `id` to a composite value in the format `<space_id>/<connector_id>` using `CompositeID{ClusterID: spaceID, ResourceID: connectorID}`.

#### Scenario: Composite id after read

- GIVEN a single matching connector is found
- WHEN read completes
- THEN `id` SHALL equal `<space_id>/<connector_id>`

### Requirement: State mapping (REQ-DS-006–REQ-DS-007)

When a single matching connector is found, the data source SHALL populate the computed attributes `connector_id`, `space_id`, `name`, `connector_type_id`, `config`, `is_deprecated`, `is_missing_secrets`, and `is_preconfigured` from the connector model. The `config` attribute SHALL be a JSON-serialized string of the connector's configuration object with null-valued keys removed; for known connector types with type-specific handlers, the config SHALL be re-marshaled through that type's schema to strip unsupported fields.

#### Scenario: Config JSON serialization

- GIVEN a connector with a non-null config
- WHEN read completes
- THEN `config` SHALL be a valid JSON string representing the connector's configuration

### Requirement: Connection (REQ-DS-008)

The data source SHALL use the provider's configured client by default. When `NewAPIClientFromSDKResource` resolves a resource-level `elasticsearch_connection` block, the data source SHALL use that scoped client for the API call.

#### Scenario: Provider-level Kibana client

- GIVEN no `elasticsearch_connection` block is present
- WHEN the API call runs
- THEN the provider-level client SHALL be used
