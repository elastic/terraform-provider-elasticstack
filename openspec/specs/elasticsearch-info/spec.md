# `elasticstack_elasticsearch_info` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/cluster/cluster_info_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_info` data source, which reads cluster metadata from the Elasticsearch root info API and exposes cluster identity, node name, tagline, and detailed version information as computed attributes.

## Schema

```hcl
data "elasticstack_elasticsearch_info" "example" {
  # Computed outputs
  cluster_name = <computed, string>
  cluster_uuid = <computed, string>
  name         = <computed, string>   # node name
  tagline      = <computed, string>

  version {
    build_date                          = <computed, string>
    build_flavor                        = <computed, string>
    build_hash                          = <computed, string>
    build_snapshot                      = <computed, bool>
    build_type                          = <computed, string>
    lucene_version                      = <computed, string>
    minimum_index_compatibility_version = <computed, string>
    minimum_wire_compatibility_version  = <computed, string>
    number                              = <computed, string>
  }

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

The data source SHALL use the Elasticsearch cluster info API (root `GET /`) to retrieve cluster metadata. When the API returns an error, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN the Elasticsearch cluster info API returns an error
- WHEN read runs
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-002)

The data source SHALL set `id` to the cluster's UUID (`cluster_uuid`) returned by the API.

#### Scenario: Computed id equals cluster_uuid

- GIVEN a successful API response
- WHEN read completes
- THEN `id` SHALL equal the `cluster_uuid` value from the response

### Requirement: State mapping — top-level fields (REQ-003)

The data source SHALL populate `cluster_uuid`, `cluster_name`, `name`, and `tagline` directly from the API response. If setting any attribute fails, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: Top-level attributes set

- GIVEN a successful API response
- WHEN read completes
- THEN `cluster_uuid`, `cluster_name`, `name`, and `tagline` SHALL reflect the response values

### Requirement: State mapping — version block (REQ-004)

The data source SHALL populate the `version` list attribute as a single-element list containing all version sub-fields: `build_date` (formatted as a string), `build_flavor`, `build_hash`, `build_snapshot`, `build_type`, `lucene_version`, `minimum_index_compatibility_version`, `minimum_wire_compatibility_version`, and `number`. If setting the `version` attribute fails, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: Version block populated

- GIVEN a successful API response
- WHEN read completes
- THEN the `version` block SHALL contain exactly one element with all sub-fields populated from the response

### Requirement: Connection (REQ-005–REQ-006)

The data source SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured, the data source SHALL use that connection to construct an Elasticsearch client for its API call.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set
- WHEN the API call runs
- THEN the client SHALL be built from that block
