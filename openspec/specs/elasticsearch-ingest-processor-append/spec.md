# `elasticstack_elasticsearch_ingest_processor_append` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_append_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch append ingest processor configuration to JSON. The append processor appends one or more values to an existing array field, or creates the field as an array if it does not exist. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_append" "example" {
  field            = <required, string>       # The field to be appended to
  value            = <required, list(string)> # One or more values to append (min 1 item)
  allow_duplicates = <optional, bool>         # Default: true. If false, skip values already present
  media_type       = <optional, string>       # Media type for encoding value (e.g. application/json)

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

The data source SHALL serialize the processor configuration as `{"append": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"append"` whose value includes all configured fields

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

The data source SHALL require `field` (string) and `value` (list of strings, minimum 1 item). If either is absent the provider SHALL return a validation error.

#### Scenario: Missing required field

- GIVEN `field` or `value` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: allow_duplicates default (REQ-005)

When `allow_duplicates` is not configured, the data source SHALL default to `true` and include `"allow_duplicates": true` in the serialized JSON.

#### Scenario: Default allow_duplicates

- GIVEN `allow_duplicates` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"allow_duplicates": true`

### Requirement: Optional media_type field (REQ-006)

When `media_type` is configured, the data source SHALL include it in the serialized JSON. When `media_type` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: media_type present

- GIVEN `media_type` is set to a non-empty string
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"media_type"` key with the configured value

#### Scenario: media_type absent

- GIVEN `media_type` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include a `"media_type"` key

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
