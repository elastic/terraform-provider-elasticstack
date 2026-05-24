## ADDED Requirements

### Requirement: Elasticsearch scoped client serverless-safe version surface

The typed Elasticsearch-scoped client returned by the provider client factory SHALL expose serverless-safe version- and flavor-gating primitives as its only public means of consulting the connected cluster's server version or build flavor. The public primitives SHALL be `EnforceMinVersion(ctx, minVersion) (bool, fwdiag.Diagnostics)`, `EnforceVersionCheck(ctx, check func(*version.Version) bool) (bool, fwdiag.Diagnostics)`, and `IsServerless(ctx) (bool, fwdiag.Diagnostics)`. `EnforceMinVersion` and `EnforceVersionCheck` SHALL short-circuit to `true` when the cluster build flavor is `"serverless"`. The Elasticsearch resource envelope SHALL continue to evaluate `entitycore.WithVersionRequirements` via `EnforceMinVersion` during Create, Read, and Update.

The Elasticsearch scoped client SHALL NOT expose `ServerVersion()` or `ServerFlavor()` as public methods. Their underlying behaviour SHALL remain available only through the serverless-safe primitives above, with any raw accessors kept package-private to `internal/clients`. An acceptance-test-only helper MAY exist in `internal/clients` to retrieve the cluster version and serverless flag together for acceptance-test skip plumbing; production code SHALL NOT use it.

#### Scenario: Resource gates on minimum version with serverless awareness
- **WHEN** an Elasticsearch entity gates an attribute on a minimum cluster version
- **THEN** it SHALL use `client.EnforceMinVersion`, `client.EnforceVersionCheck`, or declare a `entitycore.WithVersionRequirements` requirement on its model
- **AND** the resulting check SHALL succeed on serverless clusters regardless of the reported version string

#### Scenario: Resource asks the flavor question directly
- **WHEN** an Elasticsearch entity needs to know whether the cluster is serverless for a non-version reason (e.g., to omit a stateful-only request field)
- **THEN** it SHALL use `client.IsServerless(ctx)` and SHALL NOT read the raw flavor string

#### Scenario: Public surface forbids raw version accessors
- **WHEN** any production code attempts to read the Elasticsearch server version or build flavor from `*clients.ElasticsearchScopedClient`
- **THEN** no public `ServerVersion()` or `ServerFlavor()` method SHALL be available
- **AND** the consumer SHALL route its decision through `EnforceMinVersion`, `EnforceVersionCheck`, `IsServerless`, or `entitycore.WithVersionRequirements`
