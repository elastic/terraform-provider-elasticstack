## ADDED Requirements

### Requirement: Provider serves only Plugin Framework
The provider SHALL serve exclusively a Plugin Framework provider implementation. No SDK v2 provider SHALL be constructed, upgraded, or muxed.

#### Scenario: Main entry point creates PF provider directly
- **GIVEN** the provider binary starts
- **WHEN** `main.go` creates the provider server
- **THEN** it SHALL call `providerserver.NewProtocol6(NewFrameworkProvider(version))` directly
- **AND** it SHALL NOT call `tf6muxserver.NewMuxServer` or `tf5to6server.UpgradeServer`

#### Scenario: SDK provider constructor is absent
- **GIVEN** the codebase after this change
- **WHEN** searching for `func New(version string) *schema.Provider`
- **THEN** the function SHALL NOT exist in `provider/provider.go`
- **AND** `provider/provider.go` itself SHALL be removed

#### Scenario: Provider schema is PF-only
- **GIVEN** the provider configuration schema
- **WHEN** the schema is resolved
- **THEN** it SHALL be defined as `fwschema.Schema` with `Blocks`
- **AND** it SHALL NOT be defined as `map[string]*schema.Schema`

### Requirement: mux dependencies are removed
The provider SHALL remove all direct dependencies on `github.com/hashicorp/terraform-plugin-mux` and the `terraform-plugin-sdk/v2` provider construction path.

#### Scenario: go.mod lacks mux and sdk provider dependencies
- **GIVEN** `go.mod` after this change
- **WHEN** inspecting direct dependencies
- **THEN** `github.com/hashicorp/terraform-plugin-mux` SHALL NOT appear as a direct dependency (unless still required by test code)
