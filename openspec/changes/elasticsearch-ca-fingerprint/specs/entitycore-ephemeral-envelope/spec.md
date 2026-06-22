## MODIFIED Requirements

### Requirement: Connection round-trip is lossless for the documented connection fields (REQ-ENVE-007)

The envelope's connection-snapshot codec SHALL round-trip the following Elasticsearch connection fields without loss when present and known: `endpoints`, `username`, `password`, `api_key`, `bearer_token`, `es_client_authentication`, `insecure`, `ca_file`, `ca_data`, `ca_fingerprint`, `cert_file`, `cert_data`, `key_file`, `key_data`, `headers`.

The Kibana variant SHALL round-trip the equivalent set of `clientconfig.KibanaConnection` fields.

The snapshot codec SHALL NOT use `encoding/json` to marshal `terraform-plugin-framework/types` values directly. The snapshot struct SHALL use plain Go types (`string`, `[]string`, `map[string]string`, `*bool`).

#### Scenario: Boolean fields round-trip both true and false

- **GIVEN** an `elasticsearch_connection` block with `insecure = false`
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL preserve `insecure = false` (not null, not absent)

#### Scenario: List and map fields round-trip without flattening loss

- **GIVEN** an `elasticsearch_connection` block with `endpoints = ["https://a", "https://b"]` and `headers = { "X-Foo" = "bar" }`
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL preserve both endpoints in order and the header map verbatim

#### Scenario: Null connection block produces a null restored List

- **GIVEN** no `elasticsearch_connection` block (null on the model)
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL be a null `types.List` of the connection element type

#### Scenario: CA fingerprint round-trips through ephemeral Open/Close lifecycle

- **GIVEN** an ephemeral resource whose `elasticsearch_connection` block has `ca_fingerprint` set
- **WHEN** Open stores private state and Close restores the connection
- **THEN** the connection reconstructed during Close SHALL carry the same `ca_fingerprint` value
