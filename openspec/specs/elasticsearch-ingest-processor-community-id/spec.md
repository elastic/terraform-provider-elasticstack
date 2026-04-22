# `elasticstack_elasticsearch_ingest_processor_community_id` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest/processor_community_id_data_source.go`

## Purpose

Provide a data-only source that serializes an Elasticsearch community ID ingest processor configuration to JSON. The community ID processor computes the Community ID flow hash algorithm for network flow tuples such as source/destination IP and port. No API calls are made; all computation is local.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_community_id" "example" {
  source_ip        = <optional, string>  # Field containing the source IP address
  source_port      = <optional, int>     # Field containing the source port
  destination_ip   = <optional, string>  # Field containing the destination IP address
  destination_port = <optional, int>     # Field containing the destination port
  iana_number      = <optional, int>     # Field containing the IANA protocol number
  icmp_type        = <optional, int>     # Field containing the ICMP type
  icmp_code        = <optional, int>     # Field containing the ICMP code
  transport        = <optional, string>  # Field containing the transport protocol (used when iana_number absent)
  target_field     = <optional, string>  # Output field for the community ID hash
  seed             = <optional, int>     # Default: 0. Seed for the hash (0–65535)
  ignore_missing   = <optional, bool>    # Default: false. Exit quietly if field absent or null

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

The data source SHALL serialize the processor configuration as `{"community_id": {...}}` JSON and store the result in the `json` computed attribute.

#### Scenario: JSON output wraps processor type

- GIVEN valid configuration is applied
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a single top-level key `"community_id"` whose value includes all configured fields

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

### Requirement: seed default and validation (REQ-004)

`seed` SHALL default to `0` and SHALL always be included in the serialized JSON. `seed` SHALL be validated to accept only values between 0 and 65535 inclusive.

#### Scenario: Invalid seed value

- GIVEN `seed` is set to a value outside the range 0–65535
- WHEN Terraform plans or applies
- THEN the provider SHALL return a validation error

#### Scenario: Default seed

- GIVEN `seed` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"seed": 0`

### Requirement: Optional network flow fields (REQ-005)

The data source SHALL include `source_ip`, `source_port`, `destination_ip`, `destination_port`, `iana_number`, `icmp_type`, `icmp_code`, `transport`, and `target_field` in the serialized JSON only when configured, and SHALL omit each when not configured.

#### Scenario: Optional fields omitted when not set

- GIVEN none of the optional network flow fields are configured
- WHEN the data source is read
- THEN the serialized JSON SHALL NOT include those fields

### Requirement: ignore_missing default (REQ-006)

`ignore_missing` SHALL default to `false` and SHALL always be included in the serialized JSON.

#### Scenario: Default ignore_missing

- GIVEN `ignore_missing` is not explicitly configured
- WHEN the data source is read
- THEN the serialized JSON SHALL include `"ignore_missing": false`

### Requirement: Common processor fields (REQ-007)

The data source SHALL include `description`, `if`, and `tag` in the serialized JSON when configured, and SHALL omit them when not configured. `ignore_failure` SHALL always be included (default `false`). `on_failure` SHALL be included when configured as a list of at least one JSON processor object.

#### Scenario: on_failure items are JSON processor objects

- GIVEN `on_failure` contains valid JSON strings
- WHEN the data source is read
- THEN each `on_failure` entry SHALL be parsed from JSON and included as an object in the serialized processor JSON

#### Scenario: on_failure with invalid JSON

- GIVEN `on_failure` contains an entry that is not valid JSON
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic
