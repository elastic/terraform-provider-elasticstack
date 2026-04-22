## MODIFIED Requirements

### Requirement: Inline config is restricted to external-provider compatibility steps
Any in-scope `resource.TestStep` that sets `Config` SHALL also set `ExternalProviders`. Inline `Config` SHALL NOT be accepted as the ordinary fixture mechanism for in-scope acceptance steps, which use `ProtoV6ProviderFactories` and directory-backed fixtures instead. When an `ExternalProviders` step uses `Config`, that `Config` value SHALL satisfy the embedded-fixture source rules defined for external-provider compatibility steps.

#### Scenario: Ordinary inline config is rejected
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `Config` and does not set `ExternalProviders`
- **THEN** the analyzer SHALL emit a diagnostic requiring directory-backed fixtures

#### Scenario: External-provider compatibility step may use embedded config
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets both `ExternalProviders` and `Config`, and `Config` resolves to an allowed embedded fixture variable
- **THEN** the analyzer SHALL report no diagnostic for using `Config`

### Requirement: External-provider compatibility steps stay on inline config
Any in-scope `resource.TestStep` that sets `ExternalProviders` SHALL use `Config`, SHALL NOT use `ConfigDirectory`, and SHALL NOT set `ProtoV6ProviderFactories`. For these compatibility steps, `Config` SHALL resolve to a package-level `string` variable populated by a `//go:embed` directive whose embedded fixture path points to a `.tf` file under `testdata/`. The analyzer SHALL reject raw string literals, formatted strings, concatenated strings, helper-function-returned strings, and identifiers that do not resolve to that embedded-fixture declaration shape.

#### Scenario: External-provider compatibility step mixes config sources
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` together with `ConfigDirectory`
- **THEN** the analyzer SHALL emit a diagnostic requiring `Config` instead of `ConfigDirectory`

#### Scenario: External-provider compatibility step omits inline config
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and omits `Config`
- **THEN** the analyzer SHALL emit a diagnostic requiring `Config`

#### Scenario: External-provider compatibility step also sets ProtoV6 provider factories
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets both `ExternalProviders` and `ProtoV6ProviderFactories`
- **THEN** the analyzer SHALL emit a diagnostic requiring the step to use only `ExternalProviders` for backwards-compatibility wiring

#### Scenario: External-provider compatibility step uses an embedded fixture variable
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and `Config` references a package-level `string` variable populated by `//go:embed testdata/.../*.tf`
- **THEN** the analyzer SHALL report no diagnostic for the config source

#### Scenario: External-provider compatibility step uses a raw Terraform string
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and `Config` to a raw string literal, a formatted string expression, or a concatenated string expression
- **THEN** the analyzer SHALL emit a diagnostic requiring the config to come from a package-level embedded `.tf` fixture variable under `testdata/`

#### Scenario: External-provider compatibility step uses a helper-returned Terraform string
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and `Config` to a helper call that returns Terraform text
- **THEN** the analyzer SHALL emit a diagnostic requiring the config to come from a package-level embedded `.tf` fixture variable under `testdata/`

#### Scenario: External-provider compatibility step uses a non-embedded string variable
- **GIVEN** an in-scope acceptance `resource.TestStep`
- **WHEN** the step sets `ExternalProviders` and `Config` references a string variable that is not populated by `//go:embed testdata/.../*.tf`
- **THEN** the analyzer SHALL emit a diagnostic requiring the config to come from a package-level embedded `.tf` fixture variable under `testdata/`

### Requirement: Field-relationship diagnostics are actionable
When the lint rule reports a violation, the diagnostic SHALL identify the invalid `resource.TestCase` or `resource.TestStep` field relationship or invalid compatibility-step config source and SHALL direct contributors toward the accepted shape for that step.

#### Scenario: Diagnostic explains the accepted replacement
- **GIVEN** a non-compliant in-scope acceptance `resource.TestStep`
- **WHEN** the analyzer emits a diagnostic
- **THEN** the diagnostic SHALL tell the contributor whether to move `ProtoV6ProviderFactories` onto the step, switch to `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`, keep the step as an `ExternalProviders` plus `Config` compatibility case, or move static Terraform into `testdata/.../*.tf` and load it through package-level `//go:embed`

### Requirement: Repository lint enforces the rule
The analyzer SHALL be wired into repository lint execution so `make check-lint` fails on violations, and regression tests SHALL cover compliant and non-compliant step shapes for both config sourcing and provider wiring, including accepted embedded compatibility fixtures and rejected static Terraform strings defined in Go.

#### Scenario: Lint fails on a new violation
- **GIVEN** a committed in-scope acceptance step that violates the config-directory lint rule
- **WHEN** repository lint runs through `make check-lint`
- **THEN** the lint command SHALL fail in local and CI workflows

#### Scenario: Regression tests protect the accepted shapes
- **GIVEN** analyzer regression coverage for directory-backed ordinary steps, step-level `ProtoV6ProviderFactories`, external-provider compatibility steps that use embedded fixture variables, and rejected static Terraform strings defined in Go
- **WHEN** the analyzer implementation changes later
- **THEN** the regression suite SHALL continue to distinguish compliant and non-compliant step shapes as specified
