## Context

The `go-elasticsearch/v8` client exposes `Config.CertificateFingerprint` (a plain `string`
containing a hex SHA-256 fingerprint). When set, the client compares the fingerprint of the server
certificate against this value instead of (or in addition to) performing a full CA chain
validation. This is the recommended path for Elastic Cloud and self-managed deployments where the
operator reads the fingerprint from the cluster output rather than exporting a CA file.

The provider currently exposes `ca_file` / `ca_data` for full CA material and `insecure` to bypass
TLS verification entirely. There is no intermediate option that pins by fingerprint. This gap is
reported in issue #308.

All four connection block surfaces (`GetEsFWConnectionBlock`, `GetEsEphemeralConnectionBlock`,
`GetEsActionConnectionBlock`, and the provider-level block configured via `ProviderConfiguration`)
are structurally identical for the Elasticsearch connection, so the new attribute must be added to
all four to avoid confusing gaps.

## Goals / Non-Goals

**Goals:**

- Expose `ca_fingerprint` as an optional string attribute on every Elasticsearch connection block
  surface.
- Wire it to `elasticsearch.Config.CertificateFingerprint` in the client construction path.
- Provide an `ELASTICSEARCH_CA_FINGERPRINT` environment-variable override, consistent with the
  existing override pattern.
- Conflict-validate against `ca_file` / `ca_data` to prevent user error.

**Non-Goals:**

- Changing `ca_file` / `ca_data` semantics.
- Adding fingerprint support to the Kibana or Fleet connection blocks — those blocks use a
  different underlying client (`kbapi`) that does not expose a fingerprint field. This is a
  separate request.
- Validating the fingerprint format (hex string, 64 chars) at plan time in this change — useful
  but not required for correctness; can be added as a follow-up.

## Decisions

- **Naming**: `ca_fingerprint` — consistent with `ca_file` / `ca_data` prefix and the concept of
  pinning the CA/server certificate by its digest.
- **ConflictsWith `ca_file` and `ca_data`**: the three attributes represent mutually exclusive ways
  of anchoring TLS trust. Mixing them is almost certainly a configuration mistake, so a validator
  provides a useful guardrail without losing functionality.
- **No ConflictsWith `insecure`**: the go-elasticsearch client accepts both simultaneously without
  error (fingerprint check runs before the TLS chain check). While combining them is arguably
  nonsensical, it is not technically wrong and adding the conflict would silently break any user
  who happens to have both set during a migration. This can be added as a follow-up if desired.
- **Env-var name `ELASTICSEARCH_CA_FINGERPRINT`**: follows the `ELASTICSEARCH_<ATTR>` pattern used
  by all other Elasticsearch connection env-var overrides in `withEnvironmentOverrides`.
- **Ephemeral snapshot**: the `ephemeralESConnectionSnapshot` JSON struct persists the connection
  for the Close lifecycle. `CAFingerprint` must be added there so ephemeral resources restore the
  fingerprint correctly on Close.

## Risks / Trade-offs

- [Low] `CertificateFingerprint` and `CACert` (set from `ca_file`/`ca_data`) can both be set in
  the SDK config; the conflict validator prevents this at the Terraform layer but if a user bypasses
  via env-vars the SDK will use both. This is an edge case with no impact on typical usage.
- [Low] The fallback `elasticsearchConnectionBlockObjectAttrTypesFallback` map in `connection.go`
  must be kept in sync with `GetEsFWConnectionBlock`. The existing sync-once pattern means that if
  the map is missing the new key, an ImportState operation on a resource using this block may
  produce a type-mismatch error. The task list calls this out explicitly.

## Open questions

1. Should `ca_fingerprint` conflict with `insecure = true`? The go-elasticsearch client accepts
   both without error (fingerprint check runs before the standard TLS chain check), but it is
   arguably a user mistake to set both. A `ConflictsWith` validator would be a usability guardrail
   but is not strictly required for correctness.
2. Should `ELASTICSEARCH_CA_FINGERPRINT` be the correct env var name, or should it mirror what
   Elastic Cloud tooling or other providers use? (Current recommendation: yes, matches the existing
   `ELASTICSEARCH_<ATTR>` pattern in this provider.)
3. Should a future change add `ca_fingerprint` to the Kibana or Fleet connection blocks? Those
   blocks use different underlying clients that currently do not expose a fingerprint field; out of
   scope for this change, tracked as a separate request.
