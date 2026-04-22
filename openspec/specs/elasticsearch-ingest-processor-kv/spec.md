# `elasticstack_elasticsearch_ingest_processor_kv` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_kv_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_kv` data source. This data-only source accepts Elasticsearch kv (key-value) ingest processor configuration and produces a serialized JSON representation suitable for inclusion in an ingest pipeline definition. No Elasticsearch API calls are made.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_kv" "example" {
  # Processor-specific (required)
  field       = <required, string>  # field to be parsed; supports template snippets
  field_split = <required, string>  # regex pattern to split key-value pairs
  value_split = <required, string>  # regex pattern to split key from value within a pair

  # Processor-specific (optional)
  target_field   = <optional, string>      # field to insert extracted keys into; defaults to root
  include_keys   = <optional, set(string)> # keys to include; defaults to all; min 1 item
  exclude_keys   = <optional, set(string)> # keys to exclude; min 1 item
  ignore_missing = <optional, bool>        # default false
  prefix         = <optional, string>      # prefix added to extracted keys
  trim_key       = <optional, string>      # characters to trim from extracted keys
  trim_value     = <optional, string>      # characters to trim from extracted values
  strip_brackets = <optional, bool>        # default false

  # Common optional fields
  description    = <optional, string>
  if             = <optional, string>
  ignore_failure = <optional, bool>          # default false
  on_failure     = <optional, list(string)>  # each element must be valid JSON; min 1 item
  tag            = <optional, string>

  # Computed outputs
  id   = <computed, string>  # hash of the JSON output
  json = <computed, string>  # serialized processor JSON
}
```

## Requirements

### Requirement: No API calls (REQ-001)

The data source SHALL NOT make any Elasticsearch API calls. It operates entirely from the supplied configuration attributes.

#### Scenario: Read with valid configuration

- GIVEN a valid configuration for the data source
- WHEN Terraform reads (plans or applies) the data source
- THEN no Elasticsearch connection or API call SHALL be made

### Requirement: JSON output format (REQ-002)

The data source SHALL serialize the processor configuration into a JSON object wrapped under the key `"kv"`, using `json.MarshalIndent` with a single-space indent. The resulting string SHALL be stored in the `json` computed attribute.

#### Scenario: JSON wrapping

- GIVEN a configuration with `field = "message"`, `field_split = " "`, and `value_split = "="`
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object of the form `{"kv": {...}}`

### Requirement: Hash identity (REQ-003)

The data source SHALL compute `id` as a deterministic hash of the `json` output string. The same configuration inputs SHALL always produce the same `id`.

#### Scenario: Deterministic id

- GIVEN identical configuration inputs on two separate reads
- WHEN the data source is read both times
- THEN `id` SHALL be equal for both reads

### Requirement: Required processor-specific attributes (REQ-004)

The data source SHALL require `field` (string), `field_split` (string), and `value_split` (string) to be set. Omitting any of these SHALL cause a validation error before the read function runs.

#### Scenario: Missing required field

- GIVEN a configuration that omits `field`, `field_split`, or `value_split`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Optional processor-specific attributes (REQ-005)

The data source SHALL include the following optional processor-specific attributes in the JSON output when set, and omit them when not set:
- `target_field` (string): emitted as `target_field` only when set; defaults to root of document.
- `include_keys` (set of string, min 1): emitted as `include_keys` only when set; defaults to all keys.
- `exclude_keys` (set of string, min 1): emitted as `exclude_keys` only when set.
- `ignore_missing` (bool, default `false`): always emitted as `ignore_missing`.
- `prefix` (string): emitted as `prefix` only when set.
- `trim_key` (string): emitted as `trim_key` only when set.
- `trim_value` (string): emitted as `trim_value` only when set.
- `strip_brackets` (bool, default `false`): always emitted as `strip_brackets`.

#### Scenario: Optional fields absent

- GIVEN none of `target_field`, `include_keys`, `exclude_keys`, `prefix`, `trim_key`, or `trim_value` are configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include those keys

#### Scenario: include_keys filters output

- GIVEN `include_keys` is set to a non-empty list
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"include_keys"` with the configured values

### Requirement: Common processor fields (REQ-006)

The data source SHALL support the following common optional fields and include them in the JSON output when set, omitting them when not set:
- `description` (string): emitted as `description`.
- `if` (string): emitted as `if`.
- `ignore_failure` (bool, default `false`): always emitted as `ignore_failure`.
- `on_failure` (list of JSON strings, min 1): each element is parsed as JSON and emitted as an object in the `on_failure` array.
- `tag` (string): emitted as `tag`.

#### Scenario: on_failure JSON parsing error

- GIVEN an `on_failure` entry that is not valid JSON
- WHEN the data source is read
- THEN a parse error SHALL be returned in diagnostics

### Requirement: on_failure element validation (REQ-007)

Each element of `on_failure` SHALL be validated as a JSON string at plan/apply time. Invalid JSON in any element SHALL cause an error.

#### Scenario: Invalid on_failure JSON

- GIVEN `on_failure = ["not-valid-json"]`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned
