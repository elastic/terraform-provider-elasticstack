## REMOVED Requirements

### Requirement: Approved Elasticsearch client sources
Any `*clients.APIClient` value used at an in-scope sink SHALL originate from `clients.NewAPIClientFromSDKResource(...)`, `clients.MaybeNewAPIClientFromFrameworkResource(...)`, an explicitly allowlisted wrapper, or an interprocedurally inferred wrapper or factory that the analyzer can prove is helper-derived.

#### Scenario: Sink uses approved helper-derived client
- **GIVEN** an in-scope sink call in code under `internal/elasticsearch/**`
- **WHEN** the supplied `*clients.APIClient` value comes directly from an approved helper-derived source
- **THEN** the analyzer SHALL report no diagnostic

**Reason**: Elasticsearch sinks will no longer accept the broad `*clients.APIClient`, so provenance approval by custom lint is replaced by compile-time type safety at the sink boundary.

**Migration**: Resolve a typed Elasticsearch-scoped client from `ProviderClientFactory` and pass that typed client, or a narrower interface derived from it, to Elasticsearch sinks.

### Requirement: No bypass paths at sink usage
In-scope sink calls SHALL NOT use `*clients.APIClient` values created through direct construction, provider-meta casts, or ad-hoc resolution flows that bypass approved helper-derived sources.

#### Scenario: Sink uses bypassed client source
- **GIVEN** an in-scope sink call in code under `internal/elasticsearch/**`
- **WHEN** the supplied `*clients.APIClient` value was obtained through direct construction, provider-meta casting, or another non-helper-derived flow
- **THEN** the analyzer SHALL emit a diagnostic for that sink usage

**Reason**: Typed Elasticsearch sinks make bypassed broad-client sources uncallable in supported production code, so this prohibition no longer needs a separate linter contract.

**Migration**: Replace broad-client construction or provider-meta casts with factory-based typed Elasticsearch client resolution.

### Requirement: Sink enforcement scope
The lint rule SHALL validate client origin specifically at sink call sites in code under `internal/elasticsearch/**`: function arguments passed to `internal/clients/elasticsearch` functions with `*clients.APIClient` parameters, and receivers used for method calls on `*clients.APIClient`.

#### Scenario: No sink produces no finding
- **GIVEN** code under `internal/elasticsearch/**` that does not invoke an in-scope sink
- **WHEN** the analyzer evaluates that code
- **THEN** the analyzer SHALL report no issue solely because a `*clients.APIClient` value exists

**Reason**: The shared Elasticsearch sink surface is being redefined to require typed scoped clients instead of broad `*clients.APIClient`, making the old analyzer sink scope obsolete.

**Migration**: Update in-scope Elasticsearch helper and sink signatures to require the typed Elasticsearch-scoped client contract.

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

**Reason**: Typed sink boundaries eliminate the need for interprocedural wrapper provenance analysis because supported wrappers will traffic in typed Elasticsearch-scoped clients instead of broad `*clients.APIClient`.

**Migration**: Convert wrapper or factory functions to return typed Elasticsearch-scoped clients or narrower typed interfaces and remove analyzer allowlist dependencies.

### Requirement: Type-based conservative analysis
The lint rule SHALL use type information to identify in-scope sinks and `*clients.APIClient` values, and where provenance cannot be proven it SHALL treat the value as non-derived rather than assuming compliance.

#### Scenario: Uncertain provenance is treated as non-derived
- **GIVEN** an in-scope sink call whose client origin cannot be proven from type-aware analysis and provenance facts
- **WHEN** the analyzer evaluates the sink usage
- **THEN** the analyzer SHALL report a violation instead of assuming the client is approved

**Reason**: The compiler now enforces the supported client type directly, so conservative provenance analysis is no longer the mechanism that protects Elasticsearch sink usage.

**Migration**: Remove analyzer-driven provenance checks and rely on typed Elasticsearch sink signatures plus normal compilation and tests.

### Requirement: Actionable diagnostics
When the analyzer reports a violation, the diagnostic SHALL state that the sink uses a non-helper-derived client and SHALL point to the approved helper sources.

#### Scenario: Violation message directs remediation
- **GIVEN** a non-compliant sink usage
- **WHEN** the analyzer emits a diagnostic
- **THEN** the diagnostic SHALL identify the non-helper-derived sink usage and reference the approved helper-derived sources

**Reason**: Once the analyzer is removed, remediation is driven by compiler type errors and the provider factory contract rather than custom lint diagnostics.

**Migration**: Follow compile-time errors by resolving typed Elasticsearch clients through `ProviderClientFactory` and passing those typed clients to shared sinks.

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

**Reason**: The repository no longer needs a custom lint and test harness for Elasticsearch client provenance after the sink API is narrowed to typed scoped clients.

**Migration**: Delete `analysis/esclienthelperplugin`, remove its lint wiring, and cover Elasticsearch client-resolution behavior through typed APIs plus normal unit, acceptance, and build checks.
