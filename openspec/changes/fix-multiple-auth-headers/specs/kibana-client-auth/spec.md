## MODIFIED Requirements

### Requirement: Kibana auth config resolution uses method-scoped clearing

When `buildKibanaOapiConfigFromFramework` applies the Kibana provider block on top of the ES base config, it MUST detect which auth method the Kibana block introduces and clear fields from conflicting auth methods that were inherited from the ES base. Fields belonging to the same auth method group as the Kibana block's configured fields MUST NOT be cleared, to allow partial auth composition across priority levels (e.g. `username` from provider schema + `KIBANA_PASSWORD` from env are both BasicAuth and must cooperate).

Auth method groups:
- **BasicAuth**: `Username` + `Password`
- **APIKey**: `APIKey`
- **BearerToken**: `BearerToken`

Clearing rules when the Kibana block introduces a method:
- BasicAuth introduced → clear `APIKey` and `BearerToken` inherited from ES base
- APIKey introduced → clear `Username`, `Password`, and `BearerToken` inherited from ES base
- BearerToken introduced → clear `Username`, `Password`, and `APIKey` inherited from ES base

When no Kibana auth block is present, ES credentials MUST be inherited unchanged (common case).

#### Scenario: ES APIKey + Kibana BasicAuth → Kibana config has BasicAuth only
- **GIVEN** the ES block is configured with `api_key`
- **AND** the Kibana block is configured with `username` and `password`
- **WHEN** `buildKibanaOapiConfigFromFramework` resolves the Kibana config
- **THEN** the resolved `kibanaOapiConfig` SHALL have `Username` and `Password` set
- **AND** `APIKey` SHALL be empty
- **AND** `BearerToken` SHALL be empty

#### Scenario: ES APIKey + no Kibana auth block → Kibana config inherits ES APIKey
- **GIVEN** the ES block is configured with `api_key`
- **AND** no Kibana auth fields are present in the Kibana block (or there is no Kibana block)
- **WHEN** `buildKibanaOapiConfigFromFramework` resolves the Kibana config
- **THEN** the resolved `kibanaOapiConfig` SHALL have `APIKey` set to the ES APIKey
- **AND** `Username` and `Password` SHALL be empty

#### Scenario: ES BasicAuth + Kibana APIKey → Kibana config has APIKey only
- **GIVEN** the ES block is configured with `username` and `password`
- **AND** the Kibana block is configured with `api_key`
- **WHEN** `buildKibanaOapiConfigFromFramework` resolves the Kibana config
- **THEN** the resolved `kibanaOapiConfig` SHALL have `APIKey` set
- **AND** `Username` and `Password` SHALL be empty

### Requirement: Kibana env layer uses method-scoped clearing

`withNonURLEnvironmentOverrides` MUST detect which Kibana auth env vars are set using `os.LookupEnv` (to distinguish "not set" from "set to empty string") and clear fields from conflicting auth methods before applying env values. Same-method fields from lower-priority sources MUST be preserved.

Clearing rules when Kibana auth env vars are set:
- `KIBANA_USERNAME` or `KIBANA_PASSWORD` set → clear `APIKey` and `BearerToken`
- `KIBANA_API_KEY` set → clear `Username`, `Password`, and `BearerToken`
- `KIBANA_BEARER_TOKEN` set → clear `Username`, `Password`, and `APIKey`

#### Scenario: KIBANA_PASSWORD env + provider username → BasicAuth preserved, no APIKey
- **GIVEN** the provider Kibana block sets `username = "elastic"`
- **AND** the environment has `KIBANA_PASSWORD=secret`
- **AND** the ES base config has `APIKey` set
- **WHEN** the Kibana config is fully resolved
- **THEN** the resolved `kibanaOapiConfig` SHALL have `Username = "elastic"` and `Password = "secret"`
- **AND** `APIKey` SHALL be empty

#### Scenario: KIBANA_API_KEY env + provider username/password → APIKey wins, BasicAuth cleared
- **GIVEN** the provider Kibana block sets `username` and `password`
- **AND** the environment has `KIBANA_API_KEY=mykey`
- **WHEN** `withNonURLEnvironmentOverrides` is applied
- **THEN** the resolved `kibanaOapiConfig` SHALL have `APIKey = "mykey"`
- **AND** `Username` and `Password` SHALL be empty

### Requirement: Transport applies exactly one Authorization header

`transport.RoundTrip` in `internal/clients/kibanaoapi/client.go` MUST apply auth using a `switch` statement with `req.Header.Set` throughout, ensuring at most one `Authorization` header is added to every request. Priority order MUST be: `BearerToken > APIKey > BasicAuth`.

#### Scenario: APIKey and BasicAuth both set in Config → only APIKey header sent
- **GIVEN** the transport `Config` has both `APIKey` and `Username`/`Password` set
- **WHEN** a request is made
- **THEN** exactly one `Authorization` header SHALL be present: `ApiKey <key>`
- **AND** no `Basic` Authorization header SHALL be present

#### Scenario: BearerToken set → only Bearer header sent
- **GIVEN** the transport `Config` has `BearerToken` set alongside any other auth fields
- **WHEN** a request is made
- **THEN** exactly one `Authorization` header SHALL be present: `Bearer <token>`

#### Scenario: Only BasicAuth set → Basic header sent
- **GIVEN** the transport `Config` has `Username` and `Password` set, no `APIKey` or `BearerToken`
- **WHEN** a request is made
- **THEN** exactly one `Authorization` header SHALL be present with the Base64-encoded basic credential

### Requirement: Warning emitted when resolved Kibana config carries multiple auth methods

After the full Kibana config is assembled (in `newProviderKibanaOapiConfigFromFramework` and `newKibanaOapiConfigFromFramework`), if more than one auth method group is populated, a `diag.AddWarning` MUST be emitted with a message that names the conflict and directs the user to check provider configuration and environment variables.

#### Scenario: Resolved config has both APIKey and Username → warning emitted
- **GIVEN** the Kibana config resolution results in both `APIKey` and `Username` being set (e.g. same-level conflict: both `api_key` and `username` in the Kibana provider block)
- **WHEN** `newProviderKibanaOapiConfigFromFramework` finishes
- **THEN** a warning diagnostic SHALL be returned with title "Multiple Kibana authentication methods configured"

#### Scenario: Resolved config has exactly one auth method → no warning
- **GIVEN** the Kibana config resolution results in exactly one auth method group being set
- **WHEN** `newProviderKibanaOapiConfigFromFramework` finishes
- **THEN** no warning diagnostic SHALL be emitted
