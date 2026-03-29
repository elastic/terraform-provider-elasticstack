# `elasticstack_elasticsearch_ingest_processor_set_security_user` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_set_security_user_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch set security user ingest processor configuration to JSON. The set security user processor adds user-related details from the current authenticated user into a document field, such as username, roles, email, and other security properties. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_set_security_user" "example" {
  field      = <required, string>          # The field to store user information into
  properties = <optional, set(string), min 1 when set> # User-related properties to add (e.g. username, roles, email)

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

The data source SHALL serialize the processor configuration as `{"set_security_user": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"set_security_user"` whose value includes all configured fields

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

The data source SHALL require `field` (string). If `field` is absent the provider SHALL return a validation error.

#### Scenario: Missing required field

- GIVEN `field` is not configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: Optional properties field (REQ-005)

When `properties` is configured, the data source SHALL collect all values from the set and serialize them as an array in the JSON output under the `"properties"` key. When `properties` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: properties present

- GIVEN `properties` is set with one or more values
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"properties"` as an array containing all configured property names

#### Scenario: properties absent

- GIVEN `properties` is not configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include a `"properties"` key

### Requirement: ignore_failure default (REQ-006)

When `ignore_failure` is not configured, the data source SHALL default to `false` and include `"ignore_failure": false` in the serialized JSON.

#### Scenario: Default ignore_failure

- GIVEN `ignore_failure` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_failure": false`

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
