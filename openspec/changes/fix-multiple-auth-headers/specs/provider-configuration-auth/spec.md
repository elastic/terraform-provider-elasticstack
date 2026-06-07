## MODIFIED Requirements

### Requirement: ES credentials inherited into Kibana config only when no Kibana auth is configured

The provider MUST inherit ES authentication credentials into the Kibana config (via `base.toKibanaOapiConfig()`) so that users who configure a single auth source for all services do not need to repeat credentials. However, when the Kibana provider block or Kibana environment variables configure their own auth method, the inherited ES credentials MUST be replaced (not accumulated) using method-scoped clearing rules as defined in `kibana-client-auth`.

#### Scenario: No Kibana auth configured → ES credentials inherited
- **GIVEN** the `elasticsearch` block has `api_key` set
- **AND** the `kibana` block has no auth fields
- **WHEN** the provider is configured
- **THEN** the resolved Kibana config SHALL use the same `api_key` as Elasticsearch

#### Scenario: Kibana auth configured → ES credentials not carried over
- **GIVEN** the `elasticsearch` block has `api_key` set
- **AND** the `kibana` block has `username` and `password` set
- **WHEN** the provider is configured
- **THEN** the resolved Kibana config SHALL have `Username` and `Password` set
- **AND** `APIKey` SHALL NOT be present in the Kibana config
- **AND** Kibana requests SHALL carry exactly one `Authorization` header

### Requirement: `withFleetBlockFallback` does not re-introduce cleared auth methods

`withFleetBlockFallback` uses "fill if empty" guards (`if k.Username == ""`) and MUST NOT overwrite an already-populated auth field. This ensures that a cleared auth method from the ES base cannot be re-introduced through the Fleet block fallback path.

#### Scenario: Kibana BasicAuth set, Fleet block has APIKey → Fleet fallback does not overwrite
- **GIVEN** the Kibana config has `Username` and `Password` set from the Kibana provider block
- **AND** the Fleet block has `api_key` set
- **WHEN** `withFleetBlockFallback` is applied to the Kibana config
- **THEN** the Kibana config SHALL retain `Username` and `Password`
- **AND** `APIKey` SHALL NOT be set (Fleet APIKey applies only to the Fleet client config, not the Kibana config)

#### Scenario: Empty Kibana auth, Fleet block fills it via fallback
- **GIVEN** the Kibana config has no auth fields set
- **AND** the Fleet block has `username` and `password` set
- **WHEN** `withFleetBlockFallback` is applied
- **THEN** `Username` and `Password` SHALL be copied from the Fleet block into the Kibana config

### Requirement: Source priority ENV > RESOURCE > PROVIDER is enforced at every config layer

The provider configuration MUST implement `ENV > RESOURCE > PROVIDER` source priority for authentication fields at every config resolution layer (Kibana schema, Kibana env, Fleet schema, Fleet env). A higher-priority source that introduces an auth method MUST clear conflicting auth method fields from lower-priority sources at each layer. The final resolved config MAY NOT carry credentials from two different auth methods, except in the case of same-method partial composition across sources (e.g. `username` from schema + `KIBANA_PASSWORD` from env).

#### Scenario: ENV overrides PROVIDER — Kibana env APIKey overrides provider BasicAuth
- **GIVEN** the Kibana provider block has `username` and `password`
- **AND** the environment has `KIBANA_API_KEY=envkey`
- **WHEN** the Kibana config is fully resolved
- **THEN** the resolved config SHALL have `APIKey = "envkey"`
- **AND** `Username` and `Password` SHALL be empty
- **AND** Kibana HTTP requests SHALL carry exactly one `Authorization: ApiKey envkey` header

#### Scenario: Same-method partial composition across ENV and PROVIDER is preserved
- **GIVEN** the Kibana provider block has `username = "elastic"`
- **AND** the environment has `KIBANA_PASSWORD = "secret"`
- **WHEN** the Kibana config is fully resolved
- **THEN** the resolved config SHALL have `Username = "elastic"` and `Password = "secret"`
- **AND** `APIKey` SHALL be empty
- **AND** Kibana HTTP requests SHALL carry exactly one `Authorization: Basic ...` header

#### Scenario: Fleet ENV overrides Kibana PROVIDER auth
- **GIVEN** the Kibana provider block has `api_key` set (inherited into Fleet config)
- **AND** the environment has `FLEET_USERNAME=admin` and `FLEET_PASSWORD=pass`
- **WHEN** the Fleet config is fully resolved
- **THEN** the resolved Fleet config SHALL have `Username = "admin"` and `Password = "pass"`
- **AND** `APIKey` SHALL be empty in the Fleet config
- **AND** Fleet HTTP requests SHALL carry exactly one `Authorization: Basic ...` header
