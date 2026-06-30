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

### Requirement: `withFleetBlockFallback` does not fill auth when any Kibana auth method is already set

Fleet (provider block and env) is the lowest-priority auth source for Kibana resources. `withFleetBlockFallback` MUST check whether any auth method is already set in the Kibana config before filling any auth field. If any of `Username`, `Password`, `APIKey`, or `BearerToken` is non-empty, all auth field filling from the Fleet block MUST be skipped. URL, CA certs, and TLS settings are not subject to this restriction and continue to use field-level guards.

This prevents a cleared auth method from being re-introduced. For example: if the Kibana block sets BasicAuth (causing inherited ES `APIKey` to be cleared), `withFleetBlockFallback` must not refill `APIKey` from the Fleet block simply because the field is now empty.

#### Scenario: Kibana BasicAuth set, Fleet block has APIKey → Fleet auth not used
- **GIVEN** the Kibana config has `Username` and `Password` set (from the Kibana provider block, after method-scoped clearing)
- **AND** the Fleet block has `api_key` set
- **WHEN** `withFleetBlockFallback` is applied to the Kibana config
- **THEN** the Kibana config SHALL retain `Username` and `Password`
- **AND** `APIKey` SHALL NOT be set

#### Scenario: Kibana APIKey set (inherited from ES), Fleet block has BasicAuth → Fleet auth not used
- **GIVEN** the Kibana config has `APIKey` set (inherited from the ES base, no Kibana auth block)
- **AND** the Fleet block has `username` and `password` set
- **WHEN** `withFleetBlockFallback` is applied to the Kibana config
- **THEN** the Kibana config SHALL retain `APIKey`
- **AND** `Username` and `Password` SHALL NOT be set

#### Scenario: Empty Kibana auth, Fleet block fills it via fallback
- **GIVEN** the Kibana config has no auth fields set (no ES credentials, no Kibana block auth)
- **AND** the Fleet block has `username` and `password` set
- **WHEN** `withFleetBlockFallback` is applied
- **THEN** `Username` and `Password` SHALL be copied from the Fleet block into the Kibana config

### Requirement: Source priority is enforced at every config layer for Kibana resources

The provider MUST enforce the following auth source priority for Kibana-facing requests (highest to lowest):

1. `KIBANA_*` environment variables
2. Resource-level `kibana_connection` block
3. Provider-level `kibana` block
4. `FLEET_*` environment variables
5. Provider-level `fleet` block

A higher-priority source that introduces an auth method MUST clear conflicting auth method fields from lower-priority sources at each layer. The final resolved config MAY NOT carry credentials from two different auth methods, except in the case of same-method partial composition across adjacent sources (e.g. `username` from provider schema + `KIBANA_PASSWORD` from env). Fleet (env and provider block) is the lowest-priority fallback and MUST be skipped entirely for auth if any auth method is already set by a higher-priority source.

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
