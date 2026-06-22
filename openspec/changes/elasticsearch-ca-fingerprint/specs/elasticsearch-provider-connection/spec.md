## ADDED Requirements

### Requirement: CA fingerprint attribute on Elasticsearch connection blocks (REQ-007)

Every Elasticsearch connection block surface SHALL expose an optional `ca_fingerprint` string attribute that maps to `elasticsearch.Config.CertificateFingerprint` in the go-elasticsearch SDK. The attribute SHALL be present on `GetEsFWConnectionBlock` (managed resources), `GetEsEphemeralConnectionBlock` (ephemeral resources), `GetEsActionConnectionBlock` (actions), and the provider-level `ProviderConfiguration.Elasticsearch` struct.

When `ca_fingerprint` is set, the provider SHALL set `elasticsearch.Config.CertificateFingerprint` to the supplied value so the client uses the SHA-256 hex fingerprint to pin the server certificate.

`ca_fingerprint` SHALL conflict with `ca_file` and `ca_data`: mixing full-chain CA material with fingerprint pinning SHALL produce a plan-time error diagnostic.

An environment-variable override `ELASTICSEARCH_CA_FINGERPRINT` SHALL be honoured in `withEnvironmentOverrides`, consistent with the `ELASTICSEARCH_<ATTR>` pattern used for all other connection attributes.

The `ephemeralESConnectionSnapshot` struct SHALL include a `CAFingerprint` field so ephemeral resources correctly restore the fingerprint on the Close lifecycle call.

The fallback `attrTypes` map `elasticsearchConnectionBlockObjectAttrTypesFallback` in `internal/schema/connection.go` SHALL include `ca_fingerprint: types.StringType` to remain consistent with `GetEsFWConnectionBlock` and prevent type-mismatch errors during ImportState.

#### Scenario: CA fingerprint pins the connection

- GIVEN an `elasticsearch` provider block with `ca_fingerprint` set to a valid SHA-256 hex string
- WHEN the provider initialises and creates an Elasticsearch client
- THEN `elasticsearch.Config.CertificateFingerprint` SHALL equal the configured value
- AND the client SHALL use the fingerprint to verify the server certificate

#### Scenario: CA fingerprint env-var override applies

- GIVEN `ELASTICSEARCH_CA_FINGERPRINT` is set to a hex string in the environment
- AND no `ca_fingerprint` is set in the HCL configuration
- WHEN the provider initialises
- THEN `elasticsearch.Config.CertificateFingerprint` SHALL equal the value from the env-var

#### Scenario: CA fingerprint conflicts with ca_file

- GIVEN a connection block with both `ca_fingerprint` and `ca_file` set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return an error diagnostic referencing the conflict

#### Scenario: CA fingerprint conflicts with ca_data

- GIVEN a connection block with both `ca_fingerprint` and `ca_data` set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return an error diagnostic referencing the conflict

#### Scenario: CA fingerprint is preserved across ephemeral Open/Close lifecycle

- GIVEN an ephemeral resource whose `elasticsearch_connection` block has `ca_fingerprint` set
- WHEN Open stores private state and Close restores the connection
- THEN the connection reconstructed during Close SHALL carry the same `ca_fingerprint` value
