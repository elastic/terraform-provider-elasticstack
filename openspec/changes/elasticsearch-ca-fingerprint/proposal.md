## Why

Users connecting to Elasticsearch clusters that present self-signed or custom CA certificates can
already supply `ca_file` (a path to a PEM certificate) or `ca_data` (inline PEM). However, a
common deployment pattern — especially with Elastic Cloud and single-node clusters — is for the
cluster to print a **TLS certificate SHA-256 fingerprint** on first launch. Users are told to copy
that fingerprint and use it to pin the connection rather than managing a full CA chain. This "CA
fingerprint" workflow is natively supported by the underlying `go-elasticsearch/v8` client via
`Config.CertificateFingerprint` but is not currently exposed by the Terraform provider.

## What Changes

- Add a new optional `ca_fingerprint` string attribute to the Elasticsearch connection block.
- The attribute maps directly to `elasticsearch.Config.CertificateFingerprint` in the
  `go-elasticsearch/v8` SDK, which accepts a hex-encoded SHA-256 fingerprint of the server
  certificate.
- An environment-variable override `ELASTICSEARCH_CA_FINGERPRINT` will be honoured, consistent
  with the pattern used by `ELASTICSEARCH_INSECURE`, `ELASTICSEARCH_ENDPOINTS`, etc.
- Conflicts with `ca_file` and `ca_data` — mixing full-chain CA material with fingerprint pinning
  is redundant and would produce confusing behaviour.
- The attribute must be present on all four Elasticsearch connection block surfaces so users do not
  encounter inconsistencies depending on whether they use the provider-level block, a per-resource
  `elasticsearch_connection` block, an ephemeral resource, or an action.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `provider-elasticsearch-connection`: extend with `ca_fingerprint` optional attribute and
  `ELASTICSEARCH_CA_FINGERPRINT` env-var override.
- `entitycore-ephemeral-envelope`: extend ephemeral connection snapshot round-trip with
  `ca_fingerprint`.

## Impact

**Schema:**

- `internal/schema/connection_constants.go` — add `attrCAFingerprint` constant and
  `descCAFingerprint` description.
- `internal/schema/connection.go` — add `ca_fingerprint` attribute to `GetEsFWConnectionBlock`
  and its fallback `attrTypes` map.
- `internal/schema/ephemeral_connection.go` — add `ca_fingerprint` attribute to
  `GetEsEphemeralConnectionBlock`.
- `internal/schema/action_connection.go` — add `ca_fingerprint` attribute to
  `GetEsActionConnectionBlock`.

**Config structs and wiring:**

- `internal/clients/config/provider.go` — add `CAFingerprint types.String \`tfsdk:"ca_fingerprint"\`` to `ElasticsearchConnection`.
- `internal/clients/config/elasticsearch.go` — wire `CAFingerprint` to
  `config.config.CertificateFingerprint` in `newElasticsearchConfigFromFramework` and add
  `ELASTICSEARCH_CA_FINGERPRINT` lookup to `withEnvironmentOverrides`.

**Snapshot (ephemeral private state):**

- `internal/entitycore/ephemeral_connection_snapshot.go` — add `CAFingerprint` field to
  `ephemeralESConnectionSnapshot` and update the `snapshotFromElasticsearchConnection` /
  `elasticsearchConnectionFromSnapshot` helpers.

**Documentation:**

- `templates/index.md.tmpl` or equivalent — document the new attribute and env-var override.
