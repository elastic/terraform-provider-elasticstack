# `elasticstack_elasticsearch_ingest_processor_script` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_script_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch script ingest processor configuration to JSON. The script processor executes an inline or stored Painless script to transform documents. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_script" "example" {
  lang      = <optional, string>                           # Script language
  script_id = <optional, string, conflicts with: source>  # ID of a stored script; exactly one of script_id or source required
  source    = <optional, string, conflicts with: script_id> # Inline script; exactly one of script_id or source required
  params    = <optional, string>                           # JSON object string of parameters for the script

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

The data source SHALL serialize the processor configuration as `{"script": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"script"` whose value includes all configured fields

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

### Requirement: Exactly one of script_id or source (REQ-004)

The data source SHALL require exactly one of `script_id` or `source` to be configured. If both are configured or neither is configured, the provider SHALL return a validation error. `script_id` and `source` are mutually exclusive (ConflictsWith).

#### Scenario: Both script_id and source configured

- GIVEN both `script_id` and `source` are set
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

#### Scenario: Neither script_id nor source configured

- GIVEN neither `script_id` nor `source` is configured
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

### Requirement: params JSON parsing (REQ-005)

When `params` is configured, the data source SHALL parse it as a JSON object and include the parsed object in the serialized JSON. When `params` is not configured, the data source SHALL omit it from the serialized JSON. If `params` contains invalid JSON, the data source SHALL return an error diagnostic.

#### Scenario: Valid params JSON

- GIVEN `params` is set to a valid JSON object string
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"params"` as a parsed JSON object

#### Scenario: Invalid params JSON

- GIVEN `params` is set to an invalid JSON string
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic

### Requirement: Optional lang field (REQ-006)

When `lang` is configured, the data source SHALL include it in the serialized JSON. When `lang` is not configured, the data source SHALL omit it from the serialized JSON.

#### Scenario: lang present

- GIVEN `lang` is set to a non-empty string
- WHEN the data source is read
- THEN the serialized JSON SHALL include the `"lang"` key with the configured value

### Requirement: ignore_failure default (REQ-007)

When `ignore_failure` is not configured, the data source SHALL default to `false` and include `"ignore_failure": false` in the serialized JSON.

#### Scenario: Default ignore_failure

- GIVEN `ignore_failure` is not set
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_failure": false`

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
