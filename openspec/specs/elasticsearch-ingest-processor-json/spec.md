# `elasticstack_elasticsearch_ingest_processor_json` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_json_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_json` data source. This data-only source accepts Elasticsearch json ingest processor configuration and produces a serialized JSON representation suitable for inclusion in an ingest pipeline definition. No Elasticsearch API calls are made.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_json" "example" {
  # Processor-specific (required)
  field = <required, string>  # field to be parsed as JSON

  # Processor-specific (optional)
  target_field                  = <optional, string>  # field to write parsed result to
  add_to_root                   = <optional, bool>    # force parsed JSON to root of document
  add_to_root_conflict_strategy = <optional, string>  # "replace" or "merge"
  allow_duplicate_keys          = <optional, bool>    # default false

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

The data source SHALL serialize the processor configuration into a JSON object wrapped under the key `"json"`, using `json.MarshalIndent` with a single-space indent. The resulting string SHALL be stored in the `json` computed attribute.

#### Scenario: JSON wrapping

- GIVEN a configuration with `field = "string_source"`
- WHEN the data source is read
- THEN the `json` attribute SHALL contain a JSON object of the form `{"json": {...}}`

### Requirement: Hash identity (REQ-003)

The data source SHALL compute `id` as a deterministic hash of the `json` output string. The same configuration inputs SHALL always produce the same `id`.

#### Scenario: Deterministic id

- GIVEN identical configuration inputs on two separate reads
- WHEN the data source is read both times
- THEN `id` SHALL be equal for both reads

### Requirement: Required processor-specific attributes (REQ-004)

The data source SHALL require `field` (string) to be set. Omitting it SHALL cause a validation error before the read function runs.

#### Scenario: Missing required field

- GIVEN a configuration that omits `field`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Optional processor-specific attributes (REQ-005)

The data source SHALL include the following optional processor-specific attributes in the JSON output when set, and omit them when not set:
- `target_field` (string): emitted as `target_field` only when set.
- `add_to_root` (bool): emitted as `add_to_root` only when set; when `true` the parsed JSON is merged at the root of the document and `target_field` must not be set.
- `add_to_root_conflict_strategy` (string, must be `"replace"` or `"merge"`): emitted as `add_to_root_conflict_strategy` only when set.
- `allow_duplicate_keys` (bool, default `false`): emitted as `allow_duplicate_keys` only when set.

#### Scenario: add_to_root_conflict_strategy validation

- GIVEN `add_to_root_conflict_strategy` is set to a value other than `"replace"` or `"merge"`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

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
