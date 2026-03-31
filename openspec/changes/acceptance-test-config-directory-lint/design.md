## Context

Acceptance-style provider tests commonly build `resource.TestCase` values passed to `resource.Test` or `resource.ParallelTest`, whose `Steps` are anonymous `resource.TestStep` literals. The current preferred pattern is:

- ordinary provider-managed steps use `ConfigDirectory: acctest.NamedTestCaseDirectory("<case>")`
- compatibility steps that exercise a previous provider version use `ExternalProviders` together with inline `Config`

That structure is not easy to enforce with a pattern-only linter because the step literals are usually anonymous `{ ... }` elements inside `[]resource.TestStep`, and the rule depends on relationships between sibling fields on the same struct literal. The repo also already has a suitable precedent for a custom type-aware analyzer via `analysis/esclienthelper`.

## Goals / Non-Goals

**Goals:**

- Enforce the acceptance-test fixture convention for in-scope `resource.TestStep` literals in any Go test file that uses `resource.Test` or `resource.ParallelTest`.
- Allow exactly one inline-config exception path: steps that declare `ExternalProviders`.
- Require `ConfigDirectory` usage to go through `acctest.NamedTestCaseDirectory(...)` rather than lower-level helpers such as `config.TestNameDirectory()`.
- Provide clear diagnostics that explain which field combination is invalid and what the accepted replacement is.
- Integrate the rule into normal lint and CI execution with analyzer regression tests.

**Non-Goals:**

- Linting arbitrary `resource.TestStep`-shaped structs or unrelated composite literals outside actual `resource.Test` / `resource.ParallelTest` acceptance-test flows.
- Enforcing that every test step always has either `Config` or `ConfigDirectory`; import-only, refresh-only, and plan-only steps may legitimately have neither.
- Validating the semantic reason a step uses `ExternalProviders` beyond treating it as the accepted marker for previous-provider compatibility coverage.
- Replacing broader Terraform fixture organization conventions outside the defined step fields.

## Decisions

- **Custom analyzer, not `gocritic` ruleguard**: Implement the rule as a dedicated `go/analysis` analyzer using the same plugin-module pattern as `analysis/esclienthelper`. The rule needs typed detection of `resource.TestStep` literals plus sibling-field validation on the same composite literal, which is a better fit for a real analyzer than for pattern-only ruleguard checks.

- **Behavior-based scope, not path-based scope**: Analyze `_test.go` files anywhere in the repository and inspect `resource.TestStep` literals that appear within acceptance-test flows driven by `resource.Test` or `resource.ParallelTest`. This includes files like `internal/elasticsearch/index/template_test.go`, `internal/kibana/space_test.go`, and `provider/provider_test.go`, even when they are not named `*_acc_test.go`.

- **Directory-backed default path**: Any in-scope `resource.TestStep` that supplies Terraform configuration through `ConfigDirectory` must call `acctest.NamedTestCaseDirectory(...)` directly. This keeps the fixture convention explicit, consistent, and easy to audit.

- **External provider steps are the only inline-config exception**: Any in-scope `resource.TestStep` that uses `Config` must also set `ExternalProviders`. This encodes the older-provider compatibility exception mechanically instead of relying on comments or file naming.

- **Exception steps stay inline-only**: Any in-scope `resource.TestStep` that sets `ExternalProviders` must use `Config` and must not use `ConfigDirectory`. This keeps the compatibility-path contract crisp and avoids mixed-source ambiguity.

- **Field-relationship enforcement only**: The analyzer should report only on invalid combinations involving `Config`, `ConfigDirectory`, and `ExternalProviders`. It should not require or infer other test-step fields such as `ProtoV6ProviderFactories`, `ConfigVariables`, `Check`, `ImportState`, or `SkipFunc`.

- **Actionable diagnostics**: Diagnostics should tell contributors which shape is required:
  - ordinary steps should use `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`
  - compatibility steps with `ExternalProviders` should use inline `Config`
  - `config.TestNameDirectory()` and other `ConfigDirectory` helpers are not accepted in-scope

- **Migration before full enforcement**: Existing in-scope violations should be updated as part of enabling the rule so repository lint can fail only on new regressions, not on already-known drift, regardless of package path.

## Risks / Trade-offs

- **[Risk] Behavior-based scope is harder to detect than path-based scope** -> Mitigation: define in-scope tests explicitly in terms of `resource.Test` / `resource.ParallelTest` call sites and add regression tests that cover both `internal/**` and `provider/**` examples.
- **[Risk] Some current inline-config tests require non-trivial fixture extraction** -> Mitigation: treat migration as explicit implementation work and convert the in-scope tests before the analyzer is fully enforced.
- **[Risk] Direct-call requirement for `acctest.NamedTestCaseDirectory(...)` is stricter than necessary** -> Mitigation: accept that strictness in v1 to keep the convention visible and deterministic; wrappers can be reconsidered later if a real need appears.
- **[Risk] `ExternalProviders` is only a proxy for "previous provider version"** -> Mitigation: make that proxy explicit in the spec so the analyzer stays mechanical and predictable.

## Migration Plan

1. Add the new `acceptance-test-config-directory-lint` delta spec covering scope, allowed field combinations, diagnostics, and lint integration expectations.
2. Implement a dedicated analyzer package and plugin-module wrapper modeled after `analysis/esclienthelper`.
3. Update `golangci-lint` configuration so the analyzer runs in normal local and CI lint workflows.
4. Convert existing in-scope inline-config acceptance steps that are not external-provider compatibility steps to directory-backed fixtures.
5. Replace in-scope `ConfigDirectory: config.TestNameDirectory()` usage with `acctest.NamedTestCaseDirectory(...)` plus the corresponding fixture layout.
6. Add regression tests for compliant and non-compliant step shapes, then enable the rule as a failing lint check.
