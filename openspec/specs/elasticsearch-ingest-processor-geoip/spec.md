# `elasticstack_elasticsearch_ingest_processor_geoip` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest`

## Purpose

Define the schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_geoip` data source, which builds an Elasticsearch geoip ingest processor configuration and serializes it to JSON. This data source performs no API calls; it computes a JSON representation and a deterministic hash ID from the configured processor parameters. Note that the geoip processor does not expose the common processor fields (`description`, `if`, `ignore_failure`, `on_failure`, `tag`).

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_geoip" "example" {
  field         = <required, string>     # IP address field for geographical lookup
  target_field  = <optional, string>     # Field to hold the geographical information; default "geoip"
  database_file = <optional, string>     # MaxMind database file name; omitted when unset
  properties    = <optional, set(string)> # Properties to add to target_field; omitted when unset
  ignore_missing = <optional, bool>      # Silently exit if field does not exist; default false
  first_only    = <optional, bool>       # Return only first match when field is an array; default true

  # Computed outputs
  id   = <computed, string>  # Hash of the JSON output
  json = <computed, string>  # Serialized processor JSON
}
```

## Requirements

### Requirement: No API calls (REQ-001)

The data source SHALL perform no Elasticsearch API calls. It SHALL NOT require or use an `elasticsearch_connection` block.

#### Scenario: Read without a connection

- GIVEN any valid configuration for this data source
- WHEN the data source is read
- THEN no Elasticsearch API is called and no connection credentials are required

### Requirement: JSON output (REQ-002)

The data source SHALL serialize the processor configuration to a JSON string stored in `json`. The JSON SHALL be wrapped in a top-level `"geoip"` key, producing the shape `{"geoip": {...}}`.

#### Scenario: JSON wrapping

- GIVEN a valid configuration
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a top-level `"geoip"` key whose value is the processor parameters

### Requirement: Hash identity (REQ-003)

The data source SHALL set `id` to a deterministic hash of the `json` output. Two configurations that produce identical JSON SHALL produce the same `id`.

#### Scenario: Deterministic id

- GIVEN the same processor configuration applied twice
- WHEN both data sources are read
- THEN both SHALL have identical `id` values

### Requirement: Required attribute (REQ-004)

The data source SHALL require `field`. If it is absent, Terraform SHALL report a configuration error before the read function is called.

#### Scenario: Missing required attribute

- GIVEN a configuration that omits `field`
- WHEN Terraform validates the configuration
- THEN a configuration error SHALL be raised

### Requirement: Default attribute values (REQ-005)

When `target_field` is not configured, the data source SHALL use `"geoip"`. When `ignore_missing` is not configured, the data source SHALL use `false`. When `first_only` is not configured, the data source SHALL use `true`.

#### Scenario: Defaults in JSON output

- GIVEN a configuration that sets only `field`
- WHEN the data source is read
- THEN `json` SHALL include `"target_field": "geoip"`, `"ignore_missing": false`, and `"first_only": true`

### Requirement: Optional database_file field (REQ-006)

When `database_file` is configured, the data source SHALL include it in the serialized JSON. When `database_file` is not configured, it SHALL be omitted from the JSON.

#### Scenario: database_file omitted when unset

- GIVEN `database_file` is not configured
- WHEN the data source is read
- THEN the `json` output SHALL not include a `"database_file"` key

### Requirement: Optional properties field (REQ-007)

When `properties` is configured with one or more values, the data source SHALL include it in the serialized JSON. When `properties` is not configured or empty, it SHALL be omitted from the JSON.

#### Scenario: properties omitted when unset

- GIVEN `properties` is not configured
- WHEN the data source is read
- THEN the `json` output SHALL not include a `"properties"` key

### Requirement: No common processor fields (REQ-008)

The geoip processor data source SHALL NOT expose `description`, `if`, `ignore_failure`, `on_failure`, or `tag` attributes. These common processor fields are not part of the geoip processor schema.

#### Scenario: No common fields in schema

- GIVEN the data source schema definition
- WHEN inspecting available attributes
- THEN `description`, `if`, `ignore_failure`, `on_failure`, and `tag` SHALL NOT be valid attributes
