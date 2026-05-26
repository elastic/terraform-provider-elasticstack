# provider-client-factory Specification

## Purpose
TBD - created by archiving change typed-kibana-fleet-client-resolution. Update Purpose after archive.
## Requirements
### Requirement: Provider injects a client factory
The provider SHALL inject a `*clients.ProviderClientFactory` into Plugin Framework `ProviderData` and `ResourceData` as the provider-scoped client-resolution surface for resources and data sources. Covered consumers SHALL resolve typed scoped clients through factory methods rather than converting provider data or meta back into a broad `*clients.APIClient`.

#### Scenario: Framework configure receives a factory
- **WHEN** the Plugin Framework provider configures a resource or data source
- **THEN** the configured provider data SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

#### Scenario: Framework consumer resolves typed client from factory
- **WHEN** a covered Framework resource or data source needs Elasticsearch- or Kibana-derived operations
- **THEN** it SHALL obtain a typed scoped client from `*clients.ProviderClientFactory` instead of converting provider data into a broad `*clients.APIClient`

### Requirement: Factory supports phased migration
During the Kibana/Fleet typed-client phase, the `*clients.ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and SHALL also preserve explicit legacy Elasticsearch resolution methods so unconverted Elasticsearch entities continue to behave as they did before the factory migration.

#### Scenario: Kibana entity resolves typed client
- **WHEN** a Kibana or Fleet entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Kibana-scoped client whose surfaces include the Kibana OpenAPI client, SLO client, and Fleet client for their respective operations

#### Scenario: Elasticsearch entity uses transitional legacy resolution
- **WHEN** an unconverted Elasticsearch entity resolves its effective client during this phase
- **THEN** the factory SHALL expose a transitional resolution path that preserves the existing broad-client and lint-enforced Elasticsearch behavior

### Requirement: Kibana scoped client contract

The typed Kibana-scoped client returned by the factory SHALL expose the Kibana OpenAPI client, SLO client, Fleet client, Kibana auth-context helpers, and serverless-safe version-gating primitives required by covered Kibana and Fleet entities. The factory contract SHALL use the Kibana OpenAPI configuration surface as the only Kibana connection contract and SHALL NOT expose or require `github.com/disaster37/go-kibana-rest` as part of provider wiring.

The Kibana scoped client's public version-gating surface SHALL consist of `EnforceMinVersion(ctx, minVersion) (bool, diag.Diagnostics)`, `EnforceVersionCheck(ctx, check) (bool, diag.Diagnostics)`, and automatic enforcement of `entitycore.WithVersionRequirements` by the Kibana resource envelope. Both `EnforceMinVersion` and `EnforceVersionCheck` SHALL short-circuit to `true` when the server build flavor is `"serverless"`. The Kibana scoped client SHALL NOT expose `ServerVersion()` or `ServerFlavor()` as public methods; raw version and flavor accessors SHALL be package-private to `internal/clients` so that all version-gated decisions go through serverless-aware primitives.

The Kibana scoped client SHALL cache the successful `(rawVersion, flavor)` result of its first `/api/status` fetch for the lifetime of each `KibanaScopedClient` instance so that multiple version-gated decisions performed during a single resource operation share one Kibana status round-trip. Concurrent callers SHALL serialize on the in-flight fetch and observe the cached result instead of issuing parallel requests. Error results SHALL NOT be cached; a subsequent `EnforceMinVersion` or `EnforceVersionCheck` call SHALL re-attempt the request so transient failures recover naturally.

#### Scenario: Scoped client supports Kibana and Fleet operations
- **WHEN** a covered Kibana or Fleet entity uses a typed Kibana-scoped client
- **THEN** the client SHALL provide the typed client surfaces needed for Kibana HTTP workloads through the OpenAPI client, plus SLO and Fleet API operations as applicable

#### Scenario: Scoped client supports serverless-safe version gating
- **WHEN** a covered Kibana or Fleet entity performs a minimum-version check through the typed Kibana-scoped client
- **THEN** the client SHALL expose `EnforceMinVersion(ctx, minVersion)`, `EnforceVersionCheck(ctx, check)`, and SHALL enforce `entitycore.WithVersionRequirements` declared by entity models through the Kibana resource envelope
- **AND** each of these primitives SHALL short-circuit to "supported" when the server build flavor is `"serverless"`

#### Scenario: Public surface forbids raw version accessors
- **WHEN** a covered Kibana or Fleet entity attempts to read the Kibana server version or build flavor directly from the typed Kibana-scoped client
- **THEN** no public `ServerVersion()` or `ServerFlavor()` method SHALL be available on the client
- **AND** the entity SHALL instead route its decision through `EnforceMinVersion`, `EnforceVersionCheck`, or `entitycore.WithVersionRequirements`

#### Scenario: Repeated version-gating decisions share one status round-trip
- **GIVEN** a `KibanaScopedClient` instance whose first `EnforceMinVersion` or `EnforceVersionCheck` call returns a successful status
- **WHEN** subsequent `EnforceMinVersion` or `EnforceVersionCheck` calls execute on the same instance, including concurrent calls from multiple goroutines
- **THEN** the client SHALL return the cached `(rawVersion, flavor)` without issuing additional `kibanaoapi.GetKibanaStatus` requests

#### Scenario: Failed status fetches are not cached
- **GIVEN** a `KibanaScopedClient` whose first `EnforceMinVersion` or `EnforceVersionCheck` call fails (e.g. transient `kibanaoapi.GetKibanaStatus` error, missing endpoint, or HTTP 500)
- **WHEN** a subsequent `EnforceMinVersion` or `EnforceVersionCheck` call executes on the same instance
- **THEN** the client SHALL re-attempt the `kibanaoapi.GetKibanaStatus` request rather than returning the previous error from a cache

#### Scenario: Factory does not require a legacy Kibana config surface
- **WHEN** the provider client factory resolves a Kibana-scoped client from provider configuration or `kibana_connection`
- **THEN** it SHALL validate and build that client from the Kibana OpenAPI config surface without relying on a parallel legacy Kibana REST config object

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

### Requirement: Factory validates endpoint presence as a precondition
`ProviderClientFactory.GetElasticsearchClient` and `ProviderClientFactory.GetKibanaClient` SHALL validate that at least one effective endpoint is configured for the requested component before returning a scoped client. On success, the returned scoped client's accessors SHALL be safe to call unconditionally and SHALL return a non-nil, ready-to-use typed client. On failure, the factory SHALL return error diagnostics naming the configuration paths the user can set, and SHALL NOT return a partially-configured scoped client.

#### Scenario: Elasticsearch precondition fails when no ES endpoint is configured
- **GIVEN** provider configuration, `elasticsearch_connection`, and environment overrides that together produce no non-empty Elasticsearch endpoint value
- **WHEN** `GetElasticsearchClient` is called
- **THEN** the factory SHALL return an error diagnostic instructing the user to set `elasticsearch.endpoints`, `elasticsearch_connection.endpoints`, or `ELASTICSEARCH_ENDPOINTS`
- **AND** the factory SHALL NOT return a scoped client

#### Scenario: Kibana precondition fails only when both Kibana and Fleet endpoints are missing
- **GIVEN** provider configuration, `kibana_connection`, and environment overrides that together produce no non-empty Kibana endpoint value **and** no non-empty Fleet endpoint value
- **WHEN** `GetKibanaClient` is called
- **THEN** the factory SHALL return an error diagnostic instructing the user to set one of `kibana.endpoints`, `kibana_connection.endpoints`, `KIBANA_ENDPOINT`, `fleet.endpoint`, or `FLEET_ENDPOINT`
- **AND** the factory SHALL NOT return a scoped client

#### Scenario: Kibana precondition is satisfied by either Kibana or Fleet endpoint
- **GIVEN** provider configuration where exactly one of the Kibana or Fleet endpoint values is non-empty after all overlays
- **WHEN** `GetKibanaClient` is called
- **THEN** the factory SHALL return a `*KibanaScopedClient` whose `GetKibanaOapiClient()` and `GetFleetClient()` both return non-nil clients

#### Scenario: Successful factory call guarantees usable accessors
- **GIVEN** a `*ElasticsearchScopedClient` or `*KibanaScopedClient` returned by a successful factory call
- **WHEN** the consumer calls the scoped client's typed-client accessor (`GetESClient`, `GetKibanaOapiClient`, or `GetFleetClient`)
- **THEN** the accessor SHALL return a non-nil typed client without diagnostics

