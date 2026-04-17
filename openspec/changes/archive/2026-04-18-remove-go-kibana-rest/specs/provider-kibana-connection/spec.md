## MODIFIED Requirements

### Requirement: Scoped Kibana status reads use OpenAPI client

For `*clients.KibanaScopedClient`, implementations that need the Kibana server version number or build flavor SHALL obtain them from the Kibana `/api/status` response using the generated OpenAPI Kibana client (`generated/kbapi`) for the same effective scoped Kibana connection (via `internal/clients/kibanaoapi` or equivalent provider-internal wiring). The provider SHALL NOT depend on `github.com/disaster37/go-kibana-rest` anywhere in the `kibana_connection` resolution path after this change is complete, including config wiring used to build scoped clients and synthetics read paths that consume those scoped clients. When `version.build_flavor` is absent (older Kibana releases), the flavor result SHALL be an empty string, matching prior behavior for traditional deployments.

#### Scenario: Scoped version uses OpenAPI status
- **WHEN** a covered entity uses `ServerVersion()` on a `*clients.KibanaScopedClient` resolved from `kibana_connection` or provider defaults
- **THEN** the version SHALL be derived from `version.number` in the `/api/status` payload retrieved through the OpenAPI client for that scoped connection

#### Scenario: Scoped flavor uses OpenAPI status
- **WHEN** a covered entity uses `ServerFlavor()` on a `*clients.KibanaScopedClient` resolved from `kibana_connection` or provider defaults
- **THEN** the flavor SHALL be derived from `version.build_flavor` when present in that same `/api/status` payload and SHALL otherwise be an empty string

#### Scenario: Scoped connection path excludes legacy Kibana REST wiring
- **WHEN** a covered Kibana or Fleet entity resolves a provider-level or entity-local `kibana_connection`
- **THEN** the resulting scoped client path SHALL not require `github.com/disaster37/go-kibana-rest` configuration or error handling to perform version, flavor, or synthetics parameter read behavior
