# `elasticstack_elasticsearch_security_user` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/security/user_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_security_user` data source, which reads an existing Elasticsearch security user by username and exposes its attributes as computed values.

## Schema

```hcl
data "elasticstack_elasticsearch_security_user" "example" {
  username = <required, string>

  # Computed outputs
  id        = <computed, string>  # <cluster_uuid>/<username>
  full_name = <computed, string>
  email     = <computed, string>
  roles     = <computed, set(string)>
  metadata  = <computed, string>  # JSON-serialized metadata object
  enabled   = <computed, bool>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

## Requirements

### Requirement: Read API (REQ-001)

The data source SHALL use the Elasticsearch Get users API to fetch the user identified by `username`. When the API returns a non-success response, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN the Elasticsearch Get users API returns an error
- WHEN read runs
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-002)

The data source SHALL set `id` to a composite value in the format `<cluster_uuid>/<username>` by resolving the cluster UUID from the connected Elasticsearch cluster and combining it with the configured `username`.

#### Scenario: Computed id

- GIVEN a successful read of an existing user
- WHEN read completes
- THEN `id` SHALL equal `<cluster_uuid>/<username>`

### Requirement: User not found (REQ-003)

When the Elasticsearch Get users API returns nil without an error (user not found), the data source SHALL remove itself from state by calling `d.SetId("")` and return without error.

#### Scenario: User does not exist

- GIVEN the specified username does not exist in Elasticsearch
- WHEN read runs
- THEN the data source SHALL clear its id and return no error

### Requirement: State mapping (REQ-004–REQ-005)

When the user is found, the data source SHALL populate `username`, `email`, `full_name`, `roles`, `enabled`, and `metadata` from the API response. The `metadata` attribute SHALL be a JSON-serialized string produced by marshaling the user's metadata object; if marshaling fails, the data source SHALL surface an error to Terraform diagnostics.

#### Scenario: Metadata marshaling error

- GIVEN the user's metadata cannot be marshaled to JSON
- WHEN read populates state
- THEN the data source SHALL return an error diagnostic

### Requirement: Connection (REQ-006–REQ-007)

The data source SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured, the data source SHALL use that connection to construct an Elasticsearch client for its API call.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set
- WHEN the API call runs
- THEN the client SHALL be built from that block
