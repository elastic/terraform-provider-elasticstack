# elasticsearch-client-resolution-lint Specification

## Purpose
TBD - created by archiving change elasticsearch-client-resolution-lint. Update Purpose after archive.
## Requirements
### Requirement: Approved Elasticsearch client sources
Any `*clients.APIClient` value used at an in-scope sink SHALL originate from `clients.NewAPIClientFromSDKResource(...)`, `clients.MaybeNewAPIClientFromFrameworkResource(...)`, an explicitly allowlisted wrapper, or an interprocedurally inferred wrapper or factory that the analyzer can prove is helper-derived.

#### Scenario: Sink uses approved helper-derived client
- **GIVEN** an in-scope sink call in code under `internal/elasticsearch/**`
- **WHEN** the supplied `*clients.APIClient` value comes directly from an approved helper-derived source
- **THEN** the analyzer SHALL report no diagnostic

### Requirement: No bypass paths at sink usage
In-scope sink calls SHALL NOT use `*clients.APIClient` values created through direct construction, provider-meta casts, or ad-hoc resolution flows that bypass approved helper-derived sources.

#### Scenario: Sink uses bypassed client source
- **GIVEN** an in-scope sink call in code under `internal/elasticsearch/**`
- **WHEN** the supplied `*clients.APIClient` value was obtained through direct construction, provider-meta casting, or another non-helper-derived flow
- **THEN** the analyzer SHALL emit a diagnostic for that sink usage

### Requirement: Sink enforcement scope
The lint rule SHALL validate client origin specifically at sink call sites in code under `internal/elasticsearch/**`: function arguments passed to `internal/clients/elasticsearch` functions with `*clients.APIClient` parameters, and receivers used for method calls on `*clients.APIClient`.

#### Scenario: No sink produces no finding
- **GIVEN** code under `internal/elasticsearch/**` that does not invoke an in-scope sink
- **WHEN** the analyzer evaluates that code
- **THEN** the analyzer SHALL report no issue solely because a `*clients.APIClient` value exists

### Requirement: Wrapper control and inference
Wrapper sources SHALL be accepted only when the wrapper is explicitly allowlisted by fully qualified function name or when analyzer-exported provenance facts prove that the wrapper or factory returns a helper-derived `*clients.APIClient`.

#### Scenario: Wrapper policy controls sink compliance
- **GIVEN** a wrapper or factory function that returns a `*clients.APIClient` later used at an in-scope sink
- **WHEN** the wrapper is explicitly allowlisted or provenance facts prove its helper-derived return behavior
- **THEN** the analyzer SHALL allow the sink usage

#### Scenario: Unproven wrapper fails conservatively
- **GIVEN** a wrapper or factory function that returns a `*clients.APIClient` later used at an in-scope sink
- **WHEN** neither the allowlist nor provenance facts prove helper-derived behavior
- **THEN** the analyzer SHALL report the sink usage as non-compliant

### Requirement: Type-based conservative analysis
The lint rule SHALL use type information to identify in-scope sinks and `*clients.APIClient` values, and where provenance cannot be proven it SHALL treat the value as non-derived rather than assuming compliance.

#### Scenario: Uncertain provenance is treated as non-derived
- **GIVEN** an in-scope sink call whose client origin cannot be proven from type-aware analysis and provenance facts
- **WHEN** the analyzer evaluates the sink usage
- **THEN** the analyzer SHALL report a violation instead of assuming the client is approved

### Requirement: Actionable diagnostics
When the analyzer reports a violation, the diagnostic SHALL state that the sink uses a non-helper-derived client and SHALL point to the approved helper sources.

#### Scenario: Violation message directs remediation
- **GIVEN** a non-compliant sink usage
- **WHEN** the analyzer emits a diagnostic
- **THEN** the diagnostic SHALL identify the non-helper-derived sink usage and reference the approved helper-derived sources

### Requirement: Fact-backed lint and regression enforcement
The analyzer SHALL export and import provenance facts for relevant functions to improve interprocedural detection of helper-derived clients. The rule SHALL be wired into repository lint execution so `make check-lint` fails on violations, and analyzer tests SHALL cover compliant and non-compliant sink usage to prevent regression.

#### Scenario: Fact-proven wrapper remains compliant
- **GIVEN** a helper-derived wrapper or factory function whose return behavior is captured by exported provenance facts
- **WHEN** that returned client is passed to an in-scope sink
- **THEN** the analyzer SHALL report no issue without requiring explicit allowlist configuration

#### Scenario: Lint execution enforces the rule in CI
- **GIVEN** a committed violation of the client-resolution lint rule
- **WHEN** repository lint runs through `make check-lint`
- **THEN** the lint command SHALL fail in local and CI workflows

