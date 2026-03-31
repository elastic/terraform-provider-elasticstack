## ADDED Requirements

### Requirement: In-scope acceptance test step scope
The lint rule SHALL evaluate `resource.TestStep` composite literals in `_test.go` files that participate in acceptance-test flows driven by `resource.Test` or `resource.ParallelTest`. The rule SHALL apply without path-based exclusions, including files under `internal/**` and `provider/**`. The rule SHALL ignore non-`resource.TestStep` structs and code outside those acceptance-test flows.

#### Scenario: Out-of-scope test code is ignored
- **GIVEN** a Go test file outside the lint rule's defined acceptance-test scope
- **WHEN** the analyzer evaluates the file
- **THEN** the analyzer SHALL report no issue solely because the file contains a struct field named `Config`, `ConfigDirectory`, or `ExternalProviders`

#### Scenario: Provider package acceptance test is in scope
- **GIVEN** a `_test.go` file under `provider/**`
- **WHEN** the file uses `resource.Test` or `resource.ParallelTest` with `resource.TestStep` values
- **THEN** the analyzer SHALL evaluate those test steps using the same rules as any other in-scope acceptance test

### Requirement: Ordinary acceptance steps use directory-backed fixtures
Any in-scope `resource.TestStep` that supplies Terraform configuration through `ConfigDirectory` SHALL call `acctest.NamedTestCaseDirectory(...)` directly.

#### Scenario: Ordinary step uses named fixture directory
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step uses `ConfigDirectory: acctest.NamedTestCaseDirectory("create")`
- **THEN** the analyzer SHALL report no diagnostic for that configuration source

#### Scenario: Ordinary step uses a bypassed config-directory helper
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step uses `ConfigDirectory` with `config.TestNameDirectory()` or any helper other than `acctest.NamedTestCaseDirectory(...)`
- **THEN** the analyzer SHALL emit a diagnostic requiring `acctest.NamedTestCaseDirectory(...)`

### Requirement: Inline config is restricted to external-provider compatibility steps
Any in-scope `resource.TestStep` that sets `Config` SHALL also set `ExternalProviders`. Inline `Config` SHALL NOT be accepted as the ordinary fixture mechanism for in-scope acceptance steps.

#### Scenario: Ordinary inline config is rejected
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `Config` and does not set `ExternalProviders`
- **THEN** the analyzer SHALL emit a diagnostic requiring directory-backed fixtures

#### Scenario: External-provider compatibility step may use inline config
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets both `ExternalProviders` and inline `Config`
- **THEN** the analyzer SHALL report no diagnostic for using inline `Config`

### Requirement: External-provider compatibility steps stay on inline config
Any in-scope `resource.TestStep` that sets `ExternalProviders` SHALL use inline `Config` and SHALL NOT use `ConfigDirectory`.

#### Scenario: External-provider compatibility step mixes config sources
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` together with `ConfigDirectory`
- **THEN** the analyzer SHALL emit a diagnostic requiring inline `Config` instead of `ConfigDirectory`

#### Scenario: External-provider compatibility step omits inline config
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and omits `Config`
- **THEN** the analyzer SHALL emit a diagnostic requiring inline `Config`

### Requirement: Field-relationship diagnostics are actionable
When the lint rule reports a violation, the diagnostic SHALL identify the invalid `resource.TestStep` field relationship and SHALL direct contributors toward the accepted shape for that step.

#### Scenario: Diagnostic explains the accepted replacement
- **GIVEN** a non-compliant in-scope acceptance `resource.TestStep`
- **WHEN** the analyzer emits a diagnostic
- **THEN** the diagnostic SHALL tell the contributor whether to switch to `ConfigDirectory: acctest.NamedTestCaseDirectory(...)` or to keep the step as an `ExternalProviders` plus inline `Config` compatibility case

### Requirement: Repository lint enforces the rule
The analyzer SHALL be wired into repository lint execution so `make check-lint` fails on violations, and regression tests SHALL cover compliant and non-compliant step shapes.

#### Scenario: Lint fails on a new violation
- **GIVEN** a committed in-scope acceptance step that violates the config-directory lint rule
- **WHEN** repository lint runs through `make check-lint`
- **THEN** the lint command SHALL fail in local and CI workflows

#### Scenario: Regression tests protect the accepted shapes
- **GIVEN** analyzer regression coverage for directory-backed ordinary steps and external-provider compatibility steps
- **WHEN** the analyzer implementation changes later
- **THEN** the regression suite SHALL continue to distinguish compliant and non-compliant step shapes as specified
