## MODIFIED Requirements

### Requirement: Fleet auth config resolution uses method-scoped clearing

When `newFleetConfigFromFramework` applies the Fleet provider block on top of the already-resolved Kibana config, it MUST detect which auth method the Fleet block introduces and clear fields from conflicting auth methods inherited from the Kibana config. Same-method fields from lower-priority sources MUST be preserved.

Clearing rules when the Fleet block introduces a method:
- BasicAuth introduced (`Username` or `Password` set) → clear `APIKey` and `BearerToken`
- APIKey introduced → clear `Username`, `Password`, and `BearerToken`
- BearerToken introduced → clear `Username`, `Password`, and `APIKey`

When no Fleet block is configured or the Fleet block sets no auth fields, the Kibana config auth MUST be inherited unchanged.

#### Scenario: Kibana BasicAuth + Fleet APIKey block → Fleet config has APIKey only
- **GIVEN** the Kibana config has `Username` and `Password` set (resolved from ES base or Kibana block)
- **AND** the Fleet block is configured with `api_key`
- **WHEN** `newFleetConfigFromFramework` resolves the Fleet config
- **THEN** the resolved `fleetConfig` SHALL have `APIKey` set
- **AND** `Username` and `Password` SHALL be empty
- **AND** `BearerToken` SHALL be empty

#### Scenario: Kibana APIKey + no Fleet auth block → Fleet config inherits Kibana APIKey
- **GIVEN** the Kibana config has `APIKey` set
- **AND** the Fleet block sets no auth fields
- **WHEN** `newFleetConfigFromFramework` resolves the Fleet config
- **THEN** the resolved `fleetConfig` SHALL have `APIKey` set
- **AND** `Username` and `Password` SHALL be empty

#### Scenario: Kibana APIKey + Fleet BasicAuth block → Fleet config has BasicAuth only
- **GIVEN** the Kibana config has `APIKey` set
- **AND** the Fleet block is configured with `username` and `password`
- **WHEN** `newFleetConfigFromFramework` resolves the Fleet config
- **THEN** the resolved `fleetConfig` SHALL have `Username` and `Password` set
- **AND** `APIKey` SHALL be empty

### Requirement: Fleet env layer uses method-scoped clearing

`withEnvironmentOverrides` in `fleet.go` MUST detect which Fleet auth env vars are set using `os.LookupEnv` and clear fields from conflicting auth methods before applying env values. Same-method fields from lower-priority sources MUST be preserved.

Clearing rules when Fleet auth env vars are set:
- `FLEET_USERNAME` or `FLEET_PASSWORD` set → clear `APIKey` and `BearerToken`
- `FLEET_API_KEY` set → clear `Username`, `Password`, and `BearerToken`
- `FLEET_BEARER_TOKEN` set → clear `Username`, `Password`, and `APIKey`

#### Scenario: FLEET_API_KEY env + Fleet provider username/password → APIKey wins, BasicAuth cleared
- **GIVEN** the Fleet block sets `username` and `password`
- **AND** the environment has `FLEET_API_KEY=envkey`
- **WHEN** `withEnvironmentOverrides` is applied
- **THEN** the resolved `fleetConfig` SHALL have `APIKey = "envkey"`
- **AND** `Username` and `Password` SHALL be empty

#### Scenario: FLEET_PASSWORD env + Fleet provider username → BasicAuth preserved
- **GIVEN** the Fleet block sets `username = "elastic"`
- **AND** the environment has `FLEET_PASSWORD=secret`
- **WHEN** `withEnvironmentOverrides` is applied
- **THEN** the resolved `fleetConfig` SHALL have `Username = "elastic"` and `Password = "secret"`
- **AND** `APIKey` SHALL be empty

### Requirement: Warning emitted when resolved Fleet config carries multiple auth methods

After the full Fleet config is assembled in `newFleetConfigFromFramework`, if more than one auth method group is populated, a `diag.AddWarning` MUST be emitted with a message that names the conflict and directs the user to check Fleet environment variables. The same `authMethodCount` helper used for the Kibana warning MUST be reused here, counting BasicAuth only when `Username != ""`.

#### Scenario: Env-level conflict (FLEET_API_KEY and FLEET_USERNAME both set) → warning emitted
- **GIVEN** the environment has both `FLEET_API_KEY=envkey` and `FLEET_USERNAME=admin` set simultaneously
- **WHEN** `newFleetConfigFromFramework` finishes
- **THEN** a warning diagnostic SHALL be returned with title "Multiple Fleet authentication methods configured"
- **AND** the body SHALL direct the user to check Fleet environment variables for conflicting auth settings

#### Scenario: Resolved fleet config has exactly one auth method → no warning
- **GIVEN** the Fleet config resolution results in exactly one auth method group being set
- **WHEN** `newFleetConfigFromFramework` finishes
- **THEN** no warning diagnostic SHALL be emitted
