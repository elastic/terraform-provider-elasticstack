## ADDED Requirements

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
