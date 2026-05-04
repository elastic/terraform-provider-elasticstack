# test-suite-elasticsearch-typed-client Specification

## Purpose
TBD - created by archiving change typed-client-acceptance-tests. Update Purpose after archive.
## Requirements
### Requirement: Acceptance tests use typed Elasticsearch client
The listed acceptance test files SHALL use `GetESTypedClient()` to obtain a typed Elasticsearch client for preflight setup, post-test cleanup, and state assertions. They SHALL NOT call `GetESClient()` for direct Elasticsearch API interactions, except where no typed equivalent exists and a deliberate fallback is documented with a TODO comment.

#### Scenario: Test preflight uses typed client
- **GIVEN** an acceptance test that creates indices, templates, policies, or roles before Terraform applies
- **WHEN** the test performs preflight setup
- **THEN** it SHALL use `GetESTypedClient()` and typed API methods instead of `GetESClient()` and raw `esapi` calls

#### Scenario: Test cleanup uses typed client
- **GIVEN** an acceptance test that verifies resources were deleted after test completion
- **WHEN** the test performs cleanup or destroy checks
- **THEN** it SHALL use `GetESTypedClient()` and typed API methods instead of `GetESClient()` and raw `esapi` calls

#### Scenario: Test state assertion uses typed client
- **GIVEN** an acceptance test that queries Elasticsearch during test execution to validate side effects
- **WHEN** the test performs state assertions
- **THEN** it SHALL use `GetESTypedClient()` and typed API methods instead of `GetESClient()` and raw `esapi` calls

### Requirement: Typed responses replace manual JSON decoding
Where the typed Elasticsearch API provides strongly-typed response structs, acceptance test code SHALL use those structs directly. Manual `io.ReadAll`, `json.NewDecoder`, and `json.Unmarshal` from response bodies SHALL be removed.

#### Scenario: Get response uses typed struct
- **GIVEN** an acceptance test that previously read a raw HTTP response body and unmarshaled it into an ad-hoc struct or map
- **WHEN** the test is migrated to the typed client
- **THEN** the test SHALL receive a typed response struct directly from the typed API call
- **AND** manual body reading and JSON decoding code SHALL be removed

### Requirement: Preserve existing test behavior
The migration SHALL NOT modify test assertions, Terraform configurations, test step definitions, or expected test outcomes. Only the client access pattern and response handling SHALL change.

#### Scenario: Test assertions remain unchanged
- **GIVEN** an acceptance test with existing assertions
- **WHEN** the migration replaces raw client calls with typed equivalents
- **THEN** all test assertions SHALL remain identical to their pre-migration form

#### Scenario: Terraform configurations remain unchanged
- **GIVEN** an acceptance test with existing Terraform configuration blocks
- **WHEN** the migration replaces raw client calls with typed equivalents
- **THEN** all Terraform configurations in the test SHALL remain identical to their pre-migration form

### Requirement: No residual raw client calls in migrated files
After migration, none of the listed acceptance test files SHALL contain direct `GetESClient()` calls for Elasticsearch API interactions.

#### Scenario: Verify no remaining raw client usage
- **GIVEN** the listed acceptance test files have been migrated
- **WHEN** a search for `GetESClient()` is performed across those files
- **THEN** no occurrences SHALL be found except where no typed equivalent exists and a deliberate fallback is documented with a TODO comment

### Requirement: Compilation and test execution
All migrated acceptance tests SHALL compile successfully without errors, and the repository unit test suite SHALL continue to pass.

#### Scenario: Compilation succeeds after migration
- **GIVEN** the migration is complete
- **WHEN** `go test ./internal/...` is run
- **THEN** all packages SHALL compile without errors

#### Scenario: Unit tests pass after migration
- **GIVEN** the migration is complete
- **WHEN** unit tests are executed
- **THEN** all tests SHALL pass

