# `elasticstack_fleet_output` — Schema and Functional Requirements

Resource implementation: `internal/fleet/output`

Data source implementation: `internal/fleet/outputds`

## Purpose

Define schema and behavior for the Fleet output resource and data source. The resource manages Fleet outputs (Elasticsearch, Logstash, and Kafka) via the Fleet Outputs API, including Kibana space awareness. The data source lists all outputs available in a given Kibana space.

## Schema

### Resource

```hcl
resource "elasticstack_fleet_output" "example" {
  id        = <computed, string>         # internal identifier, mirrors output_id
  output_id = <optional+computed, string> # force new; Fleet-assigned or user-supplied output ID

  name  = <required, string>
  type  = <required, string>             # one of: "elasticsearch", "logstash", "kafka"
  hosts = <required, list(string)>       # at least one entry

  ca_sha256             = <optional, string>
  ca_trusted_fingerprint = <optional, string>
  default_integrations  = <optional+computed, bool> # default false
  default_monitoring    = <optional+computed, bool> # default false
  config_yaml           = <optional, string, sensitive>
  space_ids             = <optional+computed, set(string)>

  ssl {
    certificate_authorities = <optional, list(string)>  # at least one entry when provided
    certificate             = <required, string>        # non-empty
    key                     = <required, string, sensitive> # non-empty
  }

  kafka {
    auth_type         = <optional, string>   # one of: "none", "user_pass", "ssl", "kerberos"
    broker_timeout    = <optional+computed, float32>
    client_id         = <optional+computed, string>
    compression       = <optional, string>   # one of: "gzip", "snappy", "lz4", "none"
    compression_level = <optional+computed, int64> # only valid when compression = "gzip"; one of: -1, 0, 1
    connection_type   = <optional, string>   # one of: "plaintext", "encryption"; only when auth_type = "none"
    topic             = <optional, string>
    partition         = <optional, string>   # one of: "random", "round_robin", "hash"
    required_acks     = <optional, int64>    # one of: -1, 0, 1
    timeout           = <optional+computed, float32>
    version           = <optional+computed, string>
    username          = <optional, string>
    password          = <optional, string, sensitive>
    key               = <optional, string>
    headers = <optional+computed, list(object({
      key   = string
      value = string
    }))>
    hash {
      hash   = <optional, string>
      random = <optional, bool>
    }
    random {
      group_events = <optional, float64>
    }
    round_robin {
      group_events = <optional, float64>
    }
    sasl {
      mechanism = <optional, string> # one of: "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
    }
  }
}
```

### Data Source

```hcl
data "elasticstack_fleet_output" "example" {
  id       = <computed, string>  # fixed value "outputs"
  space_id = <optional, string>  # Kibana space to query; default space when omitted

  outputs = <computed, list(object({
    id                    = string
    name                  = string
    type                  = string
    hosts                 = list(string)
    ca_sha256             = string
    ca_trusted_fingerprint = string
    default_integrations  = bool
    default_monitoring    = bool
    config_yaml           = string (sensitive)
  }))>
}
```

## Requirements

### Requirement: Fleet Output CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Fleet Create output API to create outputs. The resource SHALL use the Fleet Get output API to read outputs. The resource SHALL use the Fleet Update output API to update outputs. The resource SHALL use the Fleet Delete output API to delete outputs. When the Fleet API returns a non-success status for any create, update, or delete operation, the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Fleet API returns an error on create
- WHEN the resource create runs
- THEN diagnostics SHALL contain the API error

### Requirement: Data source list API (REQ-005)

The data source SHALL use the Fleet Get outputs API to list all outputs in the specified Kibana space. When `space_id` is not set, the data source SHALL query the default space. When the Fleet API returns an error, the data source SHALL surface it to Terraform diagnostics.

#### Scenario: List outputs with space

- GIVEN `space_id = "my-space"`
- WHEN the data source read runs
- THEN the Fleet API SHALL be called with the `my-space` space context

### Requirement: Identity (REQ-006–REQ-007)

The resource SHALL expose a computed `id` attribute that mirrors `output_id`. On create, if `output_id` is not configured, the Fleet API SHALL assign the identifier; the resource SHALL store the API-assigned value in both `id` and `output_id`. When `output_id` is configured, the resource SHALL pass it to the create API as the desired identifier.

#### Scenario: Auto-assigned output_id

- GIVEN `output_id` is not set in config
- WHEN create completes successfully
- THEN `output_id` and `id` SHALL both be set to the API-returned identifier

### Requirement: Import (REQ-008)

The resource SHALL support import via `ImportStatePassthroughID` using `output_id` as the import path. When importing, the provided ID value SHALL be stored directly as `output_id`.

#### Scenario: Import by output_id

- GIVEN an existing Fleet output with id `my-output`
- WHEN `terraform import ... my-output` runs
- THEN `output_id` SHALL be `my-output` and a read cycle SHALL refresh all other attributes

### Requirement: Lifecycle — output_id forces replacement (REQ-009)

Changing `output_id` SHALL require resource replacement (`RequiresReplace`). The `output_id` attribute SHALL also use `UseStateForUnknown` so that an unknown `output_id` during plan is preserved from prior state.

#### Scenario: output_id change triggers replacement

- GIVEN `output_id` changes between plan versions
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Compatibility — Kafka requires server >= 8.13.0 (REQ-010)

When `type = "kafka"` is configured, the resource SHALL verify that the Kibana/Fleet server version is at least 8.13.0 before calling the create or update API. If the server version is lower, the resource SHALL fail with an "Unsupported version for Kafka output" error diagnostic and SHALL NOT call the Fleet API.

#### Scenario: Kafka on old server

- GIVEN `type = "kafka"` and server version < 8.13.0
- WHEN create or update runs
- THEN the resource SHALL error with "Unsupported version for Kafka output" and SHALL NOT call the Fleet API

### Requirement: Kafka compression_level constraint (REQ-011)

The `kafka.compression_level` attribute SHALL only be accepted when `kafka.compression` equals `"gzip"`. The schema validator SHALL enforce this constraint. When building the API request, `compression_level` SHALL only be sent when `compression` is `"gzip"`.

#### Scenario: compression_level with non-gzip

- GIVEN `kafka.compression = "snappy"` and `kafka.compression_level` is set
- WHEN the resource is configured
- THEN schema validation SHALL return an error

### Requirement: Kafka connection_type constraint (REQ-012)

The `kafka.connection_type` attribute SHALL only be accepted when `kafka.auth_type` equals `"none"`. The schema validator SHALL enforce this constraint.

#### Scenario: connection_type with non-none auth_type

- GIVEN `kafka.auth_type = "user_pass"` and `kafka.connection_type` is set
- WHEN the resource is configured
- THEN schema validation SHALL return an error

### Requirement: Space-aware create (REQ-013)

On create, when `space_ids` is configured with at least one space ID, the resource SHALL pass the first space ID from `space_ids` to the Fleet create API as the space context. When `space_ids` is null or unknown, the resource SHALL call the create API without a space prefix (default space).

#### Scenario: Create in named space

- GIVEN `space_ids = ["my-space"]`
- WHEN create runs
- THEN the Fleet create API SHALL be called with `my-space` as the space context

### Requirement: Space-aware read and update using state (REQ-014)

On read and update, the resource SHALL derive the operational space from the `space_ids` stored in state (not the plan). If `space_ids` in state is null or empty, the resource SHALL query using the default space. Otherwise, the resource SHALL use the first space ID from state as the space context. On update, the resource SHALL send `space_ids` from the plan in the request body so the Fleet API can adjust space membership.

#### Scenario: Update space_ids from state

- GIVEN state has `space_ids = ["space-a"]` and plan has `space_ids = ["space-b", "space-a"]`
- WHEN update runs
- THEN the update API SHALL be called using `space-a` (from state) as the space context

### Requirement: Space-aware delete (REQ-015)

On delete, the resource SHALL use the first space ID from state as the space context (same logic as read). Deleting removes the output from all spaces; to remove from specific spaces only, `space_ids` SHALL be updated instead.

#### Scenario: Delete removes globally

- GIVEN a resource with `space_ids = ["space-a"]`
- WHEN destroy runs
- THEN the Fleet delete API SHALL be called using `space-a` as the space context

### Requirement: Read — not found removes from state (REQ-016)

On read, if the Fleet API returns an error or a nil response for the output, the resource SHALL remove itself from state. This causes Terraform to plan re-creation.

#### Scenario: Output deleted outside Terraform

- GIVEN the output was manually deleted from Fleet
- WHEN read (refresh) runs
- THEN the resource SHALL be removed from state

### Requirement: State mapping — output type dispatch (REQ-017)

On read, the resource SHALL dispatch state population based on the output type discriminator. For `OutputElasticsearch`, `OutputLogstash`, and `OutputKafka` responses, the resource SHALL map all common fields (`id`, `output_id`, `name`, `type`, `hosts`, `ca_sha256`, `ca_trusted_fingerprint`, `default_integrations`, `default_monitoring`, `config_yaml`, `ssl`). For `OutputKafka`, the resource SHALL additionally map all Kafka-specific fields. If an unrecognized output type is returned, the resource SHALL surface an error diagnostic.

#### Scenario: Unknown output type

- GIVEN the API returns a type not in the known set
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

### Requirement: State mapping — SpaceIDs not returned by API (REQ-018)

The Fleet API does not return `space_ids` in output responses. On read, the resource SHALL preserve the `space_ids` value already in state. If `space_ids` is null or unknown in state, the resource SHALL set it to explicit null.

#### Scenario: space_ids preserved after read

- GIVEN `space_ids = ["my-space"]` in state
- WHEN read runs
- THEN `space_ids` SHALL remain `["my-space"]` in state after the refresh

### Requirement: State mapping — SSL null when no SSL in API response (REQ-019)

When the API response contains no SSL block (nil), the resource SHALL set `ssl` to null in state. When the API response contains an SSL block but all fields resolve to empty/nil, the resource SHALL also set `ssl` to null.

#### Scenario: No SSL in API response

- GIVEN the output has no SSL configured in Fleet
- WHEN read runs
- THEN `ssl` SHALL be null in state

### Requirement: StateUpgrade — v0 ssl list to object (REQ-020–REQ-022)

The resource schema is at version 1. Version 0 stored `ssl` as a list block. During v0 → v1 upgrade, if the raw state JSON is nil, the upgrade SHALL fail with an "Invalid raw state" error. If the raw state JSON cannot be unmarshalled, the upgrade SHALL fail with a "Failed to unmarshal raw state" error. If `ssl` in the raw state is a non-empty list, the upgrade SHALL replace it with the first element (converting the list to a single object). If `ssl` is an empty list, the upgrade SHALL remove the `ssl` key entirely. If `ssl` is not present or is already an object, the upgrade SHALL pass the state through unchanged.

#### Scenario: v0 ssl list with one element

- GIVEN v0 state with `"ssl": [{"certificate": "...", "key": "..."}]`
- WHEN the state upgrade runs
- THEN `ssl` SHALL be `{"certificate": "...", "key": "..."}`

#### Scenario: v0 empty ssl list

- GIVEN v0 state with `"ssl": []`
- WHEN the state upgrade runs
- THEN `ssl` SHALL be absent from the upgraded state

#### Scenario: Nil raw state

- GIVEN a nil raw state
- WHEN the state upgrade runs
- THEN the upgrade SHALL fail with "Invalid raw state"

### Requirement: Data source — computed id (REQ-023)

The data source SHALL set a computed `id` attribute to the fixed string `"outputs"` on every read.

#### Scenario: Data source id

- GIVEN any read of the data source
- WHEN read completes
- THEN `id` SHALL equal `"outputs"`

### Requirement: Data source — outputs list (REQ-024)

The data source SHALL populate the `outputs` list attribute with all outputs returned by the Fleet API, supporting the `OutputElasticsearch`, `OutputLogstash`, `OutputKafka`, and `OutputRemoteElasticsearch` types. Each output item SHALL expose `id`, `name`, `type`, `hosts`, `ca_sha256`, `ca_trusted_fingerprint`, `default_integrations`, `default_monitoring`, and `config_yaml`.

#### Scenario: Multiple output types in list

- GIVEN Fleet returns outputs of types elasticsearch, logstash, and kafka
- WHEN the data source read runs
- THEN `outputs` SHALL contain one entry per output, each with the correct type and common fields
