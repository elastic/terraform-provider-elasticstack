# `elasticstack_elasticsearch_ingest_processor_reroute` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_reroute_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch reroute ingest processor configuration to JSON. The reroute processor routes a document to a different target index, data stream, or index alias. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_reroute" "example" {
  destination = <optional, string> # The destination data stream, index, or index alias to route to
  dataset     = <optional, string> # The destination dataset to route to
  namespace   = <optional, string> # The destination namespace to route to

  # Common processor fields
  description    = <optional, string>
  if             = <optional, string>
  ignore_failure = <optional, bool>         # Default: false
  on_failure     = <optional, list(string)> # List of JSON-encoded processor objects (min 1 item when set)
  tag            = <optional, string>

  # Computed outputs
  id   = <computed, string>
  json = <computed, string>
}
```

## Requirements

### Requirement: JSON serialization (REQ-001)

The data source SHALL serialize the processor configuration as `{"reroute": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"reroute"` whose value includes all configured fields

### Requirement: Identity (REQ-002)

The data source SHALL compute `id` as a hash of the serialized `json` output.

#### Scenario: Deterministic id from JSON

- GIVEN a fixed configuration
- WHEN the data source is read
- THEN `id` SHALL equal the hash of the `json` attribute value, and the same configuration SHALL always produce the same `id`

### Requirement: No API calls (REQ-003)

The data source SHALL NOT make any Elasticsearch API calls. All computation SHALL be performed locally.

#### Scenario: Local-only computation

- GIVEN no Elasticsearch connection is configured
- WHEN the data source is read
- THEN it SHALL succeed and produce valid `json` and `id` outputs

### Requirement: All routing fields optional (REQ-004)

The data source SHALL have no required processor-specific fields. `destination`, `dataset`, and `namespace` are all optional. When any of these fields is configured, the data source SHALL include it in the serialized JSON. When not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: destination present

- GIVEN `destination` is set to a non-empty string
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"destination"` key with the configured value

#### Scenario: destination absent

- GIVEN `destination` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include a `"destination"` key

### Requirement: ignore_failure default (REQ-005)

When `ignore_failure` is not configured, the data source SHALL default to `false` and include `"ignore_failure": false` in the serialized JSON.

#### Scenario: Default ignore_failure

- GIVEN `ignore_failure` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_failure": false`

### Requirement: Common processor fields (REQ-006)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object, and SHALL omit non-configured entries.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
