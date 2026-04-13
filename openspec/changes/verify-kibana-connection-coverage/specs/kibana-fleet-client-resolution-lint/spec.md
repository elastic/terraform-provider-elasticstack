## ADDED Requirements

### Requirement: Approved Kibana and Fleet client sources
Any `*clients.APIClient` value used at an in-scope Kibana or Fleet sink SHALL originate from `clients.NewKibanaAPIClientFromSDKResource(...)`, `clients.MaybeNewKibanaAPIClientFromFrameworkResource(...)`, an explicitly allowlisted wrapper, or an interprocedurally inferred wrapper or factory that the analyzer can prove is helper-derived.

#### Scenario: Sink uses approved helper-derived client
- **WHEN** an in-scope Kibana or Fleet sink uses a `*clients.APIClient` value derived from an approved helper source
- **THEN** the analyzer SHALL report no diagnostic

### Requirement: No bypass paths at sink usage
In-scope Kibana and Fleet sink calls SHALL NOT use `*clients.APIClient` values created through provider-data casts, direct construction, or ad-hoc resolution flows that bypass `clients.NewKibanaAPIClientFromSDKResource(...)`, `clients.MaybeNewKibanaAPIClientFromFrameworkResource(...)`, or other approved helper-derived sources.

#### Scenario: Sink uses bypassed client source
- **WHEN** an in-scope Kibana or Fleet sink consumes a non-helper-derived `*clients.APIClient`
- **THEN** the analyzer SHALL emit a diagnostic for that sink usage

### Requirement: Sink enforcement scope
The lint rule SHALL validate client origin at concrete sink call sites in code under `internal/kibana/**` and `internal/fleet/**`. In-scope sinks SHALL include direct use of `*clients.APIClient` receivers and calls that obtain Kibana legacy, Kibana OpenAPI, SLO, or Fleet clients from `*clients.APIClient`.

#### Scenario: No sink produces no finding
- **WHEN** Kibana or Fleet code does not invoke an in-scope sink
- **THEN** the analyzer SHALL report no issue solely because a `*clients.APIClient` value exists

### Requirement: Wrapper control and conservative provenance
Wrapper sources SHALL be accepted only when the wrapper is explicitly allowlisted by fully qualified function name or when analyzer-exported provenance facts prove that the wrapper or factory returns a helper-derived `*clients.APIClient`. When provenance cannot be proven, the analyzer SHALL treat the client as non-derived.

#### Scenario: Fact-proven wrapper remains compliant
- **WHEN** a helper-derived wrapper or factory is proven by exported facts and its returned client reaches an in-scope sink
- **THEN** the analyzer SHALL report no issue without requiring explicit allowlist configuration

#### Scenario: Unproven wrapper fails conservatively
- **WHEN** a wrapper or factory reaches an in-scope sink and helper-derived provenance cannot be proven
- **THEN** the analyzer SHALL report the sink usage as non-compliant

### Requirement: Lint workflow enforcement
The Kibana/Fleet client-resolution lint rule SHALL be wired into repository lint execution so `make check-lint` fails on violations, and analyzer regression tests SHALL cover compliant and non-compliant sink usage.

#### Scenario: Lint execution enforces the rule in CI
- **WHEN** repository lint runs with a committed Kibana or Fleet client-resolution violation
- **THEN** the lint command SHALL fail in local and CI workflows
