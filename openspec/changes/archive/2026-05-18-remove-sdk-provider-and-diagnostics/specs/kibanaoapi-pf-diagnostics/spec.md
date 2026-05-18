## ADDED Requirements

### Requirement: kibanaoapi status returns PF diagnostics
`internal/clients/kibanaoapi/status.go` `GetKibanaStatus` SHALL return `github.com/hashicorp/terraform-plugin-framework/diag.Diagnostics` instead of `github.com/hashicorp/terraform-plugin-sdk/v2/diag.Diagnostics`.

#### Scenario: Kibana status error returns PF diagnostic
- **GIVEN** a call to `GetKibanaStatus` where the HTTP request fails
- **WHEN** the function returns
- **THEN** it SHALL return `("", "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(...)})`
- **AND** it SHALL NOT return `sdkdiag.Diagnostics`

#### Scenario: Kibana status success returns nil
- **GIVEN** a call to `GetKibanaStatus` that succeeds
- **WHEN** the function returns
- **THEN** it SHALL return `(versionString, flavorString, nil)`

### Requirement: kibanaoapi security role operations return PF diagnostics
All public functions in `internal/clients/kibanaoapi/security_role.go` (`GetSecurityRole`, `PutSecurityRole`, `DeleteSecurityRole`) SHALL return `fwdiag.Diagnostics` instead of `sdkdiag.Diagnostics`.

#### Scenario: GetSecurityRole HTTP error returns PF diagnostic
- **GIVEN** a call to `GetSecurityRole` where the Kibana API returns an unexpected status
- **WHEN** the function returns
- **THEN** it SHALL return `(nil, fwdiag.Diagnostics{...})`
- **AND** it SHALL NOT return `sdkdiag.Diagnostics`

#### Scenario: PutSecurityRole success returns nil
- **GIVEN** a call to `PutSecurityRole` that succeeds with HTTP 200/204
- **WHEN** the function returns
- **THEN** it SHALL return `nil` (zero `fwdiag.Diagnostics`)

#### Scenario: DeleteSecurityRole not found returns nil
- **GIVEN** a call to `DeleteSecurityRole` for a missing role
- **WHEN** the API returns HTTP 404
- **THEN** the function SHALL return `nil`

### Requirement: kibanaoapi connector search returns PF diagnostics
`internal/clients/kibanaoapi/connector.go` `SearchConnectors` SHALL return `fwdiag.Diagnostics` instead of `sdkdiag.Diagnostics`.

#### Scenario: SearchConnectors error returns PF diagnostic
- **GIVEN** a call to `SearchConnectors` that encounters an HTTP error
- **WHEN** the function returns
- **THEN** it SHALL return `(nil, fwdiag.Diagnostics{...})`

### Requirement: kibanaoapi space operations return PF diagnostics
All public functions in `internal/clients/kibanaoapi/spaces.go` (`CreateSpace`, `UpdateSpace`, `DeleteSpace`) SHALL return `fwdiag.Diagnostics` instead of `sdkdiag.Diagnostics`.

#### Scenario: CreateSpace error returns PF diagnostic
- **GIVEN** a call to `CreateSpace` that encounters an HTTP error
- **WHEN** the function returns
- **THEN** it SHALL return `(nil, fwdiag.Diagnostics{...})`

### Requirement: kibanaoapi file does not import sdk diag
No file in `internal/clients/kibanaoapi/` SHALL import `github.com/hashicorp/terraform-plugin-sdk/v2/diag`.

#### Scenario: Verify no sdk diag import in kibanaoapi
- **GIVEN** the codebase after this change
- **WHEN** searching for `terraform-plugin-sdk/v2/diag` in `internal/clients/kibanaoapi/*.go`
- **THEN** zero results SHALL be found
