## MODIFIED Requirements

### Requirement: Framework scoped Kibana client resolution
The provider SHALL expose Plugin Framework `kibana_connection` resolution through `*clients.ProviderClientFactory` methods that accept an entity-local `kibana_connection` block and return a `*clients.KibanaScopedClient`. When the block is not configured, the factory SHALL return a `*clients.KibanaScopedClient` built from provider-level defaults. When the block is configured, the factory SHALL return a `*clients.KibanaScopedClient` whose Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: Framework factory falls back to provider defaults
- **WHEN** a Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is absent
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` derived from provider configuration

#### Scenario: Framework factory builds a scoped Kibana-derived client
- **WHEN** a Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is configured
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` rebuilt from that connection for Kibana HTTP operations via the OpenAPI client, SLO, and Fleet operations

### Requirement: SDK scoped Kibana client resolution
The provider SHALL expose SDK `kibana_connection` resolution through `*clients.ProviderClientFactory` methods that accept resource or data source state and return a `*clients.KibanaScopedClient`. When the block is not configured, the factory SHALL return a `*clients.KibanaScopedClient` built from provider-level defaults. When the block is configured, the factory SHALL return a `*clients.KibanaScopedClient` whose Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: SDK factory falls back to provider defaults
- **WHEN** an SDK entity resolves its effective Kibana client through the factory and `kibana_connection` is absent
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` derived from provider configuration

#### Scenario: SDK factory builds a scoped Kibana-derived client
- **WHEN** an SDK entity resolves its effective Kibana client through the factory and `kibana_connection` is configured
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` rebuilt from that connection for Kibana HTTP operations via the OpenAPI client, SLO, and Fleet operations

## ADDED Requirements

### Requirement: Scoped Kibana status reads use OpenAPI client
For `*clients.KibanaScopedClient`, implementations that need the Kibana server version number or build flavor SHALL obtain them from the Kibana `/api/status` response using the generated OpenAPI Kibana client (`generated/kbapi`) for the same effective scoped Kibana connection (via `internal/clients/kibanaoapi` or equivalent provider-internal wiring). The provider SHALL NOT depend on `github.com/disaster37/go-kibana-rest` for this status read after this change is complete. When `version.build_flavor` is absent (older Kibana releases), the flavor result SHALL be an empty string, matching prior behavior for traditional deployments.

#### Scenario: Scoped version uses OpenAPI status
- **WHEN** a covered entity uses `ServerVersion()` on a `*clients.KibanaScopedClient` resolved from `kibana_connection` or provider defaults
- **THEN** the version SHALL be derived from `version.number` in the `/api/status` payload retrieved through the OpenAPI client for that scoped connection

#### Scenario: Scoped flavor uses OpenAPI status
- **WHEN** a covered entity uses `ServerFlavor()` on a `*clients.KibanaScopedClient` resolved from `kibana_connection` or provider defaults
- **THEN** the flavor SHALL be derived from `version.build_flavor` when present in that same `/api/status` payload and SHALL otherwise be an empty string
