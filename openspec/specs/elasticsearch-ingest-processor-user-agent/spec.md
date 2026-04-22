# `elasticstack_elasticsearch_ingest_processor_user_agent` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_user_agent_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch user agent ingest processor configuration to JSON. The user agent processor extracts browser, operating system, device, and version information from a user agent string field. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_user_agent" "example" {
  field               = <required, string>      # Field containing the user agent string
  target_field        = <optional, string>      # Output field for user agent details
  regex_file          = <optional, string>      # Regex file in config/ingest-user-agent directory
  properties          = <optional, set(string)> # Properties to add to target_field (min 1 item when set)
  extract_device_type = <optional, bool>        # Extract device type on best-effort basis. Default: false
  ignore_missing      = <optional, bool>        # Quietly exit if field is null or missing. Default: false

  # Computed outputs
  id   = <computed, string>
  json = <computed, string>
}
```

## Requirements

### Requirement: JSON serialization (REQ-001)

The data source SHALL serialize the processor configuration as `{"user_agent": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"user_agent"` whose value includes all configured fields

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

### Requirement: No common processor fields (REQ-005)

The `elasticstack_elasticsearch_ingest_processor_user_agent` data source SHALL NOT expose `description`, `if`, `ignore_failure`, `on_failure`, or `tag` attributes. These common processor fields are not part of this data source's schema.

#### Scenario: Schema does not include common fields

- GIVEN the data source schema
- WHEN inspected
- THEN it SHALL NOT include `description`, `if`, `ignore_failure`, `on_failure`, or `tag` attributes

### Requirement: ignore_missing default (REQ-006)

When `ignore_missing` is not configured, the data source SHALL default to `false` and include `"ignore_missing": false` in the serialized JSON.

#### Scenario: Default ignore_missing

- GIVEN `ignore_missing` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_missing": false`

### Requirement: Optional target_field and regex_file (REQ-007)

When `target_field` is configured, the data source SHALL include it in the serialized JSON. When `target_field` is not configured, the data source SHALL omit it from the serialized JSON. When `regex_file` is configured, the data source SHALL include it in the serialized JSON. When `regex_file` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: target_field and regex_file absent

- GIVEN neither `target_field` nor `regex_file` is configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include `"target_field"` or `"regex_file"` keys

### Requirement: Optional properties set (REQ-008)

When `properties` is configured with one or more strings, the data source SHALL include it as a list in the serialized JSON. When `properties` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: properties present

- GIVEN `properties` is set to one or more string values
- WHEN the data source is read
- THEN the serialized JSON SHALL include a `"properties"` array containing those values

#### Scenario: properties absent

- GIVEN `properties` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include a `"properties"` key

### Requirement: Optional extract_device_type (REQ-009)

When `extract_device_type` is configured to `true`, the data source SHALL include `"extract_device_type": true` in the serialized JSON. When `extract_device_type` is not configured or is `false`, the data source SHALL omit it from the serialized JSON.

#### Scenario: extract_device_type set to true

- GIVEN `extract_device_type` is set to `true`
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"extract_device_type": true`

#### Scenario: extract_device_type not set

- GIVEN `extract_device_type` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include an `"extract_device_type"` key
