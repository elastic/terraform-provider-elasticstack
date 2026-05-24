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

#### Scenario: Factory does not require a legacy Kibana config surface
- **WHEN** the provider client factory resolves a Kibana-scoped client from provider configuration or `kibana_connection`
- **THEN** it SHALL validate and build that client from the Kibana OpenAPI config surface without relying on a parallel legacy Kibana REST config object

