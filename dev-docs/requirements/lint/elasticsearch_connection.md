# Elasticsearch entities client-resolution lint — requirements

Implementation target: `internal/analysis/esclienthelper` (custom `go/analysis` analyzer, wired through `golangci-lint`)

## Scope

This document defines provider-level lint requirements that enforce helper-derived Elasticsearch client usage at concrete client-call sink points.

In scope:
- SDK and Framework resources/data sources implemented under `internal/elasticsearch/**`.
- Calls to:
  - functions in `internal/clients/elasticsearch` that accept `*clients.APIClient`.
  - methods with receiver `*clients.APIClient`.

## Requirements

- **[REQ-001] (Approved client sources)**: Any `*clients.APIClient` value used at an in-scope sink shall originate from `clients.NewAPIClientFromSDKResource(...)`, `clients.MaybeNewAPIClientFromFrameworkResource(...)`, an explicitly allowlisted wrapper, or an interprocedurally inferred wrapper/factory that the analyzer can prove is helper-derived.
- **[REQ-002] (No bypass paths)**: In-scope sink calls shall not use `*clients.APIClient` values created through direct construction, provider meta casts, or ad-hoc resolution flows that bypass REQ-001.
- **[REQ-003] (Sink enforcement)**: The lint rule shall validate client origin specifically at sink call sites:
  - function arguments passed to `internal/clients/elasticsearch` functions with `*clients.APIClient` parameters.
  - receivers used for method calls on `*clients.APIClient`.
- **[REQ-004] (Wrapper control and inference)**: Wrapper sources may be accepted by either:
  - explicit analyzer allowlist keyed by fully qualified function name, or
  - analyzer-exported provenance facts proving helper-derived return behavior across function calls.
- **[REQ-005] (Low false positives)**: The lint rule shall use type information to identify sinks and `*clients.APIClient` values and shall not rely only on identifier names.
- **[REQ-006] (Conservative correctness)**: Where provenance cannot be proven, the analyzer shall treat the value as non-derived and report at sink usage instead of assuming compliance.
- **[REQ-007] (Actionable diagnostics)**: Violations shall identify that the sink uses a non-helper-derived client and point to the approved helper sources.
- **[REQ-008] (Fact-backed enforcement)**: The analyzer shall export/import provenance facts for relevant functions to improve interprocedural detection of helper-derived clients.
- **[REQ-009] (CI enforcement)**: The lint rule shall be wired into repository lint execution (`make check-lint`) so violations fail CI.
- **[REQ-010] (Regression guardrail)**: The lint rule behavior shall be covered by analyzer tests that verify compliant and non-compliant sink usage and prevent future weakening of detection.

## Acceptance criteria

- **[AC-001] (Elasticsearch package sink compliant)**: Given a call to `internal/clients/elasticsearch` where a `*clients.APIClient` parameter is supplied with a helper-derived value, the analyzer reports no issue.
- **[AC-002] (Elasticsearch package sink violation)**: Given a call to `internal/clients/elasticsearch` where a `*clients.APIClient` parameter is supplied with a non-helper-derived value, the analyzer emits a diagnostic.
- **[AC-003] (`*clients.APIClient` receiver compliant)**: Given a method call on `*clients.APIClient` where the receiver is helper-derived, the analyzer reports no issue.
- **[AC-004] (`*clients.APIClient` receiver violation)**: Given a method call on `*clients.APIClient` where the receiver is non-helper-derived, the analyzer emits a diagnostic.
- **[AC-005] (No sink, no finding)**: Given a file or function under `internal/elasticsearch/**` that does not hit in-scope sink calls, the analyzer reports no issue.
- **[AC-006] (Wrapper policy enforcement)**: Given delegated resolution through a wrapper, analyzer behavior is:
  - allow when wrapper is explicitly allowlisted and semantically equivalent to approved helper sources.
  - allow when wrapper is not allowlisted but analyzer facts prove helper-derived return behavior.
  - fail when neither allowlist nor facts can prove helper-derived behavior.
- **[AC-007] (CI behavior)**: Given a committed violation, running `make check-lint` fails in local and CI workflows.
- **[AC-008] (Fact inference behavior)**: Given a helper-derived wrapper/factory function in scope, when the wrapper return is passed to an in-scope sink, the analyzer reports no issue without requiring explicit allowlist configuration.
