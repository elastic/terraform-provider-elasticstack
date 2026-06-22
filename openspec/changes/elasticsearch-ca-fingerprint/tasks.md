## 1. Constants

- [ ] 1.1 Add `attrCAFingerprint = "ca_fingerprint"` constant to `internal/schema/connection_constants.go`
- [ ] 1.2 Add `descCAFingerprint` description constant: `"SHA-256 hex fingerprint of the server TLS certificate used to pin the connection instead of a full CA chain"`

## 2. Config struct

- [ ] 2.1 Add `CAFingerprint types.String \`tfsdk:"ca_fingerprint"\`` to `ElasticsearchConnection` in `internal/clients/config/provider.go`

## 3. Schema — managed resources (`GetEsFWConnectionBlock`)

- [ ] 3.1 Declare `caFingerprintPath` alongside the other path variables in `GetEsFWConnectionBlock` (`internal/schema/connection.go`)
- [ ] 3.2 Add `attrCAFingerprint` as an optional `fwschema.StringAttribute` with `ConflictsWith(caFilePath, caDataPath)` validator
- [ ] 3.3 Update `elasticsearchConnectionBlockObjectAttrTypesFallback` to include `attrCAFingerprint: types.StringType` so the fallback map stays consistent with the live schema

## 4. Schema — ephemeral resources (`GetEsEphemeralConnectionBlock`)

- [ ] 4.1 Declare `caFingerprintPath` in `GetEsEphemeralConnectionBlock` (`internal/schema/ephemeral_connection.go`)
- [ ] 4.2 Add `attrCAFingerprint` as an optional `schema.StringAttribute` with `ConflictsWith(caFilePath, caDataPath)` validator, mirroring task 3.2

## 5. Schema — actions (`GetEsActionConnectionBlock`)

- [ ] 5.1 Declare `caFingerprintPath` in `GetEsActionConnectionBlock` (`internal/schema/action_connection.go`)
- [ ] 5.2 Add `attrCAFingerprint` as an optional `schema.StringAttribute` with `ConflictsWith(caFilePath, caDataPath)` validator, mirroring task 3.2

## 6. Client wiring

- [ ] 6.1 In `newElasticsearchConfigFromFramework` (`internal/clients/config/elasticsearch.go`), after the `ca_file`/`ca_data` block, add:
  ```go
  if fingerprint := esConfig.CAFingerprint.ValueString(); fingerprint != "" {
      config.config.CertificateFingerprint = fingerprint
  }
  ```
- [ ] 6.2 In `withEnvironmentOverrides`, add an `ELASTICSEARCH_CA_FINGERPRINT` env-var override that sets `c.config.CertificateFingerprint`

## 7. Ephemeral connection snapshot

- [ ] 7.1 Add `CAFingerprint string \`json:"ca_fingerprint,omitempty"\`` to `ephemeralESConnectionSnapshot` in `internal/entitycore/ephemeral_connection_snapshot.go`
- [ ] 7.2 Set `CAFingerprint: knownStringValue(conn.CAFingerprint)` in `snapshotFromElasticsearchConnection`
- [ ] 7.3 Set `CAFingerprint: typeutils.NonEmptyStringishValue(snapshot.CAFingerprint)` in `elasticsearchConnectionFromSnapshot`

## 8. Documentation

- [ ] 8.1 Add `ca_fingerprint` and `ELASTICSEARCH_CA_FINGERPRINT` to the provider documentation template (e.g. `templates/index.md.tmpl` or the relevant provider docs)

## 9. Tests

- [ ] 9.1 Write a unit test confirming that `ca_fingerprint` in provider config wires to `elasticsearch.Config.CertificateFingerprint`
- [ ] 9.2 Write a unit test confirming the `ELASTICSEARCH_CA_FINGERPRINT` env-var override applies
- [ ] 9.3 Write a unit test confirming `ca_fingerprint` conflicts with `ca_file` and with `ca_data` at the schema validator level
- [ ] 9.4 (Optional) Acceptance test for provider-level and per-resource connection block (requires a cluster serving a known fingerprint — mark with `t.Skip` if environment is unavailable)

## 10. Build and spec validation

- [ ] 10.1 Run `make build` and confirm the project compiles without errors
- [ ] 10.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ca-fingerprint --type change` and resolve any issues
- [ ] 10.3 Run `make check-openspec`
