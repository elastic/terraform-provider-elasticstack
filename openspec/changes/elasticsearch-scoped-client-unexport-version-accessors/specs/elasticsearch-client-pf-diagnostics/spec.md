## MODIFIED Requirements

### Requirement: ElasticsearchScopedClient methods return Plugin Framework diagnostics

The methods `serverInfo`, `ClusterID`, `ID`, `EnforceMinVersion`, `EnforceVersionCheck`, and `IsServerless` on `ElasticsearchScopedClient` in `internal/clients/elasticsearch_scoped_client.go` SHALL return `fwdiag.Diagnostics` instead of `diag.Diagnostics` (SDK). No method on `ElasticsearchScopedClient` SHALL import `terraform-plugin-sdk/v2/diag` for method return values. `ElasticsearchScopedClient` SHALL NOT expose `ServerVersion` or `ServerFlavor` as public methods; their package-private replacements (if any) SHALL also return `fwdiag.Diagnostics`.

#### Scenario: ElasticsearchScopedClient.EnforceMinVersion returns PF diagnostics on error
- **GIVEN** a call to `EnforceMinVersion` where cluster info retrieval fails
- **WHEN** `serverInfo` returns an error (e.g., Elasticsearch unreachable)
- **THEN** `EnforceMinVersion` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic

#### Scenario: ElasticsearchScopedClient.EnforceMinVersion returns true for serverless
- **GIVEN** a call to `EnforceMinVersion` where the cluster reports serverless flavor
- **WHEN** flavor check succeeds with "serverless"
- **THEN** `EnforceMinVersion` SHALL return `(true, nil)` unchanged in behavior

#### Scenario: ElasticsearchScopedClient.EnforceVersionCheck returns PF diagnostics on error
- **GIVEN** a call to `EnforceVersionCheck` where cluster info retrieval fails
- **WHEN** `serverInfo` returns an error
- **THEN** `EnforceVersionCheck` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic

#### Scenario: ElasticsearchScopedClient.IsServerless returns PF diagnostics on error
- **GIVEN** a call to `IsServerless` where cluster info retrieval fails
- **WHEN** `serverInfo` returns an error
- **THEN** `IsServerless` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic

#### Scenario: ElasticsearchScopedClient does not expose raw version accessors
- **WHEN** any production consumer of `*clients.ElasticsearchScopedClient` attempts to read the cluster server version or build flavor
- **THEN** no public `ServerVersion()` or `ServerFlavor()` method SHALL be available on the type
- **AND** the consumer SHALL instead obtain serverless-safe answers via `EnforceMinVersion`, `EnforceVersionCheck`, `IsServerless`, or `entitycore.WithVersionRequirements`
