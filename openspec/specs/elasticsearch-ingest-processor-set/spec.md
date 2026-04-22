# `elasticstack_elasticsearch_ingest_processor_set` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_set_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch set ingest processor configuration to JSON. The set processor inserts, upserts, or updates a field with a specified value or by copying from another field. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_set" "example" {
  field              = <required, string>                           # The field to insert, upsert, or update
  value              = <optional, string, conflicts with: copy_from> # Value to set; exactly one of value or copy_from required
  copy_from          = <optional, string, conflicts with: value>    # Origin field to copy from; exactly one of copy_from or value required
  override           = <optional, bool>                            # Default: true. If processor updates pre-existing non-null fields
  ignore_empty_value = <optional, bool>                            # Default: false. If true and value evaluates to null/empty, exit quietly
  media_type         = <optional, string>                          # Default: "application/json". Media type for encoding value

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

The data source SHALL serialize the processor configuration as `{"set": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"set"` whose value includes all configured fields

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

### Requirement: Required field and exactly one of value or copy_from (REQ-004)

The data source SHALL require `field` (string). The data source SHALL also require exactly one of `value` or `copy_from` to be configured. If both are configured or neither is configured, the provider SHALL return a validation error. `value` and `copy_from` are mutually exclusive (ConflictsWith).

#### Scenario: Missing required field

- GIVEN `field` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

#### Scenario: Both value and copy_from configured

- GIVEN both `value` and `copy_from` are set
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

#### Scenario: Neither value nor copy_from configured

- GIVEN neither `value` nor `copy_from` is configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: Defaults for override, ignore_empty_value, media_type, ignore_failure (REQ-005)

When `override` is not configured, the data source SHALL default to `true` and include `"override": true` in the serialized JSON. When `ignore_empty_value` is not configured, the data source SHALL default to `false` and include `"ignore_empty_value": false` in the serialized JSON. When `ignore_failure` is not configured, the data source SHALL default to `false` and include `"ignore_failure": false` in the serialized JSON.

#### Scenario: Default override

- GIVEN `override` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"override": true`

#### Scenario: Default ignore_empty_value

- GIVEN `ignore_empty_value` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_empty_value": false`

### Requirement: media_type field default (REQ-006)

The `media_type` attribute defaults to `"application/json"`. Because a default is set, the data source SHALL always include `"media_type"` in the serialized JSON with the configured or default value.

#### Scenario: media_type present

- GIVEN `media_type` is set to a non-empty string
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"media_type"` key with the configured value

#### Scenario: media_type uses default

- GIVEN `media_type` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"media_type": "application/json"`

### Requirement: Optional value and copy_from fields (REQ-007)

When `value` is configured, the data source SHALL include it in the serialized JSON. When `copy_from` is configured, the data source SHALL include it in the serialized JSON. When neither is configured, both SHALL be omitted from the serialized JSON (though exactly one must be provided due to REQ-004).

#### Scenario: value present

- GIVEN `value` is set
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"value"` key with the configured value

#### Scenario: copy_from present

- GIVEN `copy_from` is set
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"copy_from"` key with the configured value

### Requirement: Common processor fields (REQ-008)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object, and SHALL omit non-configured entries.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
