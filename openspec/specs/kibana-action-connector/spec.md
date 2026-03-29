# `elasticstack_kibana_action_connector` — Schema and Functional Requirements

Data source implementation: `internal/kibana/connector_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_action_connector` data source, which locates a Kibana action connector by name, space, and optional type and exposes its metadata as computed attributes.

## Schema

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

### Requirement: Read API (REQ-001)

The data source SHALL use the Kibana Get Connectors API (`GET /api/actions/connectors`) to list all connectors for the given `space_id` and then filter the results client-side by `name` and optionally by `connector_type_id`. When the Kibana API returns a non-success status, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN the Kibana Get Connectors API returns a non-success status
- WHEN the read runs
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Connector selection (REQ-002–REQ-004)

The data source SHALL match connectors by exact name equality. When `connector_type_id` is set to a non-empty string, the data source SHALL additionally filter connectors by that connector type. If no connector matches the name and optional type filter, the data source SHALL return an error diagnostic indicating the connector was not found. If more than one connector matches, the data source SHALL return an error diagnostic indicating that multiple connectors were found.

#### Scenario: No matching connector

- GIVEN no connector in the space matches the provided name (and optional type)
- WHEN read runs
- THEN the data source SHALL fail with a "not found" error

#### Scenario: Multiple matching connectors

- GIVEN more than one connector in the space shares the provided name and type
- WHEN read runs
- THEN the data source SHALL fail with a "multiple connectors found" error

### Requirement: Identity (REQ-005)

The data source SHALL set its `id` to a composite value in the format `<space_id>/<connector_id>` using `CompositeID{ClusterID: spaceID, ResourceID: connectorID}`.

#### Scenario: Composite id after read

- GIVEN a single matching connector is found
- WHEN read completes
- THEN `id` SHALL equal `<space_id>/<connector_id>`

### Requirement: State mapping (REQ-006–REQ-007)

When a single matching connector is found, the data source SHALL populate the computed attributes `connector_id`, `space_id`, `name`, `connector_type_id`, `config`, `is_deprecated`, `is_missing_secrets`, and `is_preconfigured` from the connector model. The `config` attribute SHALL be a JSON-serialized string of the connector's configuration object with null-valued keys removed; for known connector types with type-specific handlers, the config SHALL be re-marshaled through that type's schema to strip unsupported fields.

#### Scenario: Config JSON serialization

- GIVEN a connector with a non-null config
- WHEN read completes
- THEN `config` SHALL be a valid JSON string representing the connector's configuration

### Requirement: Connection (REQ-008)

The data source SHALL use the provider's configured client by default. When `NewAPIClientFromSDKResource` resolves a resource-level `elasticsearch_connection` block, the data source SHALL use that scoped client for the API call.

#### Scenario: Provider-level Kibana client

- GIVEN no `elasticsearch_connection` block is present
- WHEN the API call runs
- THEN the provider-level client SHALL be used
