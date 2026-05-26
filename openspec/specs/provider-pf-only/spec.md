# provider-pf-only Specification

## Purpose
The Terraform provider serves exclusively a Plugin Framework provider without SDK v2 mux fallback.
## Requirements
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

### Requirement: SDK v2 direct dependency is removed
The provider SHALL remove all direct `require` dependencies on `github.com/hashicorp/terraform-plugin-sdk/v2` from `go.mod` and SHALL NOT import any SDK v2 package directly in provider source or test code.

#### Scenario: go.mod lacks SDK v2 direct dependency
- **GIVEN** `go.mod` after this change
- **WHEN** inspecting the `require` block
- **THEN** `github.com/hashicorp/terraform-plugin-sdk/v2` SHALL NOT appear as a direct dependency
- **AND** it MAY remain as an `// indirect` dependency via `terraform-plugin-testing`

#### Scenario: No direct SDK v2 imports in provider source
- **GIVEN** the provider codebase after this change
- **WHEN** searching for imports of `github.com/hashicorp/terraform-plugin-sdk/v2` outside of `vendor/`
- **THEN** no provider source file SHALL import any SDK v2 package directly
- **AND** dead-code utilities depending on SDK v2 types (`helper/schema.Set`) SHALL be removed

#### Scenario: Test code uses testing helper instead of SDK v2 acctest
- **GIVEN** acceptance test files after this change
- **WHEN** importing acctest utilities
- **THEN** test files SHALL import `github.com/hashicorp/terraform-plugin-testing/helper/acctest`
- **AND** they SHALL NOT import `github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest`

