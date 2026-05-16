# `elasticstack_elasticsearch_ingest_processor_geoip` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest`

## Purpose

Define the schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_geoip` data source, which builds an Elasticsearch geoip ingest processor configuration and serializes it to JSON. This data source performs no API calls; it computes a JSON representation and a deterministic hash ID from the configured processor parameters. The geoip processor exposes the common processor fields (`description`, `if`, `ignore_failure`, `on_failure`, `tag`) as optional attributes.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_geoip" "example" {
  field         = <required, string>     # IP address field for geographical lookup
  target_field  = <optional, string>     # Field to hold the geographical information; default "geoip"
  database_file = <optional, string>     # MaxMind database file name; omitted when unset
  properties    = <optional, set(string)> # Properties to add to target_field; omitted when unset
  ignore_missing = <optional, bool>      # Silently exit if field does not exist; default false
  first_only    = <optional, bool>       # Return only first match when field is an array; default true

  # Common processor fields
  description   = <optional, string>     # Description of the processor; omitted when unset
  if            = <optional, string>     # Condition for running the processor; omitted when unset
  ignore_failure = <optional, bool>      # Ignore failures for the processor; default false
  on_failure    = <optional, list(string)> # JSON strings for on_failure actions; omitted when unset
  tag           = <optional, string>     # Identifier for the processor; omitted when unset

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

### Requirement: Common processor fields (REQ-008)

The geoip processor data source SHALL expose `description`, `if`, `ignore_failure`, `on_failure`, and `tag` as optional attributes. When configured, each SHALL be included in the serialized JSON. When not configured, each SHALL be omitted from the JSON (except `ignore_failure`, which defaults to `false` and is always included).

#### Scenario: Common fields in schema

- GIVEN the data source schema definition
- WHEN inspecting available attributes
- THEN `description`, `if`, `ignore_failure`, `on_failure`, and `tag` SHALL be valid optional attributes

#### Scenario: Common fields included in JSON when configured

- GIVEN a configuration that sets `description = "geoip lookup"`, `if = "ctx.ip != null"`, `ignore_failure = true`, `on_failure = ['{"set":{"field":"error.message","value":"geoip failed"}}']`, and `tag = "geoip-tag"`
- WHEN the data source is read
- THEN `json` SHALL include `"description": "geoip lookup"`, `"if": "ctx.ip != null"`, `"ignore_failure": true`, `"on_failure": [{"set":{"field":"error.message","value":"geoip failed"}}]`, and `"tag": "geoip-tag"`

#### Scenario: Common fields omitted when not configured

- GIVEN a configuration that does not set `description`, `if`, `on_failure`, or `tag`
- WHEN the data source is read
- THEN `json` SHALL omit `"description"`, `"if"`, `"on_failure"`, and `"tag"` keys
- AND `json` SHALL include `"ignore_failure": false`

### Requirement: description field (REQ-009)

The data source SHALL accept an optional `description` string attribute. When configured, it SHALL be included in the serialized JSON under the `"description"` key. When not configured, the key SHALL be omitted.

#### Scenario: description configured

- GIVEN `description = "Lookup geoip for client IP"`
- WHEN the data source is read
- THEN `json` SHALL include `"description": "Lookup geoip for client IP"`

### Requirement: if field (REQ-010)

The data source SHALL accept an optional `if` string attribute. When configured, it SHALL be included in the serialized JSON under the `"if"` key. When not configured, the key SHALL be omitted.

#### Scenario: if configured

- GIVEN `if = "ctx.ip != null"`
- WHEN the data source is read
- THEN `json` SHALL include `"if": "ctx.ip != null"`

### Requirement: on_failure field (REQ-011)

The data source SHALL accept an optional `on_failure` list of JSON string attributes. When configured with one or more elements, each element SHALL be parsed as JSON and included in the serialized JSON under the `"on_failure"` key as an array of objects. When not configured, the key SHALL be omitted.

#### Scenario: on_failure configured

- GIVEN `on_failure = ['{"set":{"field":"error.message","value":"geoip failed"}}']`
- WHEN the data source is read
- THEN `json` SHALL include `"on_failure": [{"set":{"field":"error.message","value":"geoip failed"}}]`

### Requirement: tag field (REQ-012)

The data source SHALL accept an optional `tag` string attribute. When configured, it SHALL be included in the serialized JSON under the `"tag"` key. When not configured, the key SHALL be omitted.

#### Scenario: tag configured

- GIVEN `tag = "geoip-lookup"`
- WHEN the data source is read
- THEN `json` SHALL include `"tag": "geoip-lookup"`
