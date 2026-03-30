# `elasticstack_elasticsearch_ingest_processor_uri_parts` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_uri_parts_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch URI parts ingest processor configuration to JSON. The URI parts processor parses a Uniform Resource Identifier (URI) string and extracts its components as an object. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_uri_parts" "example" {
  field                = <required, string>  # Field containing the URI string
  target_field         = <optional, string>  # Output field for the URI object
  keep_original        = <optional, bool>    # Copy unparsed URI to <target_field>.original. Default: true
  remove_if_successful = <optional, bool>    # Remove field after successful parsing. Default: false

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

The data source SHALL serialize the processor configuration as `{"uri_parts": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"uri_parts"` whose value includes all configured fields

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

### Requirement: Required fields (REQ-004)

The data source SHALL require `field` (string). If absent the provider SHALL return a validation error.

#### Scenario: Missing required field

- GIVEN `field` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: keep_original and remove_if_successful defaults (REQ-005)

When `keep_original` is not configured, the data source SHALL default to `true` and include `"keep_original": true` in the serialized JSON. When `remove_if_successful` is not configured, the data source SHALL default to `false` and include `"remove_if_successful": false` in the serialized JSON.

#### Scenario: Default keep_original

- GIVEN `keep_original` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"keep_original": true`

#### Scenario: Default remove_if_successful

- GIVEN `remove_if_successful` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"remove_if_successful": false`

### Requirement: Optional target_field (REQ-006)

When `target_field` is configured, the data source SHALL include it in the serialized JSON. When `target_field` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: target_field present

- GIVEN `target_field` is set to a non-empty string
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"target_field"` key with the configured value

#### Scenario: target_field absent

- GIVEN `target_field` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include a `"target_field"` key

### Requirement: Common processor fields (REQ-007)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object, and SHALL omit non-configured entries.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
