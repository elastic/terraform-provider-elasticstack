## MODIFIED Requirements

### Requirement: KibanaScopedClient methods return Plugin Framework diagnostics

The methods `EnforceMinVersion` and `EnforceVersionCheck` on `KibanaScopedClient` in `internal/clients/kibana_scoped_client.go` SHALL return `fwdiag.Diagnostics` instead of `diag.Diagnostics` (SDK). These methods SHALL consume PF diagnostics directly from `kibanaoapi.GetKibanaStatus` without bridging via `diagutil.FrameworkDiagsFromSDK`. `KibanaScopedClient` SHALL NOT expose `ServerVersion` or `ServerFlavor` as public methods; the package-private `getServerStatusRaw` helper that underpins `EnforceMinVersion` and `EnforceVersionCheck` SHALL likewise consume `fwdiag.Diagnostics` directly.

#### Scenario: KibanaScopedClient.EnforceMinVersion returns PF diagnostics on error
- **GIVEN** a call to `KibanaScopedClient.EnforceMinVersion` where Kibana is unreachable
- **WHEN** `GetKibanaOapiClient` or `GetKibanaStatus` fails
- **THEN** `EnforceMinVersion` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic
- **AND** it SHALL NOT call `diagutil.FrameworkDiagsFromSDK`

#### Scenario: KibanaScopedClient.EnforceVersionCheck returns PF diagnostics on error
- **GIVEN** a call to `KibanaScopedClient.EnforceVersionCheck` where Kibana is unreachable
- **WHEN** `GetKibanaOapiClient` or `GetKibanaStatus` fails
- **THEN** `EnforceVersionCheck` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic
- **AND** it SHALL NOT call `diagutil.FrameworkDiagsFromSDK`

#### Scenario: KibanaScopedClient does not expose raw version accessors
- **WHEN** a consumer of `*clients.KibanaScopedClient` attempts to read the Kibana server version or build flavor
- **THEN** no public `ServerVersion()` or `ServerFlavor()` method SHALL be available on the type
- **AND** the consumer SHALL instead obtain serverless-safe answers via `EnforceMinVersion`, `EnforceVersionCheck`, or `entitycore.WithVersionRequirements`
