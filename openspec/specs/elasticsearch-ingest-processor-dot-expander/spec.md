# `elasticstack_elasticsearch_ingest_processor_dot_expander` — Schema and Functional Requirements

Data source implementation: `internal/elasticsearch/ingest`

## Purpose

Define the schema and runtime behavior for the `elasticstack_elasticsearch_ingest_processor_dot_expander` data source, which builds an Elasticsearch dot_expander ingest processor configuration and serializes it to JSON. This data source performs no API calls; it computes a JSON representation and a deterministic hash ID from the configured processor parameters.

## Schema

```hcl
data "elasticstack_elasticsearch_ingest_processor_dot_expander" "example" {
  field    = <required, string>   # The field to expand into an object field; use "*" to expand all top-level fields
  path     = <optional, string>   # The field that contains the field to expand
  override = <optional, bool>     # Controls behavior when a conflicting nested object already exists; default false

  # Common processor fields
  description    = <optional, string>
  if             = <optional, string>
  ignore_failure = <optional, bool>         # default false
  on_failure     = <optional, list(string)> # min 1 item; each element is a JSON string
  tag            = <optional, string>

  # Computed outputs
  id   = <computed, string>  # Hash of the JSON output
  json = <computed, string>  # Serialized processor JSON
}
```

## Requirements

### Requirement: No API calls (REQ-001)

The data source SHALL perform no Elasticsearch API calls. It SHALL NOT require or use an `elasticsearch_connection` block.

#### Scenario: Read without a connection

- GIVEN any valid configuration for this data source
- WHEN the data source is read
- THEN no Elasticsearch API is called and no connection credentials are required

### Requirement: JSON output (REQ-002)

The data source SHALL serialize the processor configuration to a JSON string stored in `json`. The JSON SHALL be wrapped in a top-level `"dot_expander"` key, producing the shape `{"dot_expander": {...}}`.

#### Scenario: JSON wrapping

- GIVEN a valid configuration
- WHEN the data source is read
- THEN `json` SHALL contain a JSON object with a top-level `"dot_expander"` key whose value is the processor parameters

### Requirement: Hash identity (REQ-003)

The data source SHALL set `id` to a deterministic hash of the `json` output. Two configurations that produce identical JSON SHALL produce the same `id`.

#### Scenario: Deterministic id

- GIVEN the same processor configuration applied twice
- WHEN both data sources are read
- THEN both SHALL have identical `id` values

### Requirement: Required attribute (REQ-004)

The data source SHALL require `field`. If it is absent, Terraform SHALL report a configuration error before the read function is called.

#### Scenario: Missing required attribute

- GIVEN a configuration that omits `field`
- WHEN Terraform validates the configuration
- THEN a configuration error SHALL be raised

### Requirement: Default attribute values (REQ-005)

When `override` is not configured, the data source SHALL use `false` as its value. When `ignore_failure` is not configured, the data source SHALL use `false`.

#### Scenario: Defaults in JSON output

- GIVEN a configuration that sets only the required attributes
- WHEN the data source is read
- THEN `json` SHALL include `"override": false` and `"ignore_failure": false`

### Requirement: Optional path field (REQ-006)

When `path` is configured, the data source SHALL include it in the serialized JSON. When `path` is not configured, it SHALL be omitted from the JSON.

#### Scenario: path omitted when unset

- GIVEN `path` is not configured
- WHEN the data source is read
- THEN the `json` output SHALL not include a `"path"` key

### Requirement: Optional common fields (REQ-007)

When `description`, `if`, or `tag` are configured, the data source SHALL include them in the serialized JSON. When they are not configured, they SHALL be omitted from the JSON.

#### Scenario: Optional fields omitted when unset

- GIVEN `description`, `if`, and `tag` are not configured
- WHEN the data source is read
- THEN the `json` output SHALL not include `"description"`, `"if"`, or `"tag"` keys

### Requirement: on_failure JSON parsing (REQ-008)

Each element of `on_failure` SHALL be a valid JSON string. The data source SHALL parse each element as a JSON object and include it in the serialized `on_failure` array. If an element is not valid JSON, the data source SHALL return an error diagnostic.

#### Scenario: Invalid on_failure element

- GIVEN an `on_failure` element that is not valid JSON
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic and SHALL NOT set `json`
