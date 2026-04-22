# `elasticstack_elasticsearch_ingest_processor_csv` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_csv_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch CSV ingest processor configuration to JSON. The CSV processor extracts fields from a CSV line in a document. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_csv" "example" {
  field         = <required, string>       # The field to extract CSV data from
  target_fields = <required, list(string)> # Fields to assign extracted values to (min 1 item)
  separator     = <optional, string>       # Default: ",". Single character CSV separator
  quote         = <optional, string>       # Default: "\"". Single character quote character
  trim          = <optional, bool>         # Default: false. Trim whitespace from unquoted fields
  empty_value   = <optional, string>       # Value for empty fields; empty fields are skipped if not set
  ignore_missing = <optional, bool>        # Default: false. Exit quietly if field absent or null

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

The data source SHALL serialize the processor configuration as `{"csv": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"csv"` whose value includes all configured fields

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

The data source SHALL require `field` (string) and `target_fields` (list of strings, minimum 1 item). If either is absent the provider SHALL return a validation error.

#### Scenario: Missing required field

- GIVEN `field` or `target_fields` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: separator and quote defaults (REQ-005)

`separator` SHALL default to `","` and `quote` SHALL default to `"\""`. Both SHALL always be included in the serialized JSON.

#### Scenario: Default separator and quote

- GIVEN `separator` and `quote` are not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"separator": ","` and `"quote": "\""`

### Requirement: trim default (REQ-006)

`trim` SHALL default to `false` and SHALL always be included in the serialized JSON.

#### Scenario: Default trim

- GIVEN `trim` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"trim": false`

### Requirement: Optional empty_value (REQ-007)

When `empty_value` is configured, the data source SHALL include it in the serialized JSON. When `empty_value` is not configured, the data source SHALL omit it.

#### Scenario: empty_value absent

- GIVEN `empty_value` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include an `"empty_value"` key

### Requirement: ignore_missing default (REQ-008)

`ignore_missing` SHALL default to `false` and SHALL always be included in the serialized JSON.

#### Scenario: Default ignore_missing

- GIVEN `ignore_missing` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_missing": false`

### Requirement: Common processor fields (REQ-009)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
