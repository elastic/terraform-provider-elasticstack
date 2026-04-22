# `elasticstack_elasticsearch_ingest_processor_network_direction` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_network_direction_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_network_direction` data source. This data-only source accepts Elasticsearch network_direction ingest processor configuration and produces a serialized JSON representation suitable for inclusion in an ingest pipeline definition. No Elasticsearch API calls are made.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_network_direction" "example" {
  # Processor-specific (optional, but exactly one of internal_networks or internal_networks_field required)
  source_ip      = <optional, string>  # field containing source IP address
  destination_ip = <optional, string>  # field containing destination IP address
  target_field   = <optional, string>  # output field for the network direction

  # Exactly one of these must be set (mutually exclusive)
  internal_networks       = <optional, set(string)>  # list of internal networks; min 1 item
  internal_networks_field = <optional, string>        # document field to read internal_networks from

  ignore_missing = <optional, bool>    # default true

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

The data source SHALL serialize the processor configuration into a JSON object wrapped under the key `"network_direction"`, using `json.MarshalIndent` with a single-space indent. The resulting string SHALL be stored in the `json` computed attribute.

#### Scenario: JSON wrapping

- GIVEN a configuration with `internal_networks = ["private"]`
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object of the form `{"network_direction": {...}}`

### Requirement: Hash identity (REQ-003)

The data source SHALL compute `id` as a deterministic hash of the `json` output string. The same configuration inputs SHALL always produce the same `id`.

#### Scenario: Deterministic id

- GIVEN identical configuration inputs on two separate reads
- WHEN the data source is read both times
- THEN `id` SHALL be equal for both reads

### Requirement: Mutual exclusivity of internal_networks and internal_networks_field (REQ-004)

The data source SHALL require exactly one of `internal_networks` or `internal_networks_field` to be set. Providing both or neither SHALL cause a validation error before the read function runs. These two attributes are mutually exclusive (`ConflictsWith`) and form an `ExactlyOneOf` constraint.

#### Scenario: Both set

- GIVEN a configuration that sets both `internal_networks` and `internal_networks_field`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

#### Scenario: Neither set

- GIVEN a configuration that sets neither `internal_networks` nor `internal_networks_field`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Optional processor-specific attributes (REQ-005)

The data source SHALL include the following optional processor-specific attributes in the JSON output when set, and omit them when not set:
- `source_ip` (string): emitted as `source_ip` only when set.
- `destination_ip` (string): emitted as `destination_ip` only when set.
- `target_field` (string): emitted as `target_field` only when set.
- `internal_networks` (set of string, min 1): emitted as `internal_networks` only when set.
- `internal_networks_field` (string): emitted as `internal_networks_field` only when set.
- `ignore_missing` (bool, default `true`): always emitted as `ignore_missing`.

#### Scenario: ignore_missing defaults to true

- GIVEN a configuration that does not explicitly set `ignore_missing`
- WHEN the data source is read
- THEN `ignore_missing` SHALL be `true` in the serialized JSON output

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
