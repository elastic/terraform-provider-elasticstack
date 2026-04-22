# `elasticstack_elasticsearch_ingest_processor_date` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_date_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch date ingest processor configuration to JSON. The date processor parses dates from document fields and uses them to set a target field, optionally normalizing to a configured output format. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_date" "example" {
  field         = <required, string>       # The field to get the date from
  formats       = <required, list(string)> # Expected date formats (min 1 item)
  target_field  = <optional, string>       # Default: "@timestamp". Field to store the parsed date
  timezone      = <optional, string>       # Default: "UTC". Timezone for parsing
  locale        = <optional, string>       # Default: "ENGLISH". Locale for parsing month/day names
  output_format = <optional, string>       # Default: "yyyy-MM-dd'T'HH:mm:ss.SSSXXX". Format for writing the date

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

The data source SHALL serialize the processor configuration as `{"date": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"date"` whose value includes all configured fields

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

The data source SHALL require `field` (string) and `formats` (list of strings, minimum 1 item). If either is absent the provider SHALL return a validation error.

#### Scenario: Missing required field

- GIVEN `field` or `formats` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: target_field default (REQ-005)

`target_field` SHALL default to `"@timestamp"`. When `target_field` is set to a non-empty value, the data source SHALL include it in the serialized JSON.

#### Scenario: Default target_field

- GIVEN `target_field` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"target_field": "@timestamp"`

### Requirement: timezone and locale defaults (REQ-006)

`timezone` SHALL default to `"UTC"` and `locale` SHALL default to `"ENGLISH"`. Both SHALL be included in the serialized JSON when set to their default or explicitly configured values. When omitted from config, the model field uses `omitempty`, so they are only serialized when non-empty; defaults set in the schema ensure they are always included.

#### Scenario: Default timezone and locale

- GIVEN `timezone` and `locale` are not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"timezone": "UTC"` and `"locale": "ENGLISH"`

### Requirement: output_format default (REQ-007)

`output_format` SHALL default to `"yyyy-MM-dd'T'HH:mm:ss.SSSXXX"` and SHALL be included in the serialized JSON when set to its default or an explicitly configured value.

#### Scenario: Default output_format

- GIVEN `output_format` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include the default output format string

### Requirement: Common processor fields (REQ-008)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
