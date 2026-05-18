## MODIFIED Requirements

### Requirement: KibanaScopedClient methods return Plugin Framework diagnostics
The methods `ServerVersion`, `ServerFlavor`, and `EnforceMinVersion` on `KibanaScopedClient` in `internal/clients/kibana_scoped_client.go` SHALL return `fwdiag.Diagnostics` instead of `diag.Diagnostics` (SDK). These methods SHALL consume PF diagnostics directly from `kibanaoapi.GetKibanaStatus` without bridging via `diagutil.FrameworkDiagsFromSDK`.

#### Scenario: KibanaScopedClient.EnforceMinVersion returns PF diagnostics on error
- **GIVEN** a call to `KibanaScopedClient.EnforceMinVersion` where Kibana is unreachable
- **WHEN** `GetKibanaOapiClient` or `GetKibanaStatus` fails
- **THEN** `EnforceMinVersion` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic
- **AND** it SHALL NOT call `diagutil.FrameworkDiagsFromSDK`

#### Scenario: KibanaScopedClient.ServerVersion passes through PF diagnostics
- **GIVEN** a call to `KibanaScopedClient.ServerVersion`
- **WHEN** `kibanaoapi.GetKibanaStatus` returns PF diagnostics
- **THEN** `ServerVersion` SHALL return those diagnostics directly without wrapping

### Requirement: Bridging helpers removed when unused
`diagutil.SDKErrorDiag`, `diagutil.FrameworkDiagsFromSDK`, and `diagutil.SDKDiagsFromFramework` in `internal/diagutil/translation.go` SHALL be removed. All callers of these functions SHALL be eliminated as part of this change.

#### Scenario: Helpers are deleted
- **GIVEN** that all ES client functions, scoped-client methods, kibanaoapi functions, and provider factory methods have been migrated
- **WHEN** inspecting `internal/diagutil/translation.go`
- **THEN** `SDKErrorDiag`, `FrameworkDiagsFromSDK`, and `SDKDiagsFromFramework` SHALL NOT exist in the file

#### Scenario: diagutil has no SDK diag import
- **GIVEN** the codebase after this change
- **WHEN** inspecting `internal/diagutil/translation.go`
- **THEN** it SHALL NOT import `github.com/hashicorp/terraform-plugin-sdk/v2/diag`
